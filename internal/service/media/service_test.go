package media

import (
	"context"
	"encoding/hex"
	"github.com/Karaoke-Manager/karman/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	"testing"
)

func setupService(t *testing.T) Service {
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	require.NoError(t, err)
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	err = db.AutoMigrate(&model.Song{}, &model.File{}, &model.Upload{})
	require.NoError(t, err)

	store, _ := fileStore(t)
	return NewService(db, store)
}

func TestService_StoreImageFile(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	ctx := context.Background()
	svc := setupService(t)

	// in order to not blow up repository size we download the test data on the fly.
	cases := map[string]struct {
		url      string
		width    int
		height   int
		size     int64
		Checksum string
	}{
		"png":  {"https://placehold.co/930x850.png", 930, 850, 27139, "2e21529175f51f35be15f3f11bf14b69513e542a56d49133c5809fa77f07fb7f"},
		"gif":  {"https://upload.wikimedia.org/wikipedia/commons/e/ea/Test.gif", 240, 183, 7455, "f1985afbaf6a9be3c1a97c0c870ae3b04f9a653eac067895081849e7306314f3"},
		"jpeg": {"https://placehold.co/320x100.jpg", 320, 100, 2078, "8df1ae81c32d3ac74506457a107ddf7120a5af9fd73634e6d224674c8cab3060"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			resp, err := http.Get(c.url)
			require.NoError(t, err)
			defer resp.Body.Close()
			media := resp.Header.Get("Content-Type")
			file, err := svc.StoreImageFile(ctx, media, resp.Body)
			require.NoError(t, err)
			assert.Equal(t, c.width, file.Width, "Width")
			assert.Equal(t, c.height, file.Height, "Height")
			assert.Equal(t, c.size, file.Size, "Size")
			assert.Equal(t, c.Checksum, hex.EncodeToString(file.Checksum), "Checksum")
		})
	}
}
