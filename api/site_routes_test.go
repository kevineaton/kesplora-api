package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupAndConfigRoutes(t *testing.T) {
	SetupConfig()
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.Encode(map[string]string{})

	// we are going to see if there is a site already; if so, we don't
	// want to accidentally delete the user's local install, so we will skip
	// this instead of breaking stuff
	exists, err := GetSite()
	if err == nil && exists != nil && exists.ID > 0 {
		t.Skip("Site already set up or exists, so skipping config tests")
	}

	code, _, err := testEndpoint(http.MethodGet, "/setup", b, routeGetSiteConfiguration, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code)

	// send in bad data
	input := &siteConfigurationInput{}
	encoder.Encode(input)
	code, _, err = testEndpoint(http.MethodPost, "/setup", b, routeConfigureSite, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	input.Code = ""
	input.Site.Description = "# Site Description"
	input.Site.Domain = "site.kesplora.com"
	input.Site.Name = "Test Site"
	input.Site.ProjectListOptions = SiteProjectListOptionsActive
	input.Site.ShortName = "Test"
	input.Site.SiteTechnicalContact = "testing@kesplora.com"
	input.AdminUser.Email = "testing@kesplora.com"
	input.AdminUser.FirstName = "Admin"
	input.AdminUser.LastName = "User"
	input.AdminUser.Title = "Dr."
	input.AdminUser.Password = "this IS @ simple P@ssw0rd!!"

	encoder.Encode(input)
	code, _, err = testEndpoint(http.MethodPost, "/setup", b, routeConfigureSite, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusBadRequest, code)

	// make the code right
	b.Reset()
	if config.SiteCode == "" {
		// it was already configured, so let's pretend
		config.SiteCode = "test"
	}
	input.Code = config.SiteCode
	encoder.Encode(input)
	code, res, err := testEndpoint(http.MethodPost, "/setup", b, routeConfigureSite, "")
	assert.Nil(t, err)
	require.Equal(t, http.StatusOK, code, res)

	site, err := GetSite()
	assert.Nil(t, err)
	assert.Equal(t, input.Site.Description, site.Description)
	assert.Equal(t, input.Site.Domain, site.Domain)
	assert.Equal(t, input.Site.Name, site.Name)
	assert.Equal(t, input.Site.ProjectListOptions, site.ProjectListOptions)
	assert.Equal(t, input.Site.ShortName, site.ShortName)
	assert.Equal(t, input.Site.SiteTechnicalContact, site.SiteTechnicalContact)

	// login to get the token
	b.Reset()
	encoder.Encode(map[string]string{
		"email":    input.AdminUser.Email,
		"password": input.AdminUser.Password,
	})
	code, res, err = testEndpoint(http.MethodPost, "/login", b, routeUserLogin, "")
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, code, res)
	m, err := testEndpointResultToMap(res)
	assert.Nil(t, err)
	user := &User{}
	err = mapstructure.Decode(m, user)
	assert.Nil(t, err)
	assert.Equal(t, input.AdminUser.Title, user.Title)
	assert.Equal(t, input.AdminUser.FirstName, user.FirstName)
	assert.Equal(t, input.AdminUser.LastName, user.LastName)
	assert.Equal(t, input.AdminUser.Email, user.Email)
	assert.NotEqual(t, "", user.Access)
	access := user.Access

	defer DeleteUser(user.ID)
	code, res, err = testEndpoint(http.MethodGet, "/site", b, routeConfigureSite, access)
	assert.Equal(t, http.StatusOK, code, res)
	assert.Nil(t, err)

}
