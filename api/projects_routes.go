package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// routeCreateProject creates a new project for the site
func routeCreateProject(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		MustBeAdmin:     true,
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

	input := &Project{}
	render.Bind(r, input)

	// only the name is required
	if input.Name == "" {
		sendAPIError(w, api_error_project_missing_data, errors.New("missing name"), map[string]string{})
		return
	}
	input.SiteID = results.Site.ID

	// if short code is missing, try to fake one
	if input.ShortCode == "" {
		truncLimit := 16
		if len(input.Name) < 16 {
			truncLimit = len(input.Name)
		}
		input.ShortCode = strings.ToLower(strings.Replace(input.Name, " ", "_", -1))[0:truncLimit]
	}
	if input.ShortDescription == "" {
		truncLimit := 1024
		if len(input.Description) < 1024 {
			truncLimit = len(input.Description)
		}
		input.ShortDescription = input.Description[0:truncLimit]
	}

	// TODO: validate enums

	err := CreateProject(input)
	if err != nil {
		sendAPIError(w, api_error_project_save_error, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, input)
}

// routeUpdateProject updates an existing project for the site
func routeUpdateProject(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		MustBeAdmin:     true,
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

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

	input := &Project{}
	render.Bind(r, input)

	// check the fields for updating
	// TODO: validate enums
	if input.Name != "" && input.Name != found.Name {
		found.Name = input.Name
	}
	if input.ShortCode != "" && input.ShortCode != found.ShortCode {
		found.ShortCode = input.ShortCode
	}
	if input.ShortDescription != found.ShortDescription {
		found.ShortDescription = input.ShortDescription
	}
	if input.Description != found.Description {
		found.Description = input.Description
	}
	if input.Status != found.Status {
		found.Status = input.Status
	}
	if input.ShowStatus != found.ShowStatus {
		found.ShowStatus = input.ShowStatus
	}
	if input.SignupStatus != found.SignupStatus {
		found.SignupStatus = input.SignupStatus
	}
	if input.MaxParticipants != found.MaxParticipants {
		found.MaxParticipants = input.MaxParticipants
	}
	if input.ParticipantVisibility != found.ParticipantVisibility {
		found.ParticipantVisibility = input.ParticipantVisibility
	}
	if input.ParticipantMinimumAge != found.ParticipantMinimumAge {
		found.ParticipantMinimumAge = input.ParticipantMinimumAge
	}
	if input.ConnectParticipantToConsentForm != found.ConnectParticipantToConsentForm {
		found.ConnectParticipantToConsentForm = input.ConnectParticipantToConsentForm
	}

	err = UpdateProject(found)
	if err != nil {
		sendAPIError(w, api_error_project_save_error, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, found)
}

// routeGetProject gets a project for the site by id
func routeGetProject(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

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

	// now we split; if they are an admin, they can see everything
	// if not, they see only the specified fields IF the project is valid
	if results.IsAdmin {
		sendAPIJSONData(w, http.StatusOK, found)
		return
	}

	// at this point, check for user access; send the same as not found to reduce enumerating
	if found.Status != ProjectStatusActive {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// we just send a smaller version of the return here
	ret := convertProjectToUserRet(found)
	sendAPIJSONData(w, http.StatusOK, ret)
}

// routeGetProjects gets the projects for a site
func routeGetProjects(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

	found, err := GetProjectsForSite(results.Site.ID, "all")
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// now we split; if they are an admin, they can see everything
	// if not, they see only the specified fields IF the project is valid
	if results.IsAdmin {
		sendAPIJSONData(w, http.StatusOK, found)
		return
	}

	// at this point, check for user access; send the same as not found to reduce enumerating
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

// routeLinkUserAndProject links a user to a project
func routeLinkUserAndProject(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if projectIDErr != nil || userIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	found, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// now we split; if they are an admin, they can do everything
	if results.IsAdmin {
		err = LinkUserAndProject(userID, projectID)
		if err != nil {
			sendAPIError(w, api_error_project_link_err, err, map[string]string{})
			return
		}
		sendAPIJSONData(w, http.StatusOK, map[string]bool{
			"linked": true,
		})
		return
	}

	// first, make sure the user is who they say they are; obfuscate the reason
	if results.User.ID != userID {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	// at this point, check for user access; send the same as not found to reduce enumerating
	if found.Status != ProjectStatusActive {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	if found.SignupStatus == ProjectSignupStatusClosed {
		sendAPIError(w, api_error_project_signup_unavailable, err, map[string]string{})
		return
	}

	// if it requires a code, get the code
	if found.SignupStatus == ProjectSignupStatusWithCode {
		input := &ProjectUserLinkRequest{}
		render.Bind(r, input)
		if input.Code == "" || input.Code != found.ShortCode {
			// same as unavailable to prevent sniffing
			sendAPIError(w, api_error_project_signup_unavailable, errors.New("signup unavailable"), map[string]string{})
			return
		}
	}

	if found.MaxParticipants > 0 && found.ParticipantCount >= found.MaxParticipants {
		// same as unavailable to prevent sniffing
		sendAPIError(w, api_error_project_signup_unavailable, errors.New("signup unavailable"), map[string]string{})
		return
	}

	// at this point, they are good
	err = LinkUserAndProject(userID, projectID)
	if err != nil {
		sendAPIError(w, api_error_project_link_err, err, map[string]string{})
		return
	}

	// TODO: when modules and flows are built, come back and link

	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": true,
	})
}

// routeUnlinkUserAndProject unlinks a user to a project
func routeUnlinkUserAndProject(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if projectIDErr != nil || userIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	removeProgress := r.URL.Query().Get("remove") // if it's anything other than blank, we remove it

	found, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// now we split; if they are an admin, they can do everything
	if results.IsAdmin {
		err = UnlinkUserAndProject(userID, projectID)
		if err != nil {
			sendAPIError(w, api_error_project_link_err, err, map[string]string{})
			return
		}
		sendAPIJSONData(w, http.StatusOK, map[string]bool{
			"linked": false,
		})

		if removeProgress != "" {
			// TODO: remove the progress after the HTTP call is returned to the client
		}
		return
	}

	// first, make sure the user is who they say they are; obfuscate the reason
	if results.User.ID != userID {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	// at this point, check for user access; send the same as not found to reduce enumerating
	// if the project is over, we don't want users removing their information
	if found.Status != ProjectStatusActive {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// at this point, they are good
	err = UnlinkUserAndProject(userID, projectID)
	if err != nil {
		sendAPIError(w, api_error_project_unlink_err, err, map[string]string{})
		return
	}

	if removeProgress != "" {
		// TODO: remove the progress after the HTTP call is returned to the client
	}

	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": false,
	})
}
