package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// routeAdminCreateProject creates a new project for the site
func routeAdminCreateProject(w http.ResponseWriter, r *http.Request) {
	site, err := GetSiteFromContext(r.Context())
	if site == nil || err != nil {
		// this is odd, as it should have been set at the middleware
		sendAPIError(w, api_error_site_get_error, err, nil)
		return
	}

	input := &Project{}
	render.Bind(r, input)

	// only the name is required
	if input.Name == "" {
		sendAPIError(w, api_error_project_missing_data, errors.New("missing name"), map[string]string{})
		return
	}
	input.SiteID = site.ID

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

	err = CreateProject(input)
	if err != nil {
		sendAPIError(w, api_error_project_save_error, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusCreated, input)
}

// routeAdminUpdateProject updates an existing project for the site
func routeAdminUpdateProject(w http.ResponseWriter, r *http.Request) {
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
	if input.CompleteMessage != found.CompleteMessage {
		found.CompleteMessage = input.CompleteMessage
	}
	if input.CompleteRule != found.CompleteRule {
		found.CompleteRule = input.CompleteRule
	}
	if input.FlowRule != found.FlowRule {
		found.FlowRule = input.FlowRule
	}
	if input.StartRule != found.StartRule {
		found.StartRule = input.StartRule
	}
	if input.StartDate != found.StartDate {
		found.StartDate = input.StartDate
	}
	if input.EndDate != found.EndDate {
		found.EndDate = input.EndDate
	}

	err = UpdateProject(found)
	if err != nil {
		sendAPIError(w, api_error_project_save_error, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, found)
}

// routeAdminGetProject gets a project for the site by id
func routeAdminGetProject(w http.ResponseWriter, r *http.Request) {
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
	sendAPIJSONData(w, http.StatusOK, found)
}

// routeAdminGetProjects gets the projects for a site
func routeAdminGetProjects(w http.ResponseWriter, r *http.Request) {
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
	sendAPIJSONData(w, http.StatusOK, found)
}

// routeAdminLinkUserAndProject links a user to a project
func routeAdminLinkUserAndProject(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if projectIDErr != nil || userIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	_, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// if they are an admin, they can do everything, so ignore pre-reqs
	err = LinkUserAndProject(userID, projectID)
	if err != nil {
		sendAPIError(w, api_error_project_link_err, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": true,
	})
}

// routeAdminUnlinkUserAndProject unlinks a user to a project
func routeAdminUnlinkUserAndProject(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if projectIDErr != nil || userIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	removeProgress := r.URL.Query().Get("remove") // if it's anything other than blank, we remove it

	_, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// if they are an admin, they can do everything
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
		// implement when participant progress is built
	}

}
