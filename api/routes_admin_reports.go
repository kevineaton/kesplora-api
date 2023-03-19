package api

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// ReportGetCountOfUsersOnProjectByStatus gets a report of users on project by their status
func routeAdminReportGetCountOfUsersOnProjectByStatus(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}
	results, err := ReportGetCountOfUsersOnProjectByStatus(projectID)
	if err != nil {
		sendAPIError(w, api_error_reports_get, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, results)
}

// routeAdminReportGetCountOfLastUpdatedForProject gets count of users on project by last updated time
func routeAdminReportGetCountOfLastUpdatedForProject(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}
	results, err := ReportGetCountOfLastUpdatedForProject(projectID)
	if err != nil {
		sendAPIError(w, api_error_reports_get, err, map[string]string{})
		return
	}
	// we need to convert this, since the results are a slice of maps of days ago to user id, so we group on days
	processed := []ReportValueCount{}
	holder := map[int64]int64{} // days ago -> count
	for _, r := range results {
		if _, ok := holder[r.DaysAgo]; !ok {
			holder[r.DaysAgo] = 0
		}
		holder[r.DaysAgo]++
	}
	for days, count := range holder {
		processed = append(processed, ReportValueCount{Value: fmt.Sprintf("%d", days), Count: count})
	}
	sendAPIJSONData(w, http.StatusOK, processed)
}

// routeAdminReportGetCountOfStatusForProject gets all of the flows and the count of users with each status for that block
func routeAdminReportGetCountOfStatusForProject(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}
	results, err := ReportGetCountOfStatusForProject(projectID)
	if err != nil {
		sendAPIError(w, api_error_reports_get, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, results)
}

// routeAdminReportGetSubmissionCountForProject gets a report of all forms in the project and the count of their submissions for comparison
func routeAdminReportGetSubmissionCountForProject(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}
	results, err := ReportGetSubmissionCountForProject(projectID)
	if err != nil {
		sendAPIError(w, api_error_reports_get, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, results)
}

// routeAdminReportGetProjectSubmissionResponses gets the submissions responses report
func routeAdminReportGetProjectSubmissionResponses(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if projectIDErr != nil || moduleIDErr != nil || blockIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	// check the connections
	if !IsBlockInModule(moduleID, blockID) || !IsModuleInProject(projectID, moduleID) {
		sendAPIError(w, api_error_project_misconfiguration, errors.New("block or module aren't in that project"), map[string]interface{}{
			"projectID": projectID,
			"moduleID":  moduleID,
			"blockID":   blockID,
		})
		return
	}

	block, err := GetBlockByID(blockID)
	if err != nil {
		sendAPIError(w, api_error_block_not_found, err, map[string]string{})
		return
	}

	questions, err := GetBlockFormQuestionsForBlockID(blockID)
	if err != nil {
		sendAPIError(w, api_error_submission_fetch, err, map[string]string{})
		return
	}

	allResponses, err := GetBlockFormSubmissionResponsesForBlock(blockID)
	if err != nil {
		sendAPIError(w, api_error_submission_fetch, err, map[string]string{})
		return
	}

	// now we build this out
	result := ReportSubmissionResponses{
		BlockID:   block.ID,
		BlockName: block.Name,
	}

	for _, q := range questions {
		// loop on responses to map
		question := ReportSubmissionResponsesQuestion{
			QuestionID:   q.ID,
			QuestionText: q.Question,
			QuestionType: q.QuestionType,
		}

		options, _ := GetBlockFormQuestionOptionForQuestion(q.ID)
		for _, o := range options {
			rr := ReportSubmissionResponsesResponse{
				OptionID:     o.ID,
				TextResponse: o.OptionText,
				Count:        0,
			}
			question.Responses = append(question.Responses, rr)
		}
		for _, r := range allResponses {
			if r.QuestionID == q.ID {
				// we could probably put this in a map, but for now we will loop on the loop on the loop
				found := false
				for i, er := range question.Responses {
					if r.OptionID != 0 && er.OptionID == r.OptionID {
						question.Responses[i].Count++
						found = true
						break
					}
				}
				if !found {
					response := ReportSubmissionResponsesResponse{
						OptionID:     r.OptionID,
						TextResponse: r.TextResponse,
						Count:        1,
					}
					question.Responses = append(question.Responses, response)
				}
			}
		}

		result.Questions = append(result.Questions, question)
	}

	sendAPIJSONData(w, http.StatusOK, result)
}

// routeAdminReportExportProjectSubmissionResponses exports the submissions responses report
func routeAdminReportExportProjectSubmissionResponses(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	blockID, blockIDErr := strconv.ParseInt(chi.URLParam(r, "blockID"), 10, 64)
	if projectIDErr != nil || moduleIDErr != nil || blockIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	// check the connections
	if !IsBlockInModule(moduleID, blockID) || !IsModuleInProject(projectID, moduleID) {
		sendAPIError(w, api_error_project_misconfiguration, errors.New("block or module aren't in that project"), map[string]interface{}{
			"projectID": projectID,
			"moduleID":  moduleID,
			"blockID":   blockID,
		})
		return
	}

	questions, err := GetBlockFormQuestionsForBlockID(blockID)
	if err != nil {
		sendAPIError(w, api_error_submission_fetch, err, map[string]string{})
		return
	}

	submissions, err := GetBlockFormSubmissionsForBlock(blockID)
	if err != nil {
		sendAPIError(w, api_error_submission_fetch, err, map[string]string{})
		return
	}

	// for now we only allow CSV; json comes from the previous, non-export call
	questionHeader := []string{}
	responsesResult := [][]string{}

	for _, question := range questions {
		questionHeader = append(questionHeader, question.Question)
	}
	responsesResult = append(responsesResult, questionHeader)

	for _, submission := range submissions {
		responses, _ := GetBlockFormSubmissionResponsesForSubmission(submission.ID)
		if len(responses) == 0 {
			// not worth putting, so continue
			continue
		}
		// for each response, we need to make sure the response is in the same order as the question
		responseSlice := []string{}
		lastQuestion := ""
		for _, response := range responses {
			// build the slice by matching up the question text
			for i, q := range questionHeader {
				if response.QuestionText == q {
					// here's the kicker; for multiple choice, we separate with ;
					responseText := response.TextResponse
					if lastQuestion == q {
						// we are on the same question, so we append on the last one
						responseSlice[i] = fmt.Sprintf("%s;%s", responseSlice[i], responseText)
					} else {
						responseSlice = append(responseSlice, responseText)
					}
					lastQuestion = q
				}
			}
		}
		responsesResult = append(responsesResult, responseSlice)
	}

	// now create a csv
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=submissions_%d_%d_%d.csv", projectID, moduleID, blockID))
	w.WriteHeader(http.StatusOK)
	wr := csv.NewWriter(w)
	wr.WriteAll(responsesResult)

}
