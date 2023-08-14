package internal

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"strings"
	"time"

	"codello.dev/ultrastar/txt"
	"golang.org/x/net/webdav"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
)

// txtNode represents the TXT file for a song.
type txtNode model.Song

func (n *txtNode) Stat() (fs.FileInfo, error) {
	return n, nil
}

func (n *txtNode) Name() string {
	return n.TxtFileName
}

func (n *txtNode) Size() int64 {
	b := &strings.Builder{}
	if err := txt.WriteSong(b, n.Song); err != nil {
		// TODO: Log error, should not happen
		return 0
	}
	return int64(b.Len())
}

func (n *txtNode) Mode() fs.FileMode {
	return 0444
}

func (n *txtNode) ModTime() time.Time {
	return n.UpdatedAt
}

func (n *txtNode) IsDir() bool {
	return false
}

func (n *txtNode) Sys() any {
	return nil
}

func (n *txtNode) ContentType(context.Context) (string, error) {
	return "text/plain; charset=utf-8", nil
}

func (n *txtNode) Open(_ context.Context, _ song.Service, _ media.Service, flag int) (webdav.File, error) {
	if flag&(os.O_RDWR|os.O_WRONLY) != 0 {
		return nil, fs.ErrPermission
	}
	b := &bytes.Buffer{}
	_ = txt.WriteSong(b, n.Song)
	return &txtFile{
		song: (*model.Song)(n),
		r:    bytes.NewReader(b.Bytes()),
	}, nil
}

// txtFile represents a txtNode that has been opened for reading.
type txtFile struct {
	song *model.Song
	r    *bytes.Reader
}

func (f *txtFile) Close() error {
	return nil
}

func (f *txtFile) Read(b []byte) (int, error) {
	return f.r.Read(b)
}

func (f *txtFile) Seek(offset int64, whence int) (int64, error) {
	return f.r.Seek(offset, whence)
}

func (f *txtFile) Write([]byte) (n int, err error) {
	return 0, fs.ErrPermission
}

func (f *txtFile) Readdir(int) ([]fs.FileInfo, error) {
	return nil, fs.ErrInvalid
}

func (f *txtFile) Stat() (fs.FileInfo, error) {
	return (*txtNode)(f.song), nil
}
