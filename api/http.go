package api

import (
	"bytes"
	"encoding/json"
	"errors"
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
	Key     string      `json:"error,omitempty"`
}

type routePermissionsCheckOptions struct {
	MustBeAdmin       bool
	MustBeParticipant bool
	MustBeUser        int64 // the user id to compare against
	ShouldSendError   bool
}

type routePermissionsCheckResults struct {
	SiteActive bool
	SiteStatus string
	Site       *Site
	IsExpired  bool
	IsValid    bool
	IsAdmin    bool
	User       *jwtUser
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

// sendAPIError sends a JSON object for an API error; note that the systemError is not sent back
// to the client, so it is generally safe for system-level messaging; can be piped to error logs
func sendAPIError(w http.ResponseWriter, key string, systemError error, data interface{}) {
	apiErrorData := apiErrorHelper(key)
	if systemError == nil {
		systemError = errors.New(key) // we may want to have a config flag to output this
	}
	response, _ := json.Marshal(apiReturn{
		Data:    data,
		Key:     key,
		Message: apiErrorData.Message,
	})
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(apiErrorData.Code)
	w.Write(response)
}

func checkRoutePermissions(w http.ResponseWriter, r *http.Request, options *routePermissionsCheckOptions) *routePermissionsCheckResults {
	results := &routePermissionsCheckResults{}
	// first, check if the site is active
	site, err := GetSite()
	if err != nil {
		results.SiteActive = false
		results.SiteStatus = "error"
	} else {
		results.SiteActive = site.Status == SiteStatusActive
		results.SiteStatus = site.Status
	}

	if !results.SiteActive && options.ShouldSendError {
		sendAPIError(w, api_error_site_not_active, errors.New("site not active"), map[string]string{
			"siteStatus": results.SiteStatus,
		})
		return results
	}

	results.Site = site

	// check if the access token is present
	found := r.Context().Value(appContextKeyFound).(bool)
	user, userOK := r.Context().Value(appContextKeyUser).(jwtUser)

	if !found || !userOK {
		if options.ShouldSendError {
			results.IsValid = false
			sendAPIError(w, api_error_auth_missing, errors.New("error"), map[string]string{})
		}
		return results
	}

	// see if it is expired
	expired, expiredOK := r.Context().Value(appContextKeyExpired).(bool)
	if expiredOK && expired {
		results.IsValid = false
		results.IsExpired = true
		if options.ShouldSendError {
			sendAPIError(w, api_error_auth_expired, errors.New("error"), map[string]string{})
		}
		return results
	}

	results.User = &user

	// check if an admin
	if options.MustBeAdmin && user.SystemRole != UserSystemRoleAdmin {
		results.IsValid = false
		if options.ShouldSendError {
			sendAPIError(w, api_error_auth_must_admin, errors.New("error"), map[string]string{})
		}
		return results
	}
	results.IsAdmin = user.SystemRole == UserSystemRoleAdmin

	// check if a participant
	if options.MustBeParticipant && user.SystemRole != UserSystemRoleParticipant {
		results.IsValid = false
		if options.ShouldSendError {
			sendAPIError(w, api_error_auth_must_participant, errors.New("error"), map[string]string{})
		}
		return results
	}

	// check the user id
	if options.MustBeUser != 0 && options.MustBeUser != user.ID {
		results.IsValid = false
		if options.ShouldSendError {
			sendAPIError(w, api_error_auth_must_user, errors.New("error"), map[string]string{})
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
	req.Header.Add("Authorization", "Bearer: "+accessToken)
	rr := httptest.NewRecorder()
	chi := SetupAPI()
	chi.ServeHTTP(rr, req)
	return rr.Code, rr.Body, nil
}

func testEndpointResultToMap(bu *bytes.Buffer) (map[string]interface{}, error) {
	m := map[string]interface{}{}
	err := json.Unmarshal(bu.Bytes(), &m)
	mm := m["data"].(map[string]interface{})
	return mm, err
}

func testEndpointResultToSlice(bu *bytes.Buffer) ([]interface{}, error) {
	m := map[string]interface{}{}
	err := json.Unmarshal(bu.Bytes(), &m)
	mm := m["data"].([]interface{})
	return mm, err
}
