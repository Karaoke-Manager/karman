package schema

import (
	"encoding/json"
	"io/fs"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

type Upload struct {
	render.NopRenderer
	UUID   uuid.UUID         `json:"uuid"`
	Status model.UploadState `json:"status"`

	SongsTotal     int `json:"songsTotal"`
	SongsProcessed int `json:"songsProcessed"`
	Errors         int `json:"errors"`
}

func FromUpload(m *model.Upload) Upload {
	return Upload{
		UUID:           m.UUID,
		Status:         m.State,
		SongsTotal:     m.SongsTotal,
		SongsProcessed: m.SongsProcessed,
		Errors:         m.Errors,
	}
}

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

func FromUploadFileStat(stat fs.FileInfo, children []fs.FileInfo, nextMarker string) UploadFileStat {
	var entries []UploadDirEntry
	if stat.IsDir() {
		entries = make([]UploadDirEntry, len(children))
		for i, c := range children {
			entries[i] = UploadDirEntry{
				Name: c.Name(),
				Dir:  c.IsDir(),
				Size: c.Size(),
			}
		}
	}
	return UploadFileStat{
		Name:       stat.Name(),
		Size:       stat.Size(),
		Dir:        stat.IsDir(),
		Children:   entries,
		NextMarker: nextMarker,
	}
}

type UploadFileStat struct {
	render.NopRenderer
	Name       string           `json:"name"`
	Size       int64            `json:"size"`
	Dir        bool             `json:"dir"`
	Children   []UploadDirEntry `json:"children,omitempty"`
	NextMarker string           `json:"nextMarker,omitempty"`
}

func (s UploadFileStat) MarshalJSON() ([]byte, error) {
	data := map[string]any{
		"name": s.Name,
		"dir":  s.Dir,
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

type UploadDirEntry struct {
	Name string `json:"name"`
	Dir  bool   `json:"dir"`
	Size int64  `json:"size"`
}

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
