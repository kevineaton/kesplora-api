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
	api_error_auth_must_participant = "api_error_auth_must_participant"
	api_error_auth_must_user        = "api_error_auth_must_user"

	// config
	api_error_config_missing_data = "api_error_config_missing_data"
	api_error_config_invalid_code = "api_error_config_invalid_code"

	// sites
	api_error_site_not_active = "api_error_site_not_active"
	api_error_site_get_error  = "api_error_site_get_error"
	api_error_site_save_error = "api_error_site_save_error"

	// user errors
	api_error_user_not_found   = "api_error_user_not_found"
	api_error_user_general     = "api_error_user_general"
	api_error_user_cannot_save = "api_error_cannot_save"
	api_error_user_bad_data    = "api_error_user_bad_data"
	api_error_user_bad_login   = "api_error_user_bad_login"
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

	// config
	api_error_config_missing_data: {
		Code:    http.StatusBadRequest,
		Message: "the following fields are required: code, description, name, shortName, siteTechnicalContact, firstName, lastName, email, password",
	},
	api_error_config_invalid_code: {
		Code:    http.StatusBadRequest,
		Message: "that was not a valid code",
	},

	// sites
	api_error_site_not_active: {
		Code:    http.StatusForbidden,
		Message: "site is not active",
	},
	api_error_site_get_error: {
		Code:    http.StatusBadRequest,
		Message: "site is not configured or site cannot be retrieved",
	},
	api_error_site_save_error: {
		Code:    http.StatusBadRequest,
		Message: "site cannot be updated",
	},

	// user
	api_error_user_not_found: {
		Code:    http.StatusForbidden,
		Message: "user missing",
	},
	api_error_user_general: {
		Code:    http.StatusBadRequest,
		Message: "user error",
	},
	api_error_user_cannot_save: {
		Code:    http.StatusBadRequest,
		Message: "cannot save the user",
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
