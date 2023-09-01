package media

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// mockWriter is a writer that checks that a certain string is written.
type mockWriter struct {
	bytes.Buffer
	expect string
}

// Close ends the writing process and checks that the written data equals the expected data.
func (w *mockWriter) Close() error {
	if w.Buffer.String() != w.expect {
		return fmt.Errorf("got file contents %q, expected %q", w.Buffer.String(), w.expect)
	}
	return nil
}

// mockStore is a Store implementation that expects and returns a placeholder for all files.
type mockStore struct {
	placeholder string
}

// NewMockStore returns a new Store implementation that holds file contents in memory.
func NewMockStore(placeholder string) Store {
	return &mockStore{placeholder}
}

// Create opens a writer to a new file.
func (s *mockStore) Create(ctx context.Context, mediaType mediatype.MediaType, id uuid.UUID) (io.WriteCloser, error) {
	return &mockWriter{expect: s.placeholder}, nil
}

// Open returns a new reader to the mocked data.
func (s *mockStore) Open(ctx context.Context, mediaType mediatype.MediaType, id uuid.UUID) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader(s.placeholder)), nil
}
