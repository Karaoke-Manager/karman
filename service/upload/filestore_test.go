package upload

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		w, err := store.Create(ctx, id, "foobar.txt")
		if !assert.NoError(t, err) {
			return
		}
		n, err := io.WriteString(w, "Hello World")
		assert.NoError(t, err)
		assert.Equal(t, 11, n)
		require.NoError(t, w.Close())

		stat, err := os.Stat(filepath.Join(dir, id.String(), "foobar.txt"))
		require.NoError(t, err)
		assert.False(t, stat.IsDir())
		assert.Equal(t, int64(11), stat.Size())
	})

	t.Run("overwrite file", func(t *testing.T) {
		store, dir := fileStore(t)
		dir = filepath.Join(dir, id.String())
		name := filepath.Join(dir, "file.txt")

		require.NoError(t, os.Mkdir(dir, store.DirMode))
		require.NoError(t, os.WriteFile(name, []byte("Test"), store.FileMode))

		w, err := store.Create(ctx, id, "file.txt")
		if !assert.NoError(t, err) {
			return
		}
		n, err := io.WriteString(w, "Another\nValue")
		assert.NoError(t, err)
		assert.Equal(t, 13, n)
		require.NoError(t, w.Close())

		stat, err := os.Stat(name)
		require.NoError(t, err)
		assert.False(t, stat.IsDir())
		assert.Equal(t, int64(13), stat.Size())
	})

	t.Run("intermediate folders", func(t *testing.T) {
		store, dir := fileStore(t)

		w, err := store.Create(ctx, id, "my/foo/bar.txt")
		if !assert.NoError(t, err) {
			return
		}
		require.NoError(t, w.Close())
		stat, err := os.Stat(filepath.Join(dir, id.String(), "my/foo/bar.txt"))
		if !assert.NoError(t, err) {
			assert.False(t, stat.IsDir())
			assert.Equal(t, 0, stat.Size())
			assert.Equal(t, store.FileMode, stat.Mode())
		}
		dirStat, err := os.Stat(filepath.Join(dir, id.String(), "my/foo"))
		if !assert.NoError(t, err) {
			assert.True(t, dirStat.IsDir())
			assert.Equal(t, store.DirMode, dirStat.Mode())
		}
	})

	t.Run("root", func(t *testing.T) {
		store, _ := fileStore(t)
		_, err := store.Create(ctx, id, ".")
		assert.Error(t, err)
	})
}

func TestFileStore_Stat(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")

	t.Run("present", func(t *testing.T) {
		store, dir := fileStore(t)
		require.NoError(t, os.MkdirAll(filepath.Join(dir, id.String()), store.DirMode))
		require.NoError(t, os.WriteFile(filepath.Join(dir, id.String(), "foobar.txt"), []byte("Hello World"), store.FileMode))

		stat, err := store.Stat(ctx, id, "foobar.txt")
		if assert.NoError(t, err) {
			assert.False(t, stat.IsDir())
			assert.Equal(t, int64(11), stat.Size())
		}
	})

	t.Run("absent", func(t *testing.T) {
		store, _ := fileStore(t)
		_, err := store.Stat(ctx, id, "hello.txt")
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("root", func(t *testing.T) {
		store, _ := fileStore(t)
		stat, err := store.Stat(ctx, id, ".")
		if assert.NoError(t, err) {
			assert.True(t, stat.IsDir())
			assert.Equal(t, id.String(), stat.Name())
		}
	})
}

func TestFileStore_Open(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")

	t.Run("file", func(t *testing.T) {
		store, dir := fileStore(t)
		require.NoError(t, os.MkdirAll(filepath.Join(dir, id.String()), store.DirMode))
		require.NoError(t, os.WriteFile(filepath.Join(dir, id.String(), "test.txt"), []byte("Foobar"), store.FileMode))

		f, err := store.Open(ctx, id, "test.txt")
		if !assert.NoError(t, err) {
			return
		}
		data, err := io.ReadAll(f)
		assert.NoError(t, err)
		assert.Equal(t, []byte("Foobar"), data)
		assert.NoError(t, f.Close())
	})

	t.Run("absent", func(t *testing.T) {
		store, _ := fileStore(t)
		_, err := store.Open(ctx, id, "foobar")
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("folder", func(t *testing.T) {
		store, dir := fileStore(t)
		require.NoError(t, os.MkdirAll(filepath.Join(dir, id.String(), "test", "folder"), store.DirMode))

		f, err := store.Open(ctx, id, "test")
		if !assert.NoError(t, err) {
			return
		}
		stat, err := f.Stat()
		assert.NoError(t, err)
		assert.True(t, stat.IsDir())
		assert.Implements(t, (*Dir)(nil), f)
	})
}

func TestFileStore_Delete(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")

	cases := map[string]struct {
		path string
	}{
		"file":             {"foobar/test.txt"},
		"empty folder":     {"empty"},
		"recursive folder": {"foobar"},
		"absent":           {"nope"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			store, dir := fileStore(t)
			require.NoError(t, os.MkdirAll(filepath.Join(dir, id.String(), "empty"), store.DirMode))
			require.NoError(t, os.MkdirAll(filepath.Join(dir, id.String(), "foobar"), store.DirMode))
			require.NoError(t, os.WriteFile(filepath.Join(dir, id.String(), "foobar/test.txt"), []byte("Foobar"), store.FileMode))

			assert.NoError(t, store.Delete(ctx, id, c.path))
			_, err := os.Stat(filepath.Join(dir, id.String(), c.path))
			assert.ErrorIs(t, err, fs.ErrNotExist)
		})
	}
}

func TestFolderDir_Marker(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")

	store, _ := NewFileStore("./testdata")
	f, err := store.Open(ctx, id, ".")
	require.NoError(t, err)
	dir := f.(Dir)

	assert.Equal(t, "", dir.Marker())
	_, err = dir.Readdir(2)
	assert.NoError(t, err)
	assert.Equal(t, "dir2", dir.Marker())
}

func TestFolderDir_SkipTo(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")
	store, _ := NewFileStore("./testdata")

	f, err := store.Open(ctx, id, ".")
	require.NoError(t, err)
	dir := f.(Dir)

	assert.NoError(t, dir.SkipTo("def"))
	assert.Error(t, dir.SkipTo("abc"))
}

func TestFolderDir_ReadDir(t *testing.T) {
	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")
	store, _ := NewFileStore("./testdata")

	cases := map[string]struct {
		marker    string
		n         int
		len       int
		newMarker string
	}{
		"abc":  {"abc", 2, 2, "dir2"},
		"dir2": {"dir2", 0, 2, "file2.txt"},
		"go":   {"go", 0, 0, "go"},
		"all":  {"", -1, 4, "file2.txt"},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			f, err := store.Open(ctx, id, ".")
			require.NoError(t, err)
			dir := f.(Dir)

			assert.NoError(t, dir.SkipTo(c.marker))
			entries, err := dir.Readdir(c.n)
			assert.NoError(t, err)
			assert.Len(t, entries, c.len)
			assert.Equal(t, c.newMarker, dir.Marker())
		})
	}
}
