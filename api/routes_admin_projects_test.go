package api

// we need to put these in a suite to setup a site

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/suite"
)

type SuiteTestsProjectRoutes struct {
	suite.Suite
}

func TestSuiteTestsProjectRoutes(t *testing.T) {
	suite.Run(t, new(SuiteTestsProjectRoutes))
}

func (suite *SuiteTestsProjectRoutes) SetupSuite() {
	setupTesting()
}

func (suite SuiteTestsProjectRoutes) TestProjectRoutesCRUD() {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.Encode(map[string]string{})

	// first, create an admin
	admin := &User{
		SystemRole: UserSystemRoleAdmin,
	}
	err := createTestUser(admin)
	suite.Nil(err)
	defer DeleteUser(admin.ID)
	// now a user
	user := &User{
		SystemRole: UserSystemRoleUser,
	}
	err = createTestUser(user)
	suite.Nil(err)
	defer DeleteUser(user.ID)

	code, res, err := testEndpoint(http.MethodPost, "/admin/projects", b, routeAdminCreateProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)

	input := &Project{
		Name:        "Test Project",
		Description: "# Project\n\n- Cool\n- Awesome\n",
	}
	b.Reset()
	encoder.Encode(input)
	code, res, err = testEndpoint(http.MethodPost, "/admin/projects", b, routeAdminCreateProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err := testEndpointResultToMap(res)
	suite.Nil(err)
	created := &Project{}
	err = mapstructure.Decode(m, created)
	suite.Nil(err)
	suite.NotZero(created.ID)
	suite.Equal(input.Name, created.Name)
	suite.Equal(input.Description, created.Description)
	suite.Equal("test_project", created.ShortCode)
	suite.Equal(input.Description, created.ShortDescription)
	suite.Equal(ProjectStatusPending, created.Status)
	defer DeleteProject(created.ID)

	// get all projects for the site
	code, res, err = testEndpoint(http.MethodGet, "/admin/projects", b, routeAdminGetProjects, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	s, err := testEndpointResultToSlice(res)
	suite.Nil(err)
	suite.NotZero(len(s))

	// get the individual
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/projects/%d", created.ID), b, routeAdminGetProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	found := &Project{}
	err = mapstructure.Decode(m, found)
	suite.Nil(err)
	suite.NotZero(found.ID)
	suite.Equal(input.Name, found.Name)
	suite.Equal(input.Description, found.Description)
	suite.Equal("test_project", found.ShortCode)
	suite.Equal(input.Description, found.ShortDescription)
	suite.Equal(ProjectStatusPending, found.Status)

	// update
	updateInput := &Project{
		Name:                  "Updated",
		Description:           "# Nevermind",
		ShortDescription:      "# Test",
		ShortCode:             "test12345",
		Status:                ProjectStatusActive,
		ShowStatus:            ProjectShowStatusSite,
		MaxParticipants:       10,
		ParticipantMinimumAge: 18,
	}
	b.Reset()
	encoder.Encode(updateInput)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/projects/%d", created.ID), b, routeAdminUpdateProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// get it again
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/projects/%d", created.ID), b, routeAdminGetProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	found = &Project{}
	err = mapstructure.Decode(m, found)
	suite.Nil(err)
	suite.NotZero(found.ID)
	suite.Equal(updateInput.Name, found.Name)
	suite.Equal(updateInput.Description, found.Description)
	suite.Equal(updateInput.ShortCode, found.ShortCode)
	suite.Equal(updateInput.ShortDescription, found.ShortDescription)
	suite.Equal(ProjectStatusActive, found.Status)
	suite.Equal(updateInput.ShowStatus, found.ShowStatus)
	suite.Equal(updateInput.MaxParticipants, found.MaxParticipants)
	suite.Equal(updateInput.ParticipantMinimumAge, found.ParticipantMinimumAge)

	// try as a normal user

	// update and create should be blocked
	code, res, err = testEndpoint(http.MethodPost, "/admin/projects", b, routeAdminCreateProject, user.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/projects/%d", created.ID), b, routeAdminUpdateProject, user.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)

	code, res, err = testEndpoint(http.MethodGet, "/projects", b, routeParticipantGetProjects, user.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	s, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	suite.NotZero(len(s))

	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d", created.ID), b, routeAllGetProject, user.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	found = &Project{}
	err = mapstructure.Decode(m, found)
	suite.Nil(err)
	suite.NotZero(found.ID)
	suite.Equal(updateInput.Name, found.Name)
	suite.Equal(updateInput.Description, found.Description)
	suite.Equal("", found.ShortCode)
	suite.Equal(updateInput.ShortDescription, found.ShortDescription)
	suite.Equal(ProjectStatusActive, found.Status)
	suite.Equal("", found.ShowStatus)

}
