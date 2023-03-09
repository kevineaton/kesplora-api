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

	block, err := GetModuleBlockForParticipant(user.ID, projectID, moduleID, blockID)
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

	if block.UserStatus == BlockUserStatusNotStarted {
		SaveBlockUserStatusForParticipant(&BlockUserStatus{
			UserID:     user.ID,
			ProjectID:  projectID,
			ModuleID:   moduleID,
			BlockID:    blockID,
			UserStatus: BlockUserStatusStarted,
		})
		block.UserStatus = BlockUserStatusStarted
	}
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
	block, err := GetModuleBlockForParticipant(user.ID, projectID, moduleID, blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	// TODO: if block is a form, they can't use this endpoint
	if block.BlockType == BlockTypeForm {
		sendAPIError(w, api_error_block_status_form, errors.New("incorrect path"), map[string]interface{}{})
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
		sendAPIError(w, api_error_block_status_save, err, map[string]interface{}{})
		return
	}

	projectStatus, err := CheckProjectParticipantStatusForParticipant(user.ID, projectID)
	if err == nil {
		// update the user status
		err = UpdateUserAndProjectStatus(user.ID, projectID, projectStatus)
		if err != nil {
			if projectStatus == BlockUserStatusCompleted {
				// set the complete message
				project, _ := GetProjectByID(projectID)
				input.ProjectUserStatus = projectStatus
				input.ProjectCompleteMessage = project.CompleteMessage
			}
		}
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
		err = RemoveAllProgressForParticipantAndModule(user.ID, moduleID)
	} else if projectID != 0 {
		err = RemoveAllProgressForParticipantAndFlow(user.ID, projectID)
	}

	if err != nil {
		sendAPIError(w, api_error_block_status_save, err, map[string]interface{}{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"reset": true,
	})
}

func routeParticipantSaveFormResponse(w http.ResponseWriter, r *http.Request) {
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

	// make sure this project/module/block/user combo is valid
	_, err = GetModuleBlockForParticipant(user.ID, projectID, moduleID, blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	form, err := GetBlockFormByBlockID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	questions, err := GetBlockFormQuestionsForBlockID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	input := &BlockFormQestionResponseInput{}
	render.Bind(r, input)

	// TODO: check for existing submissions

	// now, create a new submission
	submission := &BlockFormSubmission{
		BlockID: blockID,
		UserID:  user.ID,
	}
	err = CreateBlockFormSubmission(submission)
	if err != nil {
		// TODO: replace with better error
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	// since these are small, loops in loops aren't a big deal
	// we will ignore the ids in the input and match it up
	for _, question := range questions {
		for _, response := range input.Responses {
			if question.ID == response.QuestionID {
				// insert it
				insert := &BlockFormSubmissionResponse{
					SubmissionID: submission.ID,
					QuestionID:   question.ID,
					OptionID:     response.OptionID,
					TextResponse: response.TextResponse,
					IsCorrect:    BlockFormSubmissionResponseIsCorrectNA,
				}
				// loop over the options for the text; short and long won't have options, so it will have to be set after
				if question.QuestionType != BlockFormQuestionTypeShort && question.QuestionType != BlockFormQuestionTypeLong {
					for _, option := range question.Options {
						if response.OptionID == option.ID {
							insert.TextResponse = option.OptionText
							insert.IsCorrect = option.OptionIsCorrect
							break
						}
					}
				} else if form.FormType == BlockFormTypeQuiz {
					insert.IsCorrect = BlockFormSubmissionResponseIsCorrectPending
				}
				err = CreateBlockFormSubmissionResponse(insert)
				if err != nil {
					fmt.Printf("\nErr on response: %+v\n%+v\n", err, insert)
				}
				submission.Responses = append(submission.Responses, *insert)

				// continue to next questions IF this isn't a multi
				if question.QuestionType != BlockFormQuestionTypeMultiple {
					break
				}
			}
		}
	}

	if form.FormType == BlockFormTypeQuiz {
		// TODO: add in calculating the whole of the submission if it's a quiz and then update
	}

	// mark as completed
	status := &BlockUserStatus{
		UserID:        user.ID,
		ProjectID:     projectID,
		ModuleID:      moduleID,
		BlockID:       blockID,
		LastUpdatedOn: time.Now().Format(timeFormatAPI),
		UserStatus:    BlockUserStatusCompleted,
	}
	err = SaveBlockUserStatusForParticipant(status)
	if err != nil {
		sendAPIError(w, api_error_block_status_save, err, map[string]interface{}{})
		return
	}

	sendAPIJSONData(w, http.StatusOK, submission)
}

func routeParticipantGetFormSubmissions(w http.ResponseWriter, r *http.Request) {
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

	_, err = GetModuleBlockForParticipant(user.ID, projectID, moduleID, blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	submissions, err := GetBlockFormSubmissionsForUser(user.ID, blockID)
	if err != nil {
		// TODO: err
		fmt.Printf("|ERR: %+v\n", err)
	}
	for i := range submissions {
		submissions[i].Responses, _ = GetBlockFormSubmissionResponsesForSubmission(submissions[i].ID)
	}

	sendAPIJSONData(w, http.StatusOK, submissions)
}

func routeParticipantDeleteSubmissions(w http.ResponseWriter, r *http.Request) {
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

	_, err = GetModuleBlockForParticipant(user.ID, projectID, moduleID, blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]interface{}{})
		return
	}

	submissions, err := GetBlockFormSubmissionsForUser(user.ID, blockID)
	if err != nil {
		fmt.Printf("err: %+v\n", err)
	}

	for i := range submissions {
		err = DeleteBlockFormSubmission(submissions[i].ID)
		if err != nil {
			fmt.Printf("err: %+v\n", err)
		}
	}

	status := &BlockUserStatus{
		UserID:        user.ID,
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
