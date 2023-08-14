package apierror

import (
	"fmt"
	"net/http"

	"github.com/Karaoke-Manager/server/internal/entity"
)

// These constants identify known problem types related to uploads.
const (
	// TypeUploadClosed indicates that a file action to an upload was rejected because the upload has already been marked for processing.
	TypeUploadClosed = ProblemTypeDomain + "/upload-closed"

	// TypeUploadFileNotFound indicates that a file was requested from an upload but the file was not found.
	TypeUploadFileNotFound = ProblemTypeDomain + "/upload-file-not-found"

	// TypeInvalidUploadPath indicates that the file path within an upload is not a valid path.
	TypeInvalidUploadPath = ProblemTypeDomain + "/invalid-upload-path"
)

// UploadClosed generates an error indicating that the upload has been marked for processing and cannot be modified.
func UploadClosed(upload entity.Upload) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeUploadClosed,
		Title:  "Upload Closed",
		Status: http.StatusConflict,
		Detail: "You cannot access the files of the upload, because the upload is already closed.",
		Fields: map[string]any{
			"upload": upload.UUID.String(),
		},
	}
}

// UploadFileNotFound generates an error indicating that a requested file within an upload was not found.
func UploadFileNotFound(upload entity.Upload, path string) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeUploadFileNotFound,
		Title:  "File not Found",
		Status: http.StatusNotFound,
		Detail: fmt.Sprintf("The file at %q cannot be found in upload %s", path, upload.UUID.String()),
		Fields: map[string]any{
			"upload": upload.UUID.String(),
			"path":   path,
		},
	}
}

// InvalidUploadPath generates an error indicating that the requested upload path within an upload is not a valid path.
func InvalidUploadPath(path string) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeInvalidUploadPath,
		Title:  "Invalid Upload Path",
		Status: http.StatusBadRequest,
		Detail: fmt.Sprintf("The path \"%s\" is not a valid path for an upload.", path),
		Fields: map[string]any{
			"path": path,
		},
	}
}
