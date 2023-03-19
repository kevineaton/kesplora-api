package api

import (
	"errors"
	"net/http"
	"time"
)

const (
	BlockUserStatusNotStarted = "not_started"
	BlockUserStatusStarted    = "started"
	BlockUserStatusCompleted  = "completed"
)

// Flow represents the flow of a project from a participant's viewpoint; keep in mind this is not a direct
// translation of the Flows table; it needs to support the joins that come from the Flows, Links, and Status tables
type Flow struct {
	UserID            int64  `json:"userId" db:"userId"`
	ProjectID         int64  `json:"projectId" db:"projectId"`
	FlowOrder         int64  `json:"flowOrder" db:"flowOrder"`
	ModuleID          int64  `json:"moduleId" db:"moduleId"`
	ModuleName        string `json:"moduleName" db:"moduleName"`
	ModuleDescription string `json:"moduleDescription" db:"moduleDescription"`
	BlockID           int64  `json:"blockId" db:"blockId"`
	BlockName         string `json:"blockName" db:"blockName"`
	BlockSummary      string `json:"blockSummary" db:"blockSummary"`
	BlockType         string `json:"blockType" db:"blockType"`
	UserStatus        string `json:"userStatus" db:"userStatus"`
	LastUpdatedOn     string `json:"lastUpdatedOn" db:"lastUpdatedOn"`
}

// BlockUserStatus represents a specific block/status entry
type BlockUserStatus struct {
	UserID        int64  `json:"userId" db:"userId"`
	ProjectID     int64  `json:"projectId" db:"projectId"`
	ModuleID      int64  `json:"moduleId" db:"moduleId"`
	BlockID       int64  `json:"blockId" db:"blockId"`
	UserStatus    string `json:"userStatus" db:"userStatus"`
	LastUpdatedOn string `json:"lastUpdatedOn" db:"lastUpdatedOn"`

	// these are needed for the save
	ProjectUserStatus      string `json:"projectUserStatus,omitempty" db:"projectUserStatus"`
	ProjectCompleteMessage string `json:"projectCompleteMessage,omitempty" db:"projectCompleteMessage"`
}

// GetProjectFlowForParticipant gets the entire flow for a project for a participant to lay out the
// flow and status for each section. Note the explicit lack of a module or project status; that can
// be calculated based upon this data.
func GetProjectFlowForParticipant(participantID, projectID int64) ([]Flow, error) {
	flow := []Flow{}
	err := config.DBConnection.Select(&flow, `SELECT f.flowOrder, m.id AS moduleId, m.name AS moduleName, m.description AS moduleDescription, 
	b.id AS blockId, b.name AS blockName, b.summary AS blockSummary, b.blockType AS blockType,
	IFNULL(bus.status, 'not_started') AS userStatus,
	IFNULL(bus.lastUpdatedOn, NOW()) AS lastUpdatedOn
	FROM (Flows f, Modules m, Blocks b, BlockModuleFlows bmf)
	LEFT JOIN BlockUserStatus bus ON (bus.projectId = f.projectId AND bus.userId = ? AND bus.blockId = b.id)
	WHERE f.projectId = ? AND
	f.moduleId = m.id AND
	m.status = 'active' AND
	m.id = bmf.moduleId AND
	bmf.blockId = b.id
	ORDER BY f.flowOrder, bmf.flowOrder`, participantID, projectID)
	for i := range flow {
		flow[i].ProjectID = projectID
		flow[i].UserID = participantID
		flow[i].processForAPI()
	}
	return flow, err
}

// SaveBlockUserStatusForParticipant creates or updates a participant's block status in the flow
func SaveBlockUserStatusForParticipant(input *BlockUserStatus) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`INSERT INTO BlockUserStatus (userId, blockId, moduleId, projectId, lastUpdatedOn, status)
	VALUES (:userId, :blockId, :moduleId, :projectId, :lastUpdatedOn, :userStatus)
	ON DUPLICATE KEY UPDATE status = :userStatus, lastUpdatedOn = :lastUpdatedOn`, input)
	return err
}

// IsModuleInProject checks if a module is in a project flow
func IsModuleInProject(projectID, moduleID int64) bool {
	c := &CountReturn{}
	err := config.DBConnection.Get(c, `SELECT COUNT(*) AS count FROM Flows WHERE projectId = ? AND moduleId = ?`, projectID, moduleID)
	return err == nil && c.Count > 0
}

// IsBlockInModule checks if a block is in a module for a flow
func IsBlockInModule(moduleID, blockID int64) bool {
	c := &CountReturn{}
	err := config.DBConnection.Get(c, `SELECT COUNT(*) AS count FROM BlockModuleFlows WHERE moduleId = ? AND blockId = ?`, moduleID, blockID)
	return err == nil && c.Count > 0
}

// we have three separate deletes for different levels; this allows the clients to not have to make
// many different calls for large projects; keep in mind that the routes may need to do things like clear
// out responses to surveys, etc

// RemoveAllProgressForParticipantAndFlow removes all progress saved for a user in a project flow
func RemoveAllProgressForParticipantAndFlow(participantID, projectID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockUserStatus WHERE userId = ? AND projectId = ?`, participantID, projectID)
	return err
}

// RemoveAllProgressForParticipantAndModule removes a participant's progress in a module
func RemoveAllProgressForParticipantAndModule(participantID, moduleID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockUserStatus WHERE userId = ? AND moduleId = ?`, participantID, moduleID)
	return err
}

// RemoveAllProgressForParticipantAndBlock removes a participant's progress in a block
func RemoveAllProgressForParticipantAndBlock(participantID, blockID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockUserStatus WHERE userId = ? AND blockId = ?`, participantID, blockID)
	return err
}

// CheckProjectParticipantStatusForParticipant checks the participant's status in a project
func CheckProjectParticipantStatusForParticipant(participantID, projectID int64) (string, error) {
	// if the user isn't even in the project, return an error
	if !IsUserInProject(participantID, projectID) {
		return "", errors.New("participant is not in project")
	}

	// get the modules/blocks and status for each
	modules, err := GetProjectFlowForParticipant(participantID, projectID)
	if err != nil {
		return "", err
	}

	// if any aren't not_started, it's at least started
	// if they are all complete, they are complete
	status := BlockUserStatusNotStarted
	allComplete := true
	for i := range modules {
		if modules[i].UserStatus != BlockUserStatusNotStarted {
			status = BlockUserStatusStarted
			if modules[i].UserStatus != BlockUserStatusCompleted {
				allComplete = false
			}
		}
		// we can short circuit here IF we found a module that has been started and
		// also found a module that isn't complete
		if !allComplete && status != BlockUserStatusNotStarted {
			break
		}
	}
	if allComplete {
		status = BlockUserStatusCompleted
	}
	return status, nil
}

func (input *Flow) processForAPI() {
	input.LastUpdatedOn, _ = parseTimeToTimeFormat(input.LastUpdatedOn, timeFormatAPI)
}

// Bind binds the data for the HTTP
func (data *Flow) Bind(r *http.Request) error {
	return nil
}

func (input *BlockUserStatus) processForDB() {
	if input.LastUpdatedOn == "" {
		input.LastUpdatedOn = time.Now().Format(timeFormatDB)
	} else {
		input.LastUpdatedOn, _ = parseTimeToTimeFormat(input.LastUpdatedOn, timeFormatDB)
	}
}

func (input *BlockUserStatus) processForAPI() {
	input.LastUpdatedOn, _ = parseTimeToTimeFormat(input.LastUpdatedOn, timeFormatAPI)
}

// Bind binds the data for the HTTP
func (data *BlockUserStatus) Bind(r *http.Request) error {
	return nil
}
