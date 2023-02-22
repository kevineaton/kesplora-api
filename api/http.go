package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
)

// avoid collisions with other keys that may enter the context
type key string

// appContextKeyUser is used for the context to find the user by access token
const appContextKeyUser key = "user"

// appContextKeyFound is used to make it easy to determine if the user was found
const appContextKeyFound key = "found"

// AppContextKeyExpired is the expired key
const appContextKeyExpired key = "expired"

// appContextSite is the key for the site
const appContextSite key = "site"

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

// sendAPIFileData sends a file's binary data
func sendAPIFileData(w http.ResponseWriter, code int, contentType string, payload []byte) {
	w.WriteHeader(code)
	w.Header().Set("Content-Type", contentType)
	w.Write(payload)
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

	// log it
	Log(LogLevelError, "http_error", fmt.Sprintf("%s := %v", key, systemError.Error()), &LogOptions{
		ExtraData: map[string]interface{}{
			"data":  data,
			"error": systemError.Error(),
		},
	})
}

// checkRoutePermissions is a helper to check the various security permissions for a route
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

// getUserFromHTTPContext is a helper to get the user id from the context; this is
// useful for getting the user after the automated checks have been run but we don't
// want to call the check again
func getUserFromHTTPContext(r *http.Request) (*jwtUser, error) {
	user, userOK := r.Context().Value(appContextKeyUser).(jwtUser)
	if !userOK {
		return nil, errors.New("invalid user")
	}
	return &user, nil
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

// processQuery searches for common query string parameters
type commonQueryParams struct {
	Start     string
	End       string
	Count     int64
	Offset    int64
	SortField string
	SortDir   string
}

func processQuery(r *http.Request) commonQueryParams {
	params := commonQueryParams{}

	startQ := r.URL.Query().Get("start")
	endQ := r.URL.Query().Get("end")
	countQ := r.URL.Query().Get("count")
	offsetQ := r.URL.Query().Get("offset")
	sortDirQ := r.URL.Query().Get("sortDir")
	sortField := r.URL.Query().Get("sort")

	start, err := parseTimeToTimeFormat(startQ, timeFormatAPI)
	if err != nil {
		start = "2017-01-01T00:00:00Z"
	}
	params.Start = start

	end, err := parseTimeToTimeFormat(endQ, timeFormatAPI)
	if err != nil {
		end = "2080-01-01 00:00:00" // TODO: fix in 2079
	}
	params.End = end

	count, err := strconv.ParseInt(countQ, 10, 64)
	if err != nil {
		count = 500 // get a lot
	}
	params.Count = count

	offset, err := strconv.ParseInt(offsetQ, 10, 64)
	if err != nil {
		offset = 0
	}
	params.Offset = offset

	sortDir := strings.ToUpper(sortDirQ)
	if sortDir != "ASC" && sortDir != "DESC" {
		sortDir = "DESC"
	}
	params.SortDir = sortDir
	params.SortField = sortField
	return params
}
