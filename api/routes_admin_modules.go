package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// routeAdminCreateModule creates a new module
func routeAdminCreateModule(w http.ResponseWriter, r *http.Request) {
	input := &Module{}
	render.Bind(r, input)
	if input.Name == "" {
		sendAPIError(w, api_error_module_missing_data, errors.New(api_error_module_missing_data), map[string]interface{}{
			"input": input,
		})
		return
	}

	err := CreateModule(input)
	if err != nil {
		sendAPIError(w, api_error_module_save_error, err, map[string]interface{}{
			"input": input,
			"error": err.Error(),
		})
		return
	}
	sendAPIJSONData(w, http.StatusCreated, input)
}

// routeAdminGetAllSiteModules gets all of the modules created on a site
func routeAdminGetAllSiteModules(w http.ResponseWriter, r *http.Request) {
	mods, err := GetAllModulesForSite()
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	sendAPIJSONData(w, http.StatusOK, mods)
}

// routeAdminGetModuleByID gets a single module
func routeAdminGetModuleByID(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	if moduleErr != nil {
		sendAPIError(w, api_error_invalid_path, moduleErr, map[string]string{})
		return
	}
	found, err := GetModuleByID(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{
			"moduleID": moduleID,
		})
		return
	}

	sendAPIJSONData(w, http.StatusOK, found)
}

// routeAdminUpdateModule updates a module
func routeAdminUpdateModule(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	if moduleErr != nil {
		sendAPIError(w, api_error_invalid_path, moduleErr, map[string]string{})
		return
	}

	found, err := GetModuleByID(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{
			"moduleID": moduleID,
		})
		return
	}

	input := &Module{}
	render.Bind(r, input)
	if input.Name != "" && input.Name != found.Name {
		found.Name = input.Name
	}
	if input.Description != found.Description {
		found.Description = input.Description
	}
	if input.Status != "" {
		found.Status = input.Status
	}

	err = UpdateModule(found)
	if err != nil {
		sendAPIError(w, api_error_module_save_error, err, map[string]interface{}{
			"input": input,
			"error": err.Error(),
		})
		return
	}
	sendAPIJSONData(w, http.StatusOK, found)

}

// routeAdminDeleteModule deletes a module and removes it from all flows
func routeAdminDeleteModule(w http.ResponseWriter, r *http.Request) {
	moduleID, moduleErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	if moduleErr != nil {
		sendAPIError(w, api_error_invalid_path, moduleErr, map[string]string{})
		return
	}

	err := DeleteModule(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{
			"moduleID": moduleID,
		})
		return
	}
	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
}

// routeAdminGetModulesOnProject gets all the modules on the platform
func routeAdminGetModulesOnProject(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, projectIDErr, map[string]string{})
		return
	}

	modules, err := GetModulesForProject(projectID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{
			"projectID": projectID,
			"error":     err.Error(),
		})
		return
	}
	// TODO: if not an admin, remove any non-active modules
	sendAPIJSONData(w, http.StatusOK, modules)
}

// routeAdminLinkModuleAndProject links a module and a project in a specific order
func routeAdminLinkModuleAndProject(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	order, orderErr := strconv.ParseInt(chi.URLParam(r, "order"), 10, 64)
	if projectIDErr != nil || moduleIDErr != nil || orderErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	_, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]interface{}{
			"projectID": projectID,
			"moduleID":  moduleID,
		})
		return
	}
	_, err = GetModuleByID(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{
			"projectID": projectID,
			"moduleID":  moduleID,
		})
		return
	}

	err = LinkModuleAndProject(projectID, moduleID, order)
	if err != nil {
		sendAPIError(w, api_error_module_link_err, err, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": true,
	})
}

// routeAdminUnlinkModuleAndProject removes a module from a project
func routeAdminUnlinkModuleAndProject(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	moduleID, moduleIDErr := strconv.ParseInt(chi.URLParam(r, "moduleID"), 10, 64)
	if projectIDErr != nil || moduleIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	_, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]interface{}{
			"projectID": projectID,
			"moduleID":  moduleID,
		})
		return
	}
	_, err = GetModuleByID(moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_not_found, err, map[string]interface{}{
			"projectID": projectID,
			"moduleID":  moduleID,
		})
		return
	}

	err = UnlinkModuleAndProject(projectID, moduleID)
	if err != nil {
		sendAPIError(w, api_error_module_unlink_err, err, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": false,
	})
}

// routeAdminUnlinkAllModulesFromProject removes all modules from a project
func routeAdminUnlinkAllModulesFromProject(w http.ResponseWriter, r *http.Request) {
	projectID, projectIDErr := strconv.ParseInt(chi.URLParam(r, "projectID"), 10, 64)
	if projectIDErr != nil {
		sendAPIError(w, api_error_invalid_path, errors.New("invalid path"), map[string]string{})
		return
	}

	_, err := GetProjectByID(projectID)
	if err != nil {
		sendAPIError(w, api_error_project_not_found, err, map[string]interface{}{
			"projectID": projectID,
		})
		return
	}

	err = UnlinkAllModulesFromProject(projectID)
	if err != nil {
		sendAPIError(w, api_error_module_unlink_err, err, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"linked": false,
	})
}
