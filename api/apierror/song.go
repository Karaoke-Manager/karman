package apierror

import (
	"errors"
	"net/http"

	"codello.dev/ultrastar/txt"

	"github.com/Karaoke-Manager/karman/model"
)

const (
	// TypeInvalidTXT indicates that the UltraStar txt data could not be parsed.
	// It is usually accompanied by a line number that caused the error.
	TypeInvalidTXT = ProblemTypeDomain + "/invalid-ultrastar-txt"

	// TypeUploadSongReadonly indicates that the song cannot be modified because it belongs to an upload.
	TypeUploadSongReadonly = ProblemTypeDomain + "/upload-song-readonly"

	// TypeMediaFileNotFound indicates that the requested media file was not found.
	TypeMediaFileNotFound = ProblemTypeDomain + "/song-media-not-found"
)

// InvalidUltraStarTXT generates an error indicating that the UltraStar data in the request could not be parsed.
func InvalidUltraStarTXT(err error) *ProblemDetails {
	var parseErr txt.ParseError
	var fields map[string]any
	if errors.As(err, &parseErr) {
		fields = map[string]any{
			"line": parseErr.Line(),
		}
	}
	return &ProblemDetails{
		Type:   TypeInvalidTXT,
		Title:  "Invalid UltraStar TXT format",
		Status: http.StatusBadRequest,
		Detail: err.Error(),
		Fields: fields,
	}
}

// UploadSongReadonly generates an error indicating that song cannot be modified because it belongs to an upload.
func UploadSongReadonly(song model.Song) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeUploadSongReadonly,
		Title:  "Songs in an upload cannot be modified.",
		Status: http.StatusConflict,
		Detail: "The song must be imported before it can be modified.",
		Fields: map[string]any{
			"uuid": song.UUID.String(),
		},
	}
}

// MediaFileNotFound generates an error indicating that the requested song exists
// but the requested media file does not.
// media indicates the type of media (cover/background/audio/video).
func MediaFileNotFound(song model.Song, media string) *ProblemDetails {
	return &ProblemDetails{
		Type:   TypeMediaFileNotFound,
		Title:  "Media File Not Found",
		Status: http.StatusNotFound,
		Detail: "The song has no " + media + ".",
		Fields: map[string]any{
			"uuid":  song.UUID.String(),
			"media": media,
		},
	}
}
