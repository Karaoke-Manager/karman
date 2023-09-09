package internal

import (
	"context"
	"io"
	"io/fs"
	"os"
	"time"

	"golang.org/x/net/webdav"

	"github.com/Karaoke-Manager/karman/core/media"
	"github.com/Karaoke-Manager/karman/core/song"
	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/pkg/mediatype"
	"github.com/Karaoke-Manager/karman/pkg/streamio"
)

// mediaNode is a node representing a media file.
type mediaNode struct {
	name string // name of the file
	file *model.File
}

func (n *mediaNode) Stat() (fs.FileInfo, error) {
	return n, nil
}

func (n *mediaNode) Name() string {
	return n.name
}

func (n *mediaNode) Size() int64 {
	return n.file.Size
}

func (*mediaNode) Mode() fs.FileMode {
	return 0444
}

func (n *mediaNode) ModTime() time.Time {
	return n.file.UpdatedAt
}

func (*mediaNode) IsDir() bool {
	return false
}

func (*mediaNode) Sys() any {
	return nil
}

func (n *mediaNode) ContentType(context.Context) (string, error) {
	return n.file.Type.String(), nil
}

func (n *mediaNode) ETag(context.Context) (string, error) {
	// TODO: Support for ETags
	return "", webdav.ErrNotImplemented
}

func (n *mediaNode) Open(ctx context.Context, _ song.Repository, mediaStore media.Store, flag int) (webdav.File, error) {
	if flag&(os.O_RDWR|os.O_WRONLY) != 0 {
		return nil, fs.ErrPermission
	}
	// TODO: determine mediatype from path and file extension.
	r, err := mediaStore.Open(ctx, mediatype.Nil, n.file.UUID)
	if err != nil {
		return nil, err
	}
	return &mediaFile{
		info: n,
		rd:   r,
		r:    streamio.NewBufferedReadSeeker(r, 0),
	}, nil
}

// mediaFile implements a mediaNode that has been opened for reading.
type mediaFile struct {
	info *mediaNode
	rd   io.ReadCloser // underlying reader
	r    io.ReadSeeker // seekable version of rd
}

func (f *mediaFile) Close() error {
	return f.rd.Close()
}

func (f *mediaFile) Read(b []byte) (int, error) {
	return f.r.Read(b)
}

func (f *mediaFile) Seek(offset int64, whence int) (int64, error) {
	return f.r.Seek(offset, whence)
}

func (f *mediaFile) Write([]byte) (n int, err error) {
	return 0, fs.ErrPermission
}

func (f *mediaFile) Readdir(int) ([]fs.FileInfo, error) {
	return nil, fs.ErrInvalid
}

func (f *mediaFile) Stat() (fs.FileInfo, error) {
	return f.info, nil
}
