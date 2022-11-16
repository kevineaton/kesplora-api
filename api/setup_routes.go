package api

import (
	"net/http"

	"github.com/go-chi/render"
)

type siteConfigurationInput struct {
	Site      Site   `json:"site"`
	Code      string `json:"code"`
	AdminUser User   `json:"user"`
}

// routeGetSiteConfiguration gets whether the site is configured or not
func routeGetSiteConfiguration(w http.ResponseWriter, r *http.Request) {
	// this route is unauthenticated
	site, err := GetSite()
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"configured": err == nil && site.Status == SiteStatusActive,
	})
}

// routeConfigureSite configures the site
func routeConfigureSite(w http.ResponseWriter, r *http.Request) {
	// this route is unauthenticated
	site, err := GetSite()
	exists := false
	if err == nil && site.Status == SiteStatusActive {
		sendAPIJSONData(w, http.StatusOK, map[string]bool{
			"configured": true,
		})
	} else if err == nil {
		exists = true
	}

	input := &siteConfigurationInput{}
	render.Bind(r, input)

	// most things are required, so we have a massive check up front so all required fields are
	// sent if there is an error
	if input.Code == "" ||
		input.Site.Description == "" ||
		input.Site.Name == "" ||
		input.Site.ShortName == "" ||
		input.Site.SiteTechnicalContact == "" ||
		input.AdminUser.FirstName == "" ||
		input.AdminUser.LastName == "" ||
		input.AdminUser.Email == "" ||
		input.AdminUser.Password == "" {
		sendAPIError(w, api_error_config_missing_data, nil, map[string]string{})
		return
	}

	if input.Code != config.SiteCode {
		sendAPIError(w, api_error_config_invalid_code, nil, map[string]string{})
		return
	}

	// set up the site
	createdSite := &Site{
		Name:                 input.Site.Name,
		ShortName:            input.Site.ShortName,
		Description:          input.Site.Description,
		ProjectListOptions:   input.Site.ProjectListOptions,
		Domain:               input.Site.Domain,
		SiteTechnicalContact: input.Site.SiteTechnicalContact,
		Status:               SiteStatusActive,
	}
	if exists {
		createdSite.ID = site.ID
		err = UpdateSite(createdSite)
	} else {
		err = CreateSite(createdSite)
	}

	if err != nil {
		sendAPIError(w, api_error_site_save_error, err, map[string]string{})
		return
	}

	// create the user
	user := &User{
		Title:     input.AdminUser.Title,
		Email:     input.AdminUser.Email,
		FirstName: input.AdminUser.FirstName,
		LastName:  input.AdminUser.LastName,
		Password:  input.AdminUser.Password,
		Status:    UserStatusActive,
	}
	err = CreateUser(user)
	if err != nil {
		sendAPIError(w, api_error_user_cannot_save, err, map[string]string{})
		return
	}

	sendAPIJSONData(w, http.StatusOK, map[string]interface{}{
		"configured": true,
		"site":       createdSite,
	})
}

// Bind binds the data for the HTTP
func (data *siteConfigurationInput) Bind(r *http.Request) error {
	return nil
}
