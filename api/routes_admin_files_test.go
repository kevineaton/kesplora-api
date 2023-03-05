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

type SuiteTestsFilesRoutes struct {
	suite.Suite
}

func TestSuiteTestsFilesRoutes(t *testing.T) {
	suite.Run(t, new(SuiteTestsFilesRoutes))
}

func (suite *SuiteTestsFilesRoutes) SetupSuite() {
	setupTesting()
}

// this test will bail if files cannot be uploaded as it would assume correct responses; if
// file uploading is enabled through the environment, then the other test will bail and this
// will run
func (suite SuiteTestsFilesRoutes) TestAdminFileRoutesImplementedCRUD() {

	// note that without AWS cedentials, many of the calls simply will not work
	// since the file upload fails; we can hack around some of it for testing, but
	// if you want to actually test the full uploads/downloads, you need to provide AWS S3
	// credentials and there may be a cost involved!

	providers, err := getAllowedFileProviders()
	if err != nil || len(providers) == 0 {
		suite.T().Skip("no providers configured")
	}

	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.Encode(map[string]string{})
	require := suite.Require()
	participantPassword := randomString(64)

	// first, create an admin
	admin := &User{
		SystemRole: UserSystemRoleAdmin,
	}
	err = createTestUser(admin)
	require.Nil(err)
	defer DeleteUser(admin.ID)
	// now a participant
	participant := &User{
		SystemRole: UserSystemRoleParticipant,
		Password:   participantPassword,
	}
	err = createTestUser(participant)
	suite.Nil(err)
	defer DeleteUser(participant.ID)

	project := &Project{}
	err = createTestProject(project)
	suite.Nil(err)
	defer DeleteProject(project.ID)

	// upload a file

	// rather than worrying about file i/o, these bytes represent a small .txt file with the string "Hello world!"
	textBytes := []byte{72, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 33}
	fileBuffer := bytes.NewBuffer(textBytes)
	code, res, err := testEndpointUpload("/admin/files", "test.txt", fileBuffer, routeAdminUploadFile, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err := testEndpointResultToMap(res)
	suite.Nil(err)
	createdFile := &File{}
	err = mapstructure.Decode(m, createdFile)
	suite.Nil(err)
	suite.NotZero(createdFile.ID, createdFile)
	suite.NotZero(createdFile.FileSize, createdFile)
	suite.Equal("test.txt", createdFile.Display)
	suite.Equal("test.txt", createdFile.RemoteKey)
	suite.Equal(".txt", createdFile.FileType)

	// get a list of files
	code, res, err = testEndpoint(http.MethodGet, "/admin/files", b, routeAdminGetFiles, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	mS, err := testEndpointResultToSlice(res)
	suite.Nil(err)
	suite.NotZero(len(mS))
	foundFiles := []File{}
	err = mapstructure.Decode(mS, &foundFiles)
	suite.Nil(err)
	suite.NotZero(len(foundFiles))
	found := false
	for _, f := range foundFiles {
		if f.ID == createdFile.ID {
			found = true
			suite.Equal(f.Display, createdFile.Display)
			suite.Equal(f.RemoteKey, createdFile.RemoteKey)
			suite.Equal(f.FileType, createdFile.FileType)
		}
	}
	suite.True(found)

	// update the meta data
	updateInput := &File{
		Display:     "Hello Text",
		Description: "A testing text file. If you see this in storage, the test didn't clean it up :( ",
		Visibility:  "users",
	}
	b.Reset()
	encoder.Encode(updateInput)
	code, res, err = testEndpoint(http.MethodPatch, fmt.Sprintf("/admin/files/%d", createdFile.ID), b, routeUpdateFileMetadata, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// get the file
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/files/%d", createdFile.ID), b, routeAdminGetFileMetaData, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	foundFile := &File{}
	err = mapstructure.Decode(m, foundFile)
	suite.Nil(err)
	suite.NotZero(foundFile.ID, foundFile)
	suite.NotZero(foundFile.FileSize, foundFile)
	suite.Equal(createdFile.FileSize, foundFile.FileSize)
	suite.Equal(updateInput.Display, foundFile.Display)
	suite.Equal(updateInput.Description, foundFile.Description)
	suite.Equal(updateInput.Visibility, foundFile.Visibility)
	suite.Equal(createdFile.RemoteKey, foundFile.RemoteKey)

	// create a file block

	moduleInput := &Module{
		Name: "Module For File Test",
	}
	b.Reset()
	encoder.Encode(moduleInput)
	code, res, err = testEndpoint(http.MethodPost, "/admin/modules", b, routeAdminCreateModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdModule := &Module{}
	err = mapstructure.Decode(m, createdModule)
	suite.Nil(err)
	suite.NotZero(createdModule.ID)
	suite.Equal(moduleInput.Name, createdModule.Name)
	suite.Equal(moduleInput.Description, createdModule.Description)
	defer DeleteModule(createdModule.ID)

	// module file block
	fileBlockInput := &Block{
		Name:       "Module File",
		Summary:    "A file block",
		BlockType:  BlockTypeFile,
		AllowReset: "yes",
		Content: BlockFile{
			FileID: createdFile.ID,
		},
	}
	b.Reset()
	encoder.Encode(fileBlockInput)
	code, res, err = testEndpoint(http.MethodPost, "/admin/blocks/file", b, routeAdminCreateBlock, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	createdFileBlock := &Block{}
	err = mapstructure.Decode(m, createdFileBlock)
	suite.Nil(err)
	suite.NotZero(createdFileBlock.ID)
	suite.Equal(fileBlockInput.Name, createdFileBlock.Name)
	suite.Equal(fileBlockInput.Summary, createdFileBlock.Summary)
	suite.Equal(fileBlockInput.BlockType, createdFileBlock.BlockType)
	suite.Equal(fileBlockInput.AllowReset, createdFileBlock.AllowReset)
	// handle the interface conversion
	createdFileBlockContent := &BlockFile{}
	unmB, err := json.Marshal(createdFileBlock.Content)
	suite.Nil(err)
	err = json.Unmarshal(unmB, createdFileBlockContent)
	suite.Nil(err)
	suite.Equal(createdFile.ID, createdFileBlockContent.FileID)
	defer DeleteBlock(createdFileBlock.ID)
	defer DeleteBlock(createdFileBlock.ID)
	defer DeleteModule(createdModule.ID)

	// put it in a module and try to get it as a participant
	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/modules/%d/blocks/%d/order/%d", createdModule.ID, createdFileBlock.ID, 1), b, routeAdminLinkBlockAndModule, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	code, res, err = testEndpoint(http.MethodPut, fmt.Sprintf("/admin/projects/%d/modules/%d/order/%d", project.ID, createdModule.ID, 1), b, routeAdminLinkModuleAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/admin/projects/%d/users/%d", project.ID, participant.ID), b, routeAdminLinkUserAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	loginParticipantInput := loginInput{
		Login:    participant.Email,
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
	participant.Access = part.Access

	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/projects/%d/modules/%d/blocks/%d", project.ID, createdModule.ID, createdFileBlock.ID), b, routeParticipantGetBlock, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	foundBlock := &Block{}
	err = mapstructure.Decode(m, foundBlock)
	suite.Nil(err)

	// check it
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/files/%d", createdFileBlockContent.FileID), b, routeParticipantGetFileMetaData, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	foundFileByParticipant := &File{}
	err = mapstructure.Decode(m, foundFileByParticipant)
	suite.Nil(err)
	suite.NotZero(foundFileByParticipant.ID, foundFileByParticipant)
	suite.NotZero(foundFileByParticipant.FileSize, foundFileByParticipant)
	suite.Equal(foundFile.FileSize, foundFileByParticipant.FileSize)
	suite.Equal(foundFile.Display, foundFileByParticipant.Display)
	suite.Equal(foundFile.Description, foundFileByParticipant.Description)
	suite.Equal(foundFile.Visibility, foundFileByParticipant.Visibility)
	suite.Equal(foundFile.RemoteKey, foundFileByParticipant.RemoteKey)

	// download it
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/files/%d/download", createdFileBlockContent.FileID), b, routeParticipantDownloadFile, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res) // we don't need to parse the file at this time, just test permissions

	// verify admin can't use that route
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/participant/files/%d/download", createdFileBlockContent.FileID), b, routeParticipantDownloadFile, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)

	// ditto for the admin
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/files/%d/download", createdFileBlockContent.FileID), b, routeAdminDownloadFile, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res) // we don't need to parse the file at this time, just test permissions

	// verify the participant can't use that route
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/admin/files/%d/download", createdFileBlockContent.FileID), b, routeAdminDownloadFile, part.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)

	// replace it slightly
	updatedTextBytes := []byte{72, 101, 108, 108, 111, 32, 119, 111, 114, 108, 100, 32}
	fileBuffer = bytes.NewBuffer(updatedTextBytes)
	code, res, err = testEndpointUpload(fmt.Sprintf("/admin/files/%d", createdFile.ID), "test.txt", fileBuffer, routeAdminReplaceFile, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusCreated, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	replacedFile := &File{}
	err = mapstructure.Decode(m, replacedFile)
	suite.Nil(err)
	suite.NotZero(replacedFile.ID, replacedFile)
	suite.NotZero(replacedFile.FileSize, replacedFile)

	// remove the user from the project
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/projects/%d/users/%d", project.ID, participant.ID), b, routeAdminUnlinkUserAndProject, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// delete
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/admin/files/%d", createdFileBlockContent.FileID), b, routeAdminDeleteFile, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// delete the block

	err = DeleteFileFromDB(createdFile.ID)
	suite.Nil(err)
	err = DeleteFileFromBucket("test.txt")
	suite.Nil(err)
}
