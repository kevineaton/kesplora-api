package api

// ReportValueCount is a generic report count holder
type ReportValueCount struct {
	Value string `json:"value" db:"value"`
	Count int64  `json:"count" db:"count"`
}

// ReportUserLastUpdatedAgo holds the days ago versus the user
type ReportUserLastUpdatedAgo struct {
	DaysAgo int64 `json:"daysAgo" db:"daysAgo"`
	UserID  int64 `json:"userId" db:"userId"`
}

// ReportBlockStatusCount is a report on a user's status in a block
type ReportBlockStatusCount struct {
	ModuleID        int64  `json:"moduleId" db:"moduleId"`
	ModuleName      string `json:"moduleName" db:"moduleName"`
	BlockID         int64  `json:"blockId" db:"blockId"`
	BlockName       string `json:"blockName" db:"blockName"`
	CompletedCount  int64  `json:"completedCount" db:"completedCount"`
	NotStartedCount int64  `json:"notStartedCount" db:"notStartedCount"`
	StartedCount    int64  `json:"startedCount" db:"startedCount"`
}

// ReportSubmissionCount is the count of number of submissions
type ReportSubmissionCount struct {
	ModuleID   int64  `json:"moduleId" db:"moduleId"`
	ModuleName string `json:"moduleName" db:"moduleName"`
	BlockID    int64  `json:"blockId" db:"blockId"`
	BlockName  string `json:"blockName" db:"blockName"`
	BlockType  string `json:"blockType" db:"blockType"`
	Count      int64  `json:"count" db:"count"`
}

// ReportSubmissionResponses is the responses
type ReportSubmissionResponses struct {
	BlockID         int64                               `json:"blockId" db:"blockId"`
	BlockName       string                              `json:"blockName" db:"blockName"`
	SubmissionCount int64                               `json:"submissionCount" db:"submissionCount"`
	Questions       []ReportSubmissionResponsesQuestion `json:"questions" db:"questions"`
}

// ReportSubmissionResponsesQuestion is a question in the response report
type ReportSubmissionResponsesQuestion struct {
	QuestionID   int64                               `json:"questionId" db:"questionId"`
	QuestionText string                              `json:"questionText" db:"questionText"`
	QuestionType string                              `json:"questionType" db:"questionType"`
	Responses    []ReportSubmissionResponsesResponse `json:"responses" db:"responses"`
}

// ReportSubmissionResponsesResponse is a response to a question in the report
type ReportSubmissionResponsesResponse struct {
	OptionID     int64  `json:"optionId" db:"optionId"`
	TextResponse string `json:"textResponse" db:"textResponse"`
	Count        int64  `json:"count" db:"count"`
}

// ReportGetCountOfUsersOnProjectByStatus gets the users on a project by their status
func ReportGetCountOfUsersOnProjectByStatus(projectID int64) ([]ReportValueCount, error) {
	results := []ReportValueCount{}
	err := config.DBConnection.Select(&results, `SELECT p.status AS value, count(*) as count
	FROM ProjectUserLinks p
	WHERE p.projectId = ?
	GROUP BY value ORDER BY count`, projectID)
	return results, err
}

// ReportGetCountOfLastUpdatedForProject gets the report of users in a project by their last updated status
func ReportGetCountOfLastUpdatedForProject(projectID int64) ([]ReportUserLastUpdatedAgo, error) {
	results := []ReportUserLastUpdatedAgo{}
	err := config.DBConnection.Select(&results, `SELECT DATEDIFF(NOW(), bs.lastUpdatedOn) AS daysAgo, bs.userId
	FROM BlockUserStatus bs
	WHERE bs.projectId = ?
	ORDER BY daysAgo`, projectID)
	return results, err
}

// ReportGetCountOfStatusForProject gets the status of the users grouped for a project
func ReportGetCountOfStatusForProject(projectID int64) ([]ReportBlockStatusCount, error) {
	results := []ReportBlockStatusCount{}
	err := config.DBConnection.Select(&results, `SELECT m.id AS moduleId, m.name AS moduleName, b.id AS blockId, b.name AS blockName, 
	(SELECT COUNT(*) FROM BlockUserStatus bs WHERE bs.projectId = ? AND bs.moduleId = m.id AND bs.blockId = b.id AND bs.status = 'completed' ) AS completedCount,
	(SELECT COUNT(*) FROM BlockUserStatus bs WHERE bs.projectId = ? AND bs.moduleId = m.id AND bs.blockId = b.id AND bs.status = 'not_started' ) AS notStartedCount,
	(SELECT COUNT(*) FROM BlockUserStatus bs WHERE bs.projectId = ? AND bs.moduleId = m.id AND bs.blockId = b.id AND bs.status = 'started' ) AS startedCount
	FROM Flows f, Modules m, Blocks b, BlockModuleFlows bmf
	WHERE 
	f.projectId = ? AND
	f.moduleId = m.id AND
	m.id = bmf.moduleId AND
	bmf.blockId = b.id
	ORDER BY f.flowOrder, bmf.flowOrder`, projectID, projectID, projectID, projectID)
	return results, err
}

// ReportGetSubmissionCountForProject gets the count of submissions for a project
func ReportGetSubmissionCountForProject(projectID int64) ([]ReportSubmissionCount, error) {
	results := []ReportSubmissionCount{}
	err := config.DBConnection.Select(&results, `SELECT m.id AS moduleId, m.name AS moduleName, b.id AS blockId, b.name AS blockName, b.blockType, COUNT(*) as count
	FROM Blocks b, Flows f, BlockModuleFlows bmf, BlockFormSubmissions s, Modules m
	WHERE f.projectId = ? AND
	f.moduleId = m.id AND
	f.moduleId = bmf.moduleId AND
	bmf.blockId = b.id AND
	b.id = s.blockId
	GROUP BY b.id, m.id, m.name, b.name, b.blockType, f.flowOrder, bmf.flowOrder
	ORDER BY f.flowOrder, bmf.flowOrder`, projectID)
	return results, err
}
