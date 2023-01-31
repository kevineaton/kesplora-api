package api

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// routeAllGetProjects gets all projects on a site
func routeAllGetProjects(w http.ResponseWriter, r *http.Request) {
	site, err := GetSiteFromContext(r.Context())
	if site == nil || err != nil {
		// this is odd, as it should have been set at the middleware
		sendAPIError(w, api_error_site_get_error, err, nil)
		return
	}

	found, err := GetProjectsForSite(site.ID, "all")
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}
	available := []ProjectAPIReturnNonAdmin{}
	for i := range found {
		if found[i].Status == ProjectStatusActive {
			available = append(available, *convertProjectToUserRet(&found[i]))
		}
	}
	if len(available) == 0 {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// we just send a smaller version of the return here
	sendAPIJSONData(w, http.StatusOK, available)
}

// routeAllGetProject gets a project for a participant
func routeAllGetProject(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, projectIDErr, map[string]string{})
		return
	}

	found, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}
	if found.Status != ProjectStatusActive {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// we just send a smaller version of the return here
	ret := convertProjectToUserRet(found)
	sendAPIJSONData(w, http.StatusOK, ret)
}

// routeAllGetConsentForm gets the consent form for a project
func routeAllGetConsentForm(w http.ResponseWriter, r *http.Request) {
	// seeing the consent form should be accessible to everyone
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, projectIDErr, nil)
		return
	}

	_, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, nil)
		return
	}

	form, err := GetConsentFormForProject(projectID)
	if err != nil {
		sendAPIError(w, api_error_consent_not_found, err, nil)
		return
	}

	sendAPIJSONData(w, http.StatusOK, form)
}

// routeAllCreateConsentResponse creates a response FOR A PARTICIPANT. This is in the `all` grouping because the user could be
// creating a new account while providing the consent
func routeAllCreateConsentResponse(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: false,
	})

	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, projectIDErr, nil)
		return
	}

	project, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, nil)
		return
	}

	input := &ConsentResponse{}
	render.Bind(r, input)
	input.ProjectID = projectID

	// first, make sure the project can even be signed up for
	if project.SignupStatus == ProjectSignupStatusClosed {
		sendAPIError(w, api_error_consent_response_code_err, errors.New("project closed"), map[string]string{})
		return
	} else if project.SignupStatus == ProjectSignupStatusWithCode && input.ProjectCode != project.ShortCode {
		sendAPIError(w, api_error_consent_response_code_err, errors.New("invalid code"), map[string]string{
			"providedCode": input.ProjectCode,
		})
		return
	}

	// check participants
	if project.MaxParticipants > 0 && project.ParticipantCount >= project.MaxParticipants {
		sendAPIError(w, api_error_consent_response_max_reached, errors.New("max participants reached"), map[string]int64{
			"max": project.MaxParticipants,
		})
		return
	}

	// check the DOB of EITHER the logged in user OR the passed in user that would be created
	if project.ParticipantMinimumAge > 0 {
		dob := time.Now()
		if results.User != nil {
			dob, _ = time.Parse(results.User.DateOfBirth, "2006-01-02")
		} else if input.User != nil && input.User.DateOfBirth != "" {
			dob, err = time.Parse(input.User.DateOfBirth, "2006-01-02")
		}
		if err != nil {
			sendAPIError(w, api_error_consent_response_not_min_age, fmt.Errorf("minimum age is %d but could not parse DOB",
				project.ParticipantMinimumAge), map[string]int64{
				"max": project.MaxParticipants,
			})
			return
		}
		years, months, _, _, _, _ := CalculateDuration(dob)
		if years < int(project.ParticipantMinimumAge) {
			sendAPIError(w, api_error_consent_response_not_min_age, fmt.Errorf("minimum age is %d, user age is %d years %d months (born %s)",
				project.ParticipantMinimumAge, years, months, dob.Format("2006-01-02")), map[string]int64{
				"max": project.MaxParticipants,
			})
			return
		}
	}

	// ok, before going any further, we need to figure out if we are dealing with a new user (results.User is nil)
	// or if they are signing up. So we split, but we need to keep in mind that even if the user is logged in,
	// if the project is set for anonymous usage, we need to create a new account with a participant id and password
	// of course, the consent form MUST have actual user information, so there will be the question of connecting them
	//
	// since this is complicated, I broke it out and there's some copy and pasting; once we are comfortable with the
	// flows and pattern, we can consolidate it and refactor it to be a little more professional

	// here's the basic flow:
	// Step 1
	//  if the user is nil:
	//    if the project is visibility == code
	//      create a participant code user
	//    else
	//      create an actual user
	//			user will have to validate their information
	//    set the user in the variable for the consent form
	//
	// Step 2
	//  if the project is visibility == code
	//    if the user is a "real user"
	//      create a participant code user
	//      set as the participant
	//
	// Step 3
	//  consent form needs real user info
	//  save consent with the consent info sent up
	//  if project.connectParticipantToConsentForm == yes
	//    connect with participant

	participant := &jwtUser{} // hold the participant

	// Step 1
	if results.User == nil {
		// the user isn't logged in, so we will create a new account
		if input.User == nil {
			sendAPIError(w, api_error_consent_response_participant_save_err, errors.New("user information must be provided"), map[string]interface{}{
				"input": input,
			})
			return
		}

		// the required fields depends on the account type, so we will need to check the data separately
		if project.ParticipantVisibility != "code" {
			// we create a full account
			if input.User.FirstName == "" || input.User.LastName == "" || input.User.Email == "" || input.User.Password == "" || input.User.DateOfBirth == "" {
				sendAPIError(w, api_error_consent_response_participant_save_err, errors.New("all fields required"), map[string]interface{}{})
				return
			}

		} else {
			// create a new account with just the code
			code := fmt.Sprintf("%d%d%d%d%d%d",
				projectID,
				rand.Intn(9),
				rand.Intn(9),
				rand.Intn(9),
				rand.Intn(9),
				time.Now().Second(),
			)
			input.User.FirstName = ""
			input.User.LastName = ""
			input.User.Email = ""
			input.User.Title = ""
			input.User.Pronouns = ""
			input.User.SystemRole = UserSystemRoleParticipant
			input.User.ParticipantCode = code
			input.User.Status = UserStatusActive
		}

		// create the actual user
		fmt.Printf("\n------------------------\nInput\n%+v\n", input.User)
		err = CreateUser(input.User)
		fmt.Printf("\nBack: %+v\n", err)
		if err != nil {
			sendAPIError(w, api_error_consent_response_participant_save_err, err, map[string]interface{}{})
			return
		}
		token, _, err := generateJWT(input.User)
		if err != nil {
			sendAPIError(w, api_error_consent_response_participant_save_err, err, map[string]interface{}{
				"error": err,
			})
			return
		}
		jwtUser, _ := parseJWT(token)
		participant = &jwtUser
		results.User = participant
	}

	// Step 2
	if project.ParticipantVisibility == "code" {
		// the results.User should contain either the logged in user OR the created user; if this is a "code" visibility
		// project, we need to decide if we need to create another account
		if participant.ID == 0 {
			// a new user is needed, so repeat the above
			code := fmt.Sprintf("%d%d%d%d%d%d",
				projectID,
				rand.Intn(9),
				rand.Intn(9),
				rand.Intn(9),
				rand.Intn(9),
				time.Now().Second(),
			)
			if input.User == nil {
				input.User = &User{
					DateOfBirth: results.User.DateOfBirth,
				}
			}
			input.User.FirstName = ""
			input.User.LastName = ""
			input.User.Email = ""
			input.User.Title = ""
			input.User.Pronouns = ""
			input.User.SystemRole = UserSystemRoleParticipant
			input.User.ParticipantCode = code
			input.User.Status = UserStatusActive
			err = CreateUser(input.User)
			if err != nil {
				sendAPIError(w, api_error_consent_response_participant_save_err, err, map[string]interface{}{})
				return
			}
			token, _, err := generateJWT(input.User)
			if err != nil {
				sendAPIError(w, api_error_consent_response_participant_save_err, err, map[string]interface{}{
					"error": err,
				})
				return
			}
			jwtUser, _ := parseJWT(token)
			participant = &jwtUser
			results.User = participant
		}
	}

	// Step 3
	// at this point, the project is open for sign ups and either there is a code and it matches or no code needed
	// the results.User will hold either the old user or the new account, and the setting will say whether to link them
	// keep in mind, that if the visibility is code but we say to connect the response, it will be possible to connect
	// the information
	if project.ConnectParticipantToConsentForm == "no" {
		input.ParticipantID = 0
	} else {
		input.ParticipantID = results.User.ID
	}

	// ok, parse and save
	err = CreateConsentResponse(input)
	if err != nil {
		sendAPIError(w, api_error_consent_response_save_err, err, map[string]string{})
		return
	}

	// once saved, link the participant to the project
	err = LinkUserAndProject(results.User.ID, project.ID)
	if err != nil {
		sendAPIError(w, api_error_project_link_err, err, map[string]string{})
		return
	}

	// TODO: if the new user status is pending, we need to send the email validation
	// email and send them through the "confirm account" process

	sendAPIJSONData(w, http.StatusOK, input)
}
