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

type SuiteTestsModulesRoutes struct {
	suite.Suite
}

func TestSuiteTestsModuleRoutes(t *testing.T) {
	suite.Run(t, new(SuiteTestsModulesRoutes))
}

func (suite *SuiteTestsModulesRoutes) SetupSuite() {
	setupTesting()
}

func (suite SuiteTestsModulesRoutes) TestModuleAdminRoutesCRUD() {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.Encode(map[string]string{})
	require := suite.Require()

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

	project := &Project{}
	err = createTestProject(project)
	suite.Nil(err)
	defer DeleteProject(project.ID)

	code, res, err := testEndpoint(http.MethodPost, "/admin/modules", b, routeAdminCreateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)
	code, res, err = testEndpoint(http.MethodPost, "/admin/modules", b, routeAdminCreateModule, user.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)

	input := &Module{
		Name:        "Module 1",
		Description: "# Module 1 - Introduction\n\n- Cool\n- Awesome\n",
		Status:      ModuleStatusActive,
	}
	b.Reset()
	encoder.Encode(input)
	code, res, err = testEndpoint(http.MethodPost, "/admin/modules", b, routeAdminCreateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err := testEndpointResultToMap(res)
	suite.Nil(err)
	module1 := &Module{}
	err = mapstructure.Decode(m, module1)
	suite.Nil(err)
	suite.NotZero(module1.ID)
	defer DeleteModule(module1.ID)

	input = &Module{
		Name:        "Module 2",
		Description: "# Module 2 - Exposure\n\n- Cool\n- Awesome\n",
		Status:      ModuleStatusActive,
	}
	b.Reset()
	encoder.Encode(input)
	code, res, err = testEndpoint(http.MethodPost, "/admin/modules", b, routeAdminCreateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	module2 := &Module{}
	err = mapstructure.Decode(m, module2)
	suite.Nil(err)
	suite.NotZero(module2.ID)
	defer DeleteModule(module2.ID)

	// get all on platform
	code, res, err = testEndpoint(http.MethodGet, "/admin/modules", b, routeAdminGetAllSiteModules, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	mS, err := testEndpointResultToSlice(res)
	suite.Nil(err)
	require.NotZero(len(mS))
	all := []Module{}
	err = mapstructure.Decode(mS, &all)
	suite.Nil(err)
	found1 := false
	found2 := false
	for i := range all {
		if all[i].ID == module1.ID {
			found1 = true
			continue
		}
		if all[i].ID == module2.ID {
			found2 = true
			continue
		}
	}
	suite.True(found1)
	suite.True(found2)

	// put them on the project in the flow
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/projects/%d/modules/%d/order/1", project.ID, module1.ID), b, routeAdminLinkModuleAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/projects/%d/modules/%d/order/2", project.ID, module2.ID), b, routeAdminLinkModuleAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// get them, verify the order
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/projects/%d/flow", project.ID), b, routeAdminGetModulesOnProject, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	mS, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	suite.NotZero(len(mS))
	all = []Module{}
	err = mapstructure.Decode(mS, &all)
	suite.Nil(err)
	found1 = false
	found2 = false
	for i := range all {
		if all[i].ID == module1.ID {
			found1 = true
			suite.False(found2) // since module 1 is first, module 2 should not have been found
			continue
		}
		if all[i].ID == module2.ID {
			found2 = true
			continue
		}
	}
	suite.True(found1)
	suite.True(found2)

	// get each individually
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/modules/%d", module1.ID), b, routeAdminGetModuleByID, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	get1 := &Module{}
	err = mapstructure.Decode(m, get1)
	suite.Nil(err)
	suite.Equal(module1.ID, get1.ID)
	suite.Equal(module1.Name, get1.Name)
	suite.Equal(module1.Status, get1.Status)
	suite.Equal(module1.Description, get1.Description)

	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/modules/%d", module2.ID), b, routeAdminGetModuleByID, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	get2 := &Module{}
	err = mapstructure.Decode(m, get2)
	suite.Nil(err)
	suite.Equal(module2.ID, get2.ID)
	suite.Equal(module2.Name, get2.Name)
	suite.Equal(module2.Status, get2.Status)
	suite.Equal(module2.Description, get2.Description)

	// update
	update := &Module{
		Name:        "Updated!",
		Description: "Not Ready",
		Status:      ModuleStatusDisabled,
	}
	b.Reset()
	encoder.Encode(update)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/modules/%d", module1.ID), b, routeAdminUpdateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(code, http.StatusOK, res)

	// get
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/modules/%d", module1.ID), b, routeAdminGetModuleByID, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	updated1 := &Module{}
	err = mapstructure.Decode(m, updated1)
	suite.Nil(err)
	suite.Equal(module1.ID, updated1.ID)
	suite.Equal(update.Name, updated1.Name)
	suite.Equal(update.Status, updated1.Status)
	suite.Equal(update.Description, updated1.Description)

	// change the order
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/projects/%d/modules/%d/order/9", project.ID, module1.ID), b, routeAdminLinkModuleAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// get to check the order
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/projects/%d/flow", project.ID), b, routeAdminGetModulesOnProject, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	mS, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	suite.NotZero(len(mS))
	all = []Module{}
	err = mapstructure.Decode(mS, &all)
	suite.Nil(err)
	found1 = false
	found2 = false
	for i := range all {
		if all[i].ID == module1.ID {
			found1 = true
			suite.True(found2) // since module 2 is first, module 1 should not have been found
			continue
		}
		if all[i].ID == module2.ID {
			found2 = true
			suite.False(found1) // since module 2 is first, module 1 should not have been found
			continue
		}
	}
	suite.True(found1)
	suite.True(found2)

	// delete module 1
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/modules/%d", module1.ID), b, routeAdminDeleteModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// make sure it's gone on a single GET
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/modules/%d", module1.ID), b, routeAdminGetModuleByID, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusNotFound, code, res)

	// make sure it's not in the list
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/projects/%d/flow", project.ID), b, routeAdminGetModulesOnProject, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	mS, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	suite.NotZero(len(mS))
	all = []Module{}
	err = mapstructure.Decode(mS, &all)
	suite.Nil(err)
	found1 = false
	found2 = false
	for i := range all {
		if all[i].ID == module1.ID {
			found1 = true
			continue
		}
		if all[i].ID == module2.ID {
			found2 = true
			continue
		}
	}
	suite.False(found1)
	suite.True(found2)

	// remove module 2 and make sure it exists but is gone from the flow
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/projects/%d/modules/%d", project.ID, module2.ID), b, routeAdminUnlinkModuleAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/modules/%d", module2.ID), b, routeAdminGetModuleByID, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// make sure it's not in the list
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/projects/%d/flow", project.ID), b, routeAdminGetModulesOnProject, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	mS, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	suite.Zero(len(mS))
	all = []Module{}
	err = mapstructure.Decode(mS, &all)
	suite.Nil(err)
	found1 = false
	found2 = false
	for i := range all {
		if all[i].ID == module1.ID {
			found1 = true
			continue
		}
		if all[i].ID == module2.ID {
			found2 = true
			continue
		}
	}
	suite.False(found1)
	suite.False(found2)

}
