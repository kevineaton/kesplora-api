package api

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

// routeParticipantGetFileMetaData gets the file's meta data
func routeParticipantGetFileMetaData(w http.ResponseWriter, r *http.Request) {
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

	// we need to check the visibility before going forward
	if file.Visibility == FileVisibilityAdmin {
		sendAPIError(w, "api_error_file_no_exist", errors.New("could not find that file"), nil)
		return
	} else if file.Visibility == FileVisibilityProject {
		// TODO: find out what blocks this file is connected to and make sure it is allowed
	}

	sendAPIJSONData(w, http.StatusOK, file)
}

// routeParticipantDownloadFile downloads a file as a participant
func routeParticipantDownloadFile(w http.ResponseWriter, r *http.Request) {
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

	// we need to check the visibility before going forward
	if file.Visibility == FileVisibilityAdmin {
		sendAPIError(w, "api_error_file_no_exist", errors.New("could not find that file"), nil)
		return
	} else if file.Visibility == FileVisibilityProject {
		// TODO: find out what blocks this file is connected to and make sure it is allowed
	}

	// user and public are fine

	data, err := GetFileFromBucket(file.RemoteKey)
	if err != nil {
		sendAPIError(w, api_error_file_download, fileIDErr, nil)
		return
	}
	// TODO: get the content type from the extension
	format := r.URL.Query().Get("format")
	if format == "base64" {
		encoded := base64.StdEncoding.EncodeToString([]byte(data))
		sendAPIJSONData(w, http.StatusOK, encoded)
		return
	}
	sendAPIFileData(w, http.StatusOK, "octet/binary", data)
}
