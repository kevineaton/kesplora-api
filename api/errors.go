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
	api_error_auth_save             = "api_error_auth_save"
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
	api_error_site_save       = "api_error_site_save"

	// user errors
	api_error_users_site       = "api_error_users_site"
	api_error_users_project    = "api_error_users_project"
	api_error_user_not_found   = "api_error_user_not_found"
	api_error_user_general     = "api_error_user_general"
	api_error_user_cannot_save = "api_error_cannot_save"
	api_error_user_bad_data    = "api_error_user_bad_data"
	api_error_user_bad_login   = "api_error_user_bad_login"
	api_error_user_bad_logout  = "api_error_user_bad_logout"

	// project errors
	api_error_project_missing_data       = "api_error_project_missing_data"
	api_error_project_save               = "api_error_project_save"
	api_error_project_no_projects_found  = "api_error_project_no_projects_found"
	api_error_project_not_found          = "api_error_project_not_found"
	api_error_project_link               = "api_error_project_link"
	api_error_project_unlink             = "api_error_project_unlink"
	api_error_project_signup_unavailable = "api_error_project_signup_unavailable"
	api_error_project_for_user           = "api_error_projects_for_user"
	api_error_project_user_not_in        = "api_error_projects_user_not_in"
	api_error_project_misconfiguration   = "api_error_project_misconfiguration"

	// consent form errors
	api_error_consent_save                         = "api_error_consent_save"
	api_error_consent_save_participants_not_zero   = "api_error_consent_save_participants_not_zero"
	api_error_consent_not_found                    = "api_error_consent_not_found"
	api_error_consent_delete_participants_not_zero = "api_error_consent_delete_participants_not_zero"
	api_error_consent_delete                       = "api_error_consent_delete"
	api_error_consent_response_project_closed      = "api_error_consent_response_project_closed"
	api_error_consent_response_code                = "api_error_consent_response_code"
	api_error_consent_response_max_reached         = "api_error_consent_response_max_reached"
	api_error_consent_response_not_min_age         = "api_error_consent_response_not_min_age"
	api_error_consent_response_participant_save    = "api_error_consent_response_participant_save"
	api_error_consent_response_save                = "api_error_consent_response_save"
	api_error_consent_response_get                 = "api_error_consent_response_get"

	// module errors
	api_error_module_missing_data = "api_error_module_missing_data"
	api_error_module_save         = "api_error_module_save"
	api_error_module_not_found    = "api_error_module_not_found"
	api_error_module_link         = "api_error_module_link"
	api_error_module_unlink       = "api_error_module_unlink"
	api_error_module_delete       = "api_error_module_delete"

	// blocks
	api_error_block_missing_data         = "api_error_block_missing_data"
	api_error_block_content_missing_data = "api_error_block_content_missing_data"
	api_error_block_save                 = "api_error_block_save"
	api_error_block_invalid_type         = "api_error_block_invalid_type"
	api_error_block_not_found            = "api_error_block_not_found"
	api_error_block_content_not_found    = "api_error_block_content_not_found"
	api_error_block_link                 = "api_error_block_link"
	api_error_block_unlink               = "api_error_block_unlink"
	api_error_block_delete               = "api_error_block_delete"

	api_error_block_status_save = "api_error_block_status_save"
	api_error_block_status_form = "api_error_block_status_form"

	api_error_submission_fetch    = "api_error_submission_fetch"
	api_error_submission_mismatch = "api_error_submission_mismatch"
	api_error_submission_missing  = "api_error_submission_missing"
	api_error_submission_create   = "api_error_submission_create"
	api_error_submission_delete   = "api_error_submission_delete"

	// file errors
	api_error_file_upload_no_provider   = "api_error_file_upload_no_provider"
	api_error_file_upload_parse_general = "api_error_file_upload_parse_general"
	api_error_file_upload_parse_form    = "api_error_file_upload_parse_form"
	api_error_file_upload_read          = "api_error_file_upload_read"
	api_error_file_upload_upload        = "api_error_file_upload_upload"
	api_error_file_upload_meta_save     = "api_error_file_upload_meta_save"
	api_error_file_no_exist             = "api_error_file_no_exist"
	api_error_file_download             = "api_error_file_download"
	api_error_file_delete_remote        = "api_error_file_delete_remote"
	api_error_file_delete_meta          = "api_error_file_delete_meta"
	api_error_file_update_meta          = "api_error_file_update_meta"

	// notes errors
	api_error_notes_not_found = "api_error_notes_not_found"
	api_error_notes_delete    = "api_error_notes_delete"
	api_error_notes_save      = "api_error_notes_save"

	// reports errors
	api_error_reports_get = "api_error_reports_get"
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
	api_error_auth_save: {
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
	api_error_site_save: {
		Code:    http.StatusBadRequest,
		Message: "site cannot be updated",
	},

	// user
	api_error_users_site: {
		Code:    http.StatusBadRequest,
		Message: "could not fetch users on platform",
	},
	api_error_users_project: {
		Code:    http.StatusBadRequest,
		Message: "could not fetch users for that project",
	},
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
	api_error_project_save: {
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
	api_error_project_link: {
		Code:    http.StatusBadRequest,
		Message: "could not link that user and project",
	},
	api_error_project_unlink: {
		Code:    http.StatusBadRequest,
		Message: "could not unlink that user and project",
	},
	api_error_project_signup_unavailable: {
		Code:    http.StatusBadRequest,
		Message: "sign up unavailable",
	},
	api_error_project_for_user: {
		Code:    http.StatusBadRequest,
		Message: "could not get projects for that user",
	},
	api_error_project_user_not_in: {
		Code:    http.StatusBadRequest,
		Message: "user is not a participant in that project",
	},
	api_error_project_misconfiguration: {
		Code:    http.StatusBadRequest,
		Message: "the passed in data results in a misconfiguration or is otherwise incorrect",
	},

	// consent and responses
	api_error_consent_save: {
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
	api_error_consent_delete: {
		Code:    http.StatusBadRequest,
		Message: "could not delete",
	},
	api_error_consent_response_project_closed: {
		Code:    http.StatusForbidden,
		Message: "project is closed",
	},
	api_error_consent_response_code: {
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
	api_error_consent_response_participant_save: {
		Code:    http.StatusBadRequest,
		Message: "participant creation failed, check your input",
	},
	api_error_consent_response_save: {
		Code:    http.StatusBadRequest,
		Message: "could not save",
	},
	api_error_consent_response_get: {
		Code:    http.StatusForbidden,
		Message: "could not find",
	},

	// modules
	api_error_module_missing_data: {
		Code:    http.StatusBadRequest,
		Message: "missing data",
	},
	api_error_module_save: {
		Code:    http.StatusBadRequest,
		Message: "could not save",
	},
	api_error_module_not_found: {
		Code:    http.StatusNotFound,
		Message: "module not found",
	},
	api_error_module_link: {
		Code:    http.StatusBadRequest,
		Message: "could not link that module and project",
	},
	api_error_module_unlink: {
		Code:    http.StatusBadRequest,
		Message: "could not unlink that module and project",
	},
	api_error_module_delete: {
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
	api_error_block_save: {
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
	api_error_block_link: {
		Code:    http.StatusBadRequest,
		Message: "could not link that module and block",
	},
	api_error_block_unlink: {
		Code:    http.StatusBadRequest,
		Message: "could not unlink that module and block",
	},
	api_error_block_delete: {
		Code:    http.StatusBadRequest,
		Message: "could not delete the block",
	},

	// flow and block statuses
	api_error_block_status_save: {
		Code:    http.StatusBadRequest,
		Message: "could not save that status",
	},
	api_error_block_status_form: {
		Code:    http.StatusBadRequest,
		Message: "to save a form, you have to call the /submissions path, not the /status path",
	},

	api_error_submission_missing: {
		Code:    http.StatusNotFound,
		Message: "that submission does not exist",
	},
	api_error_submission_mismatch: {
		Code:    http.StatusForbidden,
		Message: "that submission and the passed in data don't match",
	},
	api_error_submission_fetch: {
		Code:    http.StatusBadRequest,
		Message: "could not fetch that submission",
	},
	api_error_submission_create: {
		Code:    http.StatusBadRequest,
		Message: "could not save that submission",
	},
	api_error_submission_delete: {
		Code:    http.StatusBadRequest,
		Message: "could not delete that submission",
	},

	// files
	api_error_file_upload_no_provider: {
		Code:    http.StatusBadRequest,
		Message: "site does not support file uploads",
	},
	api_error_file_upload_parse_general: {
		Code:    http.StatusBadRequest,
		Message: "could not parse the file upload, make sure it is sent as multi-part",
	},
	api_error_file_upload_parse_form: {
		Code:    http.StatusBadRequest,
		Message: "could not parse the form, make sure the key is file",
	},
	api_error_file_upload_read: {
		Code:    http.StatusBadRequest,
		Message: "could not read the data in that file",
	},
	api_error_file_upload_upload: {
		Code:    http.StatusBadRequest,
		Message: "could not upload the file to the remote storage",
	},
	api_error_file_upload_meta_save: {
		Code:    http.StatusBadRequest,
		Message: "could not save the file meta data",
	},
	api_error_file_no_exist: {
		Code:    http.StatusForbidden,
		Message: "file does not exist or you don't have permission",
	},
	api_error_file_download: {
		Code:    http.StatusBadRequest,
		Message: "could not download the file",
	},
	api_error_file_delete_remote: {
		Code:    http.StatusBadRequest,
		Message: "could not delete the file from the provider",
	},
	api_error_file_delete_meta: {
		Code:    http.StatusBadRequest,
		Message: "could not delete the file meta data",
	},
	api_error_file_update_meta: {
		Code:    http.StatusBadRequest,
		Message: "could not update the file metadata",
	},

	// notes
	api_error_notes_not_found: {
		Code:    http.StatusNotFound,
		Message: "note or notes not found",
	},
	api_error_notes_delete: {
		Code:    http.StatusBadRequest,
		Message: "could not delete note",
	},
	api_error_notes_save: {
		Code:    http.StatusBadRequest,
		Message: "could not save that note",
	},

	// reports
	api_error_reports_get: {
		Code:    http.StatusBadRequest,
		Message: "could not fetch that report",
	},
}
