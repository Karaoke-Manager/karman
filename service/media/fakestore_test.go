package media

import (
	"context"
	"io"
	"testing"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

func Test_fakeStore(t *testing.T) {
	id := uuid.New()
	store := NewFakeStore()

	w, err := store.Create(context.TODO(), mediatype.Nil, id)
	if err != nil {
		t.Errorf("Create(ctx, nil, %q) returned an unexpected error: %s", id, err)
		return
	}
	n, err := io.WriteString(w, "Hello World")
	if err != nil {
		t.Errorf("io.WriteString(...) returned an unexpected error: %s", err)
	}
	if n != 11 {
		t.Errorf("io.WriteString(...) = %d, _, expected %d", n, 11)
	}
	if err = w.Close(); err != nil {
		t.Errorf("w.Close() returned an unexpected error: %s", err)
	}

	for i := 0; i < 2; i++ {
		r, err := store.Open(context.TODO(), mediatype.Nil, id)
		if err != nil {
			t.Errorf("[i=%d] Open(ctx, nil, %q) returned an unexpected error: %s", i, id, err)
		}
		data, err := io.ReadAll(r)
		if err != nil {
			t.Errorf("[i=%d] ReadAll() returned an unexpected error: %s", i, err)
		}
		if string(data) != "Hello World" {
			t.Errorf("[i=%d] ReadAll() = %q, _, expected %q", i, data, "Hello World")
		}
		if err = r.Close(); err != nil {
			t.Errorf("[i=%d] r.Close() returned an unexpected error: %s", i, err)
		}
	}
}
