package upload

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// fileStore creates a new FileStore using a temporary directory.
// The directory path is returned as the second parameter.
func fileStore(t *testing.T) (*FileStore, string) {
	dir, err := os.MkdirTemp("", "karman-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp(...) returned an unexpected error: %s", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})
	store, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore(%q) returned an unexpected error: %s", dir, err)
	}
	return store, dir
}

func TestNewFileStore(t *testing.T) {
	t.Parallel()

	dir, err := os.MkdirTemp("", "karman-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp(...) returned an unexpected error: %s", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	t.Run("missing root directory", func(t *testing.T) {
		_, err = NewFileStore(filepath.Join(dir, "test1"))
		if err == nil {
			t.Errorf("NewFileStore(<missing>) did not return an error, but an error was expected")
		}
	})
	t.Run("file root", func(t *testing.T) {
		path := filepath.Join(dir, "test2")
		if err = os.WriteFile(path, []byte("Hello"), 0600); err != nil {
			t.Fatalf("WriteFile(%q, %q, 0600) returned an unexpected error: %s", path, "Hello", err)
		}
		_, err = NewFileStore(path)
		if err == nil {
			t.Errorf("NewFileStore(<file>) did not return an error, but an error was expected")
		}
	})
	t.Run("success", func(t *testing.T) {
		path := filepath.Join(dir, "test3")
		if err = os.Mkdir(path, 0770); err != nil {
			t.Fatalf("Mkdir(%q, 0770) returned an unexpected error: %s", path, err)
		}
		_, err = NewFileStore(path)
		if err != nil {
			t.Fatalf("NewFileStore(%q) returned an unexpected error: %s", path, err)
		}
	})
}

func TestFileStore_Create(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")

	t.Run("new file", func(t *testing.T) {
		store, dir := fileStore(t)
		w, err := store.Create(ctx, id, "foobar.txt")
		if err != nil {
			t.Errorf("Create(ctx, %q, %q) returned an unexpected error: %s", id, "foobar.txt", err)
			return
		}
		n, err := io.WriteString(w, "Hello World")
		if err != nil {
			t.Errorf("WriteString(...) returned an unexpected error: %s", err)
			return
		}
		if n != 11 {
			t.Errorf("WriteString(...) write %d bytes, expected %d", n, 11)
		}
		if err = w.Close(); err != nil {
			t.Errorf("Close() returned an unexpected error: %s", err)
		}

		stat, err := os.Stat(filepath.Join(dir, id.String(), "foobar.txt"))
		if err != nil {
			t.Fatalf("os.Stat(...) returned an unexpected error: %s", err)
		}
		if stat.IsDir() {
			t.Errorf("os.Stat(...) indicates a directory, expected file")
		}
		if stat.Size() != 11 {
			t.Errorf("os.Stat(...) indicates a file size of %d, expected %d", stat.Size(), 11)
		}
	})

	t.Run("overwrite file", func(t *testing.T) {
		store, dir := fileStore(t)
		dir = filepath.Join(dir, id.String())
		name := filepath.Join(dir, "file.txt")

		if err := os.Mkdir(dir, store.DirMode); err != nil {
			t.Fatalf("os.Mkdir(%q, ...) returned an unexpected error: %s", dir, err)
		}
		if err := os.WriteFile(name, []byte("Test"), store.FileMode); err != nil {
			t.Fatalf("os.WriteFile(%q, ...) returned an unexpected error: %s", name, err)
		}

		w, err := store.Create(ctx, id, "file.txt")
		if err != nil {
			t.Errorf("Create(ctx, %q, %q) returned an unexpected error: %s", id, "file.txt", err)
			return
		}
		n, err := io.WriteString(w, "Another\nValue")
		if err != nil {
			t.Errorf("WriteString(...) returned an unexpected error: %s", err)
		}
		if n != 13 {
			t.Errorf("WriteString(...) wrote %d bytes, expected %d", n, 13)
		}
		if err = w.Close(); err != nil {
			t.Errorf("Close() returned an unexpected error: %s", err)
			return
		}

		stat, err := os.Stat(name)
		if err != nil {
			t.Fatalf("os.Stat(%q) returned an unexpected error: %s", name, err)
		}
		if stat.IsDir() {
			t.Errorf("Stat(%q) indicates a directory, expected file", name)
		}
		if stat.Size() != 13 {
			t.Errorf("Stat(%q) indicates a size of %d, expected %d", name, stat.Size(), 13)
		}
	})

	t.Run("intermediate folders", func(t *testing.T) {
		store, dir := fileStore(t)

		w, err := store.Create(ctx, id, "my/foo/bar.txt")
		if err != nil {
			t.Errorf("Create(ctx, %q, %q) returned an unexpected error: %s", id, "my/foo/bar.txt", err)
			return
		}
		if err = w.Close(); err != nil {
			t.Errorf("Close() returned an unexpected error: %s", err)
			return
		}
		stat, err := os.Stat(filepath.Join(dir, id.String(), "my/foo/bar.txt"))
		if err != nil {
			t.Fatalf("os.Stat(...) returned an unexpected error: %s", err)
		}
		if stat.IsDir() {
			t.Errorf("f.Stat() indicates a directory, expected file")
		}
		if stat.Size() != 0 {
			t.Errorf("Stat() indicates a size of %d, expected %d", stat.Size(), 0)
		}
		if stat.Mode() != store.FileMode {
			t.Errorf("f.Stat() indicates mode %#o, expected %#o", stat.Mode(), store.FileMode)
		}
		dirStat, err := os.Stat(filepath.Join(dir, id.String(), "my/foo"))
		if err != nil {
			t.Fatalf("os.Stat(...) returned an unexpected error: %s", err)
		}
		if !dirStat.IsDir() {
			t.Errorf("dir.Stat() indicates a file, expected directory")
		}
		if dirStat.Mode().Perm() != store.DirMode {
			t.Errorf("dir.Stat() indicates mode %#o, expected %#o", dirStat.Mode().Perm(), store.DirMode)
		}
	})

	t.Run("root", func(t *testing.T) {
		store, _ := fileStore(t)
		_, err := store.Create(ctx, id, ".")
		if err == nil {
			t.Errorf("Create(ctx, %q, %q) did not return an error, but an error was expected", id, ".")
		}
	})
}

func TestFileStore_Stat(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")

	t.Run("present", func(t *testing.T) {
		store, dir := fileStore(t)
		if err := os.MkdirAll(filepath.Join(dir, id.String()), store.DirMode); err != nil {
			t.Fatalf("os.MkdirAll(...) returned an unexpected error: %s", err)
		}
		if err := os.WriteFile(filepath.Join(dir, id.String(), "foobar.txt"), []byte("Hello World"), store.FileMode); err != nil {
			t.Fatalf("os.WriteFile(...) returned an unexpected error: %s", err)
		}

		stat, err := store.Stat(ctx, id, "foobar.txt")
		if err != nil {
			t.Errorf("Stat(ctx, %q, %q) returned an unexpected error: %s", id, "foobar.txt", err)
			return
		}
		if stat.IsDir() {
			t.Errorf("Stat(ctx, %q, %q) indicates a directory, expected file", id, "foobar.txt")
		}
		if stat.Size() != 11 {
			t.Errorf("Stat(ctx, %q, %q) indicates a file size of %d, expected %d", id, "foobar.txt", stat.Size(), 11)
		}
	})

	t.Run("absent", func(t *testing.T) {
		store, _ := fileStore(t)
		_, err := store.Stat(ctx, id, "hello.txt")
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("Stat(ctx, %q, %q) returned an unexpected error: %s, expected fs.ErrNotExist", id, "hello.txt", err)
		}
	})

	t.Run("root", func(t *testing.T) {
		store, _ := fileStore(t)
		stat, err := store.Stat(ctx, id, ".")
		if err != nil {
			t.Errorf("Stat(ctx, %q, %q) returned an unexpected error: %s", id, ".", err)
			return
		}
		if !stat.IsDir() {
			t.Errorf("Stat(ctx, %q, %q) indicates a file, expected directory", id, ".")
		}
		if stat.Name() != id.String() {
			t.Errorf("Stat(ctx, %q, %q) indicates a filename of %q, expected %q", id, ".", stat.Name(), id.String())
		}
	})
}

func TestFileStore_Open(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")

	t.Run("file", func(t *testing.T) {
		store, dir := fileStore(t)
		if err := os.MkdirAll(filepath.Join(dir, id.String()), store.DirMode); err != nil {
			t.Fatalf("os.MkdirAll(...) returned an unexpected error: %s", err)
		}
		if err := os.WriteFile(filepath.Join(dir, id.String(), "test.txt"), []byte("Foobar"), store.FileMode); err != nil {
			t.Fatalf("os.WriteFile(...) returned an unexpected error: %s", err)
		}

		f, err := store.Open(ctx, id, "test.txt")
		if err != nil {
			t.Errorf("Open(ctx, %q, %q) returned an unexpected error: %s", id, "test.txt", err)
			return
		}
		data, err := io.ReadAll(f)
		if err != nil {
			t.Errorf("Read returned an unexpected error: %s", err)
		}
		if string(data) != "Foobar" {
			t.Errorf("Read returned data %q, expected %q", data, "Foobar")
		}
		if err = f.Close(); err != nil {
			t.Errorf("Close() returned an unexpected error: %s", err)
		}
	})

	t.Run("absent", func(t *testing.T) {
		store, _ := fileStore(t)
		_, err := store.Open(ctx, id, "foobar")
		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("Open(ctx, %q, %q) returned an unexpected error: %s, expected fs.ErrNotExist", id, "foobar", err)
		}
	})

	t.Run("folder", func(t *testing.T) {
		store, dir := fileStore(t)
		if err := os.MkdirAll(filepath.Join(dir, id.String(), "test", "folder"), store.DirMode); err != nil {
			t.Fatalf("os.MkdirAll(...) returned an unexpected error: %s", err)
		}

		f, err := store.Open(ctx, id, "test")
		if err != nil {
			t.Errorf("Open(ctx, %q, %q) returned an unexpected error: %s", id, "test", err)
			return
		}
		stat, err := f.Stat()
		if err != nil {
			t.Errorf("f.Stat() returned an unexpected error: %s", err)
		}
		if !stat.IsDir() {
			t.Errorf("f.Stat() indicates a file, expected directory")
		}
		if _, ok := f.(Dir); !ok {
			t.Errorf("f does not implement interface Dir, expected f to implement Dir")
		}
	})
}

func TestFileStore_Delete(t *testing.T) {
	t.Parallel()

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
			if err := os.MkdirAll(filepath.Join(dir, id.String(), "empty"), store.DirMode); err != nil {
				t.Fatalf("os.MkdirAll(...) returned an unexpected error: %s", err)
			}
			if err := os.MkdirAll(filepath.Join(dir, id.String(), "foobar"), store.DirMode); err != nil {
				t.Fatalf("os.MkdirAll(...) returned an unexpected error: %s", err)
			}
			if err := os.WriteFile(filepath.Join(dir, id.String(), "foobar/test.txt"), []byte("Foobar"), store.FileMode); err != nil {
				t.Fatalf("os.WriteFile(...) returned an unexpected error: %s", err)
			}

			if err := store.Delete(ctx, id, c.path); err != nil {
				t.Errorf("Delete(ctx, %q, %q) returned an unexpected error: %s", id, c.path, err)
			}
			_, err := os.Stat(filepath.Join(dir, id.String(), c.path))
			if !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("Delete(ctx, %q, %q) [2nd time] returned an unexpected error: %s, expected fs.ErrNotExist", id, c.path, err)
			}
		})
	}
}

func TestFolderDir_Marker(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")

	store, _ := NewFileStore("./testdata")
	f, err := store.Open(ctx, id, ".")
	if err != nil {
		t.Errorf("Open(ctx, %q, %q) returned an unexpected error: %s", id, ".", err)
		return
	}
	dir := f.(Dir)

	if dir.Marker() != "" {
		t.Errorf("dir.Marker() = %q, expected %q", dir.Marker(), "")
	}
	_, err = dir.Readdir(2)
	if err != nil {
		t.Errorf("dir.Readdir(2) returned an unexpected error: %s", err)
	}
	if dir.Marker() != "dir2" {
		t.Errorf("dir.Marker() = %q, expected %q", dir.Marker(), "dir2")
	}
}

func TestFolderDir_SkipTo(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	id := uuid.MustParse("e4d7ec99-77e0-4595-815a-18f3811c1b9d")
	store, _ := NewFileStore("./testdata")

	f, err := store.Open(ctx, id, ".")
	if err != nil {
		t.Errorf("Open(ctx, %q, %q) returned an unexpected error: %s", id, ".", err)
		return
	}
	dir := f.(Dir)

	if err := dir.SkipTo("def"); err != nil {
		t.Errorf("dir.SkipTo(%q) returned an unexpected error: %s", "def", err)
	}
	if err := dir.SkipTo("abc"); err == nil {
		t.Errorf("dir.SkipTo(%q) did not return an error, but an error was expected", "abc")
	}
}

func TestFolderDir_ReadDir(t *testing.T) {
	t.Parallel()

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
			if err != nil {
				t.Errorf("Open(ctx, %q, %q) returned an unexpected error: %s", id, ".", err)
				return
			}
			dir := f.(Dir)

			if err := dir.SkipTo(c.marker); err != nil {
				t.Errorf("dir.SkipTo(%q) returned an unexpected error: %s", c.marker, err)
			}
			entries, err := dir.Readdir(c.n)
			if err != nil {
				t.Errorf("dir.Readdir(%d) returned an unexpected error: %s", c.n, err)
			}
			if len(entries) != c.len {
				t.Errorf("dir.Readdir(%d) returned %d entries, expected %d", c.n, len(entries), c.len)
			}
			if dir.Marker() != c.newMarker {
				t.Errorf("dir.Marker() = %q, expected %q", dir.Marker(), c.newMarker)
			}
		})
	}
}
