package api

// we need to put these in a suite to setup a site

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/suite"
)

type SuiteTestsUserRoutes struct {
	suite.Suite
}

func TestSuiteTestsUserRoutes(t *testing.T) {
	suite.Run(t, new(SuiteTestsUserRoutes))
}

func (suite *SuiteTestsUserRoutes) SetupSuite() {
	setupTesting()
}

func (suite SuiteTestsUserRoutes) TestUserAuthRoutes() {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.Encode(map[string]string{})
	require := suite.Require()

	// since we don't have sign ups yet, we will just create a
	// user and go from there
	plainPassword := "test_P@ssword!"
	updatedPlainPassword := "test_updateP@$$W0Rd!"
	user := &User{
		FirstName:  "Admin",
		LastName:   "Admin",
		Status:     "active",
		Email:      "admin_test@kesplora.com",
		SystemRole: UserSystemRoleAdmin,
		Password:   plainPassword,
	}
	err := CreateUser(user)
	require.Nil(err)
	defer DeleteUser(user.ID)

	code, res, err := testEndpoint(http.MethodPost, "/login", b, routeUserLogin, "")
	suite.Nil(err)
	suite.Equal(http.StatusBadRequest, code, res)

	b.Reset()
	encoder.Encode(&map[string]string{
		"login":    user.Email,
		"password": plainPassword,
	})
	code, res, err = testEndpoint(http.MethodPost, "/login", b, routeUserLogin, "")
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err := testEndpointResultToMap(res)
	suite.Nil(err)
	loggedIn := &User{}
	mapstructure.Decode(m, loggedIn)
	suite.Equal(user.ID, loggedIn.ID)
	suite.Equal(user.FirstName, loggedIn.FirstName)
	suite.Equal(user.LastName, loggedIn.LastName)
	suite.Equal(user.Email, loggedIn.Email)
	suite.Equal(user.SystemRole, loggedIn.SystemRole)

	// get the profile
	code, res, err = testEndpoint(http.MethodGet, "/me", b, routeGetUserProfile, loggedIn.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	profile := &User{}
	mapstructure.Decode(m, profile)
	suite.Equal(user.ID, profile.ID)
	suite.Equal(user.FirstName, profile.FirstName)
	suite.Equal(user.LastName, profile.LastName)
	suite.Equal(user.Email, profile.Email)
	suite.Equal(user.SystemRole, profile.SystemRole)

	// update the profile
	updated := &User{
		Title:     "Dr.",
		FirstName: "Updated",
		LastName:  "Updated",
		Pronouns:  "they/them",
		Email:     "updated_test@kesplora.com",
		Password:  updatedPlainPassword,
	}
	b.Reset()
	encoder.Encode(updated)
	code, res, err = testEndpoint(http.MethodPatch, "/me", b, routeUpdateUserProfile, loggedIn.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)

	// get it again
	code, res, err = testEndpoint(http.MethodGet, "/me", b, routeGetUserProfile, loggedIn.Access)
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	updatedProfile := &User{}
	mapstructure.Decode(m, updatedProfile)
	suite.Equal(user.ID, updatedProfile.ID)
	suite.Equal(updated.FirstName, updatedProfile.FirstName)
	suite.Equal(updated.LastName, updatedProfile.LastName)
	suite.Equal(updated.Email, updatedProfile.Email)
	suite.Equal(user.SystemRole, updatedProfile.SystemRole)
	suite.Equal(updated.Title, updatedProfile.Title)
	suite.Equal(updated.Pronouns, updatedProfile.Pronouns)

	b.Reset()
	encoder.Encode(&map[string]string{
		"login":    user.Email,
		"password": plainPassword,
	})
	code, res, err = testEndpoint(http.MethodPost, "/login", b, routeUserLogin, "")
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
	b.Reset()
	encoder.Encode(&map[string]string{
		"login":    updated.Email,
		"password": updatedPlainPassword,
	})
	code, res, err = testEndpoint(http.MethodPost, "/login", b, routeUserLogin, "")
	suite.Nil(err)
	suite.Equal(http.StatusForbidden, code, res)
}

func (suite SuiteTestsUserRoutes) TestUserAuthRefresh() {
	b := new(bytes.Buffer)
	encoder := json.NewEncoder(b)
	encoder.Encode(map[string]string{})
	require := suite.Require()

	plainPassword := "test_P@ssword!"
	user := &User{
		FirstName:  "Admin",
		LastName:   "Admin",
		Status:     "active",
		Email:      "admin_test@kesplora.com",
		SystemRole: UserSystemRoleAdmin,
		Password:   plainPassword,
	}
	err := CreateUser(user)
	require.Nil(err)
	defer DeleteUser(user.ID)

	b.Reset()
	encoder.Encode(&map[string]string{
		"login":    user.Email,
		"password": plainPassword,
	})
	code, res, err := testEndpoint(http.MethodPost, "/login", b, routeUserLogin, "")
	suite.Nil(err)
	suite.Equal(http.StatusOK, code, res)
	found := &User{}
	m, err := testEndpointResultToMap(res)
	suite.Nil(err)
	err = mapstructure.Decode(m, found)
	suite.Nil(err)
	suite.NotEqual("", found.Access)
	suite.NotEqual("", found.Refresh)
	suite.NotEqual("", found.Expires)

	// check the db for the refresh token
	foundToken, err := getTokenForUser(found.ID, tokenTypeRefresh)
	suite.Nil(err)
	require.NotNil(foundToken)
	require.NotEqual("", foundToken.Token)

	// we need to sleep 1 second to allow the expires to change
	time.Sleep(1 * time.Second)

	// since we don't have cookies in our test suite, we just send up the body
	refreshInput := &refreshTokenInput{
		Refresh: foundToken.Token,
	}
	b.Reset()
	encoder.Encode(refreshInput)

	code, res, err = testEndpoint(http.MethodPost, "/me/refresh", b, routeUserRefreshAccess, "")
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)
	found2 := &User{}
	m, err = testEndpointResultToMap(res)
	suite.Nil(err)
	err = mapstructure.Decode(m, found2)
	suite.Nil(err)
	suite.NotEqual(found.Access, found2.Access)
	suite.Equal(found.Refresh, found2.Refresh)
	suite.NotEqual(found.Expires, found2.Expires)

	// logout; that should delete the refresh token so refresh should fail
	code, res, err = testEndpoint(http.MethodPost, "/logout", b, routeUserRefreshAccess, found.Access)
	suite.Nil(err)
	require.Equal(http.StatusOK, code, res)

	b.Reset()
	encoder.Encode(refreshInput)
	code, res, err = testEndpoint(http.MethodPost, "/me/refresh", b, routeUserRefreshAccess, "")
	suite.Nil(err)
	suite.Equal(http.StatusUnauthorized, code, res)
}
