package media

import (
	"bytes"
	"context"
	"io"
	"io/fs"

	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

// closeBuffer adds a Close function to bytes.Buffer.
type closeBuffer struct {
	bytes.Buffer
}

// Close is a noop.
func (b *closeBuffer) Close() error {
	return nil
}

// fakeStore is an in-memory Store implementation.
type fakeStore struct {
	files map[uuid.UUID]*closeBuffer
}

// NewFakeStore returns a new Store implementation that holds file contents in memory.
func NewFakeStore() Store {
	return &fakeStore{make(map[uuid.UUID]*closeBuffer)}
}

// Create opens a writer to a new file.
func (s *fakeStore) Create(ctx context.Context, mediaType mediatype.MediaType, id uuid.UUID) (io.WriteCloser, error) {
	buf := &closeBuffer{bytes.Buffer{}}
	s.files[id] = buf
	return buf, nil
}

// Open returns a new reader to the data of the file.
func (s *fakeStore) Open(ctx context.Context, mediaType mediatype.MediaType, id uuid.UUID) (io.ReadCloser, error) {
	buf, ok := s.files[id]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
