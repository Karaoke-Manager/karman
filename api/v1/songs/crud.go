package songs

import (
	"net/http"

	"codello.dev/ultrastar/txt"
	"github.com/lmittmann/tint"

	"github.com/Karaoke-Manager/karman/api/apierror"
	"github.com/Karaoke-Manager/karman/api/middleware"
	"github.com/Karaoke-Manager/karman/api/schema"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

// Create implements the POST /v1/songs endpoint.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	data, err := txt.NewReader(r.Body).ReadSong()
	if err != nil {
		_ = render.Render(w, r, apierror.InvalidUltraStarTXT(err))
		h.logger.WarnContext(r.Context(), "Could not parse UltraStar TXT.", tint.Err(err))
		return
	}
	song := model.Song{Song: data}
	h.songSvc.ParseArtists(r.Context(), &song)
	if err = h.songRepo.CreateSong(r.Context(), &song); err != nil {
		h.logger.ErrorContext(r.Context(), "Could not create song.", tint.Err(err))
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	render.SetStatus(r, http.StatusCreated)
	s := schema.FromSong(song)
	_ = render.Render(w, r, &s)
}

// Find implements the GET /v1/songs endpoint.
func (h *Handler) Find(w http.ResponseWriter, r *http.Request) {
	pagination := middleware.MustGetPagination(r.Context())
	songs, total, err := h.songRepo.FindSongs(r.Context(), pagination.Limit, pagination.Offset)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Could not list songs.", "limit", pagination.Limit, "offset", pagination.Offset, tint.Err(err))
		_ = render.Render(w, r, apierror.ErrInternalServerError)
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
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	resp := schema.FromSong(song)
	_ = render.Render(w, r, &resp)
}

// Update implements the PATCH /v1/songs/{uuid} endpoint.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	song := MustGetSong(r.Context())
	update := schema.FromSong(song)
	if err := render.Bind(r, &update); err != nil {
		_ = render.Render(w, r, apierror.BindError(err))
		return
	}
	update.Apply(&song)
	if err := h.songRepo.UpdateSong(r.Context(), &song); err != nil {
		h.logger.ErrorContext(r.Context(), "Could not update song.", "uuid", song.UUID, tint.Err(err))
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}

// Delete implements the DELETE /v1/songs/{uuid} endpoint.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := middleware.MustGetUUID(r.Context())
	if _, err := h.songRepo.DeleteSong(r.Context(), id); err != nil {
		h.logger.ErrorContext(r.Context(), "Could not delete song.", "uuid", id, tint.Err(err))
		_ = render.Render(w, r, apierror.ErrInternalServerError)
		return
	}
	_ = render.NoContent(w, r)
}
