package api

import "net/http"

const (
	ProjectStatusPending  = "pending"
	ProjectStatusActive   = "active"
	ProjectStatusDisabled = "disabled"

	ProjectShowStatusSite   = "site"
	ProjectShowStatusDirect = "direct"
	ProjectShowStatusNo     = "no"

	ProjectSignupStatusOpen     = "open"
	ProjectSignupStatusWithCode = "with_code"
	ProjectSignupStatusClosed   = "closed"

	ProjectParticipantVisibilityCode  = "code"
	ProjectParticipantVisibilityEmail = "email"
	ProjectParticipantVisibilityFull  = "full"
)

// Project is a major research project, which will have flows associated with it. Participants work through the project flows.
// Projects can be viewed as meta-info about the activities conducted in the project.
type Project struct {
	ID                              int64  `json:"id" db:"id"`
	SiteID                          int64  `json:"siteId" db:"siteId"`
	Name                            string `json:"name" db:"name"`
	ShortCode                       string `json:"shortCode" db:"shortCode"`
	ShortDescription                string `json:"shortDescription" db:"shortDescription"`
	Description                     string `json:"description" db:"description"`
	Status                          string `json:"status" db:"status"`
	ShowStatus                      string `json:"showStatus" db:"showStatus"`
	SignupStatus                    string `json:"signupStatus" db:"signupStatus"`
	MaxParticipants                 int64  `json:"maxParticipants" db:"maxParticipants"`
	ParticipantVisibility           string `json:"participantVisibility" db:"participantVisibility"`
	ParticipantMinimumAge           string `json:"participantMinimumAge" db:"participantMinimumAge"`
	ConnectParticipantToConsentForm string `json:"connectParticipantToConsentForm" db:"connectParticipantToConsentForm"`
}

// CreateProject creates a new project
func CreateProject(input *Project) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO Projects SET
		siteId = :siteId,
		name = :name,
		shortCode = :shortCode,
		shortDescription = :shortDescription,
		description = :description,
		status = :status,
		showStatus = :showStatus,
		signupStatus = :signupStatus,
		maxParticipants = :maxParticipants,
		participantMinimumAge = :participantMinimumAge,
		connectParticipantToConsentForm = :connectParticipantToConsentForm`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateProject updates an existing project
func UpdateProject(input *Project) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE Projects SET
		siteId = :siteId,
		name = :name,
		shortCode = :shortCode,
		shortDescription = :shortDescription,
		description = :description,
		status = :status,
		showStatus = :showStatus,
		signupStatus = :signupStatus,
		maxParticipants = :maxParticipants,
		participantMinimumAge = :participantMinimumAge,
		connectParticipantToConsentForm = :connectParticipantToConsentForm
		WHERE id = :id`, input)
	return err
}

// GetProjectByID gets a single project by its id
func GetProjectByID(projectID int64) (*Project, error) {
	project := &Project{}
	defer project.processForAPI()
	err := config.DBConnection.Get(project, "SELECT p.* FROM Projects p WHERE p.id = ?", projectID)
	return project, err
}

// GetProjectsForSite gets all of the projects for a site, optionally filtered by status
func GetProjectsForSite(siteID int64, status string) ([]Project, error) {
	projects := []Project{}
	var err error
	if status == "" || status == "all" {
		err = config.DBConnection.Select(&projects, `SELECT p.* FROM Projects p WHERE siteId = ? ORDER BY name`, siteID)
	} else {
		err = config.DBConnection.Select(&projects, `SELECT p.* FROM Projects p WHERE siteId = ? AND status = ? ORDER BY name`, siteID, status)
	}
	if err != nil {
		return projects, err
	}
	for i := range projects {
		projects[i].processForAPI()
	}
	return projects, err
}

// DeleteProject deletes a project. Note that this probably
func DeleteProject(projectID int64) error {
	_, err := config.DBConnection.Exec("DELETE FROM Projects WHERE id = ?", projectID)
	if err != nil {
		return err
	}

	// TODO: as more entities are built out, add the delete calls here
	return nil
}

func (input *Project) processForDB() {
	if input.Status == "" {
		input.Status = ProjectStatusPending
	}
	if input.ShowStatus == "" {
		input.ShowStatus = ProjectShowStatusSite
	}
	if input.SignupStatus == "" {
		input.SignupStatus = ProjectSignupStatusOpen
	}
	if input.ParticipantVisibility == "" {
		input.ParticipantVisibility = ProjectParticipantVisibilityCode
	}
	if input.ConnectParticipantToConsentForm == "" {
		input.ConnectParticipantToConsentForm = Yes
	}
}

func (input *Project) processForAPI() {
}

// Bind binds the data for the HTTP
func (data *Project) Bind(r *http.Request) error {
	return nil
}
