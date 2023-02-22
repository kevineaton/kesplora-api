package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	BlockFormTypeSurvey = "survey"
	BlockFormTypeQuiz   = "quiz"

	BlockFormQuestionTypeExplanation = "explanation" // just text, no response allowed
	BlockFormQuestionTypeMultiple    = "multiple"    // multiple choice
	BlockFormQuestionTypeSingle      = "single"      // single choice
	BlockFormQuestionTypeShort       = "short"       // short text
	BlockFormQuestionTypeLong        = "long"        // long text
	BlockFormQuestionTypeLikert5     = "likert5"     // likert5
	BlockFormQuestionTypeLikert7     = "likert7"     // likert7

	BlockFormSubmissionResultsNA         = "na"
	BlockFormSubmissionResultsNeedsInput = "needs_input"
	BlockFormSubmissionResultsPassed     = "passed"
	BlockFormSubmissionResultsFailed     = "failed"

	BlockFormSubmissionResponseIsCorrectNA      = "na"
	BlockFormSubmissionResponseIsCorrectPending = "pending"
	BlockFormSubmissionResponseIsCorrectYes     = "yes"
	BlockFormSubmissionResponseIsCorrectNo      = "no"
)

// BlockForm is a form block, such as a survey or quiz for a check on learning
type BlockForm struct {
	BlockID       int64               `json:"blockId" db:"blockId"`
	FormType      string              `json:"formType" db:"formType"`
	AllowResubmit string              `json:"allowResubmit" db:"allowResubmit"`
	Questions     []BlockFormQuestion `json:"questions"`
}

// BlockFormQuestion is a question in a form
type BlockFormQuestion struct {
	ID           int64  `json:"id" db:"id"`
	BlockID      int64  `json:"blockId" db:"blockId"`
	QuestionType string `json:"questionType" db:"questionType"`
	Question     string `json:"question" db:"question"`
	FormOrder    int64  `json:"formOrder" db:"formOrder"`

	// needed for the return
	Options []BlockFormQuestionOption `json:"options"`
}

// BlockFormQuestionOption are the options for a single or multiple choice question
type BlockFormQuestionOption struct {
	ID              int64  `json:"id" db:"id"`
	QuestionID      int64  `json:"questionId" db:"questionId"`
	OptionText      string `json:"optionText" db:"optionText"`
	OptionOrder     int64  `json:"optionOrder" db:"optionOrder"`
	OptionIsCorrect string `json:"optionIsCorrect" db:"optionIsCorrect"`
}

// BlockFormSubmission is a completed submission which will like in responses
type BlockFormSubmission struct {
	ID        int64  `json:"id" db:"id"`
	BlockID   int64  `json:"blockId" db:"blockId"`
	UserID    int64  `json:"userId" db:"userId"`
	Submitted string `json:"submitted" db:"submitted"`
	Results   string `json:"results" db:"results"`
	// needed for the return
	Responses []BlockFormSubmissionResponse `json:"responses"`
}

// BlockFormSubmissionResponse is a user's response to a question in a form; it can be
// either an option (single/multi) or text (optionId is 0 and question type is short/long)
type BlockFormSubmissionResponse struct {
	ID           int64  `json:"id" db:"id"`
	SubmissionID int64  `json:"submissionId" db:"submissionId"`
	QuestionID   int64  `json:"questionId" db:"questionId"`
	OptionID     int64  `json:"optionId" db:"optionId"`
	TextResponse string `json:"textResponse" db:"textResponse"`
	IsCorrect    string `json:"isCorrect" db:"isCorrect"`
}

// BlockFormQestionResponseInput is a helper input for sending many responses at once to the API
type BlockFormQestionResponseInput struct {
	Responses []BlockFormSubmissionResponse `json:"responses"`
}

//
// Forms
//

// SaveBlockForm creates or updates the limited sub-meta-data
func SaveBlockForm(input *BlockForm) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`INSERT INTO BlockForm SET blockId = :blockId, formType = :formType ON DUPLICATE KEY UPDATE formType = :formType`, input)
	return err
}

// GetBlockFormByBlockID gets the form metadata but not the rest of the data
func GetBlockFormByBlockID(blockID int64) (*BlockForm, error) {
	form := &BlockForm{}
	defer form.processForAPI()
	err := config.DBConnection.Get(form, `SELECT * FROM BlockForm WHERE blockId = ?`, blockID)
	return form, err
}

// DeleteBlockFormByBlockID deletes a form and all of the questions / options / responses
func DeleteBlockFormByBlockID(blockID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockForm WHERE blockId = ?`, blockID)
	if err != nil {
		return err
	}
	err = DeleteBlockFormQuestionsForBlock(blockID)
	return err
}

//
// Questions
//

// CreateBlockFormQuestion creates a new question
func CreateBlockFormQuestion(input *BlockFormQuestion) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO BlockFormQuestions SET 
	blockId = :blockId,
	questionType = :questionType,
	question = :question,
	formOrder = :formOrder`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateBlockFormQuestion updates most fields in the question
func UpdateBlockFormQuestion(input *BlockFormQuestion) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE BlockFormQuestions SET 
	questionType = :questionType,
	question = :question,
	formOrder = :formOrder
	WHERE id = :id`, input)
	return err

}

// DeleteBlockFormQuestion deletes a question and all responses / options
func DeleteBlockFormQuestion(questionID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockFormQuestions WHERE id = ?`, questionID)
	// TODO: backfill
	return err
}

// DeleteBlockFormQuestionsForBlock deletes all questions and all responses / options for a block
func DeleteBlockFormQuestionsForBlock(blockID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockFormQuestions WHERE blockId = ?`, blockID)
	// TODO: backfill
	return err
}

// GetBlockFormQuestionsForBlockID gets the questions and options for a block
func GetBlockFormQuestionsForBlockID(blockID int64) ([]BlockFormQuestion, error) {
	questions := []BlockFormQuestion{}
	err := config.DBConnection.Select(&questions, `SELECT * FROM BlockFormQuestions WHERE blockId = ? ORDER BY formOrder`, blockID)
	if err != nil {
		return questions, err
	}
	for i := range questions {
		options, err := GetBlockFormQuestionOptionForQuestion(questions[i].ID)
		if err == nil {
			questions[i].Options = options
		}
		questions[i].processForAPI()
	}
	return questions, err
}

//
// Question Options
//

// CreateBlockFormQuestionOption creates a new option
func CreateBlockFormQuestionOption(input *BlockFormQuestionOption) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO BlockFormQuestionOptions SET 
	questionId = :questionId,
	optionText = :optionText,
	optionOrder = :optionOrder,
	optionIsCorrect = :optionIsCorrect`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateBlockFormQuestionOption updates most fields for an option
func UpdateBlockFormQuestionOption(input *BlockFormQuestionOption) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE BlockFormQuestionOptions SET 
	optionText = :optionText,
	optionOrder = :optionOrder,
	optionIsCorrect = :optionIsCorrect
	WHERE id = :id`, input)
	return err
}

// DeleteBlockFormQuestionOption deletes an option
func DeleteBlockFormQuestionOption(id int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockFormQuestionOptions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	// TODO: backfill
	return nil
}

// GetBlockFormQuestionOptionForQuestion gets the options for a question
func GetBlockFormQuestionOptionForQuestion(questionID int64) ([]BlockFormQuestionOption, error) {
	options := []BlockFormQuestionOption{}
	err := config.DBConnection.Select(&options, `SELECT * FROM BlockFormQuestionOptions WHERE questionId = ? ORDER BY optionOrder`, questionID)
	for i := range options {
		options[i].processForAPI()
	}
	return options, err
}

func HandleSaveBlockForm(content *BlockForm) error {
	// first, create/save the block form
	err := SaveBlockForm(content)
	if err != nil {
		return err
	}
	// here we need to loop through the questions, then
	// through the options; if the id is 0, if create; if not
	// update
	errors := []error{}
	for _, question := range content.Questions {
		// update or create
		question.BlockID = content.BlockID
		if question.ID == 0 {
			err := CreateBlockFormQuestion(&question)
			if err != nil {
				errors = append(errors, err)
			}
		} else {
			err := UpdateBlockFormQuestion(&question)
			if err != nil {
				errors = append(errors, err)
			}
		}
		// now handle the options
		for _, option := range question.Options {
			option.QuestionID = question.ID
			if option.ID == 0 {
				err := CreateBlockFormQuestionOption(&option)
				if err != nil {
					errors = append(errors, err)
				}
			} else {
				err := UpdateBlockFormQuestionOption(&option)
				if err != nil {
					errors = append(errors, err)
				}
			}
		}
	}
	if len(errors) != 0 {
		// join the errors
		errStringBuffer := strings.Builder{}
		for i := range errors {
			errStringBuffer.WriteString(errors[i].Error() + "; ")
		}
		return fmt.Errorf("errors found: %s", errStringBuffer.String())
	}
	return nil
}

//
// Submissions
//

// CreateBlockFormSubmission creates a new submission to link in individual responses
func CreateBlockFormSubmission(input *BlockFormSubmission) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO BlockFormSubmissions SET
		blockId = :blockId,
		userId = :userId,
		submitted = :submitted,
		results = :results
	`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateBlockFormSubmission updates a submission's information; note that we allow changing
// most of the fields, but really the only fields that should change it the results field
func UpdateBlockFormSubmission(input *BlockFormSubmission) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE BlockFormSubmissions SET
		blockId = :blockId,
		userId = :userId,
		submitted = :submitted,
		results = :results
		WHERE id = :id
	`, input)
	return err
}

// GetBlockFormSubmissionByID gets a specific response
func GetBlockFormSubmissionByID(id int64) (*BlockFormSubmission, error) {
	submission := &BlockFormSubmission{}
	defer submission.processForAPI()
	err := config.DBConnection.Get(submission, `SELECT * FROM BlockFormSubmissions WHERE id = ?`, id)
	return submission, err
}

// GetBlockFormSubmissionByID gets a specific response
func GetBlockFormSubmissionsForUser(userID, blockID int64) ([]BlockFormSubmission, error) {
	submissions := []BlockFormSubmission{}
	err := config.DBConnection.Select(&submissions, `SELECT * FROM BlockFormSubmissions WHERE userId = ? AND blockId = ? ORDER BY submitted`, userID, blockID)
	for i := range submissions {
		submissions[i].processForAPI()
	}
	return submissions, err
}

// DeleteBlockFormSubmission deletes a submission and connected responses
func DeleteBlockFormSubmission(submissionID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockFormSubmissions WHERE id = ?`, submissionID)
	if err != nil {
		return err
	}
	err = DeleteBlockFormSubmissionResponses(submissionID)
	return err
}

//
// Submission Responses
//

// CreateBlockFormSubmissionResponse creates a new individual question response
func CreateBlockFormSubmissionResponse(input *BlockFormSubmissionResponse) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO BlockFormSubmissionResponses SET
		submissionId = :submissionId,
		questionId = :questionId,
		optionId = :optionId,
		textResponse = :textResponse,
		isCorrect = :isCorrect
	`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateBlockFormSubmissionResponse updates an individual response; note that we allow changing
// most of the fields, but really the only fields that should change are the isCorrect and text
// fields
func UpdateBlockFormSubmissionResponse(input *BlockFormSubmissionResponse) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE BlockFormSubmissionResponses SET
		submissionId = :blockId,
		questionId = :questionId,
		optionId = :optionId,
		textResponse = :textResponse,
		isCorrect = :isCorrect
		WHERE id = :id
	`, input)
	return err
}

// GetBlockFormSubmissionResponsesForSubmission gets all of the responses for a submission
func GetBlockFormSubmissionResponsesForSubmission(submissionID int64) ([]BlockFormSubmissionResponse, error) {
	responses := []BlockFormSubmissionResponse{}
	err := config.DBConnection.Select(&responses, `SELECT * FROM BlockFormSubmissionResponses WHERE submissionId = ?`, submissionID)
	for i := range responses {
		responses[i].processForAPI()
	}
	return responses, err
}

// DeleteBlockFormSubmissionResponses deletes all of the responses for a submission
func DeleteBlockFormSubmissionResponses(submissionID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockFormSubmissionResponses WHERE submissionId = ?`, submissionID)
	return err
}

//
// Binds
//

// Bind binds the data for the HTTP
func (data *BlockForm) Bind(r *http.Request) error {
	return nil
}

// Bind binds the data for the HTTP
func (data *BlockFormQuestion) Bind(r *http.Request) error {
	return nil
}

// Bind binds the data for the HTTP
func (data *BlockFormQuestionOption) Bind(r *http.Request) error {
	return nil
}

// Bind binds the data for the HTTP
func (data *BlockFormSubmission) Bind(r *http.Request) error {
	return nil
}

// Bind binds the data for the HTTP
func (data *BlockFormSubmissionResponse) Bind(r *http.Request) error {
	return nil
}

// Bind binds the data for the HTTP
func (data *BlockFormQestionResponseInput) Bind(r *http.Request) error {
	return nil
}

//
// Processors
//

func (input *BlockForm) processForDB() {
	if input.FormType == "" {
		input.FormType = BlockFormTypeSurvey
	}
}

func (input *BlockForm) processForAPI() {
	// nothing to do
}

func (input *BlockFormQuestion) processForDB() {
	if input.QuestionType == "" {
		input.QuestionType = BlockFormQuestionTypeShort
	}
}

func (input *BlockFormQuestion) processForAPI() {
	// nothing to do
}

func (input *BlockFormQuestionOption) processForDB() {
	if input.OptionIsCorrect == "" {
		input.OptionIsCorrect = "na"
	}
}

func (input *BlockFormQuestionOption) processForAPI() {
	// nothing to do
}

func (input *BlockFormSubmission) processForDB() {
	if input.Submitted == "" {
		input.Submitted = time.Now().Format(timeFormatDB)
	} else {
		input.Submitted, _ = parseTimeToTimeFormat(input.Submitted, timeFormatDB)
	}
	if input.Results == "" {
		input.Results = BlockFormSubmissionResultsNA
	}
}

func (input *BlockFormSubmission) processForAPI() {
	input.Submitted, _ = parseTimeToTimeFormat(input.Submitted, timeFormatAPI)
}

func (input *BlockFormSubmissionResponse) processForDB() {
	if input.IsCorrect == "" {
		input.IsCorrect = BlockFormSubmissionResponseIsCorrectNA
	}
}

func (input *BlockFormSubmissionResponse) processForAPI() {
	// nothing to do
}
