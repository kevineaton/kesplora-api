package api

import (
	"fmt"
	"math/rand"
	"net/http"
)

const (
	ModuleStatusActive   = "active"
	ModuleStatusPending  = "pending"
	ModuleStatusDisabled = "disabled"
)

// Module is a module that contains Blocks and are organized into Flows for a Project
type Module struct {
	ID            int64  `json:"id" db:"id"`
	Name          string `json:"name" db:"name"`
	Status        string `json:"status" db:"status"`
	Description   string `json:"description" db:"description"`
	FlowOrder     int64  `json:"flowOrder" db:"flowOrder"`         // used in getting for a project
	ProjectsCount int64  `json:"projectsCount" db:"projectsCount"` // used in getting for platform to identify if it's in a project
}

// CreateModule creates a module as a standalone "box"
func CreateModule(input *Module) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO Modules SET
	name = :name,
	status = :status,
	description = :description`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateModule updates a single module
func UpdateModule(input *Module) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE Modules SET
	name = :name,
	status = :status,
	description = :description
	WHERE id = :id`, input)
	return err
}

// DeleteModule removes a module from flows and then deletes the module
func DeleteModule(moduleID int64) error {
	// remove from all flows
	_, err := config.DBConnection.Exec(`DELETE FROM Flows WHERE moduleId = ?`, moduleID)
	if err != nil {
		return err
	}

	_, err = config.DBConnection.Exec(`DELETE FROM BlockModuleFlows WHERE moduleId = ?`, moduleID)
	if err != nil {
		return err
	}

	// delete the module
	_, err = config.DBConnection.Exec(`DELETE FROM Modules WHERE id = ?`, moduleID)
	return err
}

// GetModuleByID gets a single module
func GetModuleByID(moduleID int64) (*Module, error) {
	mod := &Module{}
	defer mod.processForAPI()
	err := config.DBConnection.Get(mod, `SELECT m.*, IFNULL((SELECT COUNT(*) FROM Flows f WHERE f.moduleId = m.id GROUP BY f.projectId), 0) AS projectsCount
	FROM Modules m WHERE m.id = ?`, moduleID)
	return mod, err
}

// GetModulesForProject gets all of the modules for a project
func GetModulesForProject(projectID int64) ([]Module, error) {
	mods := []Module{}
	err := config.DBConnection.Select(&mods, `SELECT m.*, o.flowOrder , IFNULL((SELECT COUNT(*) FROM Flows f WHERE f.moduleId = m.id GROUP BY f.projectId), 0) AS projectsCount
		FROM Modules m, Flows o 
		WHERE o.projectId = ? AND  o.moduleId = m.id ORDER BY o.flowOrder`, projectID)
	if err != nil {
		return mods, err
	}
	for i := range mods {
		mods[i].processForAPI()
	}
	return mods, nil
}

// GetAllModulesForSite gets all the modules on the site, needed for the building of the flows interface
func GetAllModulesForSite() ([]Module, error) {
	mods := []Module{}
	err := config.DBConnection.Select(&mods, `SELECT m.*, IFNULL((SELECT COUNT(*) FROM Flows f WHERE f.moduleId = m.id GROUP BY f.projectId), 0) AS projectsCount
	FROM Modules m 
		ORDER BY m.name`)
	if err != nil {
		return mods, err
	}
	for i := range mods {
		mods[i].processForAPI()
	}
	return mods, nil
}

// LinkModuleAndProject saves a connection between a module and a project
func LinkModuleAndProject(projectID, moduleID, order int64) error {
	_, err := config.DBConnection.Exec(`INSERT INTO Flows
	(projectId, moduleId, flowOrder)
	VALUES 
	(?, ?, ?) ON DUPLICATE KEY UPDATE flowOrder = ?`, projectID, moduleID, order, order)
	return err
}

// UnlinkModuleAndProject removes the module from a project
func UnlinkModuleAndProject(projectID, moduleID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Flows WHERE projectId = ? AND moduleId = ?`, projectID, moduleID)
	return err
}

// UnlinkAllModulesFromProject removes all modules from a project
func UnlinkAllModulesFromProject(projectID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Flows WHERE projectId = ?`, projectID)
	return err
}

// createTestModule is a helper to create a test module
func createTestModule(defaults *Module, projectID int64, moduleOrder int64) error {
	if defaults.Name == "" {
		defaults.Name = fmt.Sprintf("Test Module %d", rand.Intn(9999999))
	}
	if defaults.Description == "" {
		defaults.Description = "# Test Site\n\n *Do Not Use!*\n"
	}
	if defaults.Status == "" {
		defaults.Status = ModuleStatusActive
	}
	err := CreateModule(defaults)
	if err != nil {
		return err
	}
	if projectID != 0 {
		err = LinkModuleAndProject(projectID, defaults.ID, moduleOrder)
		if err != nil {
			return err
		}
	}
	return nil
}

func (input *Module) processForDB() {
	if input.Status == "" {
		input.Status = ModuleStatusPending
	}
}

func (input *Module) processForAPI() {

}

// Bind binds the data for the HTTP
func (data *Module) Bind(r *http.Request) error {
	return nil
}
