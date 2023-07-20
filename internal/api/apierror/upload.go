package apierror

import (
	"fmt"
	"github.com/Karaoke-Manager/karman/internal/model"
	"net/http"
)

const (
	TypeUploadClosed       = ProblemTypeDomain + "/upload-closed"
	TypeUploadFileNotFound = ProblemTypeDomain + "/upload-file-not-found"
	TypeInvalidUploadPath  = ProblemTypeDomain + "/invalid-upload-path"
)

func UploadClosed(upload model.Upload) *ProblemDetails {
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

func UploadFileNotFound(upload model.Upload, path string) *ProblemDetails {
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
