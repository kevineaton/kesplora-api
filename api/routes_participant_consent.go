package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

func routeParticipantGetConsentResponse(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{ // TODO: we could probably have a helper to just get the user
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

	if results.User.ID != response.ParticipantID {
		sendAPIError(w, api_error_consent_response_get_err, errors.New("unavailable"), nil)
		return
	}
	sendAPIJSONData(w, http.StatusOK, response)
}

// routeParticipantDeleteConsentResponse deletes the consent form AND erases the user from the project // TODO: refactor
func routeParticipantDeleteConsentResponse(w http.ResponseWriter, r *http.Request) {
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

	if results.User.ID != response.ParticipantID {
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
