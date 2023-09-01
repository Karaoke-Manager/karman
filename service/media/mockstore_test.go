package media

import (
	"context"
	"io"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

func Test_mockStore_Create(t *testing.T) {
	store := NewMockStore("test")
	id := uuid.New()

	t.Run("success", func(t *testing.T) {
		w, err := store.Create(context.TODO(), mediatype.Nil, id)
		if err != nil {
			t.Errorf("Create(ctx, nil, %q) returned an unexpected error: %s", id, err)
		}
		if _, err = io.WriteString(w, "test"); err != nil {
			t.Errorf("w.Write(...) returned an unexpected error: %s", err)
		}
		if err = w.Close(); err != nil {
			t.Errorf("w.Close() returned an unexpected error: %s", err)
		}
	})
	t.Run("success", func(t *testing.T) {
		w, err := store.Create(context.TODO(), mediatype.Nil, id)
		if err != nil {
			t.Errorf("Create(ctx, nil, %q) returned an unexpected error: %s", id, err)
		}
		if _, err = io.WriteString(w, "hello world"); err != nil {
			t.Errorf("w.Write(...) returned an unexpected error: %s", err)
		}
		if err = w.Close(); err == nil {
			t.Errorf("w.Close() returned no error, but an error was expected")
		}
	})
}

func Test_mockStore_Open(t *testing.T) {
	store := NewMockStore("test")
	id := uuid.New()

	r, err := store.Open(context.TODO(), mediatype.Nil, id)
	if err != nil {
		t.Errorf("Open(ctx, nil, %q) returned an unexpected error: %s", id, err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Errorf("r.Read(...) returned an unexpected error: %s", err)
	}
	if string(data) != "test" {
		t.Errorf("r.Read(...) returned %s, expected %s", string(data), "test")
	}
	if err = r.Close(); err != nil {
		t.Errorf("r.Close() returned an unexpected error: %s", err)
	}
}
