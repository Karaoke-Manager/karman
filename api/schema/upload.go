package schema

import (
	"encoding/json"
	"io/fs"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// Upload is the response schema for model.Upload.
// The struct contains JSON tags for completeness.
// These however, are not used during marshalling.
type Upload struct {
	render.NopRenderer
	UUID   uuid.UUID         `json:"uuid"`
	Status model.UploadState `json:"status"`

	SongsTotal     int `json:"songsTotal"`
	SongsProcessed int `json:"songsProcessed"`
	Errors         int `json:"errors"`
}

// FromUpload generates a response schema, describing m.
func FromUpload(m *model.Upload) Upload {
	return Upload{
		UUID:           m.UUID,
		Status:         m.State,
		SongsTotal:     m.SongsTotal,
		SongsProcessed: m.SongsProcessed,
		Errors:         m.Errors,
	}
}

// MarshalJSON marshals u into a JSON string.
// The fields included depend on the upload status.
func (u *Upload) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"uuid":   u.UUID,
		"status": u.Status,
	}
	if u.Status == model.UploadStateProcessing || u.Status == model.UploadStateDone {
		data["songsTotal"] = u.SongsTotal
		data["errors"] = u.Errors
	}
	if u.Status == model.UploadStateProcessing {
		data["songsProcessed"] = u.SongsProcessed
	}
	return json.Marshal(data)
}

// FromUploadFileStat creates an UploadFileStat from the specified file infos.
// If nextMarker is not empty, it will also be included.
func FromUploadFileStat(stat fs.FileInfo, children []fs.FileInfo, nextMarker string) UploadFileStat {
	s := UploadFileStat{
		Name:       stat.Name(),
		Size:       stat.Size(),
		Dir:        stat.IsDir(),
		NextMarker: nextMarker,
	}
	if stat.IsDir() {
		s.Children = make([]UploadDirEntry, len(children))
		for i, c := range children {
			s.Children[i] = UploadDirEntry{
				Name: c.Name(),
				Dir:  c.IsDir(),
				Size: c.Size(),
			}
		}
	}
	return s
}

// UploadFileStat describes the response schema for a file listing in an upload.
type UploadFileStat struct {
	render.NopRenderer
	Name       string           `json:"name"`
	Size       int64            `json:"size"`
	Dir        bool             `json:"dir"`
	Children   []UploadDirEntry `json:"children,omitempty"`
	NextMarker string           `json:"nextMarker,omitempty"`
}

// MarshalJSON marshals s into JSON data.
// The included fields mainly depend on whether s is a directory or a file.
func (s UploadFileStat) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"dir": s.Dir,
	}
	if s.Name != "" {
		data["name"] = s.Name
	}
	if s.Dir {
		data["children"] = s.Children
		if s.NextMarker != "" {
			data["nextMarker"] = s.NextMarker
		}
	} else {
		data["size"] = s.Size
	}
	return json.Marshal(data)
}

// UploadDirEntry describes the subschema for files in a directory in an upload.
// This schema should be used only with UploadFileStat.
//
// It differs from UploadFileStat mostly in the fact, that folders do not list their children recursively.
type UploadDirEntry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
	Size int64  `json:"size"`
}

// MarshalJSON marshals e into a JSON string.
// The file size is only included for files.
func (e UploadDirEntry) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"name": e.Name,
		"dir":  e.Dir,
	}
	if !e.Dir {
		data["size"] = e.Size
	}
	return json.Marshal(data)
}

// UploadProcessingError describes an item in an error listing for uploads.
type UploadProcessingError struct {
	render.NopRenderer
	File    string `json:"file"`
	Message string `json:"message"`
}

// FromUploadProcessingError creates an UploadProcessingError describing err.
func FromUploadProcessingError(err *model.UploadProcessingError) UploadProcessingError {
	return UploadProcessingError{
		File:    err.File,
		Message: err.Message,
	}
}
