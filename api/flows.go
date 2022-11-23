package api

// Flow represents the flow of a project. This is currently unused and is a holder in case
// future work needs to apply specific logic to flow steps.
type Flow struct {
	ProjectID int64 `json:"projectId" db:"projectId"`
	ModuleID  int64 `json:"moduleId" db:"moduleId"`
	FlowOrder int64 `json:"flowOrder" db:"flowOrder"`
}
