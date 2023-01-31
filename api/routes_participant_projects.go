package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

// routeParticipantGetBlock gets the block information scoped to a user/project/module
func routeParticipantGetBlock(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil || user == nil {
		// this should not have happened
		sendAPIError(w, api_error_auth_missing, errors.New("missing user"), nil)
		return
	}

	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if projectIDErr != nil || moduleIDErr != nil || blockIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	connected := IsUserInProject(user.ID, projectID)
	if !connected {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	block, err := GetModuleBlockForparticipant(user.ID, projectID, moduleID, blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	content, err := handleBlockGet(block.BlockType, block.ID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	block.Content = content
	sendAPIJSONData(w, http.StatusOK, block)
}

// routeParticipantSaveBlockStatus updates a participant's status
func routeParticipantSaveBlockStatus(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil || user == nil {
		// this should not have happened
		sendAPIError(w, api_error_auth_missing, errors.New("missing user"), nil)
		return
	}

	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if projectIDErr != nil || moduleIDErr != nil || blockIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	status := chi.URLParam(r, "status")
	if status != BlockUserStatusCompleted && status != BlockUserStatusNotStarted && status != BlockUserStatusStarted {
		sendAPIError(w, api_error_invalid_path, fmt.Errorf("invalid status: %s", status), map[string]string{})
		return
	}

	connected := IsUserInProject(user.ID, projectID)
	if !connected {
		sendAPIError(w, api_error_project_not_found, err, map[string]string{})
		return
	}

	// make sure this project/module/block/user combo is valid
	_, err = GetModuleBlockForparticipant(user.ID, projectID, moduleID, blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	input := &BlockUserStatus{
		UserID:        user.ID,
		ProjectID:     projectID,
		ModuleID:      moduleID,
		BlockID:       blockID,
		LastUpdatedOn: time.Now().Format(timeFormatAPI),
		UserStatus:    status,
	}
	err = SaveBlockUserStatusForParticipant(input)
	if err != nil {
		sendAPIError(w, api_error_block_status_save_err, err, map[string]interface{}{})
		return
	}

	sendAPIJSONData(w, http.StatusOK, input)
}

// routeParticipantRemoveBlockStatus removes the status entry for a project/module/block scope
func routeParticipantRemoveBlockStatus(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil || user == nil {
		// this should not have happened
		sendAPIError(w, api_error_auth_missing, errors.New("missing user"), nil)
		return
	}

	// this endpoint is used by all three possible paths, so we look at where the 0s are
	// therefore, we don't care about path errors; the default value will be a 0 if it is
	// not a valid number

	projectID, _ := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	moduleID, _ := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, _ := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)

	// we don't really need to care if the user is a participant or not as a delete is
	// going to end up the same regardless
	if blockID != 0 {
		err = RemoveAllProgressForParticipantAndBlock(user.ID, blockID)
	} else if moduleID != 0 {
		err = RemoveAllProgressForParticipantAndModule(user.ID, blockID)
	} else if projectID != 0 {
		err = RemoveAllProgressForParticipantAndFlow(user.ID, blockID)
	}

	if err != nil {
		sendAPIError(w, api_error_block_status_save_err, err, map[string]interface{}{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"reset": true,
	})
}
