package schema

import (
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/Karaoke-Manager/karman/pkg/render"
)

type Song struct {
	render.NopRenderer
}

func NewSongFromModel(song model.Song) Song {

}
