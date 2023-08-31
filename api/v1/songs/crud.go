package songs

import (
	"net/http"

	"codello.dev/ultrastar/txt"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// Create implements the POST /v1/songs endpoint.
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	data, err := txt.NewReader(r.Body).ReadSong()
	if err != nil {
		_ = render.Render(w, r, apierror.InvalidUltraStarTXT(err))
		return
	}
	song := model.Song{Song: data}
	if err = c.songRepo.CreateSong(r.Context(), &song); err != nil {
		_ = render.Render(w, r, apierror.ServiceError(err))
		return
	}
	render.SetStatus(r, http.StatusCreated)
	s := schema.FromSong(song)
	_ = render.Render(w, r, &s)
}

// Find implements the GET /v1/songs endpoint.
func (c *Controller) Find(w http.ResponseWriter, r *http.Request) {
	pagination := middleware.MustGetPagination(r.Context())
	songs, total, err := c.songRepo.FindSongs(r.Context(), pagination.Limit, pagination.Offset)
	if err != nil {
		_ = render.Render(w, r, apierror.ServiceError(err))
		return
	}

	resp := schema.List[*schema.Song]{
		Items:  make([]*schema.Song, len(songs)),
		Offset: pagination.Offset,
		Limit:  pagination.RequestLimit,
		Total:  total,
	}
	for i, song := range songs {
		s := schema.FromSong(song)
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
	update.Apply(&song)
	if err := c.songRepo.UpdateSong(r.Context(), &song); err != nil {
		// TODO: Check for validation errors?
		_ = render.Render(w, r, apierror.ServiceError(err))
		return
	}
	_ = render.NoContent(w, r)
}

// Delete implements the DELETE /v1/songs/{uuid} endpoint.
func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	id := middleware.MustGetUUID(r.Context())
	if _, err := c.songRepo.DeleteSong(r.Context(), id); err != nil {
		_ = render.Render(w, r, apierror.ServiceError(err))
		return
	}
	_ = render.NoContent(w, r)
}
