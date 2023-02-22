package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/render"
)

type loginInput struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type refreshTokenInput struct {
	Refresh string `json:"refresh"`
}

// routeAllUserLogin attempts to login a user
func routeAllUserLogin(w http.ResponseWriter, r *http.Request) {
	input := loginInput{}
	render.Bind(r, &input)
	if input.Login == "" || input.Password == "" {
		sendAPIError(w, api_error_user_bad_data, nil, map[string]string{
			"login": input.Login,
		})
		return
	}

	// we break this here in case we want to separate it later
	user, err := AttemptLoginForUser(input.Login, input.Password)
	if err != nil || user == nil || user.ID == 0 {
		sendAPIError(w, api_error_user_bad_login, nil, map[string]string{})
		return
	}

	// generate the tokens
	accessToken, accessExpires, refreshToken, err := userGenerateTokens(user, true)
	if err != nil {
		sendAPIError(w, api_error_user_bad_login, err, map[string]string{})
		return
	}

	// now generate the cookies
	accessCookie, refreshCookie := generateCookies(accessToken.Token, refreshToken.Token)
	if err == nil {
		http.SetCookie(w, accessCookie)
		http.SetCookie(w, refreshCookie)
	}
	user.Access = accessToken.Token
	user.Expires = accessExpires
	user.Refresh = refreshToken.Token

	sendAPIJSONData(w, http.StatusOK, user)
}

// routeAllGetUserProfile gets a user's profile based upon their JWT
func routeAllGetUserProfile(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}

	user, err := GetUserByID(results.User.ID)
	if err != nil {
		sendAPIError(w, api_error_user_not_found, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, user)
}

// routeAllUpdateUserProfile updates a profile based upon their JWT
func routeAllUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}
	user, err := GetUserByID(results.User.ID)
	if err != nil {
		sendAPIError(w, api_error_user_not_found, err, map[string]string{})
		return
	}

	input := &User{}
	render.Bind(r, input)

	if input.Title != "" {
		user.Title = input.Title
	}
	if input.FirstName != "" {
		user.FirstName = input.FirstName
	}
	if input.LastName != "" {
		user.LastName = input.LastName
	}
	if input.Email != "" {
		user.Email = input.Email
	}
	if input.Pronouns != "" {
		user.Pronouns = input.Pronouns
	}
	if input.DateOfBirth != "" {
		user.DateOfBirth = input.DateOfBirth
	}
	err = UpdateUser(user)
	if err != nil {
		sendAPIError(w, api_error_user_general, err, map[string]string{})
	}
	sendAPIJSONData(w, http.StatusOK, user)
}

// routeAllUserRefreshAccess is a bit of a bear, but handles refreshing the access token for the user
func routeAllUserRefreshAccess(w http.ResponseWriter, r *http.Request) {
	refreshToken := ""
	refreshCookie, err := r.Cookie(tokenTypeRefresh)
	if err == nil && refreshCookie != nil {
		refreshToken = refreshCookie.Value
	}
	if refreshToken == "" {
		// maybe it was posted through an integration of some sort?
		input := &refreshTokenInput{}
		render.Bind(r, input)
		refreshToken = input.Refresh
	}

	if refreshToken == "" {
		sendAPIError(w, api_error_auth_missing, errors.New("missing refresh"), map[string]string{})
		return
	}

	// split; the pattern is ker_id^rand1_rand2, so we split on ^, then on _ in the first
	parts := strings.Split(refreshToken, "^")
	if len(parts) < 2 {
		sendAPIError(w, api_error_auth_malformed, errors.New("auth malformed"), map[string]string{})
		return
	}
	subParts := strings.Split(parts[0], "_")
	if len(subParts) != 2 || subParts[0] != "ker" {
		sendAPIError(w, api_error_auth_malformed, errors.New("auth malformed"), map[string]string{})
		return
	}
	userID, err := strconv.ParseInt(subParts[1], 10, 64)
	if err != nil {
		sendAPIError(w, api_error_auth_malformed, err, map[string]string{})
		return
	}

	// get the token in the DB and make sure they match
	foundToken, err := getTokenForUser(userID, tokenTypeRefresh)
	if err != nil {
		sendAPIError(w, api_error_auth_missing, err, map[string]string{})
		return
	}
	if foundToken.Token != refreshToken {
		sendAPIError(w, api_error_auth_missing, errors.New("missing refresh"), map[string]string{})
		return
	}
	parsedExpires, err := parseTime(foundToken.ExpiresOn)
	if err != nil {
		sendAPIError(w, api_error_auth_malformed, err, map[string]string{})
		return
	}
	if parsedExpires.Before(time.Now()) {
		sendAPIError(w, api_error_auth_expired, fmt.Errorf("expired at %s", parsedExpires.Format(timeFormatAPI)), map[string]string{})
		return
	}

	// get the user, make sure their account is still valid
	foundUser, err := GetUserByID(userID)
	if err != nil {
		sendAPIError(w, api_error_user_not_found, err, map[string]string{})
		return
	}
	if foundUser.Status != UserStatusActive {
		sendAPIError(w, api_error_user_not_found, err, map[string]string{})
		return
	}

	// generate new access tokens and extend the refresh token expires
	foundToken.ExpiresOn = time.Now().Add(tokenExpiresMinutesRefresh * time.Minute).Format(timeFormatDB)
	err = saveTokenForUser(foundToken)
	if err != nil {
		sendAPIError(w, api_error_auth_save_err, err, map[string]string{})
		return
	}

	accessToken, accessExpires, _, err := userGenerateTokens(foundUser, false)
	if err != nil {
		sendAPIError(w, api_error_user_bad_login, err, map[string]string{})
		return
	}

	// now generate the cookies
	accessCookie, refreshCookie := generateCookies(accessToken.Token, foundToken.Token)
	if err == nil {
		http.SetCookie(w, accessCookie)
		http.SetCookie(w, refreshCookie)
	}
	foundUser.Access = accessToken.Token
	foundUser.Expires = accessExpires
	foundUser.Refresh = refreshToken
	sendAPIJSONData(w, http.StatusOK, foundUser)
}

// routeAllUserLogout logs out a user
func routeAllUserLogout(w http.ResponseWriter, r *http.Request) {
	results := checkRoutePermissions(w, r, &routePermissionsCheckOptions{
		ShouldSendError: true,
	})
	if !results.IsValid {
		return
	}
	err := deleteTokenForUser(results.User.ID, tokenTypeRefresh)
	if err != nil {
		sendAPIError(w, api_error_user_bad_logout, err, map[string]string{})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"loggedIn": false,
	})
}

// Bind binds the data for the HTTP
func (data *loginInput) Bind(r *http.Request) error {
	return nil
}

// Bind binds the data for the HTTP
func (data *refreshTokenInput) Bind(r *http.Request) error {
	return nil
}
