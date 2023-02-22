package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/suite"
)

// this file will attempt to create a project for the site with one of
// each block in two modules, then register a few participants and go through the
// full process; this will be a complete start-to-finish test, although it will
// not hit every possible endpoint; the goal is to provide a "basic set" test
// to ensure that the most common use case is covered

type SuiteTestsFullSetup struct {
	suite.Suite
}

func TestSuiteTestsFullSetup(t *testing.T) {
	suite.Run(t, new(SuiteTestsFullSetup))
}

func (suite *SuiteTestsFullSetup) SetupSuite() {
	setupTesting()
}

// TODO: build this out
func (suite SuiteTestsFullSetup) TestBasicFlow() {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.Encode(map[string]string{})
	require := suite.Require()

	// we assume the site is configured already
	site, err := GetSite()
	require.Nil(err)
	require.NotZero(site.ID)

	// create an admin
	admin := &User{
		SystemRole: UserSystemRoleAdmin,
	}
	err = createTestUser(admin)
	require.Nil(err)
	defer DeleteUser(admin.ID)

	// set up new project
	input := &Project{
		Name:        "Test Project",
		Description: "# Project\n\n- Cool\n- Awesome\n",
	}
	b.Reset()
	encoder.Encode(input)
	code, res, err := testEndpoint(http.MethodPost, "/admin/projects", b, routeAdminCreateProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err := testEndpointResultToMap(res)
	suite.Nil(err)
	createdProject := &Project{}
	err = mapstructure.Decode(m, createdProject)
	suite.Nil(err)
	suite.NotZero(createdProject.ID)
	suite.Equal(input.Name, createdProject.Name)
	suite.Equal(input.Description, createdProject.Description)
	suite.Equal("test_project", createdProject.ShortCode)
	suite.Equal(input.Description, createdProject.ShortDescription)
	suite.Equal(ProjectStatusPending, createdProject.Status)
	defer DeleteProject(createdProject.ID)

	// create 3 modules with appropriate blocks

	// module 1
	module1Input := &Module{
		Name: "Module 1",
	}
	b.Reset()
	encoder.Encode(module1Input)
	code, res, err = testEndpoint(http.MethodPost, "/admin/modules", b, routeAdminCreateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule1 := &Module{}
	err = mapstructure.Decode(m, createdModule1)
	suite.Nil(err)
	suite.NotZero(createdModule1.ID)
	defer DeleteModule(createdModule1.ID)

	// module 1 text block

	// module 1 presentation block

	// module 2

	// module 2 external block

	// module 2 text block

	// module 3

	// module 3 form block

	// module 3 text block 1

	// module 3 text block 2

	// save the flow

	// save the consent

	// participant sign up

	// participant completes each block in turn

	// project is complete

	// clean up

}
