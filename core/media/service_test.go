package media

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/nolog"
	"github.com/Karaoke-Manager/karman/test"
)

func TestService_StoreFile(t *testing.T) {
	t.Parallel()

	store, _ := fileStore(t)
	svc := NewService(nolog.Logger, NewFakeRepository(), store)

	// in order to not blow up repository size we download the test data on the fly.
	cases := map[string]struct {
		file     string
		media    mediatype.MediaType
		duration time.Duration
		width    int
		height   int
		size     int64
		Checksum string
	}{
		"png":  {"test.png", mediatype.ImagePNG, 0, 930, 850, 27139, "2e21529175f51f35be15f3f11bf14b69513e542a56d49133c5809fa77f07fb7f"},
		"gif":  {"test.gif", mediatype.ImageGIF, 0, 240, 183, 7455, "f1985afbaf6a9be3c1a97c0c870ae3b04f9a653eac067895081849e7306314f3"},
		"jpeg": {"test.jpg", mediatype.ImageJPEG, 0, 320, 100, 2078, "8df1ae81c32d3ac74506457a107ddf7120a5af9fd73634e6d224674c8cab3060"},
		"mp3":  {"test.mp3", mediatype.AudioMPEG, 42*time.Second + 83263728*time.Nanosecond, 0, 0, 733645, "9a2270d5964f64981fb1e91dd13e5941262817bdce873cf357c92adbef906b5d"},
		"mp4":  {"test.mp4", mediatype.VideoMP4, 10 * time.Second, 1920, 1080, 9452, "7c6fdbefbd753782d31e987903411f93d216e23bff3fe3eec9ee3a6577996c64"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f := test.MustOpen(t, fmt.Sprintf("testdata/%s", c.file))
			file, err := svc.StoreFile(context.TODO(), c.media, f)
			if err != nil {
				t.Errorf("StoreFile(ctx, %q, f) returned an unexpected error: %s", c.media, err)
				return
			}
			if file.Duration != c.duration {
				t.Errorf("StoreFile(ctx, %q, f) yielded file.Duration = %s, expected %s", c.media, file.Duration, c.duration)
			}
			if file.Width != c.width {
				t.Errorf("StoreFile(ctx, %q, f) yielded file.Width = %d, expected %d", c.media, file.Width, c.width)
			}
			if file.Height != c.height {
				t.Errorf("StoreFile(ctx, %q, f) yielded file.Height = %d, expected %d", c.media, file.Height, c.height)
			}
			if file.Size != c.size {
				t.Errorf("StoreFile(ctx, %q, f) yielded file.Size = %d, expected %d", c.media, file.Size, c.size)
			}
			sum := hex.EncodeToString(file.Checksum)
			if sum != c.Checksum {
				t.Errorf("StoreFile(ctx, %q, f) yielded file.Checksum = %s, expected %s", c.media, sum, c.Checksum)
			}
		})
	}
}

func TestService_DeleteFile(t *testing.T) {
	t.Parallel()

	store := NewMemStore()
	repo := NewFakeRepository().(*fakeRepo)
	svc := NewService(nolog.Logger, repo, store)

	id := uuid.New()
	w, _ := store.Create(context.TODO(), mediatype.Nil, id)
	_, _ = io.WriteString(w, "Hello World")
	_ = w.Close()
	repo.files[id] = model.File{Model: model.Model{UUID: id}}

	err := svc.DeleteFile(context.TODO(), id)
	if err != nil {
		t.Errorf("DeleteFile(ctx, %q) returned an unexpected error: %s", id, err)
	}
	_, err = store.Open(context.TODO(), mediatype.Nil, id)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("DeleteFile(ctx, %q) did not actually delete the file from the store: %s", id, err)
	}
	_, err = repo.GetFile(context.TODO(), id)
	if !errors.Is(err, core.ErrNotFound) {
		t.Errorf("DeleteFile(ctx, %q) did not actuall delete the file from the database: %s", id, err)
	}
}
