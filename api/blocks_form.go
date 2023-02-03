package api

import (
	"net/http"
	"time"
)

const (
	BlockFormTypeSurvey = "survey"
	BlockFormTypeQuiz   = "quiz"

	BlockFormQuestionTypeMultiple = "multiple" // multiple choice
	BlockFormQuestionTypeSingle   = "single"   // single choice
	BlockFormQuestionTypeShort    = "short"    // short text
	BlockFormQuestionTypeLong     = "long"     // long text
)

// BlockForm is a form block, such as a survey or quiz for a check on learning
type BlockForm struct {
	BlockID  int64  `json:"blockId" db:"blockId"`
	FormType string `json:"formType" db:"formType"`
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

// BlockFormQuestionResponse is a user's response to a question in a form; it can be
// either an option (single/multi) or text (optionId is 0 and question type is short/long)
type BlockFormQuestionResponse struct {
	ID           int64  `json:"id" db:"id"`
	BlockID      int64  `json:"blockId" db:"blockId"`
	QuestionId   int64  `json:"" db:""`
	UserID       int64  `json:"userId" db:"userId"`
	OptionID     int64  `json:"optionId" db:"optionId"`
	Submitted    string `json:"submitted" db:"submitted"`
	TextResponse string `json:"textResponse" db:"textResponse"`
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
	// TODO: backfill
	return nil
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
	question = :question
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
	question = :question
	formOrder = :formOrder
	WHERE id = :id`, input)
	return err

}

// DeleteBlockFormQuestion deletes a question and all responses / options
func DeleteBlockFormQuestion(id int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockFormQuestions WHERE id = ?`, id)
	// TODO: backfill
	return err
}

// DeleteBlockFormQuestionsForBlock deletes all questions and all responses / options for a block
func DeleteBlockFormQuestionsForBlock(id int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockFormQuestions WHERE blockId = ?`, id)
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

//
// Responses
//

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
func (data *BlockFormQuestionResponse) Bind(r *http.Request) error {
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

func (input *BlockFormQuestionResponse) processForDB() {
	if input.Submitted == "" {
		input.Submitted = time.Now().Format(timeFormatDB)
	} else {
		input.Submitted, _ = parseTimeToTimeFormat(input.Submitted, timeFormatDB)
	}
}

func (input *BlockFormQuestionResponse) processForAPI() {
	input.Submitted, _ = parseTimeToTimeFormat(input.Submitted, timeFormatAPI)
}
