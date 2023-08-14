package songs

import (
	"errors"
	"net/http"

	"codello.dev/ultrastar/txt"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/server/internal/api/apierror"
	"github.com/Karaoke-Manager/server/internal/api/middleware"
	"github.com/Karaoke-Manager/server/internal/model"
	"github.com/Karaoke-Manager/server/internal/schema"
	"github.com/Karaoke-Manager/server/pkg/render"
)

// Create implements the POST /v1/songs endpoint.
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	data, err := txt.ReadSong(r.Body)
	if err != nil {
		_ = render.Render(w, r, apierror.InvalidUltraStarTXT(err))
		return
	}
	song := &model.Song{Song: *data}
	if err = c.songSvc.CreateSong(r.Context(), song); err != nil {
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	render.SetStatus(r, http.StatusCreated)
	s := schema.FromSong(song)
	_ = render.Render(w, r, &s)
}

// Find implements the GET /v1/songs endpoint.
func (c *Controller) Find(w http.ResponseWriter, r *http.Request) {
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

// Get implements the GET /v1/songs/{uuid} endpoint.
func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	resp := schema.FromSong(song)
	_ = render.Render(w, r, &resp)
}

// Update implements the PATCH /v1/songs/{uuid} endpoint.
func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	update := schema.FromSong(song)
	if err := render.Bind(r, &update); err != nil {
		_ = render.Render(w, r, apierror.BindError(err))
		return
	}
	update.Apply(song)
	if err := c.songSvc.UpdateSongData(r.Context(), song); err != nil {
		// TODO: Check for validation errors?
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// Delete implements the DELETE /v1/songs/{uuid} endpoint.
func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	id := middleware.MustGetUUID(r.Context())
	if err := c.songSvc.DeleteSong(r.Context(), id); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			_ = render.Render(w, r, apierror.DBError(err))
			return
		}
	}
	_ = render.NoContent(w, r)
}
