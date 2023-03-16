package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

//
// Note that although this is an _all_ file, the calls are from within /admin or /participant so only logged in AND
// it's assumed that they are logged in
//

// routeAllGetMyNotes gets all of the current user's notes
func routeAllGetMyNotes(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil {
		sendAPIError(w, api_error_auth_missing, err, map[string]string{})
		return
	}

	noteType := r.URL.Query().Get("type")
	filter := &NoteSelectOptions{
		NoteType: noteType,
	}

	// TODO: add in the project/module/block filtering

	notes, err := GetAllNotesForUser(user.ID, noteType, filter)
	if err != nil {
		sendAPIError(w, api_error_notes_not_found, err, map[string]string{})
		return
	}

	sendAPIJSONData(w, http.StatusOK, notes)
}

// routeAllCreateNote creates a new note for a user
func routeAllCreateNote(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil {
		sendAPIError(w, api_error_auth_missing, err, map[string]string{})
		return
	}

	input := &Note{}
	render.Bind(r, input)
	input.UserID = user.ID

	if input.Title == "" || input.Body == "" {
		sendAPIError(w, api_error_notes_save, errors.New("title and body required"), map[string]interface{}{
			"input": input,
		})
		return
	}

	// if the user is an admin, set it to journal
	if user.SystemRole == UserSystemRoleAdmin {
		input.ProjectID = 0
		input.ModuleID = 0
		input.BlockID = 0
		input.Visibility = NoteVisibilityPrivate
		input.NoteType = NoteTypeJournal
	} else {
		if input.ProjectID != 0 {
			if !IsUserInProject(user.ID, input.ProjectID) {
				input.ProjectID = 0
				input.ModuleID = 0
				input.BlockID = 0
				input.NoteType = NoteTypeJournal
			}
			if input.ModuleID != 0 && !IsModuleInProject(input.ProjectID, input.ModuleID) {
				input.ModuleID = 0
				input.BlockID = 0
			}
			if input.BlockID != 0 && !IsBlockInModule(input.ModuleID, input.BlockID) {
				input.BlockID = 0
			}
		}
	}

	err = CreateNote(input)
	if err != nil {
		sendAPIError(w, api_error_notes_save, err, map[string]interface{}{
			"input": input,
		})
		return
	}

	sendAPIJSONData(w, http.StatusCreated, input)
}

// routeAllGetMyNoteByID gets a single note for a user by its id
func routeAllGetMyNoteByID(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil {
		sendAPIError(w, api_error_auth_missing, err, map[string]string{})
		return
	}

	noteID, noteIDErr := strconv.ParseInt(chi.URLParam(r, "noteID"), 10, 64)
	if noteIDErr != nil {
		sendAPIError(w, api_error_invalid_path, err, map[string]string{})
		return
	}

	note, err := GetNote(noteID)
	if err != nil {
		sendAPIError(w, api_error_notes_not_found, err, map[string]string{})
		return
	}

	if note.UserID != user.ID {
		sendAPIError(w, api_error_notes_not_found, err, map[string]string{})
		return
	}

	sendAPIJSONData(w, http.StatusOK, note)
}

// routeAllUpdateNoteByID update a single note for a user by its id
func routeAllUpdateNoteByID(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil {
		sendAPIError(w, api_error_auth_missing, err, map[string]string{})
		return
	}

	noteID, noteIDErr := strconv.ParseInt(chi.URLParam(r, "noteID"), 10, 64)
	if noteIDErr != nil {
		sendAPIError(w, api_error_invalid_path, err, map[string]string{})
		return
	}

	note, err := GetNote(noteID)
	if err != nil {
		sendAPIError(w, api_error_notes_not_found, err, map[string]string{})
		return
	}

	if note.UserID != user.ID {
		sendAPIError(w, api_error_notes_not_found, err, map[string]string{})
		return
	}

	input := &Note{}
	render.Bind(r, input)

	if input.Title != "" {
		note.Title = input.Title
	}
	if input.Body != "" {
		note.Body = input.Body
	}
	if input.Visibility != "" {
		note.Visibility = input.Visibility
	}
	if input.NoteType != "" {
		note.NoteType = input.NoteType
	}

	err = UpdateNote(note)
	if err != nil {
		sendAPIError(w, api_error_notes_save, err, map[string]interface{}{
			"input": input,
		})
		return
	}

	sendAPIJSONData(w, http.StatusOK, note)
}

// routeAllDeleteMyNoteByID deletes a single note
func routeAllDeleteMyNoteByID(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromHTTPContext(r)
	if err != nil {
		sendAPIError(w, api_error_auth_missing, err, map[string]string{})
		return
	}

	noteID, noteIDErr := strconv.ParseInt(chi.URLParam(r, "noteID"), 10, 64)
	if noteIDErr != nil {
		sendAPIError(w, api_error_invalid_path, err, map[string]string{})
		return
	}

	note, err := GetNote(noteID)
	if err != nil {
		sendAPIError(w, api_error_notes_not_found, err, map[string]string{})
		return
	}

	if note.UserID != user.ID {
		sendAPIError(w, api_error_notes_not_found, err, map[string]string{})
		return
	}

	err = DeleteNoteByID(noteID)
	if err != nil {
		sendAPIError(w, api_error_notes_delete, err, map[string]string{})
		return
	}

	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
}
