package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// routeParticipantGetProjects gets all projects on a site for a participant
func routeParticipantGetProjects(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil || user == nil {
		// this should not have happened
		sendAPIError(w, api_error_auth_missing, errors.New("missing user"), nil)
		return
	}

	found, err := GetProjectsForParticipant(user.ID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}
	available := []ProjectAPIReturnNonAdmin{}
	for i := range found {
		available = append(available, *convertProjectToUserRet(&found[i]))
	}

	sendAPIJSONData(w, http.StatusOK, available)
}

// routeParticipantGetProject gets the project information for a single project that
// a participant is connected to
func routeParticipantGetProject(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil || user == nil {
		// this should not have happened
		sendAPIError(w, api_error_auth_missing, errors.New("missing user"), nil)
		return
	}

	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	found, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	connected := IsUserInProject(user.ID, projectID)
	if !connected {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}
	available := convertProjectToUserRet(found)
	sendAPIJSONData(w, http.StatusOK, available)
}

// routeParticipantGetProjectFlow gets the user's flow for a project
func routeParticipantGetProjectFlow(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil || user == nil {
		// this should not have happened
		sendAPIError(w, api_error_auth_missing, errors.New("missing user"), nil)
		return
	}

	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	_, err = GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	connected := IsUserInProject(user.ID, projectID)
	if !connected {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	flow, err := GetProjectFlowForParticipant(user.ID, projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, flow)
}

// routeParticipantUnlinkUserAndProject unlinks a user to a project
func routeParticipantUnlinkUserAndProject(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil || user == nil {
		// this should not have happened
		sendAPIError(w, api_error_auth_missing, errors.New("missing user"), nil)
		return
	}

	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	removeProgress := r.URL.Query().Get("remove") // if it's anything other than blank, we remove it

	found, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// at this point, check for user access; send the same as not found to reduce enumerating
	// if the project is over, we don't want users removing their information
	if found.Status != ProjectStatusActive {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// at this point, they are good; we don't care if they don't actually belong
	// to the project, since it wouldn't matter
	err = UnlinkUserAndProject(user.ID, projectID)
	if err != nil {
		sendAPIError(w, api_error_project_unlink_err, err, map[string]string{})
		return
	}

	if removeProgress != "" {
		err = RemoveAllProgressForParticipantAndFlow(user.ID, projectID)
		if err != nil {
			sendAPIError(w, api_error_project_unlink_err, err, map[string]string{})
			return
		}
	}

	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": false,
	})
}
