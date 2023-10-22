package api

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	// first, dedupe on the user if
	deDupe := map[int64]int64{}
	for _, r := range results {
		if _, ok := deDupe[r.UserID]; !ok {
			deDupe[r.UserID] = r.DaysAgo
		}
	}

	// we need to convert this, since the results are a slice of maps of days ago to user id, so we group on days
	processed := []ReportValueCount{}
	holder := map[int64]int64{} // days ago -> count
	for _, daysAgo := range deDupe {
		if _, ok := holder[daysAgo]; !ok {
			holder[daysAgo] = 0
		}
		holder[daysAgo]++
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
		if question.QuestionType != BlockFormQuestionTypeExplanation {
			questionHeader = append(questionHeader, question.Question)
		}
	}
	responsesResult = append(responsesResult, questionHeader)

	// so, because of the way the questions are set up, we need to loop over the responses in the submissions,
	// match with the header, add to the  results, and then move on, which means this can be large and
	// it's probably a good area to revisit

	for _, submission := range submissions {
		responses, _ := GetBlockFormSubmissionResponsesForSubmission(submission.ID)
		if len(responses) == 0 {
			// not worth putting, so continue
			continue
		}
		responsesSlice := []string{}

		// first, we need to loop over the responses and for any of the multiples, we need to string concat them
		// with pipes
		questionsToResponses := map[string]string{}
		for _, response := range responses {
			current, ok := questionsToResponses[response.QuestionText]
			if !ok {
				questionsToResponses[response.QuestionText] = response.TextResponse
			} else {
				// it's probably a multiple if there's a bunch in here
				current += "|" + strings.Trim(response.TextResponse, "\t")
				questionsToResponses[response.QuestionText] = current
			}
		}
		for _, question := range questionHeader {
			for questionText, responseText := range questionsToResponses {
				if questionText == question {
					responsesSlice = append(responsesSlice, strings.Trim(responseText, "\t"))
				}
			}
		}

		responsesResult = append(responsesResult, responsesSlice)
	}

	// now create a csv
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=submissions_%d_%d_%d.csv", projectID, moduleID, blockID))
	w.WriteHeader(http.StatusOK)
	wr := csv.NewWriter(w)
	wr.WriteAll(responsesResult)

}
