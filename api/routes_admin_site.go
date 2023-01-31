package api

import (
	"net/http"

	"github.com/go-chi/render"
)

// routeAdminUpdateSite updates the site
func routeAdminUpdateSite(w http.ResponseWriter, r *http.Request) {
	// validity checked in middleware of router
	site, err := GetSite()
	if err != nil {
		sendAPIError(w, api_error_site_get_error, err, map[string]string{})
		return
	}

	input := &Site{}
	render.Bind(r, input)

	if input.Description != "" {
		site.Description = input.Description
	}
	if input.Domain != "" {
		site.Domain = input.Domain
	}
	if input.Name != "" {
		site.Name = input.Name
	}
	if input.ProjectListOptions != "" {
		site.ProjectListOptions = input.ProjectListOptions
	}
	if input.ShortName != "" {
		site.ShortName = input.ShortName
	}
	if input.SiteTechnicalContact != "" {
		site.SiteTechnicalContact = input.SiteTechnicalContact
	}
	if input.Status != "" {
		site.Status = input.Status
	}
	err = UpdateSite(site)
	if err != nil {
		sendAPIError(w, api_error_site_save_error, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, site)
}
