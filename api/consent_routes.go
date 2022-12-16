package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// routeSaveConsentForm creates OR updates the consent form
func routeSaveConsentForm(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		MustBeAdmin:     true,
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

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

	input := &ConsentForm{}
	render.Bind(r, input)

	// If the project already has people in it, we want to block this UNLESS the override flag is sent
	if project.ParticipantCount > 0 && !input.OverrideSaveIfParticipants {
		sendAPIError(w, api_error_consent_save_participants_not_zero, fmt.Errorf("no override and non-zero participants: %d", project.ParticipantCount), map[string]interface{}{
			"project": project,
		})
		return
	}

	// pretty much all of this is optional, so we can just save it
	input.ProjectID = projectID
	err = SaveConsentFormForProject(input)
	if err != nil {
		sendAPIError(w, api_error_consent_save_err, err, nil)
		return
	}

	sendAPIJSONData(w, http.StatusOK, input)
}

// routeGetConsentForm gets a consent form
func routeGetConsentForm(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

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
		fmt.Printf("\n%+v\n", err)
		sendAPIError(w, api_error_consent_not_found, err, nil)
		return
	}

	sendAPIJSONData(w, http.StatusOK, form)
}

// routeDeleteConsentForm deletes a consent form
func routeDeleteConsentForm(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		MustBeAdmin:     true,
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

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

	// if there are participants, no deletion is allowed
	if project.ParticipantCount > 0 {
		sendAPIError(w, api_error_consent_delete_participants_not_zero, fmt.Errorf("non-zero participants: %d", project.ParticipantCount), map[string]interface{}{
			"project": project,
		})
		return
	}
	err = DeleteConsentFormForProject(projectID)
	if err != nil {
		sendAPIError(w, api_error_consent_delete_err, err, nil)
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
}

//
// Response
//

// routeCreateConsentResponse creates a response FOR A USER
func routeCreateConsentResponse(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

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

	// and DOB
	if project.ParticipantMinimumAge > 0 {
		dob, _ := time.Parse(results.User.DateOfBirth, "2006-01-02")
		years, months, _, _, _, _ := CalculateDuration(dob)
		if years < int(project.ParticipantMinimumAge) {
			sendAPIError(w, api_error_consent_response_not_min_age, fmt.Errorf("minimum age is %d, user age is %d years %d months (born %s)",
				project.ParticipantMinimumAge, years, months, dob.Format("2006-01-02")), map[string]int64{
				"max": project.MaxParticipants,
			})
			return
		}
	}

	// at this point, the project is open for sign ups and either there is a code and it matches or no code needed

	// depending on the project settings, we need to check a few things here
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

	sendAPIJSONData(w, http.StatusOK, input)
}

// routeGetConsentResponses gets the Consent responses. Admin only.
func routeGetConsentResponses(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		MustBeAdmin:     true,
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

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

	responses, err := GetConsentResponsesForProject(projectID)
	if err != nil {
		sendAPIError(w, api_error_consent_response_get_err, err, nil)
		return
	}
	sendAPIJSONData(w, http.StatusOK, responses)
}

// routeGetConsentResponse gets a response and then validates it is visible to the caller
func routeGetConsentResponse(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	responseID, responseIDErr := strconv.ParseInt(chi.URLParam(r, "responseID"), 10, 64)
	if projectIDErr != nil || responseIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), nil)
		return
	}

	_, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, nil)
		return
	}

	// if the user, make sure it is their's, but it's hard when it's anonymous though
	response, err := GetConsentResponseByID(responseID)
	if err != nil {
		sendAPIError(w, api_error_consent_response_get_err, err, nil)
		return
	}

	if results.User.SystemRole == UserSystemRoleAdmin {
		// return
		sendAPIJSONData(w, http.StatusOK, response)
		return
	}

	if results.User.ID != response.ParticipantID {
		sendAPIError(w, api_error_consent_response_get_err, errors.New("unavailable"), nil)
		return
	}
	sendAPIJSONData(w, http.StatusOK, response)
}

// routeDeleteConsentResponse deletes the consent form AND erases the user from the project
func routeDeleteConsentResponse(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	responseID, responseIDErr := strconv.ParseInt(chi.URLParam(r, "responseID"), 10, 64)
	if projectIDErr != nil || responseIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), nil)
		return
	}

	_, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, nil)
		return
	}

	// if the user, make sure it is their's; hard when it's anonymous though; this should
	// remove the user from the project entirely, removing progress, etc
	response, err := GetConsentResponseByID(responseID)
	if err != nil {
		sendAPIError(w, api_error_consent_response_get_err, err, nil)
		return
	}

	if results.User.SystemRole != UserSystemRoleAdmin && results.User.ID != response.ParticipantID {
		sendAPIError(w, api_error_consent_response_get_err, errors.New("unavailable"), nil)
		return
	}

	err = RemoveUserFromProjectCompletely(response.ParticipantID, projectID)
	if err != nil {
		sendAPIError(w, api_error_project_unlink_err, err, nil)
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"removed": true,
	})
}
