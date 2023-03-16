package api

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// routeAdminCreateBlock creates a new block id and then processes the content for saving
func routeAdminCreateBlock(w http.ResponseWriter, r *http.Request) {
	blockType := chi.URLParam(r, "blockType")
	if !isValidBlockType(blockType) {
		sendAPIError(w, api_error_block_invalid_type, errors.New("invalid type"), map[string]string{
			"blockType": blockType,
		})
		return
	}

	input := &Block{}
	render.Bind(r, input)

	if input.Name == "" {
		sendAPIError(w, api_error_block_missing_data, errors.New("missing data"), map[string]interface{}{
			"input": input,
		})
		return
	}

	// validate the content before committing to a save
	err := handleBlockRequiredFields(blockType, input.Content)
	if err != nil {
		sendAPIError(w, api_error_block_content_missing_data, err, map[string]interface{}{
			"input": input,
		})
		return
	}

	err = CreateBlock(input)
	if err != nil {
		sendAPIError(w, api_error_block_save, err, map[string]interface{}{
			"input": input,
		})
		return
	}

	// now, hand off the save depending on the body of the content
	data, err := handleBlockSave(blockType, input.ID, input.Content)
	if err != nil {
		sendAPIError(w, api_error_block_save, err, map[string]interface{}{
			"input": input,
		})
		return
	}
	input.Content = data
	sendAPIJSONData(w, http.StatusCreated, input)
}

// routeAdminGetBlocksOnSite gets the meta data about all modules on the platform
func routeAdminGetBlocksOnSite(w http.ResponseWriter, r *http.Request) {
	blocks, err := GetBlocksForSite()
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, blocks)
}

// routeAdminGetBlocksForModule gets the blocks on a module
func routeAdminGetBlocksForModule(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	if moduleIDErr != nil {
		sendAPIError(w, api_error_invalid_path, moduleIDErr, map[string]string{})
		return
	}

	_, err := GetModuleByID(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{})
		return
	}

	blocks, err := GetBlocksForModule(moduleID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, blocks)
}

// routeAdminUnlinkAllBlocksFromModule unlinks all blocks from a module
func routeAdminUnlinkAllBlocksFromModule(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	if moduleIDErr != nil {
		sendAPIError(w, api_error_invalid_path, moduleIDErr, map[string]string{})
		return
	}

	_, err := GetModuleByID(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{})
		return
	}

	err = UnlinkAllBlocksFromModule(moduleID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"unlinkedAll": true,
	})
}

// routeAdminGetBlock gets the block and content
func routeAdminGetBlock(w http.ResponseWriter, r *http.Request) {
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if blockIDErr != nil {
		sendAPIError(w, api_error_invalid_path, blockIDErr, map[string]string{})
		return
	}

	block, err := GetBlockByID(blockID)
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

// routeAdminUpdateBlock updates a block and its content
func routeAdminUpdateBlock(w http.ResponseWriter, r *http.Request) {
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if blockIDErr != nil {
		sendAPIError(w, api_error_invalid_path, blockIDErr, map[string]string{})
		return
	}

	block, err := GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	input := &Block{}
	render.Bind(r, input)
	if block.Name != "" && block.Name != input.Name {
		block.Name = input.Name
	}
	if block.Summary != "" && block.Summary != input.Summary {
		block.Summary = input.Summary
	}
	err = UpdateBlock(block)
	if err != nil {
		sendAPIError(w, api_error_block_save, err, map[string]interface{}{
			"input": input,
		})
		return
	}
	// now, we take the content and send it straight through to saving
	if input.Content != nil {
		content, err := handleBlockSave(block.BlockType, block.ID, input.Content)
		if err != nil {
			sendAPIError(w, api_error_block_save, err, map[string]interface{}{
				"input": input,
			})
			return
		}
		block.Content = content
	}
	sendAPIJSONData(w, http.StatusOK, block)
}

// routeAdminDeleteBlock deletes a block and associated content
func routeAdminDeleteBlock(w http.ResponseWriter, r *http.Request) {
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if blockIDErr != nil {
		sendAPIError(w, api_error_invalid_path, blockIDErr, map[string]string{})
		return
	}

	block, err := GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}
	err = handleBlockDelete(block.BlockType, blockID)
	if err != nil {
		sendAPIError(w, api_error_block_delete, err, map[string]string{})
		return
	}
	err = DeleteBlock(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_delete, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
}

// routeAdminLinkBlockAndModule links a module and a block
func routeAdminLinkBlockAndModule(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	order, orderErr := strconv.ParseInt(chi.URLParam(r, "order"), 10, 64)
	if blockIDErr != nil || moduleIDErr != nil || orderErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	// make sure the module and block both exist
	_, err := GetModuleByID(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{})
		return
	}
	_, err = GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	err = LinkBlockAndModule(moduleID, blockID, order)
	if err != nil {
		sendAPIError(w, api_error_block_link, err, map[string]int64{
			"moduleID": moduleID,
			"blockID":  blockID,
			"order":    order,
		})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": true,
	})
}

// routeAdminUnlinkBlockAndModule unlinks a module and a block
func routeAdminUnlinkBlockAndModule(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if blockIDErr != nil || moduleIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	err := UnlinkBlockAndModule(moduleID, blockID)
	if err != nil {
		sendAPIError(w, api_error_block_unlink, err, map[string]int64{
			"moduleID": moduleID,
			"blockID":  blockID,
		})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": false,
	})
}

//
// Form Submission Special Cases
//

// routeAdminGetUserSubmissions gets a user's list of submissions for a form
func routeAdminGetUserSubmissions(w http.ResponseWriter, r *http.Request) {
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if blockIDErr != nil || userIDErr != nil || projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalidPath"), map[string]string{})
		return
	}

	if !IsUserInProject(userID, projectID) {
		sendAPIError(w, api_error_projects_user_not_in, errors.New("user not in that project"), map[string]string{})
		return
	}

	_, err := GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	submissions, err := GetBlockFormSubmissionsForUser(userID, blockID)
	if err != nil {
		sendAPIError(w, api_error_submission_fetch, err, map[string]string{})
		return
	}
	for i := range submissions {
		submissions[i].Responses, _ = GetBlockFormSubmissionResponsesForSubmission(submissions[i].ID)
	}
	sendAPIJSONData(w, http.StatusOK, submissions)
}

// routeAdminDeleteUserSubmissions deletes all submissions for a user
func routeAdminDeleteUserSubmissions(w http.ResponseWriter, r *http.Request) {
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if blockIDErr != nil || userIDErr != nil || projectIDErr != nil || moduleIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalidPath"), map[string]string{})
		return
	}

	if !IsUserInProject(userID, projectID) {
		sendAPIError(w, api_error_projects_user_not_in, errors.New("user not in that project"), map[string]string{})
		return
	}

	_, err := GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	submissions, err := GetBlockFormSubmissionsForUser(userID, blockID)
	if err != nil {
		sendAPIError(w, api_error_submission_fetch, err, map[string]string{})
		return
	}
	for i := range submissions {
		DeleteBlockFormSubmission(submissions[i].ID)
	}

	status := &BlockUserStatus{
		UserID:        userID,
		ProjectID:     projectID,
		ModuleID:      moduleID,
		BlockID:       blockID,
		LastUpdatedOn: time.Now().Format(timeFormatAPI),
		UserStatus:    BlockUserStatusNotStarted,
	}
	err = SaveBlockUserStatusForParticipant(status)
	if err != nil {
		sendAPIError(w, api_error_block_status_save, err, map[string]interface{}{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
}

// routeAdminGetUserSubmission gets a submission for a user
func routeAdminGetUserSubmission(w http.ResponseWriter, r *http.Request) {
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	_, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	submissionID, submissionIDErr := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if blockIDErr != nil || userIDErr != nil || projectIDErr != nil || moduleIDErr != nil || submissionIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalidPath"), map[string]string{})
		return
	}

	if !IsUserInProject(userID, projectID) {
		sendAPIError(w, api_error_projects_user_not_in, errors.New("user not in that project"), map[string]string{})
		return
	}

	_, err := GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	sub, err := GetBlockFormSubmissionByID(submissionID)
	if err != nil {
		sendAPIError(w, api_error_submission_fetch, err, map[string]string{})
		return
	}
	sub.Responses, _ = GetBlockFormSubmissionResponsesForSubmission(sub.ID)

	sendAPIJSONData(w, http.StatusOK, sub)
}

// routeAdminDeleteUserSubmission deletes a submission for a user
func routeAdminDeleteUserSubmission(w http.ResponseWriter, r *http.Request) {
	userID, userIDErr := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	submissionID, submissionIDErr := strconv.ParseInt(chi.URLParam(r, "submissionID"), 10, 64)
	if blockIDErr != nil || userIDErr != nil || projectIDErr != nil || moduleIDErr != nil || submissionIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalidPath"), map[string]string{})
		return
	}

	if !IsUserInProject(userID, projectID) {
		sendAPIError(w, api_error_projects_user_not_in, errors.New("user not in that project"), map[string]string{})
		return
	}

	_, err := GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	sub, err := GetBlockFormSubmissionByID(submissionID)
	if err != nil {
		sendAPIError(w, api_error_submission_fetch, err, map[string]string{})
		return
	}
	if sub.BlockID != blockID || sub.UserID != userID {
		sendAPIError(w, api_error_submission_mismatch, err, map[string]string{})
		return
	}

	err = DeleteBlockFormSubmission(submissionID)
	if err != nil {
		sendAPIError(w, api_error_submission_delete, err, map[string]string{})
		return
	}

	submissions, err := GetBlockFormSubmissionsForUser(userID, blockID)
	if err != nil {
		sendAPIError(w, api_error_submission_fetch, err, map[string]string{})
		return
	}

	if len(submissions) == 0 {
		status := &BlockUserStatus{
			UserID:        userID,
			ProjectID:     projectID,
			ModuleID:      moduleID,
			BlockID:       blockID,
			LastUpdatedOn: time.Now().Format(timeFormatAPI),
			UserStatus:    BlockUserStatusNotStarted,
		}
		err = SaveBlockUserStatusForParticipant(status)
		if err != nil {
			sendAPIError(w, api_error_block_status_save, err, map[string]interface{}{})
			return
		}
	}

	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
}
