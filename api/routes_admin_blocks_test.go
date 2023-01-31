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

type SuiteTestsBlocksRoutes struct {
	suite.Suite
}

func TestSuiteTestsBlocksRoutes(t *testing.T) {
	suite.Run(t, new(SuiteTestsBlocksRoutes))
}

func (suite *SuiteTestsBlocksRoutes) SetupSuite() {
	setupTesting()
}

func (suite SuiteTestsBlocksRoutes) TestSimpleBlocksAdminRoutesCRUD() {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.Encode(map[string]string{})
	require := suite.Require()

	// first, create an admin
	admin := &User{
		SystemRole: UserSystemRoleAdmin,
	}
	err := createTestUser(admin)
	require.Nil(err)
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

	module1 := &Module{}
	err = createTestModule(module1, project.ID, 0)
	suite.Nil(err)
	require.NotZero(module1.ID)
	defer DeleteModule(module1.ID)

	code, res, err := testEndpoint(http.MethodPost, "/admin/blocks/text", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusBadRequest, code, res)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/mooo", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/text", b, routeAdminCreateBlock, user.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)

	// create one of each and verify each; yes, these are huge
	blockTextContentInput := &BlockText{
		Text: "# Test!\n\n## Another!",
	}
	blockTextInput := &Block{
		Name:      "Text Block",
		Summary:   "This is a simple test block",
		BlockType: BlockTypeText,
		Content:   blockTextContentInput,
	}
	b.Reset()
	encoder.Encode(blockTextInput)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/text", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusCreated, code, res)
	blockText := &Block{}
	m, err := testEndpointResultToMap(res)
	suite.Nil(err)
	err = mapstructure.Decode(m, blockText)
	suite.Nil(err)
	blockTextContent := &BlockText{}
	err = mapstructure.Decode(blockText.Content, blockTextContent)
	suite.Nil(err)
	suite.NotZero(blockText.ID)
	suite.Equal(blockTextInput.Name, blockText.Name)
	suite.Equal(blockTextInput.Summary, blockText.Summary)
	suite.Equal(blockTextInput.BlockType, blockText.BlockType)
	suite.Equal(blockTextContentInput.Text, blockTextContent.Text)
	defer DeleteBlock(blockText.ID)
	defer handleBlockDelete(blockText.BlockType, blockText.ID)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/blocks/%d", blockText.ID), b, routeAdminGetBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	blockTextFound := &Block{}
	blockTextContentFound := &BlockText{}
	err = mapstructure.Decode(m, blockTextFound)
	suite.Nil(err)
	err = mapstructure.Decode(blockTextFound.Content, blockTextContentFound)
	suite.Nil(err)
	suite.Equal(blockText.ID, blockTextFound.ID)
	suite.Equal(blockTextInput.Name, blockTextFound.Name)
	suite.Equal(blockTextInput.Summary, blockTextFound.Summary)
	suite.Equal(blockTextInput.BlockType, blockTextFound.BlockType)
	suite.Equal(blockTextContentInput.Text, blockTextContentFound.Text)

	blockExternalContentInput := &BlockExternal{
		ExternalLink: "https://api.kesplora.com",
	}
	blockExternalInput := &Block{
		Name:      "External Block",
		Summary:   "This is a simple external block",
		BlockType: BlockTypeExternal,
		Content:   blockExternalContentInput,
	}
	b.Reset()
	encoder.Encode(blockExternalInput)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/external", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusCreated, code, res)
	blockExternal := &Block{}
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	err = mapstructure.Decode(m, blockExternal)
	suite.Nil(err)
	blockExternalContent := &BlockExternal{}
	err = mapstructure.Decode(blockExternal.Content, blockExternalContent)
	suite.Nil(err)
	suite.NotZero(blockExternal.ID)
	suite.Equal(blockExternalInput.Name, blockExternal.Name)
	suite.Equal(blockExternalInput.Summary, blockExternal.Summary)
	suite.Equal(blockExternalInput.BlockType, blockExternal.BlockType)
	suite.Equal(blockExternalContentInput.ExternalLink, blockExternalContent.ExternalLink)
	defer DeleteBlock(blockText.ID)
	defer handleBlockDelete(blockText.BlockType, blockText.ID)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/blocks/%d", blockExternal.ID), b, routeAdminGetBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	blockExternalFound := &Block{}
	blockExternalContentFound := &BlockExternal{}
	err = mapstructure.Decode(m, blockExternalFound)
	suite.Nil(err)
	err = mapstructure.Decode(blockExternalFound.Content, blockExternalContentFound)
	suite.Nil(err)
	suite.Equal(blockExternal.ID, blockExternalFound.ID)
	suite.Equal(blockExternalInput.Name, blockExternalFound.Name)
	suite.Equal(blockExternalInput.Summary, blockExternalFound.Summary)
	suite.Equal(blockExternalInput.BlockType, blockExternalFound.BlockType)
	suite.Equal(blockExternalContentInput.ExternalLink, blockExternalContentFound.ExternalLink)

	blockPresentationContentInput := &BlockPresentation{
		EmbedLink:        "https://api.kesplora.com",
		PresentationType: "pdf",
	}
	blockPresentationInput := &Block{
		Name:      "Presentation Block",
		Summary:   "This is a simple external block",
		BlockType: BlockTypePresentation,
		Content:   blockPresentationContentInput,
	}
	b.Reset()
	encoder.Encode(blockPresentationInput)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/presentation", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusCreated, code, res)
	blockPresentation := &Block{}
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	err = mapstructure.Decode(m, blockPresentation)
	suite.Nil(err)
	blockPresentationContent := &BlockPresentation{}
	err = mapstructure.Decode(blockPresentation.Content, blockPresentationContent)
	suite.Nil(err)
	suite.NotZero(blockExternal.ID)
	suite.Equal(blockPresentationInput.Name, blockPresentation.Name)
	suite.Equal(blockPresentationInput.Summary, blockPresentation.Summary)
	suite.Equal(blockPresentationInput.BlockType, blockPresentation.BlockType)
	suite.Equal(blockPresentationContentInput.EmbedLink, blockPresentationContent.EmbedLink)
	suite.Equal(blockPresentationContentInput.PresentationType, blockPresentationContent.PresentationType)
	defer DeleteBlock(blockText.ID)
	defer handleBlockDelete(blockText.BlockType, blockText.ID)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/blocks/%d", blockPresentation.ID), b, routeAdminGetBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	blockPresentationFound := &Block{}
	blockPresentationContentFound := &BlockPresentation{}
	err = mapstructure.Decode(m, blockPresentationFound)
	suite.Nil(err)
	err = mapstructure.Decode(blockPresentationFound.Content, blockPresentationContentFound)
	suite.Nil(err)
	suite.Equal(blockPresentation.ID, blockPresentationFound.ID)
	suite.Equal(blockPresentationInput.Name, blockPresentationFound.Name)
	suite.Equal(blockPresentationInput.Summary, blockPresentationFound.Summary)
	suite.Equal(blockPresentationInput.BlockType, blockPresentationFound.BlockType)
	suite.Equal(blockPresentationContentInput.EmbedLink, blockPresentationContentFound.EmbedLink)
	suite.Equal(blockPresentationContentInput.PresentationType, blockPresentationContentFound.PresentationType)

	// update each and verify each
	updateExternal := &Block{
		Name:    "Updated External",
		Summary: "Updated External Summary",
		Content: BlockExternal{
			ExternalLink: "https://kesplora.com",
		},
	}
	b.Reset()
	encoder.Encode(updateExternal)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/blocks/%d", blockExternal.ID), b, routeAdminUpdateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/blocks/%d", blockExternal.ID), b, routeAdminGetBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	blockExternalFoundUpdated := &Block{}
	blockExternalContentFoundUpdated := &BlockExternal{}
	err = mapstructure.Decode(m, blockExternalFoundUpdated)
	suite.Nil(err)
	err = mapstructure.Decode(blockExternalFoundUpdated.Content, blockExternalContentFoundUpdated)
	suite.Nil(err)
	suite.Equal(blockExternal.ID, blockExternalFoundUpdated.ID)
	suite.Equal(updateExternal.Name, blockExternalFoundUpdated.Name, m)
	suite.Equal(updateExternal.Summary, blockExternalFoundUpdated.Summary)
	suite.Equal(BlockTypeExternal, blockExternalFoundUpdated.BlockType)
	suite.Equal("https://kesplora.com", blockExternalContentFoundUpdated.ExternalLink)

	updateText := &Block{
		Name:    "Updated Text",
		Summary: "Updated Text Summary",
		Content: BlockText{
			Text: "# Updated",
		},
	}
	b.Reset()
	encoder.Encode(updateText)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/blocks/%d", blockText.ID), b, routeAdminUpdateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/blocks/%d", blockText.ID), b, routeAdminGetBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	blockTextFoundUpdated := &Block{}
	blockTextContentFoundUpdated := &BlockText{}
	err = mapstructure.Decode(m, blockTextFoundUpdated)
	suite.Nil(err)
	err = mapstructure.Decode(blockTextFoundUpdated.Content, blockTextContentFoundUpdated)
	suite.Nil(err)
	suite.Equal(blockText.ID, blockTextFoundUpdated.ID)
	suite.Equal(updateText.Name, blockTextFoundUpdated.Name, m)
	suite.Equal(updateText.Summary, blockTextFoundUpdated.Summary)
	suite.Equal(BlockTypeText, blockTextFoundUpdated.BlockType)
	suite.Equal("# Updated", blockTextContentFoundUpdated.Text)

	updatePresentation := &Block{
		Name:    "Updated Presentation",
		Summary: "Updated Presentation Summary",
		Content: BlockPresentation{
			EmbedLink:        "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			PresentationType: BlockPresentationTypeYoutube,
		},
	}
	b.Reset()
	encoder.Encode(updatePresentation)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/blocks/%d", blockPresentation.ID), b, routeAdminUpdateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/blocks/%d", blockPresentation.ID), b, routeAdminGetBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	blockPresentationFoundUpdated := &Block{}
	blockPresentationContentFoundUpdated := &BlockPresentation{}
	err = mapstructure.Decode(m, blockPresentationFoundUpdated)
	suite.Nil(err)
	err = mapstructure.Decode(blockPresentationFoundUpdated.Content, blockPresentationContentFoundUpdated)
	suite.Nil(err)
	suite.Equal(blockPresentation.ID, blockPresentationFoundUpdated.ID)
	suite.Equal(updatePresentation.Name, blockPresentationFoundUpdated.Name, m)
	suite.Equal(updatePresentation.Summary, blockPresentationFoundUpdated.Summary)
	suite.Equal(BlockTypePresentation, blockPresentationFoundUpdated.BlockType)
	suite.Equal("https://www.youtube.com/watch?v=dQw4w9WgXcQ", blockPresentationContentFoundUpdated.EmbedLink)
	suite.Equal(BlockPresentationTypeYoutube, blockPresentationContentFoundUpdated.PresentationType)

	// get for the site
	code, res, err = testEndpoint(http.MethodGet, "/admin/blocks", b, routeAdminGetBlocksOnSite, user.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodGet, "/admin/blocks", b, routeAdminGetBlocksOnSite, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	s, err := testEndpointResultToSlice(res)
	suite.Nil(err)
	require.NotZero(len(s))
	foundText := false
	foundPresentation := false
	foundExternal := false
	for i := range s {
		t := &Block{}
		err = mapstructure.Decode(s[i], t)
		suite.Nil(err)
		if t.ID == blockExternal.ID {
			foundExternal = true
		}
		if t.ID == blockPresentation.ID {
			foundPresentation = true
		}
		if t.ID == blockText.ID {
			foundText = true
		}
		// since this may be run against a live instance, and there may be a lot, once we have them
		// all we can break the loop
		if foundExternal && foundPresentation && foundText {
			break
		}
	}
	suite.True(foundText)
	suite.True(foundPresentation)
	suite.True(foundExternal)

	// link text and external to module
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/1", module1.ID, blockExternal.ID), b, routeAdminLinkBlockAndModule, user.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/1", module1.ID, blockExternal.ID), b, routeAdminLinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/2", module1.ID, blockText.ID), b, routeAdminLinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// get for module, make sure presentation isn't there
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/modules/%d/blocks", module1.ID), b, routeAdminGetBlocksForModule, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	s, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	require.Equal(2, len(s))
	foundText = false
	foundPresentation = false
	foundExternal = false
	for i := range s {
		t := &Block{}
		err = mapstructure.Decode(s[i], t)
		suite.Nil(err)
		if t.ID == blockExternal.ID {
			foundExternal = true
		}
		if t.ID == blockPresentation.ID {
			foundPresentation = true
		}
		if t.ID == blockText.ID {
			foundText = true
		}
	}
	suite.True(foundText)
	suite.False(foundPresentation)
	suite.True(foundExternal)

	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/projects/%d/flow", project.ID), b, routeAdminGetModulesOnProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// remove external, make sure it's only text now
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/modules/%d/blocks/%d", module1.ID, blockExternal.ID), b, routeAdminUnlinkBlockAndModule, user.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/modules/%d/blocks/%d", module1.ID, blockExternal.ID), b, routeAdminUnlinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/modules/%d/blocks", module1.ID), b, routeAdminGetBlocksForModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	s, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	require.Equal(1, len(s))
	foundText = false
	foundPresentation = false
	foundExternal = false
	for i := range s {
		t := &Block{}
		err = mapstructure.Decode(s[i], t)
		suite.Nil(err)
		if t.ID == blockExternal.ID {
			foundExternal = true
		}
		if t.ID == blockPresentation.ID {
			foundPresentation = true
		}
		if t.ID == blockText.ID {
			foundText = true
		}
	}
	suite.True(foundText)
	suite.False(foundPresentation)
	suite.False(foundExternal)

	// delete each and verify
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/blocks/%d", blockExternal.ID), b, routeAdminDeleteBlock, user.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/blocks/%d", blockExternal.ID), b, routeAdminDeleteBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/blocks/%d", blockExternal.ID), b, routeAdminGetBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusNotFound, code, res)

	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/blocks/%d", blockPresentation.ID), b, routeAdminDeleteBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/blocks/%d", blockPresentation.ID), b, routeAdminGetBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusNotFound, code, res)

	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/blocks/%d", blockText.ID), b, routeAdminDeleteBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/blocks/%d", blockText.ID), b, routeAdminGetBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusNotFound, code, res)

	code, res, err = testEndpoint(http.MethodGet, "/admin/blocks", b, routeAdminGetBlocksOnSite, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	s, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	foundText = false
	foundPresentation = false
	foundExternal = false
	for i := range s {
		t := &Block{}
		err = mapstructure.Decode(s[i], t)
		suite.Nil(err)
		if t.ID == blockExternal.ID {
			foundExternal = true
		}
		if t.ID == blockPresentation.ID {
			foundPresentation = true
		}
		if t.ID == blockText.ID {
			foundText = true
		}
		// we have to loop through all, so no early breaks this time
	}
	suite.False(foundText)
	suite.False(foundPresentation)
	suite.False(foundExternal)
}
