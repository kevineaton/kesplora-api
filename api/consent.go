package api

import (
	"net/http"
	"time"
)

// ConsentForm is the consent form for a project
type ConsentForm struct {
	ProjectID                     int64  `json:"projectId" db:"projectId"`
	ContentInMarkdown             string `json:"contentInMarkdown" db:"contentInMarkdown"`
	ContactInformationDisplay     string `json:"contactInformationDisplay" db:"contactInformationDisplay"`
	InstitutionInformationDisplay string `json:"institutionInformationDisplay" db:"institutionInformationDisplay"`

	OverrideSaveIfParticipants bool `json:"overrideSaveIfParticipants,omitempty"` // used to allow updating a consent form after the project started
}

// ConsentResponse is the response to a consent form
type ConsentResponse struct {
	ID                                    int64  `json:"id" db:"id"`
	ProjectID                             int64  `json:"projectId" db:"projectId"`
	DateConsented                         string `json:"dateConsented" db:"dateConsented"`
	ConsentStatus                         string `json:"consentStatus" db:"consentStatus"`
	ParticipantComments                   string `json:"participantComments" db:"participantComments"`
	ResearcherComments                    string `json:"researcherComments" db:"researcherComments"`
	ParticipantProvidedFirstName          string `json:"participantProvidedFirstName" db:"participantProvidedFirstName"`
	ParticipantProvidedLastName           string `json:"participantProvidedLastName" db:"participantProvidedLastName"`
	ParticipantProvidedContactInformation string `json:"participantProvidedContactInformation" db:"participantProvidedContactInformation"`
	ParticipantID                         int64  `json:"participantId" db:"participantId"` // will be 0 if the project specifies to not link them

	ProjectCode string `json:"projectCode,omitempty"` // used for signup when the project needs a code

	// these are used when the project must be anonymous, so a new account is created during consent
	// and tied to a participant code
	User *User `json:"user"`
}

const (
	ConsentResponseStatusAccepted         = "accepted"
	ConsentResponseStatusAcceptedForOther = "accepted_for_other"
	ConsentResponseStatusDeclined         = "declined"
)

//
// Forms
//

// SaveConsentFormForProject creates or updates a consent form for a project
func SaveConsentFormForProject(input *ConsentForm) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`INSERT INTO ConsentForms SET projectId = :projectId, contentInMarkdown = :contentInMarkdown, 
		contactInformationDisplay = :contactInformationDisplay, institutionInformationDisplay = :institutionInformationDisplay
		ON DUPLICATE KEY UPDATE contentInMarkdown = :contentInMarkdown, 
		contactInformationDisplay = :contactInformationDisplay, institutionInformationDisplay = :institutionInformationDisplay`, input)
	return err
}

// GetConsentFormForProject gets the project's consent form
func GetConsentFormForProject(projectID int64) (*ConsentForm, error) {
	data := &ConsentForm{}
	defer data.processForAPI()
	err := config.DBConnection.Get(data, `SELECT c.* FROM ConsentForms c WHERE projectId = ?`, projectID)
	return data, err
}

// DeleteConsentFormForProject deletes the consent form for a project and all responses
func DeleteConsentFormForProject(projectID int64) error {
	err := DeleteConsentResponsesForProject(projectID)
	if err != nil {
		return err
	}

	_, err = config.DBConnection.Exec(`DELETE FROM ConsentForms WHERE projectId = ?`, projectID)

	return err
}

//
// Responses
//

// CreateConsentResponse creates a new response
func CreateConsentResponse(input *ConsentResponse) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO ConsentResponses SET 
		projectId = :projectId,
		dateConsented = :dateConsented,
		consentStatus = :consentStatus,
		participantComments = :participantComments,
		researcherComments = :researcherComments,
		participantId = :participantId,
		participantProvidedFirstName = :participantProvidedFirstName,
		participantProvidedLastName = :participantProvidedLastName,
		participantProvidedContactInformation = :participantProvidedContactInformation`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// GetConsentResponsesForProject gets all of the responses for a project
func GetConsentResponsesForProject(projectID int64) ([]ConsentResponse, error) {
	data := []ConsentResponse{}
	err := config.DBConnection.Select(&data, `SELECT cr.* FROM ConsentResponses cr 
		WHERE cr.projectId = ? ORDER BY cr.dateConsented`, projectID)
	for i := range data {
		data[i].processForAPI()
	}
	return data, err
}

// GetConsentResponseByID gets a single response
func GetConsentResponseByID(responseID int64) (*ConsentResponse, error) {
	data := &ConsentResponse{}
	defer data.processForAPI()
	err := config.DBConnection.Get(data, `SELECT cr.* FROM ConsentResponses cr WHERE cr.id = ?`, responseID)
	return data, err
}

// DeleteConsentesponse deletes a single response
func DeleteConsentesponse(responseID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM ConsentResponses WHERE id = ?`, responseID)
	return err
}

// DeleteConsentesponseForParticipant deletes a response for a specific participant; this could be dangerous
// depending on the study protocol!
func DeleteConsentesponseForParticipant(userID, projectID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM ConsentResponses WHERE projectId = ? AND participantId = ?`, projectID, userID)
	return err
}

// DeleteConsentResponsesForProject deletes all responses for a project
func DeleteConsentResponsesForProject(projectID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM ConsentResponses WHERE projectId = ?`, projectID)
	return err
}

//
// binders and processors
//

func (input *ConsentForm) processForDB() {
}

func (input *ConsentForm) processForAPI() {
}

func (input *ConsentResponse) processForDB() {
	if input.DateConsented == "" {
		input.DateConsented = time.Now().Format(timeFormatDB)
	} else {
		input.DateConsented, _ = parseTimeToTimeFormat(input.DateConsented, timeFormatDB)
	}
	if input.ConsentStatus == "" {
		input.ConsentStatus = ConsentResponseStatusDeclined // don't auto accept
	}
}

func (input *ConsentResponse) processForAPI() {
	input.DateConsented, _ = parseTimeToTimeFormat(input.DateConsented, timeFormatAPI)
	input.ProjectCode = ""
}

// Bind binds the data for the HTTP
func (data *ConsentForm) Bind(r *http.Request) error {
	return nil
}

// Bind binds the data for the HTTP
func (data *ConsentResponse) Bind(r *http.Request) error {
	return nil
}
