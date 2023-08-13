package media

import (
	"context"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/Karaoke-Manager/karman/internal/entity"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupService(t *testing.T) Service {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	err = db.AutoMigrate(&entity.Song{}, &entity.File{}, &entity.Upload{})
	require.NoError(t, err)

	store, _ := fileStore(t)
	return NewService(db, store)
}

func TestService_StoreFile(t *testing.T) {
	ctx := context.Background()
	svc := setupService(t)

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
			f, err := os.Open("testdata/" + c.file)
			require.NoErrorf(t, err, "Could not load test file: %s", c.file)
			defer f.Close()
			file, err := svc.StoreFile(ctx, c.media, f)
			if assert.NoError(t, err) {
				assert.Equal(t, c.duration, file.Duration, "Duration")
				assert.Equal(t, c.width, file.Width, "Width")
				assert.Equal(t, c.height, file.Height, "Height")
				assert.Equal(t, c.size, file.Size, "Size")
				assert.Equal(t, c.Checksum, hex.EncodeToString(file.Checksum), "Checksum")
			}
		})
	}
}
