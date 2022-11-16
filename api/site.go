package api

import (
	"net/http"
	"time"
)

const (
	SiteStatusPending  = "pending"
	SiteStatusActive   = "active"
	SiteStatusDisabled = "disabled"

	SiteProjectListOptionsAll    = "show_all"
	SiteProjectListOptionsActive = "show_active"
	SiteProjectListOptionsNone   = "show_none"
)

// Site is an available location or installation
type Site struct {
	ID                   int64  `json:"id" db:"id"` // in the current set up, there should only be one site
	CreatedOn            string `json:"createdOn" db:"createdOn"`
	ShortName            string `json:"shortName" db:"shortName"`
	Name                 string `json:"name" db:"name"`
	Description          string `json:"description" db:"description"`
	Domain               string `json:"domain" db:"domain"`
	Status               string `json:"status" db:"status"`                         // pending, active, disabled
	ProjectListOptions   string `json:"projectListOptions" db:"projectListOptions"` // show_all, show_active, show_none
	SiteTechnicalContact string `json:"siteTechnicalContact" db:"siteTechnicalContact"`
}

// GetSite gets the site from the DB
func GetSite() (*Site, error) {
	site := &Site{}
	defer site.processForAPI()
	err := config.DBConnection.Get(site, `SELECT * FROM Site LIMIT 1`)
	return site, err
}

// GetSiteByID gets the site by the id
func GetSiteByID(id int64) (*Site, error) {
	site := &Site{}
	defer site.processForAPI()
	err := config.DBConnection.Get(site, `SELECT * FROM Site WHERE id = ? LIMIT 1`, id)
	return site, err
}

// CreateSite creates a site; this should be called relatively infrequently
func CreateSite(input *Site) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO Site SET
	createdOn = :createdOn,
	shortName = :shortName,
	name = :name,
	description = :description,
	domain = :domain,
	status = :status,
	projectListOptions = :projectListOptions,
	siteTechnicalContact = :siteTechnicalContact`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateSite updates a site
func UpdateSite(input *Site) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE Site SET
	createdOn = :createdOn,
	shortName = :shortName,
	name = :name,
	description = :description,
	domain = :domain,
	status = :status,
	projectListOptions = :projectListOptions,
	siteTechnicalContact = :siteTechnicalContact
	WHERE id = :id`, input)
	return err
}

// deleteSiteByID deletes a site and really should only be used for testing purposes; if you
// want to make a site unavailable, it's better to mark it as disabled
func deleteSiteByID(siteID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Site WHERE id = ?`, siteID)
	if err != nil {
		return err
	}
	// as other models are built out, we need to delete them; this is effectively, in this initial version, a DB wipe
	return err
}

// deleteSiteBy deletes a site and really should only be used for testing purposes; if you
// want to make a site unavailable, it's better to mark it as disabled
func deleteSite() error {
	_, err := config.DBConnection.Exec(`DELETE FROM Site`)
	if err != nil {
		return err
	}
	// as other models are built out, we need to delete them; this is effectively, in this initial version, a DB wipe
	return err
}

func (input *Site) processForDB() {
	if input.Status == "" {
		input.Status = SiteStatusPending
	}
	if input.CreatedOn == "" {
		input.CreatedOn = time.Now().Format(timeFormatDB)
	} else {
		input.CreatedOn, _ = parseTimeToTimeFormat(input.CreatedOn, timeFormatDB)
	}
	if input.ProjectListOptions == "" {
		input.ProjectListOptions = SiteProjectListOptionsActive
	}
	if input.Status == "" {
		input.Status = SiteStatusPending
	}
}

func (input *Site) processForAPI() {
	input.CreatedOn, _ = parseTimeToTimeFormat(input.CreatedOn, timeFormatAPI)
}

// Bind binds the data for the HTTP
func (data *Site) Bind(r *http.Request) error {
	return nil
}
