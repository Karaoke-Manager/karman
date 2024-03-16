package media

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/nolog"
)

// fileStore creates a new FileStore in a temporary directory.
func fileStore(t *testing.T) (*FileStore, string) {
	dir, err := os.MkdirTemp("", "karman-test-*")
	if err != nil {
		t.Fatalf("fileStore() could not create temporary directory: %s", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	store, err := NewFileStore(nolog.Logger, dir)
	if err != nil {
		t.Fatalf("fileStore() could not create FileStore instance: %s", err)
	}
	return store, dir
}

func TestNewFileStore(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "karman-test-*")
	if err != nil {
		t.Fatalf("could not create temporary directory: %s", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	t.Run("missing root directory", func(t *testing.T) {
		path := filepath.Join(dir, "test1")
		_, err = NewFileStore(nolog.Logger, path)
		if err == nil {
			t.Errorf("NewFileStore(%q) did not return an error, but an error was expected", path)
		}
	})
	t.Run("file root", func(t *testing.T) {
		path := filepath.Join(dir, "test2")
		err = os.WriteFile(path, []byte("Hello"), 0600)
		if err != nil {
			t.Fatalf("cold not create temporary file at %s: %s", path, err)
		}
		_, err = NewFileStore(nolog.Logger, path)
		if err == nil {
			t.Errorf("NewFileStore(%q) did not return an error, but an error was expected", path)
		}
	})
	t.Run("success", func(t *testing.T) {
		path := filepath.Join(dir, "test3")
		err = os.Mkdir(path, 0770)
		if err != nil {
			t.Fatalf("could not create new directory at %s", path)
		}
		_, err = NewFileStore(nolog.Logger, path)
		if err != nil {
			t.Errorf("NewFileStore(%q) returned an unexpected error: %s", path, err)
		}
	})
}

func TestFileStore_Create(t *testing.T) {
	t.Parallel()

	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")

	t.Run("new file", func(t *testing.T) {
		store, dir := fileStore(t)
		w, err := store.Create(context.TODO(), mediatype.Nil, id)
		if err != nil {
			t.Errorf("Create(ctx, \"\", %q) returned an unexpected error: %s", id, err)
			return
		}
		n, err := io.WriteString(w, "Hello World")
		if err != nil {
			t.Errorf("WriteString(...) returned an unexpected error: %s", err)
			return
		}
		if n != 11 {
			t.Errorf("WriteString(...) wrote %d bytes, expected %d", n, 11)
		}
		if err = w.Close(); err != nil {
			t.Errorf("Close() returned an unexpected error: %s", err)
		}

		path := filepath.Join(dir, "e4", id.String())
		stat, err := os.Stat(path)
		if err != nil {
			t.Errorf("Stat(%q) returned an unexpected error: %s", path, err)
			return
		}
		if stat.IsDir() {
			t.Errorf("Stat(%q) indicates a directory, file expected", path)
		}
		if stat.Size() != 11 {
			t.Errorf("Stat(%q) indicates a size of %d, expected %d", path, stat.Size(), 11)
		}
	})

	t.Run("overwrite file", func(t *testing.T) {
		store, dir := fileStore(t)
		path := filepath.Join(dir, "e4", id.String())
		if err := os.Mkdir(filepath.Dir(path), 0770); err != nil {
			t.Fatalf("could not create directory at %q: %s", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte("Test"), store.FileMode); err != nil {
			t.Fatalf("could not create file at %q: %s", path, err)
		}

		w, err := store.Create(context.TODO(), mediatype.Nil, id)
		if err != nil {
			t.Errorf("Create(ctx, \"\", %q) returned an unexpected error: %s", id, err)
		}
		n, err := io.WriteString(w, "Another\nValue")
		if err != nil {
			t.Errorf("WriteString(...) returned an unexpected error: %s", err)
			return
		}
		if n != 13 {
			t.Errorf("WriteString(...) wrote %d bytes, expected %d", n, 13)
		}
		if err = w.Close(); err != nil {
			t.Errorf("Close() returned an unexpected error: %s", err)
			return
		}

		stat, err := os.Stat(path)
		if err != nil {
			t.Errorf("Stat(%q) returned an unexpected error: %s", path, err)
			return
		}
		if stat.IsDir() {
			t.Errorf("Stat(%q) indicates a directory, file expected", path)
		}
		if stat.Size() != 13 {
			t.Errorf("Stat(%q) indicates a size of %d, expected %d", path, stat.Size(), 13)
		}
	})
}

func TestFileStore_Open(t *testing.T) {
	t.Parallel()

	store, dir := fileStore(t)
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")
	path := filepath.Join(dir, "e4", id.String())
	if err := os.Mkdir(filepath.Dir(path), 0770); err != nil {
		t.Fatalf("Mkdir(%q) returned an unexpected error: %s", path, err)
	}
	if err := os.WriteFile(path, []byte("Hello World"), 0600); err != nil {
		t.Fatalf("WriteFile(...) returned an unexpected error: %s", err)
	}

	t.Run("read file", func(t *testing.T) {
		r, err := store.Open(context.TODO(), mediatype.Nil, id)
		if err != nil {
			t.Errorf("Open(ctx, \"\", %q) returned an unexpected error: %s", id, err)
		}
		data, err := io.ReadAll(r)
		if err != nil {
			t.Errorf("ReadAll(...) returned an unexpected error: %s", err)
		}
		if string(data) != "Hello World" {
			t.Errorf("ReadAll(...) returned %q, expected %q", data, "Hello World")
		}
		if err = r.Close(); err != nil {
			t.Errorf("Close() returned an unexpected error: %s", err)
		}
	})

	t.Run("non existing", func(t *testing.T) {
		_, err := store.Open(context.TODO(), mediatype.Nil, uuid.New())
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("Open() returned an unexpected error: %s, exected fs.ErrNotExist", err)
		}
	})
}

func TestFileStore_Delete(t *testing.T) {
	t.Parallel()

	store, dir := fileStore(t)
	id := uuid.MustParse("9c04D6a5-4848-4d57-b128-e8cd4090089b")
	path := filepath.Join(dir, "9c", id.String())

	if err := os.Mkdir(filepath.Dir(path), 0770); err != nil {
		t.Fatalf("could not create directory at %q: %s", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte("Test"), store.FileMode); err != nil {
		t.Fatalf("could not create file at %q: %s", path, err)
	}

	ok, err := store.Delete(context.TODO(), mediatype.Nil, id)
	if err != nil {
		t.Errorf("Delete(ctx, nil, %q) returned an unexpected error: %s", id, err)
	}
	if !ok {
		t.Errorf("Delete(ctx, nil, %q) = %t, _, expected %t", id, ok, true)
	}
	// Repeat delete to test idempotency
	ok, err = store.Delete(context.TODO(), mediatype.Nil, id)
	if err != nil {
		t.Errorf("Delete(ctx, nil, %q) [2nd time] returned an unexpected error: %s", id, err)
	}
	if ok {
		t.Errorf("Delete(ctx, nil, %q) = %t, _ [2nd time], expected %t", id, ok, false)
	}
}
