package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
)

// avoid collisions with other keys that may enter the context
type key string

// appContextKeyUser is used for the context to find the user by access token
const appContextKeyUser key = "user"

// appContextKeyFound is used to make it easy to determine if the user was found
const appContextKeyFound key = "found"

// AppContextKeyExpired is the expired key
const appContextKeyExpired key = "expired"

// apiReturn is the primary return shape for JSON-based returns; depending if it's an error or a success, the
// actual JSON fields may differ
type apiReturn struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type routePermissionsCheckOptions struct {
	MustBeAdmin       bool
	MustBeParticipant bool
	MustBeUser        int64 // the user id to compare against
	ShouldSendError   bool
}

type routePermissionsCheckResults struct {
	IsExpired bool
	IsValid   bool
	User      *jwtUser
}

// sendAPIJSONData sends a JSON object for a successful API call
func sendAPIJSONData(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(apiReturn{
		Data: payload,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// sendAPIError sends a JSON object for an API error
func sendAPIError(w http.ResponseWriter, key string, data interface{}) {
	apiErrorData := apiErrorHelper(key)
	response, _ := json.Marshal(apiReturn{
		Data:    data,
		Error:   key,
		Message: apiErrorData.Message,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErrorData.Code)
	w.Write(response)
}

func checkRoutePermissions(w http.ResponseWriter, r *http.Request, options *routePermissionsCheckOptions) *routePermissionsCheckResults {
	// check if the access token is present
	found := r.Context().Value(appContextKeyFound).(bool)
	user, userOK := r.Context().Value(appContextKeyUser).(jwtUser)
	results := &routePermissionsCheckResults{}

	if !found || !userOK {
		if options.ShouldSendError {
			results.IsValid = false
			sendAPIError(w, api_error_auth_missing, map[string]string{})
		}
		return results
	}

	// see if it is expired
	expired, expiredOK := r.Context().Value(appContextKeyExpired).(bool)
	if expiredOK && expired {
		results.IsValid = false
		results.IsExpired = true
		if options.ShouldSendError {
			sendAPIError(w, api_error_auth_expired, map[string]string{})
		}
		return results
	}

	results.User = &user

	// check if an admin
	if options.MustBeAdmin && user.SystemRole != UserSystemRoleAdmin {
		results.IsValid = false
		if options.ShouldSendError {
			sendAPIError(w, api_error_auth_must_admin, map[string]string{})
		}
		return results
	}

	// check if a participant
	if options.MustBeParticipant && user.SystemRole != UserSystemRoleParticipant {
		results.IsValid = false
		if options.ShouldSendError {
			sendAPIError(w, api_error_auth_must_participant, map[string]string{})
		}
		return results
	}

	// check the user id
	if options.MustBeUser != 0 && options.MustBeUser != user.ID {
		results.IsValid = false
		if options.ShouldSendError {
			sendAPIError(w, api_error_auth_must_user, map[string]string{})
		}
		return results
	}

	results.IsValid = true

	return results
}

func testEndpoint(method string, endpoint string, data io.Reader, handler http.HandlerFunc, accessToken string) (code int, body *bytes.Buffer, err error) {
	req, err := http.NewRequest(method, endpoint, data)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", "Header: "+accessToken)
	rr := httptest.NewRecorder()
	chi := SetupAPI()
	chi.ServeHTTP(rr, req)
	return rr.Code, rr.Body, nil
}
