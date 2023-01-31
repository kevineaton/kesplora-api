package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// routeAdminSaveConsentForm creates OR updates the consent form
func routeAdminSaveConsentForm(w http.ResponseWriter, r *http.Request) {
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

// routeAdminDeleteConsentForm deletes a consent form
func routeAdminDeleteConsentForm(w http.ResponseWriter, r *http.Request) {
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
// Responses
//

// routeAdminGetConsentResponses gets the Consent responses. Admin only.
func routeAdminGetConsentResponses(w http.ResponseWriter, r *http.Request) {
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

// routeAdminGetConsentResponse gets a response and then validates it is visible to the caller
func routeAdminGetConsentResponse(w http.ResponseWriter, r *http.Request) {
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

	sendAPIJSONData(w, http.StatusOK, response)
}

// routeAdminDeleteConsentResponse deletes the consent form AND erases the user from the project
func routeAdminDeleteConsentResponse(w http.ResponseWriter, r *http.Request) {
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

	response, err := GetConsentResponseByID(responseID)
	if err != nil {
		sendAPIError(w, api_error_consent_response_get_err, err, nil)
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
