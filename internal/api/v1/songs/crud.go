package songs

import (
	"github.com/Karaoke-Manager/go-ultrastar/txt"
	"github.com/Karaoke-Manager/karman/internal/api/apierror"
	"github.com/Karaoke-Manager/karman/internal/api/middleware"
	"github.com/Karaoke-Manager/karman/internal/schema"
	"github.com/Karaoke-Manager/karman/pkg/render"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	song, err := txt.ReadSong(r.Body)
	if err != nil {
		_ = render.Render(w, r, apierror.InvalidUltraStarTXT(err))
		return
	}
	resp, err := c.svc.CreateSong(r.Context(), song)
	if err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	s := schema.FromSong(resp)
	_ = render.Render(w, r, &s)
}

func (c *Controller) Find(w http.ResponseWriter, r *http.Request) {
	pagination := middleware.MustGetPagination(r.Context())
	songs, total, err := c.svc.FindSongs(r.Context(), pagination.Limit, pagination.Offset)
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

func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	resp := schema.FromSong(song)
	_ = render.Render(w, r, &resp)
}

func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	if song.UploadID != nil {
		_ = render.Render(w, r, apierror.UploadSongReadonly(song))
		return
	}
	update := schema.FromSong(song)
	if err := render.Bind(r, &update); err != nil {
		_ = render.Render(w, r, apierror.BindError(err))
		return
	}
	update.Apply(&song)
	if err := c.svc.SaveSong(r.Context(), &song); err != nil {
		// TODO: Check for validation errors?
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	uuid := chi.URLParam(r, "uuid")
	if err := c.svc.DeleteSongByUUID(r.Context(), uuid); err != nil {
		_ = render.Render(w, r, apierror.DBError(err))
		return
	}
	_ = render.NoContent(w, r)
}
