package api

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SuiteTestsSiteDB struct {
	suite.Suite
}

func TestSuiteTestsSiteDB(t *testing.T) {
	suite.Run(t, new(SuiteTestsSiteDB))
}

func (suite *SuiteTestsSiteDB) SetupSuite() {
	setupTesting()
}

// this test will bail if files cannot be uploaded as it would assume correct responses; if
// file uploading is enabled through the environment, then the other test will bail and this
// will run
func (suite SuiteTestsSiteDB) TestSiteDBCalls() {
	require := suite.Require()
	currentSite, err := GetSite()
	require.Nil(err)

	originalSite, err := GetSiteByID(currentSite.ID)
	suite.Nil(err)

	createdSite := &Site{
		ShortName:   "test",
		Name:        "Testing",
		Description: "Created during testing",
		Domain:      "test.kesplora.com",
	}

	err = CreateSite(createdSite)
	suite.Nil(err)
	defer DeleteSiteByID(createdSite.ID)

	foundCreatedSite, err := GetSiteByID(createdSite.ID)
	suite.Nil(err)
	suite.Equal(createdSite.ID, foundCreatedSite.ID)
	suite.Equal(createdSite.ShortName, foundCreatedSite.ShortName)
	suite.Equal(createdSite.Name, foundCreatedSite.Name)
	suite.Equal(createdSite.Description, foundCreatedSite.Description)
	suite.Equal(createdSite.Domain, foundCreatedSite.Domain)

	updateSite := &Site{
		ID:          createdSite.ID,
		Name:        "Testing Updated",
		Description: "Testing description updated",
		ShortName:   "test_updated",
		Domain:      "test2.kesplora.com",
	}
	err = UpdateSite(updateSite)
	suite.Nil(err)
	foundUpdatedSite, err := GetSiteByID(createdSite.ID)
	suite.Nil(err)
	suite.Equal(createdSite.ID, foundUpdatedSite.ID)
	suite.Equal(updateSite.ShortName, foundUpdatedSite.ShortName)
	suite.Equal(updateSite.Name, foundUpdatedSite.Name)
	suite.Equal(updateSite.Description, foundUpdatedSite.Description)
	suite.Equal(updateSite.Domain, foundUpdatedSite.Domain)

	err = DeleteSiteByID(createdSite.ID)
	suite.Nil(err)

	foundSiteAfterDelete, err := GetSite()
	suite.Nil(err)
	suite.Equal(originalSite.ID, foundSiteAfterDelete.ID)
}
