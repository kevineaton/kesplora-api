package api

import (
	"bytes"
	"encoding/json"
	"fmt"
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
//
// yes this is huge
//

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

	participantPassword := randomString(64)

	// we are going to create a "found" holder here to help us out
	foundParticipantBlockInModule := map[int64]map[int64]bool{}

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
	projectInput := &Project{
		Name:            "Test Project",
		Description:     "# Project\n\n- Cool\n- Awesome\n",
		FlowRule:        ProjectFlowRuleFree,
		CompleteMessage: "",
		SignupStatus:    ProjectSignupStatusWithCode,
		ShortCode:       "testing",
	}
	b.Reset()
	encoder.Encode(projectInput)
	code, res, err := testEndpoint(http.MethodPost, "/admin/projects", b, routeAdminCreateProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err := testEndpointResultToMap(res)
	suite.Nil(err)
	createdProject := &Project{}
	err = mapstructure.Decode(m, createdProject)
	suite.Nil(err)
	suite.NotZero(createdProject.ID)
	suite.Equal(projectInput.Name, createdProject.Name)
	suite.Equal(projectInput.Description, createdProject.Description)
	suite.Equal(projectInput.ShortCode, createdProject.ShortCode)
	suite.Equal(projectInput.Description, createdProject.ShortDescription)
	suite.Equal(ProjectStatusPending, createdProject.Status)
	defer DeleteProject(createdProject.ID)

	projectInput.CompleteRule = ProjectCompleteRuleBlocked
	projectInput.CompleteMessage = "You completed the project"
	projectInput.Status = "active"
	b.Reset()
	encoder.Encode(projectInput)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/projects/%d", createdProject.ID), b, routeAdminUpdateProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	updatedProject := &Project{}
	err = mapstructure.Decode(m, updatedProject)
	suite.Nil(err)
	suite.Equal(projectInput.CompleteRule, updatedProject.CompleteRule)
	suite.Equal(projectInput.CompleteMessage, updatedProject.CompleteMessage)
	suite.Equal(projectInput.Status, updatedProject.Status)

	// for helper sake, just fill it in
	createdProject.CompleteRule = updatedProject.CompleteRule
	createdProject.CompleteMessage = updatedProject.CompleteMessage
	createdProject.Status = updatedProject.Status

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
	suite.Equal(module1Input.Name, createdModule1.Name)
	suite.Equal(module1Input.Description, createdModule1.Description)
	defer DeleteModule(createdModule1.ID)
	foundParticipantBlockInModule[createdModule1.ID] = map[int64]bool{}

	// module 1 text block
	module1TextBlockInput := &Block{
		Name:       "Module 1 Text Block",
		Summary:    "A text block",
		BlockType:  BlockTypeText,
		AllowReset: "yes",
		Content: BlockText{
			Text: "Text content",
		},
	}
	b.Reset()
	encoder.Encode(module1TextBlockInput)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/text", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule1TextBlock := &Block{}
	err = mapstructure.Decode(m, createdModule1TextBlock)
	suite.Nil(err)
	suite.NotZero(createdModule1TextBlock.ID)
	suite.Equal(module1TextBlockInput.Name, createdModule1TextBlock.Name)
	suite.Equal(module1TextBlockInput.Summary, createdModule1TextBlock.Summary)
	suite.Equal(module1TextBlockInput.BlockType, createdModule1TextBlock.BlockType)
	suite.Equal(module1TextBlockInput.AllowReset, createdModule1TextBlock.AllowReset)
	// handle the interface conversion
	createdModule1TextBlockText := &BlockText{}
	unmB, err := json.Marshal(createdModule1TextBlock.Content)
	suite.Nil(err)
	err = json.Unmarshal(unmB, createdModule1TextBlockText)
	suite.Nil(err)
	suite.Equal("Text content", createdModule1TextBlockText.Text)
	defer DeleteBlock(createdModule1TextBlock.ID)
	// don't forget to link it
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/%d", createdModule1.ID, createdModule1TextBlock.ID, 1), b, routeAdminLinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	foundParticipantBlockInModule[createdModule1.ID][createdModule1TextBlock.ID] = false

	// module 1 embed block
	module1EmbedBlockInput := &Block{
		Name:       "Module 1 Embed Video Block",
		Summary:    "A youtube block",
		BlockType:  BlockTypeEmbed,
		AllowReset: "yes",
		Content: BlockEmbed{
			EmbedType: BlockEmbedTypeYoutube,
			EmbedLink: "somevid",
		},
	}
	b.Reset()
	encoder.Encode(module1EmbedBlockInput)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/embed", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule1EmbedBlock := &Block{}
	err = mapstructure.Decode(m, createdModule1EmbedBlock)
	suite.Nil(err)
	suite.NotZero(createdModule1EmbedBlock.ID)
	suite.Equal(module1EmbedBlockInput.Name, createdModule1EmbedBlock.Name)
	suite.Equal(module1EmbedBlockInput.Summary, createdModule1EmbedBlock.Summary)
	suite.Equal(module1EmbedBlockInput.BlockType, createdModule1EmbedBlock.BlockType)
	suite.Equal(module1EmbedBlockInput.AllowReset, createdModule1EmbedBlock.AllowReset)
	// handle the interface conversion
	createdModule1EmbedBlockEmbed := &BlockEmbed{}
	unmB, err = json.Marshal(createdModule1EmbedBlock.Content)
	suite.Nil(err)
	err = json.Unmarshal(unmB, createdModule1EmbedBlockEmbed)
	suite.Nil(err)
	suite.Equal(BlockEmbedTypeYoutube, createdModule1EmbedBlockEmbed.EmbedType)
	suite.Equal("somevid", createdModule1EmbedBlockEmbed.EmbedLink)
	defer DeleteBlock(createdModule1EmbedBlock.ID)
	// don't forget to link it
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/%d", createdModule1.ID, createdModule1EmbedBlock.ID, 2), b, routeAdminLinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	foundParticipantBlockInModule[createdModule1.ID][createdModule1EmbedBlock.ID] = false

	// module 2
	module2Input := &Module{
		Name:        "Module 2",
		Description: "The Second Module",
		Status:      "active",
	}
	b.Reset()
	encoder.Encode(module2Input)
	code, res, err = testEndpoint(http.MethodPost, "/admin/modules", b, routeAdminCreateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule2 := &Module{}
	err = mapstructure.Decode(m, createdModule2)
	suite.Nil(err)
	suite.NotZero(createdModule2.ID)
	suite.Equal(module2Input.Name, createdModule2.Name)
	suite.Equal(module2Input.Description, createdModule2.Description)
	suite.Equal(module2Input.Status, createdModule2.Status)
	defer DeleteModule(createdModule2.ID)
	foundParticipantBlockInModule[createdModule2.ID] = map[int64]bool{}

	// module 2 external block
	module2ExternalBlockInput := &Block{
		Name:       "Module 2 External Block",
		Summary:    "External!",
		BlockType:  BlockTypeExternal,
		AllowReset: "yes",
		Content: BlockExternal{
			ExternalLink: "somewhere.com",
		},
	}
	b.Reset()
	encoder.Encode(module2ExternalBlockInput)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/external", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule2ExternalBlock := &Block{}
	err = mapstructure.Decode(m, createdModule2ExternalBlock)
	suite.Nil(err)
	suite.NotZero(createdModule2ExternalBlock.ID)
	suite.Equal(module2ExternalBlockInput.Name, createdModule2ExternalBlock.Name)
	suite.Equal(module2ExternalBlockInput.Summary, createdModule2ExternalBlock.Summary)
	suite.Equal(module2ExternalBlockInput.BlockType, createdModule2ExternalBlock.BlockType)
	suite.Equal(module2ExternalBlockInput.AllowReset, createdModule2ExternalBlock.AllowReset)
	// handle the interface conversion
	createdModule2TextBlockExternal := &BlockExternal{}
	unmB, err = json.Marshal(createdModule2ExternalBlock.Content)
	suite.Nil(err)
	err = json.Unmarshal(unmB, createdModule2TextBlockExternal)
	suite.Nil(err)
	suite.Equal("somewhere.com", createdModule2TextBlockExternal.ExternalLink)
	defer DeleteBlock(createdModule2ExternalBlock.ID)
	// don't forget to link it
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/%d", createdModule2.ID, createdModule2ExternalBlock.ID, 1), b, routeAdminLinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	foundParticipantBlockInModule[createdModule2.ID][createdModule2ExternalBlock.ID] = false

	// module 2 text block
	module2TextBlockInput := &Block{
		Name:       "Module 2 Text Block",
		Summary:    "Another text block",
		BlockType:  BlockTypeText,
		AllowReset: "yes",
		Content: BlockText{
			Text: "More text content",
		},
	}
	b.Reset()
	encoder.Encode(module2TextBlockInput)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/text", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule2TextBlock := &Block{}
	err = mapstructure.Decode(m, createdModule2TextBlock)
	suite.Nil(err)
	suite.NotZero(createdModule1TextBlock.ID)
	suite.Equal(module2TextBlockInput.Name, createdModule2TextBlock.Name)
	suite.Equal(module2TextBlockInput.Summary, createdModule2TextBlock.Summary)
	suite.Equal(module2TextBlockInput.BlockType, createdModule2TextBlock.BlockType)
	suite.Equal(module2TextBlockInput.AllowReset, createdModule2TextBlock.AllowReset)
	// handle the interface conversion
	createdModule2TextBlockText := &BlockText{}
	unmB, err = json.Marshal(createdModule2TextBlock.Content)
	suite.Nil(err)
	err = json.Unmarshal(unmB, createdModule2TextBlockText)
	suite.Nil(err)
	suite.Equal("More text content", createdModule2TextBlockText.Text)
	defer DeleteBlock(createdModule2TextBlock.ID)
	// don't forget to link it
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/%d", createdModule2.ID, createdModule2TextBlock.ID, 2), b, routeAdminLinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	foundParticipantBlockInModule[createdModule2.ID][createdModule2TextBlock.ID] = false

	// module 3
	module3Input := &Module{
		Name:        "Module 3",
		Description: "The Third Module",
		Status:      "pending",
	}
	b.Reset()
	encoder.Encode(module3Input)
	code, res, err = testEndpoint(http.MethodPost, "/admin/modules", b, routeAdminCreateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule3 := &Module{}
	err = mapstructure.Decode(m, createdModule3)
	suite.Nil(err)
	suite.NotZero(createdModule3.ID)
	suite.Equal(module3Input.Name, createdModule3.Name)
	suite.Equal(module3Input.Description, createdModule3.Description)
	suite.Equal(module3Input.Status, createdModule3.Status)
	defer DeleteModule(createdModule3.ID)
	foundParticipantBlockInModule[createdModule3.ID] = map[int64]bool{}

	// module 3 form block
	module3FormInput := &Block{
		Name:       "Module 3 Form Block",
		Summary:    "Fill This Out",
		BlockType:  BlockTypeForm,
		AllowReset: "yes",
		Content: BlockForm{
			FormType:      BlockFormTypeSurvey,
			AllowResubmit: "yes",
			Questions: []BlockFormQuestion{
				{
					Question:     "What is your name?",
					QuestionType: BlockFormQuestionTypeShort,
					FormOrder:    1,
				},
				{
					Question:     "What is your quest?",
					QuestionType: BlockFormQuestionTypeLong,
					FormOrder:    2,
				},
				{
					Question:     "What is the meaning of life?",
					QuestionType: BlockFormQuestionTypeSingle,
					FormOrder:    3,
					Options: []BlockFormQuestionOption{
						{
							OptionText:      "42",
							OptionIsCorrect: "yes",
						},
						{
							OptionText:      "24",
							OptionIsCorrect: "no",
						},
					},
				},
				{
					Question:     "What do you like?",
					QuestionType: BlockFormQuestionTypeMultiple,
					FormOrder:    4,
					Options: []BlockFormQuestionOption{
						{
							OptionText:      "Dogs",
							OptionIsCorrect: "yes",
						},
						{
							OptionText:      "Cats",
							OptionIsCorrect: "yes",
						},
						{
							OptionText:      "Platypodes",
							OptionIsCorrect: "yes",
						},
					},
				},
			},
		},
	}
	b.Reset()
	encoder.Encode(module3FormInput)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/form", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule3FormBlock := &Block{}
	err = mapstructure.Decode(m, createdModule3FormBlock)
	suite.Nil(err)
	suite.NotZero(createdModule3FormBlock.ID)
	suite.Equal(module3FormInput.Name, createdModule3FormBlock.Name)
	suite.Equal(module3FormInput.Summary, createdModule3FormBlock.Summary)
	suite.Equal(module3FormInput.BlockType, createdModule3FormBlock.BlockType)
	suite.Equal(module3FormInput.AllowReset, createdModule3FormBlock.AllowReset)
	// handle the interface conversion
	createdModule3FormBlockForm := &BlockForm{}
	unmB, err = json.Marshal(createdModule3FormBlock.Content)
	suite.Nil(err)
	err = json.Unmarshal(unmB, createdModule3FormBlockForm)
	suite.Nil(err)
	suite.Equal(4, len(createdModule3FormBlockForm.Questions))
	defer DeleteBlock(createdModule3FormBlock.ID)
	// don't forget to link it
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/%d", createdModule3.ID, createdModule3FormBlock.ID, 1), b, routeAdminLinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	foundParticipantBlockInModule[createdModule3.ID][createdModule3FormBlock.ID] = false

	// test the update of the form
	questions := createdModule3FormBlockForm.Questions
	questions = append(questions, BlockFormQuestion{
		Question:     "What else do you like?",
		QuestionType: BlockFormQuestionTypeMultiple,
		FormOrder:    5,
		Options: []BlockFormQuestionOption{
			{
				OptionText: "Dogs",
			},
			{
				OptionText: "Cats",
			},
			{
				OptionText: "Platypodes",
			},
		},
	})
	questions[2].Options[1].OptionText = "24?"
	b.Reset()
	encoder.Encode(&Block{
		Content: BlockForm{
			Questions: questions,
		},
	})
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/blocks/%d", createdModule3FormBlock.ID), b, routeAdminUpdateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// module 3 text block 1
	module3TextBlock1Input := &Block{
		Name:       "Module 3 Text Block 1",
		Summary:    "Another text block in 3-1",
		BlockType:  BlockTypeText,
		AllowReset: "yes",
		Content: BlockText{
			Text: "Yep, you guessed it, more text content",
		},
	}
	b.Reset()
	encoder.Encode(module3TextBlock1Input)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/text", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule3Text1Block := &Block{}
	err = mapstructure.Decode(m, createdModule3Text1Block)
	suite.Nil(err)
	suite.NotZero(createdModule1TextBlock.ID)
	suite.Equal(module3TextBlock1Input.Name, createdModule3Text1Block.Name)
	suite.Equal(module3TextBlock1Input.Summary, createdModule3Text1Block.Summary)
	suite.Equal(module3TextBlock1Input.BlockType, createdModule3Text1Block.BlockType)
	suite.Equal(module3TextBlock1Input.AllowReset, createdModule3Text1Block.AllowReset)
	// handle the interface conversion
	createdModule3TextBlock1Text := &BlockText{}
	unmB, err = json.Marshal(createdModule3Text1Block.Content)
	suite.Nil(err)
	err = json.Unmarshal(unmB, createdModule3TextBlock1Text)
	suite.Nil(err)
	suite.Equal("Yep, you guessed it, more text content", createdModule3TextBlock1Text.Text)
	defer DeleteBlock(createdModule3Text1Block.ID)
	// don't forget to link it
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/%d", createdModule3.ID, createdModule3Text1Block.ID, 2), b, routeAdminLinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	foundParticipantBlockInModule[createdModule3.ID][createdModule3Text1Block.ID] = false

	// module 3 text block 2
	module3TextBlock2Input := &Block{
		Name:       "Module 3 Text Block 2",
		Summary:    "Another text block in 3-2",
		BlockType:  BlockTypeText,
		AllowReset: "yes",
		Content: BlockText{
			Text: "What? Yep, more text content",
		},
	}
	b.Reset()
	encoder.Encode(module3TextBlock2Input)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/text", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule3Text2Block := &Block{}
	err = mapstructure.Decode(m, createdModule3Text2Block)
	suite.Nil(err)
	suite.NotZero(createdModule1TextBlock.ID)
	suite.Equal(module3TextBlock2Input.Name, createdModule3Text2Block.Name)
	suite.Equal(module3TextBlock2Input.Summary, createdModule3Text2Block.Summary)
	suite.Equal(module3TextBlock2Input.BlockType, createdModule3Text2Block.BlockType)
	suite.Equal(module3TextBlock2Input.AllowReset, createdModule3Text2Block.AllowReset)
	// handle the interface conversion
	createdModule3TextBlock2Text := &BlockText{}
	unmB, err = json.Marshal(createdModule3Text2Block.Content)
	suite.Nil(err)
	err = json.Unmarshal(unmB, createdModule3TextBlock2Text)
	suite.Nil(err)
	suite.Equal("What? Yep, more text content", createdModule3TextBlock2Text.Text)
	defer DeleteBlock(createdModule3Text2Block.ID)
	// don't forget to link it
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/%d", createdModule3.ID, createdModule3Text2Block.ID, 3), b, routeAdminLinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	foundParticipantBlockInModule[createdModule3.ID][createdModule3Text2Block.ID] = false

	// update each to set status to active
	activeMap := map[string]string{"status": "active"}
	b.Reset()
	encoder.Encode(&activeMap)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/modules/%d", createdModule1.ID), b, routeAdminUpdateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	b.Reset()
	encoder.Encode(&activeMap)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/modules/%d", createdModule2.ID), b, routeAdminUpdateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	b.Reset()
	encoder.Encode(&activeMap)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/modules/%d", createdModule3.ID), b, routeAdminUpdateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// save the flow
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/projects/%d/modules/%d/order/%d", createdProject.ID, createdModule1.ID, 1), b, routeAdminLinkModuleAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/projects/%d/modules/%d/order/%d", createdProject.ID, createdModule2.ID, 2), b, routeAdminLinkModuleAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/projects/%d/modules/%d/order/%d", createdProject.ID, createdModule3.ID, 3), b, routeAdminLinkModuleAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// save the consent
	consentFormInput := &ConsentForm{
		ContentInMarkdown:             "Sign up at your own risk",
		ContactInformationDisplay:     "Contact us",
		InstitutionInformationDisplay: "Some Org",
	}
	b.Reset()
	encoder.Encode(consentFormInput)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/admin/projects/%d/consent", createdProject.ID), b, routeAdminSaveConsentForm, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdConsentForm := &ConsentForm{}
	err = mapstructure.Decode(m, createdConsentForm)
	suite.Nil(err)
	suite.Equal(consentFormInput.ContactInformationDisplay, createdConsentForm.ContactInformationDisplay)
	suite.Equal(consentFormInput.ContentInMarkdown, createdConsentForm.ContentInMarkdown)
	suite.Equal(consentFormInput.InstitutionInformationDisplay, createdConsentForm.InstitutionInformationDisplay)
	suite.Equal(createdProject.ID, createdConsentForm.ProjectID)
	defer DeleteConsentFormForProject(createdProject.ID)

	// not logged in user user visits site, gets the site and list of projects
	code, res, err = testEndpoint(http.MethodGet, "/site", b, routeAllGetSite, "")
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	foundSite := &Site{}
	err = mapstructure.Decode(m, foundSite)
	suite.Nil(err)

	code, res, err = testEndpoint(http.MethodGet, "/projects", b, routeAllGetProjects, "")
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	mS, err := testEndpointResultToSlice(res)
	suite.Nil(err)
	projects := []Project{}
	err = mapstructure.Decode(mS, &projects)
	suite.Nil(err)
	suite.NotZero(len(projects))
	foundCreatedProject := false
	for _, p := range projects {
		if p.ID == createdProject.ID {
			foundCreatedProject = true
		}
	}
	suite.True(foundCreatedProject)

	// not logged in user gets the project and consent
	// note we use the created project id since this could, theoretically, be run against a running active db
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d", createdProject.ID), b, routeAllGetProject, "")
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	foundProject := &Project{}
	err = mapstructure.Decode(m, foundProject)
	suite.Nil(err)
	suite.Equal(createdProject.Name, foundProject.Name)

	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent", createdProject.ID), b, routeAllGetConsentForm, "")
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	foundConsent := &ConsentForm{}
	err = mapstructure.Decode(m, foundConsent)
	suite.Nil(err)
	suite.Equal(createdConsentForm.ContactInformationDisplay, foundConsent.ContactInformationDisplay)
	suite.Equal(createdConsentForm.InstitutionInformationDisplay, foundConsent.InstitutionInformationDisplay)
	suite.Equal(createdConsentForm.ContentInMarkdown, foundConsent.ContentInMarkdown)

	// participant sign up
	consentResponseInput := &ConsentResponse{
		ConsentStatus:                         ConsentResponseStatusAccepted,
		ParticipantProvidedFirstName:          "First Name",
		ParticipantProvidedLastName:           "First Name",
		ParticipantProvidedContactInformation: "First Name",
		ProjectCode:                           createdProject.ShortCode,
		ParticipantComments:                   "First Name",
	}
	b.Reset()
	encoder.Encode(consentResponseInput)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent/responses", createdProject.ID), b, routeAllCreateConsentResponse, "")
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)

	// should fail since missing the code; try again with the code; should still fail without a password
	consentResponseInput.ProjectCode = projectInput.ShortCode
	b.Reset()
	encoder.Encode(consentResponseInput)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent/responses", createdProject.ID), b, routeAllCreateConsentResponse, "")
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)

	consentResponseInput.User = &User{
		Password: participantPassword,
	}
	b.Reset()
	encoder.Encode(consentResponseInput)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent/responses", createdProject.ID), b, routeAllCreateConsentResponse, "")
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	consentResponseResult := &ConsentResponse{}
	err = mapstructure.Decode(m, consentResponseResult)
	suite.Nil(err)

	// user for use later
	createdParticipant := consentResponseResult.User
	require.NotNil(createdParticipant)
	defer DeleteUser(createdParticipant.ID)

	// log in the user
	loginParticipantInput := loginInput{
		Login:    createdParticipant.ParticipantCode,
		Password: participantPassword,
	}
	b.Reset()
	encoder.Encode(loginParticipantInput)
	code, res, err = testEndpoint(http.MethodPost, "/login", b, routeAllUserLogin, "")
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	part := &User{}
	err = mapstructure.Decode(m, part)
	suite.Nil(err)

	// get the project and meta data, ensure no errors
	code, res, err = testEndpoint(http.MethodGet, "/participant/projects", b, routeParticipantGetProjects, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	mS, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	suite.NotZero(len(mS))

	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d", createdProject.ID), b, routeParticipantGetProject, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	partProject := &Project{}
	err = mapstructure.Decode(m, partProject)
	suite.Nil(err)

	// participant completes each block in turn
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/flow", createdProject.ID), b, routeParticipantGetProjectFlow, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	mS, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	suite.NotZero(len(mS))
	flow := []Flow{}
	err = mapstructure.Decode(mS, &flow)
	suite.Nil(err)
	suite.NotZero(len(flow))

	// this bit is a slog; we are going to loop over each entry in the flow,
	// get the block, check the block, complete the block, try to reset the block,
	// then move on, keeping track of the status the whole way; of course, the form
	// will need a different bit of logic

	count := 0
	for _, f := range flow {
		// get the block
		code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d", createdProject.ID, f.ModuleID, f.BlockID), b, routeParticipantGetBlock, part.Access)
		suite.Nil(err)
		suite.Equal(http.StatusOK, code, res)
		m, err = testEndpointResultToMap(res)
		suite.Nil(err)
		block := &Block{}
		err = mapstructure.Decode(m, block)
		suite.Nil(err)

		// check it

		// if it's not a form, we can just toggle some stuff
		if block.BlockType != BlockTypeForm {
			// mark complete
			code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/status/%s", createdProject.ID, f.ModuleID, f.BlockID, BlockUserStatusCompleted), b, routeParticipantSaveBlockStatus, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)

			// make sure it's complete
			code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d", createdProject.ID, f.ModuleID, f.BlockID), b, routeParticipantGetBlock, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)
			suite.Nil(err)
			m, err = testEndpointResultToMap(res)
			suite.Nil(err)
			completedBlock := &Block{}
			err = mapstructure.Decode(m, completedBlock)
			suite.Nil(err)
			suite.Equal(BlockUserStatusCompleted, completedBlock.UserStatus)

			// reset it
			code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/status/%s", createdProject.ID, f.ModuleID, f.BlockID, BlockUserStatusNotStarted), b, routeParticipantSaveBlockStatus, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)

			// verify; note that GETting the block should update the status to started
			code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d", createdProject.ID, f.ModuleID, f.BlockID), b, routeParticipantGetBlock, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)
			m, err = testEndpointResultToMap(res)
			suite.Nil(err)
			resetBlock := &Block{}
			err = mapstructure.Decode(m, resetBlock)
			suite.Nil(err)
			suite.Equal(BlockUserStatusStarted, resetBlock.UserStatus)

			// complete it again
			code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/status/%s", createdProject.ID, f.ModuleID, f.BlockID, BlockUserStatusCompleted), b, routeParticipantSaveBlockStatus, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)
		} else {
			// let's answer the questions...
			currentQuestions := &BlockForm{}
			unmB, err = json.Marshal(block.Content)
			suite.Nil(err)
			err = json.Unmarshal(unmB, currentQuestions)
			suite.Nil(err)
			suite.Equal(5, len(currentQuestions.Questions))
			responses := []BlockFormSubmissionResponse{}
			for _, q := range currentQuestions.Questions {
				response := BlockFormSubmissionResponse{
					QuestionID: q.ID,
				}
				if q.QuestionType == BlockFormQuestionTypeShort {
					response.TextResponse = "Dr. Doom"
				} else if q.QuestionType == BlockFormQuestionTypeLong {
					response.TextResponse = "Domination of all!"
				} else if q.QuestionType == BlockFormQuestionTypeSingle {
					response.OptionID = q.Options[0].ID
				} else if q.QuestionType == BlockFormQuestionTypeMultiple {
					response.OptionID = q.Options[0].ID
				}
				responses = append(responses, response)
			}
			send := &BlockFormQestionResponseInput{
				Responses: responses,
			}

			// fail on the wrong endpoint
			code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/status/%s", createdProject.ID, f.ModuleID, f.BlockID, BlockUserStatusCompleted), b, routeParticipantSaveBlockStatus, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusBadRequest, code, res)

			b.Reset()
			encoder.Encode(send)
			code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/submissions", createdProject.ID, f.ModuleID, f.BlockID), b, routeParticipantSaveFormResponse, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)

			// get the response, make sure everything is accurate with the sent responses
			code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/submissions", createdProject.ID, f.ModuleID, f.BlockID), b, routeParticipantGetFormSubmissions, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)
			mS, err = testEndpointResultToSlice(res)
			suite.Nil(err)
			foundSubmissions := []BlockFormSubmission{}
			err = mapstructure.Decode(mS, &foundSubmissions)
			suite.Nil(err)
			require.Equal(1, len(foundSubmissions))
			suite.Equal(5, len(foundSubmissions[0].Responses))

			// yes, loops in loops are generally not great, but it's a test and limited data
			for _, originalResponse := range responses {
				for _, foundResponse := range foundSubmissions[0].Responses {
					if originalResponse.QuestionID == foundResponse.QuestionID {
						suite.Equal(originalResponse.OptionID, foundResponse.OptionID)
					}
				}
			}

			// delete the form submission
			code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/submissions", createdProject.ID, f.ModuleID, f.BlockID), b, routeParticipantDeleteSubmissions, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)

			// get it, it should be gone
			code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/submissions", createdProject.ID, f.ModuleID, f.BlockID), b, routeParticipantGetFormSubmissions, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)
			mS, err = testEndpointResultToSlice(res)
			suite.Nil(err)
			suite.Zero(len(mS))

			// resubmit just for safety
			b.Reset()
			encoder.Encode(send)
			code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/submissions", createdProject.ID, f.ModuleID, f.BlockID), b, routeParticipantSaveFormResponse, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)

			code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/submissions", createdProject.ID, f.ModuleID, f.BlockID), b, routeParticipantGetFormSubmissions, part.Access)
			suite.Nil(err)
			suite.Equal(http.StatusOK, code, res)
			mS, err = testEndpointResultToSlice(res)
			suite.Nil(err)
			suite.NotZero(len(mS))

		}

		// set in map
		foundParticipantBlockInModule[f.ModuleID][f.BlockID] = true
		count++
	}
	suite.NotZero(count)

	modCount := 0
	for mID := range foundParticipantBlockInModule {
		for _, v := range foundParticipantBlockInModule[mID] {
			suite.True(v)
			modCount++
		}
	}

	suite.Equal(7, modCount)

	// get the flow and make sure they are all completed now
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/flow", createdProject.ID), b, routeParticipantGetProjectFlow, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	mS, err = testEndpointResultToSlice(res)
	suite.Nil(err)
	suite.NotZero(len(mS))
	completedFlow := []Flow{}
	err = mapstructure.Decode(mS, &completedFlow)
	suite.Nil(err)
	suite.NotZero(len(flow))
	count = 0
	for _, f := range completedFlow {
		suite.Equal(BlockUserStatusCompleted, f.UserStatus)
		count++
	}
	suite.Equal(7, modCount)

	// project is complete, so check the status
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d", createdProject.ID), b, routeParticipantGetProject, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	finalProject := &Project{}
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	err = mapstructure.Decode(m, finalProject)
	suite.Nil(err)
	suite.Equal(ProjectStatusCompleted, finalProject.ParticipantStatus)

	// clean up with some fun deletes

	// first, have the user change all their status in the modules
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d/status", createdProject.ID, createdModule1.ID, createdModule1TextBlock.ID), b, routeParticipantRemoveBlockStatus, part.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	// make sure it reset
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d", createdProject.ID, createdModule1.ID, createdModule1TextBlock.ID), b, routeParticipantGetBlock, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	block := &Block{}
	err = mapstructure.Decode(m, block)
	suite.Nil(err)
	suite.Equal(BlockUserStatusStarted, block.UserStatus)

	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/participant/projects/%d/modules/%d/status", createdProject.ID, createdModule1.ID), b, routeParticipantRemoveBlockStatus, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	// verify that the other general one is reset
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d", createdProject.ID, createdModule1.ID, createdModule1EmbedBlock.ID), b, routeParticipantGetBlock, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	block = &Block{}
	err = mapstructure.Decode(m, block)
	suite.Nil(err)
	suite.Equal(BlockUserStatusStarted, block.UserStatus)

	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/participant/projects/%d/status", createdProject.ID), b, routeParticipantRemoveBlockStatus, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	// all should be reset, so grab a block from m2 and m3
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d", createdProject.ID, createdModule2.ID, createdModule2ExternalBlock.ID), b, routeParticipantGetBlock, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	block = &Block{}
	err = mapstructure.Decode(m, block)
	suite.Nil(err)
	suite.Equal(BlockUserStatusStarted, block.UserStatus)

	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d", createdProject.ID, createdModule3.ID, createdModule3Text1Block.ID), b, routeParticipantGetBlock, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	block = &Block{}
	err = mapstructure.Decode(m, block)
	suite.Nil(err)
	suite.Equal(BlockUserStatusStarted, block.UserStatus)

	// now delete their participation
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/participant/projects/%d", createdProject.ID), b, routeParticipantUnlinkUserAndProject, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d", createdProject.ID), b, routeParticipantGetProject, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)

	// admin deletes the consent
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/projects/%d/consent", createdProject.ID), b, routeAdminDeleteConsentForm, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// admin removes a block from a module
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/modules/%d/blocks/%d", createdModule1.ID, createdModule3Text1Block.ID), b, routeAdminUnlinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// admin removes a module from the flow
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/projects/%d/modules/%d", createdProject.ID, createdModule1.ID), b, routeAdminUnlinkModuleAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// admin deletes all blocks in a module
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/modules/%d/blocks", createdModule1.ID), b, routeAdminUnlinkAllBlocksFromModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// admin unlinks all modules from a project
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/projects/%d/flow", createdProject.ID), b, routeAdminUnlinkAllModulesFromProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// admin deletes the project// TODO: implement

}
