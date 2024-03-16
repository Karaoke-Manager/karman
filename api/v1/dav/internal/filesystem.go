package internal

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/net/webdav"

	"github.com/Karaoke-Manager/karman/core"
	"github.com/Karaoke-Manager/karman/core/media"
	songsvc "github.com/Karaoke-Manager/karman/core/song"
	"github.com/Karaoke-Manager/karman/model"
)

// flatFS implements a [webdav.FileSystem] that serves songs in a flat hierarchy:
// Each song is contained in a folder that contains the TXT file and the media files.
type flatFS struct {
	logger     *slog.Logger
	songRepo   songsvc.Repository
	songSvc    songsvc.Service
	mediaStore media.Store
}

// NewFlatFS creates a new [webdav.FileSystem] that serves songs in a flat hierarchy:
// The root directory contains a folder for each song which in turn contains all the song's files.
func NewFlatFS(
	logger *slog.Logger,
	songRepo songsvc.Repository,
	songSvc songsvc.Service,
	mediaStore media.Store,
) webdav.FileSystem {
	return &flatFS{logger, songRepo, songSvc, mediaStore}
}

// Mkdir is not allowed.
func (s *flatFS) Mkdir(ctx context.Context, name string, _ fs.FileMode) error {
	s.logger.WarnContext(ctx, "Creating WebDAV directories is not allowed.", "path", name)
	return fs.ErrPermission
}

// RemoveAll is not allowed.
func (s *flatFS) RemoveAll(ctx context.Context, name string) error {
	s.logger.WarnContext(ctx, "Deleting WebDAV files is not allowed.", "path", name)
	return fs.ErrPermission
}

// Rename is not allowed.
func (s *flatFS) Rename(ctx context.Context, name, _ string) error {
	s.logger.WarnContext(ctx, "Renaming WebDAV files is not allowed.", "path", name)
	return fs.ErrPermission
}

// find returns a node value for the specified name, or (nil, fs.ErrNotExist) if no such file exists in s.
func (s *flatFS) find(ctx context.Context, name string) (node, error) {
	name = strings.TrimSuffix(name, "/")
	if name == "" {
		return rootNode{}, nil
	}

	folder, filename, ok := strings.Cut(name, "/")

	if !strings.HasSuffix(folder, ")") {
		s.logger.WarnContext(ctx, "Tried to access unexpected WebDAV song.", "path", name)
		return nil, fs.ErrNotExist
	}
	idx := strings.LastIndex(folder, " (")
	if idx < 0 {
		s.logger.WarnContext(ctx, "Tried to access unexpected WebDAV song.", "path", name)
		return nil, fs.ErrNotExist
	}
	rawUUID := folder[idx+2 : len(folder)-1]
	id, err := uuid.Parse(rawUUID)
	if err != nil {
		s.logger.WarnContext(ctx, "Tried to access unexpected WebDAV song.", "path", name)
		return nil, fs.ErrNotExist
	}

	song, err := s.songRepo.GetSong(ctx, id)
	if errors.Is(err, core.ErrNotFound) {
		return nil, fs.ErrNotExist
	} else if err != nil {
		return nil, err
	}
	s.songSvc.Prepare(ctx, &song)

	if !ok {
		return songNode(song), nil
	}

	if filename == song.TxtFileName {
		return txtNode(song), nil
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
	default:
		s.logger.WarnContext(ctx, "Tried to access unexpected WebDAV song file.", "path", name)
	}
	if file != nil {
		return &mediaNode{
			name: filename,
			file: file,
		}, nil
	}
	return nil, fs.ErrNotExist
}

// Stat returns a [fs.FileInfo] for the specified file name, or an error.
func (s *flatFS) Stat(ctx context.Context, name string) (fs.FileInfo, error) {
	ref, err := s.find(ctx, name)
	if err != nil {
		return nil, err
	}
	return ref.Stat()
}

// OpenFile opens the named file and returns it, or an error.
// As writing files is not allowed, any write flag will cause an error to be returned.
func (s *flatFS) OpenFile(ctx context.Context, name string, flag int, _ fs.FileMode) (webdav.File, error) {
	ref, err := s.find(ctx, name)
	if errors.Is(err, fs.ErrNotExist) && (flag&(os.O_RDWR|os.O_WRONLY) != 0) {
		return nil, fs.ErrPermission
	}
	if err != nil {
		return nil, err
	}
	return ref.Open(ctx, s.songRepo, s.mediaStore, flag)
}

// node represents a single, existing file in the virtual file system.
type node interface {
	// Stat returns a fs.FileInfo for the node, or an error.
	// Usually a node implements [fs.FileInfo] and just returns itself here.
	Stat() (fs.FileInfo, error)
	// Open attempts to open the node using the specified services and flag.
	Open(ctx context.Context, songRepo songsvc.Repository, mediaStore media.Store, flag int) (webdav.File, error)
}
