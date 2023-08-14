package dav

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/net/webdav"
	"gorm.io/gorm"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/media"
	"github.com/Karaoke-Manager/karman/service/song"
)

type node interface {
	Stat() (fs.FileInfo, error)
	Open(ctx context.Context, songSvc song.Service, mediaSvc media.Service, flag int) (webdav.File, error)
}

type flatFS struct {
	songSvc  song.Service
	mediaSvc media.Service
}

func NewFlatFS(songSvc song.Service, mediaSvc media.Service) webdav.FileSystem {
	return &flatFS{songSvc, mediaSvc}
}

func (s *flatFS) Mkdir(_ context.Context, _ string, _ fs.FileMode) error {
	return fs.ErrPermission
}

func (s *flatFS) RemoveAll(_ context.Context, _ string) error {
	return fs.ErrPermission
}

func (s *flatFS) Rename(_ context.Context, _, _ string) error {
	return fs.ErrPermission
}

func (s *flatFS) find(ctx context.Context, name string) (node, error) {
	name = strings.TrimSuffix(name, "/")
	if name == "" {
		return rootNode{}, nil
	}

	folder, filename, ok := strings.Cut(name, "/")

	if !strings.HasSuffix(folder, ")") {
		return nil, fs.ErrNotExist
	}
	idx := strings.LastIndex(folder, " (")
	if idx < 0 {
		return nil, fs.ErrNotExist
	}
	rawUUID := folder[idx+2 : len(folder)-1]
	id, err := uuid.Parse(rawUUID)
	if err != nil {
		return nil, fs.ErrNotExist
	}

	song, err := s.songSvc.GetSong(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fs.ErrNotExist
	} else if err != nil {
		return nil, err
	}

	if !ok {
		return (*songNode)(song), nil
	}

	if filename == song.TxtFileName {
		return (*txtNode)(song), nil
	}

	var file *model.File
	switch filename {
	case song.CoverFileName:
		file = song.CoverFile
	case song.AudioFileName:
		file = song.AudioFile
	case song.VideoFileName:
		file = song.VideoFile
	case song.BackgroundFileName:
		file = song.BackgroundFile
	}
	if file != nil {
		return &mediaNode{
			name: filename,
			file: file,
		}, nil
	}
	return nil, fs.ErrNotExist
}

func (s *flatFS) Stat(ctx context.Context, name string) (fs.FileInfo, error) {
	ref, err := s.find(ctx, name)
	if err != nil {
		return nil, err
	}
	return ref.Stat()
}

func (s *flatFS) OpenFile(ctx context.Context, name string, flag int, perm fs.FileMode) (webdav.File, error) {
	ref, err := s.find(ctx, name)
	if errors.Is(err, fs.ErrNotExist) && (flag&(os.O_RDWR|os.O_WRONLY) != 0) {
		return nil, fs.ErrPermission
	}
	if err != nil {
		return nil, err
	}
	return ref.Open(ctx, s.songSvc, s.mediaSvc, flag)
}
