package api

import (
	"fmt"
	"math/rand"
	"net/http"
)

const (
	ProjectStatusPending   = "pending"
	ProjectStatusActive    = "active"
	ProjectStatusDisabled  = "disabled"
	ProjectStatusCompleted = "completed"

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
// Projects can be viewed as meta-info about the activities conducted in the project. The shortCode is used with the `SignupStatus` field and should NOT
// be returned in GETs for non-admins
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
	ParticipantMinimumAge           int64  `json:"participantMinimumAge" db:"participantMinimumAge"`
	ConnectParticipantToConsentForm string `json:"connectParticipantToConsentForm" db:"connectParticipantToConsentForm"`
	ParticipantCount                int64  `json:"participantCoun" db:"participantCount"`
}

// ProjectAPIReturnNonAdmin is a much-reduced project return struct for non-admins
type ProjectAPIReturnNonAdmin struct {
	ID               int64  `json:"id" db:"id"`
	Name             string `json:"name" db:"name"`
	ShortDescription string `json:"shortDescription" db:"shortDescription"`
	Description      string `json:"description" db:"description"`
	Status           string `json:"status" db:"status"`
	SignupStatus     string `json:"signupStatus" db:"signupStatus"`
}

// ProjectUserLinkRequest holds extra request options for joining a project, such as if a code is needed
type ProjectUserLinkRequest struct {
	Code string `json:"code"`
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
	err := config.DBConnection.Get(project, `SELECT p.*, (SELECT COUNT(*) FROM ProjectUserLinks l WHERE l.projectId = p.id) AS participantCount
	FROM Projects p WHERE p.id = ?`, projectID)
	return project, err
}

// GetProjectsForSite gets all of the projects for a site, optionally filtered by status
func GetProjectsForSite(siteID int64, status string) ([]Project, error) {
	projects := []Project{}
	var err error
	if status == "" || status == "all" {
		err = config.DBConnection.Select(&projects, `SELECT p.*, (SELECT COUNT(*) FROM ProjectUserLinks l WHERE l.projectId = p.id) AS participantCount
		FROM Projects p WHERE p.siteId = ? ORDER BY p.name`, siteID)
	} else {
		err = config.DBConnection.Select(&projects, `SELECT p.*, (SELECT COUNT(*) FROM ProjectUserLinks l WHERE l.projectId = p.id) AS participantCount
		FROM Projects p WHERE p.siteId = ? AND p.status = ? ORDER BY p.name`, siteID, status)
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

// LinkUserAndProject links a user to a project
func LinkUserAndProject(userID, projectID int64) error {
	_, err := config.DBConnection.Exec("INSERT INTO ProjectUserLinks (userId, projectId) VALUES (?,?) ON DUPLICATE KEY UPDATE userId = userId", userID, projectID)
	return err
}

// UnlinkUserAndProject unlinks a user and a project
func UnlinkUserAndProject(userID, projectID int64) error {
	_, err := config.DBConnection.Exec("DELETE FROM ProjectUserLinks WHERE userId = ? AND projectId = ?", userID, projectID)
	return err
}

// createTestProject is used for tests to create a test project
func createTestProject(defaults *Project) error {
	if defaults.Name == "" {
		defaults.Name = fmt.Sprintf("Test Project %d", rand.Intn(9999999))
	}
	if defaults.Description == "" {
		defaults.Description = "# Test Site\n\n *Do Not Use!*\n"
	}
	if defaults.ShortDescription == "" {
		defaults.ShortDescription = "# Test Site\n\n *Do Not Use!*\n"
	}
	if defaults.Status == "" {
		defaults.Status = ProjectStatusActive
	}
	return CreateProject(defaults)
}

func convertProjectToUserRet(input *Project) *ProjectAPIReturnNonAdmin {
	ret := &ProjectAPIReturnNonAdmin{
		ID:               input.ID,
		Name:             input.Name,
		ShortDescription: input.ShortDescription,
		Description:      input.Description,
		Status:           input.Status,
		SignupStatus:     input.SignupStatus,
	}
	// if signup is allowed BUT max participants is reached, signup is blocked
	if input.MaxParticipants > 0 && input.ParticipantCount >= input.MaxParticipants {
		ret.SignupStatus = ProjectSignupStatusClosed
	}
	return ret
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

// Bind binds the data for the HTTP
func (data *ProjectUserLinkRequest) Bind(r *http.Request) error {
	return nil
}
