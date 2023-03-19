package api

import (
	"io"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
)

// routeAdminUploadFile receives an uploaded multipart-file and creates a stub in the DB; note that if no
// providers are available, this will fail as not supported
func routeAdminUploadFile(w http.ResponseWriter, r *http.Request) {
	_, err := getAllowedFileProviders()
	if err != nil {
		sendAPIError(w, api_error_file_upload_no_provider, err, nil)
		return
	}
	admin, _ := getUserFromHTTPContext(r)

	err = r.ParseMultipartForm(MaxFileSizeMB << 20)
	if err != nil {
		sendAPIError(w, api_error_file_upload_parse_general, err, nil)
		return
	}

	file, headers, err := r.FormFile("file")
	if err != nil {
		sendAPIError(w, api_error_file_upload_parse_form, err, nil)
		return
	}
	data, err := io.ReadAll(file)
	if err != nil {
		sendAPIError(w, api_error_file_upload_read, err, nil)
		return
	}

	// TODO: split here on AWS or whatever future providers we support
	err = UploadFileToBucket(headers.Filename, data)
	if err != nil {
		sendAPIError(w, api_error_file_upload_upload, err, nil)
		return
	}

	// now save the data in the DB
	fileInput := &File{
		RemoteKey:      headers.Filename,
		Display:        headers.Filename,
		UploadedBy:     admin.ID,
		FileSize:       headers.Size,
		FileType:       filepath.Ext(headers.Filename),
		Visibility:     FileVisibilityAdmin,
		LocationSource: FileLocationSourceAWS,
	}
	err = CreateFileInDB(fileInput)
	if err != nil {
		// TODO: we probably want to delete the object, right?
		sendAPIError(w, api_error_file_upload_meta_save, err, map[string]interface{}{
			"fileInput": fileInput,
		})
		return
	}
	sendAPIJSONData(w, http.StatusCreated, fileInput)
}

// routeAdminGetFileMetaData gets the file's meta data
func routeAdminGetFiles(w http.ResponseWriter, r *http.Request) {
	_, err := getAllowedFileProviders()
	if err != nil {
		sendAPIError(w, api_error_file_upload_no_provider, err, nil)
		return
	}

	params := processQuery(r)

	// find the file in the DB to prevent a jump to AWS
	files, err := GetFilesFromDB(params.SortField, params.SortDir, params.Count, params.Offset)
	if err != nil {
		sendAPIError(w, api_error_file_no_exist, err, nil)
		return
	}

	sendAPIJSONData(w, http.StatusOK, files)
}

// routeAdminReplaceFile replaces a file that has been uploaded
func routeAdminReplaceFile(w http.ResponseWriter, r *http.Request) {
	_, err := getAllowedFileProviders()
	if err != nil {
		sendAPIError(w, api_error_file_upload_no_provider, err, nil)
		return
	}
	admin, _ := getUserFromHTTPContext(r)

	fileID, fileIDErr := strconv.ParseInt(chi.URLParam(r, "fileID"), 10, 64)
	if fileIDErr != nil {
		sendAPIError(w, api_error_invalid_path, fileIDErr, nil)
		return
	}

	// find the file in the DB to prevent a jump to AWS
	existingFile, err := GetFileFromDB(fileID)
	if err != nil {
		sendAPIError(w, api_error_file_no_exist, err, nil)
		return
	}

	err = r.ParseMultipartForm(MaxFileSizeMB << 20)
	if err != nil {
		sendAPIError(w, api_error_file_upload_parse_general, err, nil)
		return
	}

	file, headers, err := r.FormFile("file")
	if err != nil {
		sendAPIError(w, api_error_file_upload_parse_form, err, nil)
		return
	}
	data, err := io.ReadAll(file)
	if err != nil {
		sendAPIError(w, api_error_file_upload_read, err, nil)
		return
	}

	key := existingFile.RemoteKey

	// TODO: split here on AWS or whatever future providers we support
	err = UploadFileToBucket(key, data)
	if err != nil {
		sendAPIError(w, api_error_file_upload_upload, err, nil)
		return
	}

	// now save the data in the DB
	fileInput := &File{
		ID:         fileID,
		RemoteKey:  key,
		Display:    headers.Filename,
		UploadedBy: admin.ID,
		FileSize:   headers.Size,
		FileType:   filepath.Ext(headers.Filename),
	}
	err = UpdateFileInDB(fileInput)
	if err != nil {
		sendAPIError(w, api_error_file_upload_meta_save, err, map[string]interface{}{
			"fileInput": fileInput,
		})
		return
	}
	sendAPIJSONData(w, http.StatusCreated, fileInput)
}

// routeAdminGetFileMetaData gets the file's meta data
func routeAdminGetFileMetaData(w http.ResponseWriter, r *http.Request) {
	_, err := getAllowedFileProviders()
	if err != nil {
		sendAPIError(w, api_error_file_upload_no_provider, err, nil)
		return
	}
	fileID, fileIDErr := strconv.ParseInt(chi.URLParam(r, "fileID"), 10, 64)
	if fileIDErr != nil {
		sendAPIError(w, api_error_invalid_path, fileIDErr, nil)
		return
	}

	// find the file in the DB to prevent a jump to AWS
	file, err := GetFileFromDB(fileID)
	if err != nil {
		sendAPIError(w, api_error_file_no_exist, err, nil)
		return
	}

	sendAPIJSONData(w, http.StatusOK, file)
}

// routeAdminDeleteFile deletes a file
func routeAdminDeleteFile(w http.ResponseWriter, r *http.Request) {
	_, err := getAllowedFileProviders()
	if err != nil {
		sendAPIError(w, api_error_file_upload_no_provider, err, nil)
		return
	}
	fileID, fileIDErr := strconv.ParseInt(chi.URLParam(r, "fileID"), 10, 64)
	if fileIDErr != nil {
		sendAPIError(w, api_error_invalid_path, fileIDErr, nil)
		return
	}

	file, err := GetFileFromDB(fileID)
	if err != nil {
		sendAPIError(w, api_error_file_no_exist, err, nil)
		return
	}

	// TODO: if this belongs to a module, error out as a conflict

	// first delete from bucket
	err = DeleteFileFromBucket(file.RemoteKey)
	if err != nil {
		sendAPIError(w, api_error_file_delete_remote, err, nil)
		return
	}

	// now delete from db
	err = DeleteFileFromDB(file.ID)
	if err != nil {
		sendAPIError(w, api_error_file_delete_meta, err, nil)
		return
	}

	sendAPIJSONData(w, http.StatusOK, map[string]bool{
		"deleted": true,
	})
}

// routeUpdateFileMetadata updates the metadata for a file
func routeUpdateFileMetadata(w http.ResponseWriter, r *http.Request) {
	_, err := getAllowedFileProviders()
	if err != nil {
		sendAPIError(w, api_error_file_upload_no_provider, err, nil)
		return
	}
	fileID, fileIDErr := strconv.ParseInt(chi.URLParam(r, "fileID"), 10, 64)
	if fileIDErr != nil {
		sendAPIError(w, api_error_invalid_path, fileIDErr, nil)
		return
	}

	file, err := GetFileFromDB(fileID)
	if err != nil {
		sendAPIError(w, api_error_file_no_exist, err, nil)
		return
	}

	input := &File{}
	render.Bind(r, input)

	if input.Display != "" {
		file.Display = input.Display
	}
	if input.Description != "" {
		file.Description = input.Description
	}
	if input.Visibility != "" {
		file.Visibility = input.Visibility
	}

	err = UpdateFileInDB(file)
	if err != nil {
		sendAPIError(w, "api_error_file_update_meta", err, nil)
		return
	}
	sendAPIJSONData(w, http.StatusOK, file)
}

// routeAdminDownloadFile
func routeAdminDownloadFile(w http.ResponseWriter, r *http.Request) {
	_, err := getAllowedFileProviders()
	if err != nil {
		sendAPIError(w, api_error_file_upload_no_provider, err, nil)
		return
	}
	fileID, fileIDErr := strconv.ParseInt(chi.URLParam(r, "fileID"), 10, 64)
	if fileIDErr != nil {
		sendAPIError(w, api_error_invalid_path, fileIDErr, nil)
		return
	}

	// find the file in the DB to prevent a jump to AWS
	file, err := GetFileFromDB(fileID)
	if err != nil {
		sendAPIError(w, api_error_file_no_exist, fileIDErr, nil)
		return
	}

	data, err := GetFileFromBucket(file.RemoteKey)
	if err != nil {
		sendAPIError(w, api_error_file_download, fileIDErr, nil)
		return
	}
	// TODO: get the content type from the extension
	sendAPIFileData(w, http.StatusOK, "octet/binary", data)
}
