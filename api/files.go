package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	MaxFileSizeMB = 128

	FileVisibilityAdmin   = "admin"
	FileVisibilityUsers   = "users"
	FileVisibilityProject = "project"
	FileVisibilityPublic  = "public"

	FileLocationSourceAWS   = "aws"
	FileLocationSourceOther = "other"
)

// File is a DB entry for a file that is hosted somewhere
type File struct {
	ID             int64  `json:"id" db:"id"`
	RemoteKey      string `json:"remoteKey" db:"remoteKey"`
	Display        string `json:"display" db:"display"`
	Description    string `json:"description" db:"description"`
	FileType       string `json:"fileType" db:"fileType"`
	UploadedOn     string `json:"uploadedOn" db:"uploadedOn"`
	UploadedBy     int64  `json:"uploadedBy" db:"uploadedBy"`
	Visibility     string `json:"visibility" db:"visibility"`
	Filesize       int64  `json:"fileSize" db:"fileSize"`
	LocationSource string `json:"locationSource" db:"locationSource"`
}

// CreateFileInDB creates a file entry in the DB
func CreateFileInDB(input *File) error {
	input.processForDB()
	defer input.processForAPI()
	res, err := config.DBConnection.NamedExec(`INSERT INTO Files SET
		remoteKey = :remoteKey,
		display = :display,
		description = :description,
		fileType = :fileType,
		uploadedOn = :uploadedOn,
		uploadedBy = :uploadedBy,
		visibility = :visibility,
		fileSize = :fileSize,
		locationSource = :locationSource
	`, input)
	if err != nil {
		return err
	}
	input.ID, _ = res.LastInsertId()
	return nil
}

// UpdateFileInDB updates a file's metadata in the DB
func UpdateFileInDB(input *File) error {
	input.processForDB()
	defer input.processForAPI()
	_, err := config.DBConnection.NamedExec(`UPDATE Files SET
		remoteKey = :remoteKey,
		display = :display,
		description = :description,
		fileType = :fileType,
		uploadedOn = :uploadedOn,
		uploadedBy = :uploadedBy,
		visibility = :visibility,
		fileSize = :fileSize,
		locationSource = :locationSource
		WHERE id = :id
	`, input)
	return err
}

// UpdateFileVisibilityFromAdminOnly is a simple helper for when a block saves with a file; only the
// file id is set, so if the file visibility is set to admin, we need to set it to project; if it's not
// admin only, then we don't change it (we don't want to make it more restrictive)
func UpdateFileVisibilityFromAdminOnly(fileID int64, visibility string) error {
	_, err := config.DBConnection.Exec(`UPDATE Files SET visibility = ? WHERE id = ? AND visibility = 'admin'`, visibility, fileID)
	return err
}

// DeleteFileFromDB deletes a file's metadata from the DB
func DeleteFileFromDB(id int64) error {
	_, err := config.DBConnection.Exec(`DELETE FROM Files WHERE id = ?`, id)
	return err
}

// GetFileFromDB gets a file's metadata from the DB
func GetFileFromDB(id int64) (*File, error) {
	file := &File{}
	defer file.processForAPI()
	err := config.DBConnection.Get(file, `SELECT * FROM Files WHERE id = ?`, id)
	return file, err
}

// GetFilesFromDB gets a list of files from the DB; since this can be pretty large, we add sort and offset
func GetFilesFromDB(sortBy, sortDir string, count, offset int64) ([]File, error) {
	files := []File{}

	// we are going to sprintf the sort and use an allowed list so it is safe

	// sort dir is only ASC / DESC
	if strings.ToUpper(sortDir) != SortDESC {
		sortDir = SortASC
	}

	// since Go doesn't fallthrough, we put the valid sort by's first, then everything else will hit the defaul
	switch strings.ToLower(sortBy) {
	case "remotekey":
	case "uploadedon":
	case "uploadedby":
	default:
		sortBy = "uploadedOn"
	}

	query := fmt.Sprintf(`SELECT * FROM Files ORDER BY %s %s LIMIT ? OFFSET ?`, sortBy, sortDir)
	err := config.DBConnection.Select(&files, query, count, offset)
	for i := range files {
		files[i].processForAPI()
	}
	return files, err
}

func getAllowedFileProviders() ([]string, error) {
	// for now, just AWS
	providers := []string{}
	if config.AWSS3Client != nil {
		providers = append(providers, "AWS")
	}
	if len(providers) == 0 {
		return providers, errors.New("none")
	}
	return providers, nil
}

func (input *File) processForDB() {
	if input.UploadedOn == "" {
		input.UploadedOn = time.Now().Format(timeFormatDB)
	} else {
		input.UploadedOn, _ = parseTimeToTimeFormat(input.UploadedOn, timeFormatDB)
	}
	if input.Visibility == "" {
		input.Visibility = FileVisibilityAdmin
	}
	if input.LocationSource == "" {
		input.LocationSource = FileLocationSourceAWS
	}
}

func (input *File) processForAPI() {
	input.UploadedOn, _ = parseTimeToTimeFormat(input.UploadedOn, timeFormatAPI)
}

// Bind binds the data for the HTTP
func (data *File) Bind(r *http.Request) error {
	return nil
}
