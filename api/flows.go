package api

import (
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
}

// GetProjectFlowForParticipant gets the entire flow for a project for a participant to lay out the
// flow and status for each section. Note the explicit lack of a module or project status; that can
// be calculated based upon this data.
func GetProjectFlowForParticipant(participantID, projectID int64) ([]Flow, error) {
	flow := []Flow{}
	err := config.DBConnection.Select(&flow, `SELECT f.flowOrder, m.id AS moduleId, m.name AS moduleName, m.description AS moduleDescription, 
	b.id AS blockId, b.name AS blockName, b.summary AS blockSummary, 
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

// we have three separate delets for different levels; this allows the clients to not have to make
// many different calls for large projects; keep in mind that the routes may need to do things like clear
// out responses to surveys, etc

// RemoveAllProgressForParticipantAndFlow removes all progress saved for a user in a project flow
func RemoveAllProgressForParticipantAndFlow(participantID, projectID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockUserStatus WHERE participantId = ? AND projectId = ?`, participantID, projectID)
	return err
}

// RemoveAllProgressForParticipantAndModule removes a participant's progress in a module
func RemoveAllProgressForParticipantAndModule(participantID, moduleID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockUserStatus WHERE participantId = ? AND moduleId = ?`, participantID, moduleID)
	return err
}

// RemoveAllProgressForParticipantAndBlock removes a participant's progress in a block
func RemoveAllProgressForParticipantAndBlock(participantID, blockID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM BlockUserStatus WHERE participantId = ? AND blockId = ?`, participantID, blockID)
	return err
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
