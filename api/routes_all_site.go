package api

import "net/http"

// routeAllGetSite gets the site
func routeAllGetSite(w http.ResponseWriter, r *http.Request) {
	// this route is unauthenticated
	site, err := GetSite()
	if err != nil {
		sendAPIError(w, api_error_site_get_error, err, map[string]string{})
		return
	}
	if site.Status != SiteStatusActive {
		sendAPIError(w, api_error_site_not_active, nil, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, site)
}
