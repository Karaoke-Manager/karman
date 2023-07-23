package songs

import (
	"errors"
	"github.com/Karaoke-Manager/go-ultrastar/txt"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/api/middleware"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"gorm.io/gorm"
	"net/http"
)

func (c Controller) Create(w http.ResponseWriter, r *http.Request) {
	data, err := txt.ReadSong(r.Body)
	if err != nil {
		_ = render.Render(w, r, apierror.InvalidUltraStarTXT(err))
		return
	}
	song := model.NewSong()
	c.songSvc.UpdateSongFromData(&song, data)
	if err = c.songSvc.SaveSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	s := schema.FromSong(song)
	_ = render.Render(w, r, &s)
}

func (c Controller) Find(w http.ResponseWriter, r *http.Request) {
	pagination := middleware.MustGetPagination(r.Context())
	songs, total, err := c.songSvc.FindSongs(r.Context(), pagination.Limit, pagination.Offset)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}

	resp := schema.List[*schema.Song]{
		Items:  make([]*schema.Song, len(songs)),
		Offset: pagination.Offset,
		Limit:  pagination.RequestLimit,
		Total:  total,
	}
	for i, upload := range songs {
		s := schema.FromSong(upload)
		resp.Items[i] = &s
	}
	_ = render.Render(w, r, &resp)
}

func (c Controller) Get(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	resp := schema.FromSong(song)
	_ = render.Render(w, r, &resp)
}

func (c Controller) Update(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	update := schema.FromSong(song)
	if err := render.Bind(r, &update); err != nil {
		_ = render.Render(w, r, apierror.BindError(err))
		return
	}
	update.Apply(&song)
	if err := c.songSvc.SaveSong(r.Context(), &song); err != nil {
		// TODO: Check for validation errors?
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

func (c Controller) Delete(w http.ResponseWriter, r *http.Request) {
	id := middleware.MustGetUUID(r.Context())
	if err := c.songSvc.DeleteSongByUUID(r.Context(), id); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			_ = render.Render(w, r, apierror.DBError(err))
			return
		}
	}
	_ = render.NoContent(w, r)
}
