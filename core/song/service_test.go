package song

import (
	"context"
	"slices"
	"testing"

	"codello.dev/ultrastar"

	"github.com/Karaoke-Manager/karman/model"
)

func Test_service_ParseArtists(t *testing.T) {
	t.Parallel()

	svc := NewService()
	song := model.Song{
		Song: ultrastar.Song{
			Artist: "Foo, Bar",
		},
	}

	svc.ParseArtists(context.TODO(), &song)
	if !slices.Equal(song.Artists, []string{"Foo", "Bar"}) {
		t.Errorf("ParseArtists(%q) = %v, expected %v", song.Artist, song.Artists, []string{"Foo", "Bar"})
	}
}

func Test_service_Prepare(t *testing.T) {
	t.Parallel()

	svc := NewService()
	song := model.Song{
		Song: ultrastar.Song{
			Artist: "Queen",
		},
		Artists:   []string{"Foo", "Bar"},
		AudioFile: &model.File{},
	}

	svc.Prepare(context.TODO(), &song)
	if len(song.Artist) == 0 {
		t.Errorf("Prepare() did not set song.Artist, expected non-zero value")
	}
	if song.Artist == "Queen" {
		t.Errorf("Prepare() procuded song.Artist = %q, expected different value", song.Artist)
	}
	if song.TxtFileName == "" {
		t.Errorf("Prepare() did not set song.TxtFileName, expected non-zero value")
	}
	if song.AudioFileName == "" {
		t.Errorf("Prepare() did not set song.AudioFileName, expected non-zero value")
	}
	if song.VideoFileName != "" {
		t.Errorf("Prepare() set song.VideoFileName = %q, expected empty string", song.VideoFileName)
	}
}
