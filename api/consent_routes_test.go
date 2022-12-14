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

type SuiteTestsConsentRoutes struct {
	suite.Suite
}

func TestSuiteTestConsentRoutes(t *testing.T) {
	suite.Run(t, new(SuiteTestsConsentRoutes))
}

func (suite *SuiteTestsConsentRoutes) SetupSuite() {
	setupTesting()
}

func (suite SuiteTestsConsentRoutes) TestConsentAdminRoutes() {
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
	user1 := &User{
		SystemRole: UserSystemRoleUser,
	}
	err = createTestUser(user1)
	suite.Nil(err)
	defer DeleteUser(user1.ID)

	user2 := &User{
		SystemRole: UserSystemRoleUser,
	}
	err = createTestUser(user2)
	suite.Nil(err)
	defer DeleteUser(user2.ID)

	project := &Project{
		SignupStatus:    ProjectSignupStatusWithCode,
		ShortCode:       "test_code",
		MaxParticipants: 1,
	}
	err = createTestProject(project)
	suite.Nil(err)
	defer DeleteProject(project.ID)
	defer DeleteConsentFormForProject(project.ID)

	// for this test, we are only testing the admin flows and connections; user flows will
	// be tested elsewhere and with better flows

	code, res, err := testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent", project.ID), b, routeSaveConsentForm, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res) // currently, we allow everything to be blank
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent", project.ID), b, routeSaveConsentForm, user1.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/projects/%d/consent", project.ID), b, routeDeleteConsentForm, user1.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent/responses", project.ID), b, routeGetConsentResponses, user1.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)

	formInput := &ConsentForm{
		ContentInMarkdown:             "# Consent to Participate\n",
		ContactInformationDisplay:     "Contact: test@kesplora.com",
		InstitutionInformationDisplay: "Sponsored by Testing Unlimited Plus Ultra",
	}
	b.Reset()
	encoder.Encode(formInput)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent", project.ID), b, routeSaveConsentForm, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)

	// now get it
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent", project.ID), b, routeGetConsentForm, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	m, err := testEndpointResultToMap(res)
	suite.Nil(err)
	foundFromAdmin := &ConsentForm{}
	err = mapstructure.Decode(m, foundFromAdmin)
	suite.Nil(err)
	suite.Equal(project.ID, foundFromAdmin.ProjectID)
	suite.Equal(formInput.ContentInMarkdown, foundFromAdmin.ContentInMarkdown)
	suite.Equal(formInput.ContactInformationDisplay, foundFromAdmin.ContactInformationDisplay)
	suite.Equal(formInput.InstitutionInformationDisplay, foundFromAdmin.InstitutionInformationDisplay)

	// make sure the user can see it
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent", project.ID), b, routeGetConsentForm, user1.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	foundFromUser := &ConsentForm{}
	err = mapstructure.Decode(m, foundFromUser)
	suite.Nil(err)
	suite.Equal(project.ID, foundFromUser.ProjectID)
	suite.Equal(formInput.ContentInMarkdown, foundFromUser.ContentInMarkdown)
	suite.Equal(formInput.ContactInformationDisplay, foundFromUser.ContactInformationDisplay)
	suite.Equal(formInput.InstitutionInformationDisplay, foundFromUser.InstitutionInformationDisplay)

	// allow user 1 to respond and check paths
	user1ResponseInput := &ConsentResponse{
		ConsentStatus:                         ConsentResponseStatusAccepted,
		ParticipantProvidedFirstName:          "First",
		ParticipantProvidedLastName:           "Last",
		ParticipantProvidedContactInformation: "test.user@kesplora.com",
	}
	b.Reset()
	encoder.Encode(user1ResponseInput)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent/responses", project.ID), b, routeCreateConsentResponse, user1.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)

	// now set the code correctly
	user1ResponseInput.ProjectCode = project.ShortCode
	b.Reset()
	encoder.Encode(user1ResponseInput)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent/responses", project.ID), b, routeCreateConsentResponse, user1.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	user1Response := &ConsentResponse{}
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	err = mapstructure.Decode(m, user1Response)
	suite.Nil(err)
	suite.NotZero(user1Response.ID)
	suite.Equal(project.ID, user1Response.ProjectID)
	suite.Equal(user1ResponseInput.ParticipantProvidedFirstName, user1Response.ParticipantProvidedFirstName)
	suite.Equal(user1ResponseInput.ParticipantProvidedLastName, user1Response.ParticipantProvidedLastName)
	suite.Equal(user1ResponseInput.ParticipantProvidedContactInformation, user1Response.ParticipantProvidedContactInformation)

	// max was hit, make sure user 2 can't enroll
	user2ResponseInput := &ConsentResponse{
		ProjectCode: project.ShortCode,
	}
	b.Reset()
	encoder.Encode(user2ResponseInput)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent/responses", project.ID), b, routeCreateConsentResponse, user2.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)

	// make sure viewing is limited to admin and user 1
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent/responses/%d", project.ID, user1Response.ID), b, routeGetConsentResponse, user1.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent/responses/%d", project.ID, user1Response.ID), b, routeGetConsentResponse, user2.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent/responses/%d", project.ID, user1Response.ID), b, routeGetConsentResponse, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// make sure getting all is limited to the admin
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent/responses", project.ID), b, routeGetConsentResponses, user1.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent/responses", project.ID), b, routeGetConsentResponses, user2.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent/responses", project.ID), b, routeGetConsentResponses, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	s, err := testEndpointResultToSlice(res)
	suite.Nil(err)
	require.NotZero(len(s))

	// try to change it without an override; should fail
	formInput.ContactInformationDisplay = "Updated!"
	b.Reset()
	encoder.Encode(formInput)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent", project.ID), b, routeSaveConsentForm, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusForbidden, code, res)

	// similarly, deleting should fail
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/projects/%d/consent", project.ID), b, routeDeleteConsentForm, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusForbidden, code, res)

	// again with an override
	formInput.ContactInformationDisplay = "Updated!"
	formInput.OverrideSaveIfParticipants = true
	b.Reset()
	encoder.Encode(formInput)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent", project.ID), b, routeSaveConsentForm, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)

	// ditto with deleting
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/projects/%d/consent/responses/%d", project.ID, user1Response.ID), b, routeDeleteConsentResponse, user2.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/projects/%d/consent/responses/%d", project.ID, user1Response.ID), b, routeDeleteConsentResponse, user1.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// delete the form
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/projects/%d/consent", project.ID), b, routeDeleteConsentForm, admin.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)

}

func (suite SuiteTestsConsentRoutes) TestConsentAdminBadRoutes() {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.Encode(map[string]string{})
	admin := &User{
		SystemRole: UserSystemRoleAdmin,
	}
	err := createTestUser(admin)
	suite.Nil(err)
	defer DeleteUser(admin.ID)

	code, res, err := testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent", -1), b, routeSaveConsentForm, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%s/consent", "a"), b, routeSaveConsentForm, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent", -1), b, routeGetConsentForm, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%s/consent", "a"), b, routeGetConsentForm, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/projects/%d/consent", -1), b, routeDeleteConsentForm, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/projects/%s/consent", "a"), b, routeDeleteConsentForm, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)

	// responses
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%d/consent/responses", -1), b, routeCreateConsentResponse, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodPost, fmt.Sprintf("/projects/%s/consent/responses", "a"), b, routeCreateConsentResponse, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent/responses", -1), b, routeGetConsentResponses, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%s/consent/responses", "a"), b, routeGetConsentResponses, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%d/consent/responses/%d", -1, -1), b, routeGetConsentResponse, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodGet, fmt.Sprintf("/projects/%s/consent/responses/%s", "a", "b"), b, routeGetConsentResponse, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/projects/%d/consent/responses/%d", -1, -1), b, routeDeleteConsentResponse, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	code, res, err = testEndpoint(http.MethodDelete, fmt.Sprintf("/projects/%s/consent/responses/%s", "a", "b"), b, routeDeleteConsentResponse, admin.Access)
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)
}
