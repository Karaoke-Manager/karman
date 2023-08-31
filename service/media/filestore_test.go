package media

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

func fileStore(t *testing.T) (*FileStore, string) {
	dir, err := os.MkdirTemp("", "karman-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	store, err := NewFileStore(dir)
	require.NoError(t, err)
	return store, dir
}

func TestNewFileStore(t *testing.T) {
	dir, err := os.MkdirTemp("", "karman-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	t.Run("missing root directory", func(t *testing.T) {
		_, err = NewFileStore(filepath.Join(dir, "test1"))
		assert.Error(t, err)
	})
	t.Run("file root", func(t *testing.T) {
		path := filepath.Join(dir, "test2")
		err = os.WriteFile(path, []byte("Hello"), 0660)
		require.NoError(t, err)
		_, err = NewFileStore(path)
		assert.Error(t, err)
	})
	t.Run("success", func(t *testing.T) {
		path := filepath.Join(dir, "test3")
		err = os.Mkdir(path, 0770)
		require.NoError(t, err)
		_, err = NewFileStore(path)
		assert.NoError(t, err)
	})
}

func TestFileStore_Create(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")

	t.Run("new file", func(t *testing.T) {
		store, dir := fileStore(t)
		w, err := store.Create(ctx, mediatype.Nil, id)
		require.NoError(t, err)
		n, err := io.WriteString(w, "Hello World")
		assert.NoError(t, err)
		assert.Equal(t, 11, n)
		require.NoError(t, w.Close())

		stat, err := os.Stat(filepath.Join(dir, "e4", id.String()))
		require.NoError(t, err)
		assert.False(t, stat.IsDir())
		assert.Equal(t, int64(11), stat.Size())
	})

	t.Run("overwrite file", func(t *testing.T) {
		store, dir := fileStore(t)
		require.NoError(t, os.Mkdir(filepath.Join(dir, "e4"), 0770))
		f, err := os.Create(filepath.Join(dir, "e4", id.String()))
		require.NoError(t, err)
		_, err = f.WriteString("Test")
		require.NoError(t, err)
		require.NoError(t, f.Close())

		w, err := store.Create(ctx, mediatype.Nil, id)
		require.NoError(t, err)
		n, err := io.WriteString(w, "Another\nValue")
		assert.NoError(t, err)
		assert.Equal(t, 13, n)
		require.NoError(t, w.Close())

		stat, err := os.Stat(filepath.Join(dir, "e4", id.String()))
		require.NoError(t, err)
		assert.False(t, stat.IsDir())
		assert.Equal(t, int64(13), stat.Size())
	})
}

func TestFileStore_Open(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")
	store, dir := fileStore(t)
	require.NoError(t, os.Mkdir(filepath.Join(dir, "e4"), 0770))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "e4", id.String()), []byte("Hello World"), 0660))

	t.Run("read file", func(t *testing.T) {
		r, err := store.Open(ctx, mediatype.Nil, id)
		assert.NoError(t, err)
		data, err := io.ReadAll(r)
		assert.NoError(t, err)
		assert.Equal(t, "Hello World", string(data))
		require.NoError(t, r.Close())
	})

	t.Run("non existing", func(t *testing.T) {
		_, err := store.Open(ctx, mediatype.Nil, uuid.New())
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}
