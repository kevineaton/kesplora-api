package api

import "net/http"

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

	// auth
	api_error_auth_missing          = "api_error_auth_missing"
	api_error_auth_expired          = "api_error_auth_expired"
	api_error_auth_must_admin       = "api_error_auth_must_admin"
	api_error_auth_must_particiapnt = "api_error_auth_must_particiapnt"
	api_error_auth_must_user        = "api_error_auth_must_user"

	// user errors
	api_error_user_not_found = "api_error_user_not_found"
	api_error_user_general   = "api_error_user_general"
	api_error_user_bad_data  = "api_error_user_bad_data"
	api_error_user_bad_login = "api_error_user_bad_login"
)

// apiErrors is a mapping of keys to data
var apiErrors = map[string]apiError{
	// general
	api_error_not_implemented: {
		Code:    http.StatusNotImplemented,
		Message: "route not implemented",
	},

	// auth
	api_error_auth_missing: {
		Code:    http.StatusUnauthorized,
		Message: "authorization missing",
	},
	api_error_auth_expired: {
		Code:    419,
		Message: "authorization expired",
	},

	// user
	api_error_user_not_found: {
		Code:    http.StatusForbidden,
		Message: "user missing",
	},
	api_error_user_general: {
		Code:    http.StatusBadRequest,
		Message: "user save error",
	},
	api_error_user_bad_data: {
		Code:    http.StatusBadRequest,
		Message: "user missing data",
	},
	api_error_user_bad_login: {
		Code:    http.StatusForbidden,
		Message: "user login failed",
	},
}
