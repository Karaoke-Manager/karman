package internal

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"golang.org/x/net/webdav"

	"github.com/Karaoke-Manager/karman/service/media"
	songsvc "github.com/Karaoke-Manager/karman/service/song"
)

// rootNode represents the root directory of a flatFS.
type rootNode struct{}

func (n rootNode) Stat() (fs.FileInfo, error) {
	return n, nil
}

func (rootNode) Name() string {
	return ""
}

func (rootNode) Size() int64 {
	return 0
}

func (rootNode) Mode() fs.FileMode {
	return fs.ModeDir | 0555
}

func (rootNode) ModTime() time.Time {
	return time.Now()
}

func (rootNode) IsDir() bool {
	return true
}

func (rootNode) Sys() any {
	return nil
}

func (rootNode) Open(ctx context.Context, songRepo songsvc.Repository, _ media.Store, flag int) (webdav.File, error) {
	if flag&(os.O_RDWR|os.O_WRONLY) != 0 {
		return nil, fs.ErrInvalid
	}
	return &rootDir{ctx: ctx, songRepo: songRepo}, nil
}

// rootDir is a rootNode that has been opened for reading.
type rootDir struct {
	ctx      context.Context
	pos      int64
	songRepo songsvc.Repository
}

func (*rootDir) Close() error {
	webdav.NewMemFS()
	return nil
}

func (*rootDir) Read([]byte) (int, error) {
	return 0, fs.ErrInvalid
}

func (*rootDir) Write([]byte) (n int, err error) {
	return 0, fs.ErrInvalid
}

func (f *rootDir) Seek(offset int64, whence int) (int64, error) {
	fmt.Printf("Seeking %d whence %d\n", offset, whence)
	npos := f.pos
	switch whence {
	case io.SeekStart:
		npos = offset
	case io.SeekCurrent:
		npos += offset
	case io.SeekEnd:
		_, total, err := f.songRepo.FindSongs(f.ctx, 0, 0)
		if err != nil {
			return f.pos, err
		}
		npos = total + offset
	default:
		npos = -1
	}
	if npos < 0 {
		return 0, fs.ErrInvalid
	}
	f.pos = npos
	return f.pos, nil
}

func (f *rootDir) Readdir(count int) ([]fs.FileInfo, error) {
	if count <= 0 {
		count = -1
	}
	// FIXME: We should probably paginate database request for large databases or provide a more hierarchical FS
	songs, total, err := f.songRepo.FindSongs(f.ctx, count, f.pos)
	infos := make([]fs.FileInfo, len(songs))
	for i, song := range songs {
		infos[i] = songNode(song)
	}
	f.pos += int64(len(songs))
	if err != nil {
		return infos, err
	}
	if count > 0 && f.pos >= total {
		return infos, io.EOF
	}
	return infos, nil
}

func (f *rootDir) Stat() (fs.FileInfo, error) {
	return rootNode{}, nil
}
