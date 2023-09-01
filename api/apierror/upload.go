package apierror

import (
	"fmt"
	"net/http"

	"github.com/Karaoke-Manager/karman/model"
)

// These constants identify known problem types related to uploads.
const (
	// TypeUploadState indicates that a file action to an upload was rejected because the upload has already been marked for processing.
	TypeUploadState = ProblemTypeDomain + "/upload-state"

	// TypeUploadFileNotFound indicates that a file was requested from an upload but the file was not found.
	TypeUploadFileNotFound = ProblemTypeDomain + "/upload-file-not-found"

	// TypeInvalidUploadPath indicates that the file path within an upload is not a valid path.
	TypeInvalidUploadPath = ProblemTypeDomain + "/invalid-upload-path"
)

// UploadState generates an error indicating that the upload is not in the correct state to perform this action.
func UploadState(upload model.Upload) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeUploadState,
		Title:  "Invalid Upload State",
		Status: http.StatusConflict,
		Detail: "This action cannot be performed in the current upload state.",
		Fields: map[string]any{
			"uuid": upload.UUID.String(),
		},
	}
}

// UploadFileNotFound generates an error indicating that a requested file within an upload was not found.
func UploadFileNotFound(upload model.Upload, path string) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeUploadFileNotFound,
		Title:  "File not Found",
		Status: http.StatusNotFound,
		Detail: fmt.Sprintf("The file at %q cannot be found in upload %s", path, upload.UUID.String()),
		Fields: map[string]any{
			"uuid": upload.UUID.String(),
			"path": path,
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
