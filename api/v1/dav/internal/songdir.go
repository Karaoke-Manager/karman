package internal

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"

	"golang.org/x/net/webdav"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
)

// songNode represents the directory for a song.
type songNode model.Song

func (n *songNode) Stat() (fs.FileInfo, error) {
	return n, nil
}

func (n *songNode) Name() string {
	return fmt.Sprintf("%s - %s (%s)", n.Artist, n.Title, n.UUID)
}

func (n *songNode) Size() int64 {
	return 0
}

func (n *songNode) Mode() fs.FileMode {
	return fs.ModeDir | 0555
}

func (n *songNode) ModTime() time.Time {
	return n.UpdatedAt
}

func (n *songNode) IsDir() bool {
	return true
}

func (n *songNode) Sys() any {
	return nil
}

func (n *songNode) Open(_ context.Context, _ song.Service, _ media.Service, flag int) (webdav.File, error) {
	if flag&(os.O_RDWR|os.O_WRONLY) != 0 {
		return nil, fs.ErrInvalid
	}
	return &songDir{
		pos:  0,
		song: (*model.Song)(n),
	}, nil
}

// songDir represents a songNode that has been opened for reading.
type songDir struct {
	pos  int64
	song *model.Song
}

func (*songDir) Close() error {
	return nil
}

func (*songDir) Read([]byte) (int, error) {
	return 0, fs.ErrInvalid
}

func (*songDir) Write([]byte) (n int, err error) {
	return 0, fs.ErrInvalid
}

func (f *songDir) Seek(offset int64, whence int) (int64, error) {
	npos := f.pos
	switch whence {
	case io.SeekStart:
		npos = offset
	case io.SeekCurrent:
		npos += offset
	case io.SeekEnd:
		total := int64(1) // txt file always exists
		if f.song.AudioFile != nil {
			total++
		}
		if f.song.CoverFile != nil {
			total++
		}
		if f.song.VideoFile != nil {
			total++
		}
		if f.song.BackgroundFile != nil {
			total++
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

func (f *songDir) Readdir(count int) ([]fs.FileInfo, error) {
	if count <= 0 {
		count = -1
	}
	infos := make([]fs.FileInfo, 0, 5)
	fileIndex := int64(0)
	if f.pos == fileIndex && len(infos) != count {
		infos = append(infos, (*txtNode)(f.song))
		f.pos++
	}
	fileIndex++
	if f.song.AudioFile != nil {
		if f.pos == fileIndex && len(infos) != count {
			infos = append(infos, &mediaNode{
				name: f.song.AudioFileName,
				file: f.song.AudioFile,
			})
			f.pos++
		}
		fileIndex++
	}
	if f.song.CoverFile != nil {
		if f.pos == fileIndex && len(infos) != count {
			infos = append(infos, &mediaNode{
				name: f.song.CoverFileName,
				file: f.song.CoverFile,
			})
			f.pos++
		}
		fileIndex++
	}
	if f.song.VideoFile != nil {
		if f.pos == fileIndex && len(infos) != count {
			infos = append(infos, &mediaNode{
				name: f.song.VideoFileName,
				file: f.song.VideoFile,
			})
			f.pos++
		}
		fileIndex++
	}
	if f.song.BackgroundFile != nil {
		if f.pos == fileIndex && len(infos) != count {
			infos = append(infos, &mediaNode{
				name: f.song.BackgroundFileName,
				file: f.song.BackgroundFile,
			})
			f.pos++
		}
		fileIndex++
	}
	if count > 0 && f.pos >= fileIndex {
		return infos, io.EOF
	}
	return infos, nil
}

func (f *songDir) Stat() (fs.FileInfo, error) {
	return (*songNode)(f.song), nil
}
