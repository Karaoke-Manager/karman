package upload

import (
	"context"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_CreateFile(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)
	w, err := svc.CreateFile(ctx, data.OpenUpload, "foobar.txt")
	assert.NoError(t, err)
	assert.NoError(t, w.Close())
}

func TestService_StatFile(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)
	w, err := svc.CreateFile(ctx, data.OpenUpload, "foo/bar.txt")
	if !assert.NoError(t, err) || !assert.NoError(t, w.Close()) {
		return
	}

	t.Run("present", func(t *testing.T) {
		stat, err := svc.StatFile(ctx, data.OpenUpload, "foo/bar.txt")
		assert.NoError(t, err)
		assert.Equal(t, "bar.txt", stat.Name())
		assert.False(t, stat.IsDir())
	})
	t.Run("absent", func(t *testing.T) {
		_, err := svc.StatFile(ctx, data.OpenUpload, "abc")
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})
	t.Run("directory", func(t *testing.T) {
		stat, err := svc.StatFile(ctx, data.OpenUpload, "foo")
		assert.NoError(t, err)
		assert.Equal(t, "foo", stat.Name())
		assert.True(t, stat.IsDir())
	})
}

func TestService_OpenDir(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)
	w, err := svc.CreateFile(ctx, data.OpenUpload, "foo/bar.txt")
	if !assert.NoError(t, err) || !assert.NoError(t, w.Close()) {
		return
	}

	t.Run("present", func(t *testing.T) {
		dir, err := svc.OpenDir(ctx, data.OpenUpload, "foo")
		if !assert.NoError(t, err) {
			return
		}
		entries, err := dir.ReadDir(0)
		assert.NoError(t, err)
		assert.Len(t, entries, 1)
	})

	t.Run("absent", func(t *testing.T) {
		_, err := svc.OpenDir(ctx, data.OpenUpload, "abc")
		assert.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("file", func(t *testing.T) {
		_, err := svc.OpenDir(ctx, data.OpenUpload, "foo/bar.txt")
		assert.Error(t, err)
	})
}

func TestService_DeleteFile(t *testing.T) {
	ctx := context.Background()

	t.Run("absent", func(t *testing.T) {
		svc, data := setupService(t, true)
		err := svc.DeleteFile(ctx, data.OpenUpload, "absent.txt")
		assert.NoError(t, err)
	})

	t.Run("present", func(t *testing.T) {
		svc, data := setupService(t, true)
		w, err := svc.CreateFile(ctx, data.OpenUpload, "foo/bar.txt")
		if !assert.NoError(t, err) || !assert.NoError(t, w.Close()) {
			return
		}
		err = svc.DeleteFile(ctx, data.OpenUpload, "foo/bar.txt")
		assert.NoError(t, err)
	})

	t.Run("directory", func(t *testing.T) {
		svc, data := setupService(t, true)
		w, err := svc.CreateFile(ctx, data.OpenUpload, "foo/bar.txt")
		if !assert.NoError(t, err) || !assert.NoError(t, w.Close()) {
			return
		}
		err = svc.DeleteFile(ctx, data.OpenUpload, "foo")
		assert.NoError(t, err)
	})
}
