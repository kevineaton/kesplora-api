package api

import (
	"net/http"
	"time"
)

const (
	NoteTypeJournal = "journal"
	NoteTypeProject = "project"

	NoteVisibilityPrivate = "private"
	NoteVisibilityAdmins  = "admins"
)

// Note represents a note that a user can enter into the system, often as a part of a training or project
type Note struct {
	ID         int64  `json:"id" db:"id"`
	UserID     int64  `json:"userId" db:"userId"`
	CreatedOn  string `json:"createdOn" db:"createdOn"`
	NoteType   string `json:"noteType" db:"noteType"`
	ProjectID  int64  `json:"projectId" db:"projectId"`
	ModuleID   int64  `json:"moduleId" db:"moduleId"`
	BlockID    int64  `json:"blockId" db:"blockId"`
	Visibility string `json:"visibility" db:"visibility"`
	Title      string `json:"title" db:"title"`
	Body       string `json:"body" db:"body"`

	// needed for the gets
	ProjectName string `json:"projectName" db:"projectName"`
	ModuleName  string `json:"moduleName" db:"moduleName"`
	BlockName   string `json:"blockName" db:"blockName"`
}

// NoteSelectOptions is a helper to store all the possible combinations of values in a selector rather than
// building a whole bunch of strings or funcs
type NoteSelectOptions struct {
	NoteType  string `json:"noteType"`
	ProjectID int64  `json:"projectId"`
	ModuleID  int64  `json:"moduleId"`
	BlockID   int64  `json:"blockId"`
}

// CreateNote creates a new note
func CreateNote(input *Note) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO Notes SET
	userId = :userId,
	createdOn = :createdOn,
	noteType = :noteType,
	projectId = :projectId,
	moduleId = :moduleId,
	blockId = :blockId,
	visibility = :visibility,
	title = :title,
	body = :body`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateNote updates a single note
func UpdateNote(input *Note) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE Notes SET
	userId = :userId,
	createdOn = :createdOn,
	noteType = :noteType,
	projectId = :projectId,
	moduleId = :moduleId,
	blockId = :blockId,
	visibility = :visibility,
	title = :title,
	body = :body
	WHERE id = :id`, input)
	return err
}

// GetNote gets a single note by its id
func GetNote(noteID int64) (*Note, error) {
	note := &Note{}
	defer note.processForAPI()
	err := config.DBConnection.Get(note, `SELECT * FROM Notes WHERE id = ?`, noteID)
	return note, err
}

// GetAllNotesForUser gets non-project notes for a user
func GetAllNotesForUser(userID int64, noteType string, filter *NoteSelectOptions) ([]Note, error) {
	notes := []Note{}
	filter.process()
	var err error

	if noteType == "all" || noteType == "" {
		err = config.DBConnection.Select(&notes, `SELECT n.*, 
			IFNULL(p.name, '') AS projectName, 
			IFNULL(m.name, '') AS moduleName,
			IFNULL(b.name, '') AS blockName
			FROM Notes n
			LEFT JOIN Projects p ON n.projectId = p.id
			LEFT JOIN Modules m ON n.moduleId = m.id
			LEFT JOIN Blocks b ON n.blockId = b.id
			WHERE n.userId = ?
			ORDER BY createdOn DESC`, userID)
	} else {
		err = config.DBConnection.Select(&notes, `SELECT n.*, 
			IFNULL(p.name, '') AS projectName, 
			IFNULL(m.name, '') AS moduleName,
			IFNULL(b.name, '') AS blockName
			FROM Notes n
			LEFT JOIN Projects p ON n.projectId = p.id
			LEFT JOIN Modules m ON n.moduleId = m.id
			LEFT JOIN Blocks b ON n.blockId = b.id
			WHERE n.userId = ? AND n.noteType = ? AND n.projectId = ? AND n.moduleId = ? AND n.blockId = ?
			ORDER BY createdOn DESC;`, userID, noteType, filter.ProjectID, filter.ModuleID, filter.BlockID)
	}
	for i := range notes {
		notes[i].processForAPI()
	}
	return notes, err
}

// DeleteNoteByID deletes a single note
func DeleteNoteByID(noteID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Notes WHERE id = ?`, noteID)
	return err
}

// DeleteNotesForUser deletes all notes tied to a user
func DeleteNotesForUser(userID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Notes WHERE userID = ?`, userID)
	return err
}

// DeleteJournalNotesForUser deletes all journal notes for a user
func DeleteJournalNotesForUser(userID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Notes WHERE userID = ? AND noteType = 'journal'`, userID)
	return err
}

// DeleteProjectNotesForUser deletes the notes for a user tied to a specific project
func DeleteProjectNotesForUser(userID, projectID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Notes WHERE userId = ? AND projectId = ?`, userID, projectID)
	return err
}

// DeleteProjectNotesForProject deletes all notes on a project; should only be called if a project is deleted
func DeleteProjectNotesForProject(projectID int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Notes WHERE projectId = ?`, projectID)
	return err
}

func (input *NoteSelectOptions) process() {
	if input.NoteType == NoteTypeJournal {
		input.ProjectID = 0
		input.ModuleID = 0
		input.BlockID = 0
	}
}

func (input *Note) processForDB() {
	if input.CreatedOn == "" {
		input.CreatedOn = time.Now().Format(timeFormatDB)
	} else {
		input.CreatedOn, _ = parseTimeToTimeFormat(input.CreatedOn, timeFormatDB)
	}
	if input.NoteType == "" {
		input.NoteType = NoteTypeJournal
	}
	if input.Visibility == "" {
		input.Visibility = NoteVisibilityPrivate
	}
	if input.ProjectID == 0 {
		input.ModuleID = 0
		input.BlockID = 0
		input.NoteType = NoteTypeJournal
	}
}

func (input *Note) processForAPI() {
	input.CreatedOn, _ = parseTimeToTimeFormat(input.CreatedOn, timeFormatAPI)
}

// Bind binds the data for the HTTP
func (data *Note) Bind(r *http.Request) error {
	return nil
}
