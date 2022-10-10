package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
)

// apiReturn is the primary return shape for JSON-based returns; depending if it's an error or a success, the
// actual JSON fields may differ
type apiReturn struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
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

// apiErrorHelper is a helper to find the correct api error data
func apiErrorHelper(key string) apiError {
	if found, foundOK := apiErrors[key]; foundOK {
		return found
	}
	return apiError{
		Code:    http.StatusInternalServerError,
		Message: "unknown error",
	}
}

// apiError holds an API error dataset
type apiError struct {
	Code    int
	Message string
}

const (
	api_error_not_implemented = "api_error_not_implemented"
)

// apiErrors is a mapping of keys to data
var apiErrors = map[string]apiError{
	// general
	api_error_not_implemented: {
		Code:    http.StatusNotImplemented,
		Message: "route not implemented",
	},

	// auth
}

func testEndpoint(method string, endpoint string, data io.Reader, hendler http.HandlerFunc, accessToken string) (code int, body *bytes.Buffer, err error) {
	req, err := http.NewRequest(method, endpoint, data)
	if err != nil {
		return http.StatusInternalServerError, nil, err
	}

	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("Authorization", "Beader: "+accessToken)
	rr := httptest.NewRecorder()
	chi := SetupAPI()
	chi.ServeHTTP(rr, req)
	return rr.Code, rr.Body, nil
}
