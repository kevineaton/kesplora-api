package api

import (
	"net/http"

	"github.com/go-chi/render"
)

type loginInput struct {
	ParticipantCode string `json:"participantCode"`
	Email           string `json:"email"`
	Password        string `json:"password"`
}

// routeUserLogin attempts to login a user
func routeUserLogin(w http.ResponseWriter, r *http.Request) {
	input := loginInput{}
	render.Bind(r, &input)
	if input.Email == "" && input.ParticipantCode == "" {
		sendAPIError(w, api_error_user_bad_data, nil, map[string]string{})
		return
	}
	if input.Password == "" {
		sendAPIError(w, api_error_user_bad_data, nil, map[string]string{})
		return
	}

	// we break this here in case we want to separate it later
	user := &User{}
	var err error
	if input.Email != "" {
		user, err = AttemptLoginForUser(input.Email, input.Password)
	} else {
		user, err = AttemptLoginForUser(input.ParticipantCode, input.Password)
	}
	if err != nil || user == nil || user.ID == 0 {
		sendAPIError(w, api_error_user_bad_login, nil, map[string]string{})
		return
	}

	// generate the tokens and cookies
	accessCookie, refreshCookie, err := userGenerateTokens(user)
	if err == nil {
		http.SetCookie(w, accessCookie)
		http.SetCookie(w, refreshCookie)
	}

	sendAPIJSONData(w, http.StatusOK, user)
}

// routeGetUserProfile gets a user's profile based upon their JWT
func routeGetUserProfile(w http.ResponseWriter, r *http.Request) {
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

// routeUpdateUserProfile updates a profile based upon their JWT
func routeUpdateUserProfile(w http.ResponseWriter, r *http.Request) {
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

// Bind binds the data for the HTTP
func (data *loginInput) Bind(r *http.Request) error {
	return nil
}
