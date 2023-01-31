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

	// general
	api_error_invalid_path = "api_error_invalid_path"

	// auth
	api_error_auth_missing          = "api_error_auth_missing"
	api_error_auth_expired          = "api_error_auth_expired"
	api_error_auth_save_err         = "api_error_auth_expired_save_err"
	api_error_auth_malformed        = "api_error_auth_malformed"
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
	api_error_user_bad_logout  = "api_error_user_bad_logout"

	// project errors
	api_error_project_missing_data       = "api_error_project_missing_data"
	api_error_project_save_error         = "api_error_project_save_error"
	api_error_project_no_projects_found  = "api_error_project_no_projects_found"
	api_error_project_not_found          = "api_error_project_not_found"
	api_error_project_link_err           = "api_error_project_link_err"
	api_error_project_unlink_err         = "api_error_project_unlink_err"
	api_error_project_signup_unavailable = "api_error_project_signup_unavailable"

	// consent form errors
	api_error_consent_save_err                      = "api_error_consent_save_err"
	api_error_consent_save_participants_not_zero    = "api_error_consent_save_participants_not_zero"
	api_error_consent_not_found                     = "api_error_consent_not_found"
	api_error_consent_delete_participants_not_zero  = "api_error_consent_delete_participants_not_zero"
	api_error_consent_delete_err                    = "api_error_consent_delete_err"
	api_error_consent_response_project_closed       = "api_error_consent_response_project_closed"
	api_error_consent_response_code_err             = "api_error_consent_response_code_err"
	api_error_consent_response_max_reached          = "api_error_consent_response_max_reached"
	api_error_consent_response_not_min_age          = "api_error_consent_response_not_min_age"
	api_error_consent_response_participant_save_err = "api_error_consent_response_participant_save_err"
	api_error_consent_response_save_err             = "api_error_consent_response_save_err"
	api_error_consent_response_get_err              = "api_error_consent_response_get_err"

	// module errors
	api_error_module_missing_data = "api_error_module_missing_data"
	api_error_module_save_error   = "api_error_module_save_error"
	api_error_module_not_found    = "api_error_module_not_found"
	api_error_module_link_err     = "api_error_module_link_err"
	api_error_module_unlink_err   = "api_error_module_unlink_err"
	api_error_module_delete_err   = "api_error_module_delete_err"

	// blocks
	api_error_block_missing_data         = "api_error_block_missing_data"
	api_error_block_content_missing_data = "api_error_block_content_missing_data"
	api_error_block_save_error           = "api_error_block_save_error"
	api_error_block_invalid_type         = "api_error_block_invalid_type"
	api_error_block_not_found            = "api_error_block_not_found"
	api_error_block_content_not_found    = "api_error_block_content_not_found"
	api_error_block_link_err             = "api_error_block_link_err"
	api_error_block_unlink_err           = "api_error_block_unlink_err"
	api_error_block_delete_err           = "api_error_block_delete_err"

	api_error_block_status_save_err = "api_error_block_status_save_err"
)

// apiErrors is a mapping of keys to data
var apiErrors = map[string]apiError{
	// general
	api_error_not_implemented: {
		Code:    http.StatusNotImplemented,
		Message: "route not implemented",
	},
	api_error_invalid_path: {
		Code:    http.StatusBadRequest,
		Message: "invalid path",
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
	api_error_auth_malformed: {
		Code:    http.StatusUnauthorized,
		Message: "authorization malformed",
	},
	api_error_auth_save_err: {
		Code:    http.StatusUnauthorized,
		Message: "authorization could not be saved",
	},
	api_error_auth_must_admin: {
		Code:    http.StatusForbidden,
		Message: "must be an admin",
	},
	api_error_auth_must_participant: {
		Code:    http.StatusForbidden,
		Message: "must be a participant",
	},
	api_error_auth_must_user: {
		Code:    http.StatusForbidden,
		Message: "must be a user",
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
	api_error_user_bad_logout: {
		Code:    http.StatusBadRequest,
		Message: "user logout failed",
	},

	// projects
	api_error_project_missing_data: {
		Code:    http.StatusBadRequest,
		Message: "missing data",
	},
	api_error_project_save_error: {
		Code:    http.StatusBadRequest,
		Message: "could not save that project",
	},
	api_error_project_no_projects_found: {
		Code:    http.StatusForbidden,
		Message: "no projects available",
	},
	api_error_project_not_found: {
		Code:    http.StatusForbidden,
		Message: "project not found or unavailable to user",
	},
	api_error_project_link_err: {
		Code:    http.StatusBadRequest,
		Message: "could not link that user and project",
	},
	api_error_project_unlink_err: {
		Code:    http.StatusBadRequest,
		Message: "could not unlink that user and project",
	},
	api_error_project_signup_unavailable: {
		Code:    http.StatusBadRequest,
		Message: "sign up unavailable",
	},

	// consent and responses
	api_error_consent_save_err: {
		Code:    http.StatusBadRequest,
		Message: "could not save",
	},
	api_error_consent_save_participants_not_zero: {
		Code:    http.StatusForbidden,
		Message: "project already underway",
	},
	api_error_consent_not_found: {
		Code:    http.StatusBadRequest,
		Message: "could not find",
	},
	api_error_consent_delete_participants_not_zero: {
		Code:    http.StatusForbidden,
		Message: "project already underway",
	},
	api_error_consent_delete_err: {
		Code:    http.StatusBadRequest,
		Message: "could not delete",
	},
	api_error_consent_response_project_closed: {
		Code:    http.StatusForbidden,
		Message: "project is closed",
	},
	api_error_consent_response_code_err: {
		Code:    http.StatusForbidden,
		Message: "code does not match",
	},
	api_error_consent_response_max_reached: {
		Code:    http.StatusForbidden,
		Message: "max participants reached",
	},
	api_error_consent_response_not_min_age: {
		Code:    http.StatusForbidden,
		Message: "minimum age not met",
	},
	api_error_consent_response_participant_save_err: {
		Code:    http.StatusBadRequest,
		Message: "participant creation failed, check your input",
	},
	api_error_consent_response_save_err: {
		Code:    http.StatusBadRequest,
		Message: "could not save",
	},
	api_error_consent_response_get_err: {
		Code:    http.StatusForbidden,
		Message: "could not find",
	},

	// modules
	api_error_module_missing_data: {
		Code:    http.StatusBadRequest,
		Message: "missing data",
	},
	api_error_module_save_error: {
		Code:    http.StatusBadRequest,
		Message: "could not save",
	},
	api_error_module_not_found: {
		Code:    http.StatusNotFound,
		Message: "module not found",
	},
	api_error_module_link_err: {
		Code:    http.StatusBadRequest,
		Message: "could not link that module and project",
	},
	api_error_module_unlink_err: {
		Code:    http.StatusBadRequest,
		Message: "could not unlink that module and project",
	},
	api_error_module_delete_err: {
		Code:    http.StatusBadRequest,
		Message: "could not delete the module",
	},

	// blocks
	api_error_block_missing_data: {
		Code:    http.StatusBadRequest,
		Message: "missing data",
	},
	api_error_block_content_missing_data: {
		Code:    http.StatusBadRequest,
		Message: "missing content data",
	},
	api_error_block_save_error: {
		Code:    http.StatusBadRequest,
		Message: "could not save",
	},
	api_error_block_invalid_type: {
		Code:    http.StatusBadRequest,
		Message: "that is not a valid block type",
	},
	api_error_block_not_found: {
		Code:    http.StatusNotFound,
		Message: "block not found",
	},
	api_error_block_content_not_found: {
		Code:    http.StatusNotFound,
		Message: "block content not found",
	},
	api_error_block_link_err: {
		Code:    http.StatusBadRequest,
		Message: "could not link that module and block",
	},
	api_error_block_unlink_err: {
		Code:    http.StatusBadRequest,
		Message: "could not unlink that module and block",
	},
	api_error_block_delete_err: {
		Code:    http.StatusBadRequest,
		Message: "could not delete the block",
	},

	// flow and block statuses
	api_error_block_status_save_err: {
		Code:    http.StatusBadRequest,
		Message: "could not save that status",
	},
}
