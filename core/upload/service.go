package upload

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strings"

	"codello.dev/ultrastar/txt"
	"github.com/google/uuid"

	"github.com/Karaoke-Manager/karman/core/song"
	"github.com/Karaoke-Manager/karman/model"
)

type service struct {
	logger *slog.Logger
	repo   Repository
	store  Store

	songRepo    song.Repository
	songService song.Service
}

// NewService creates a new Service instance using the supplied repo and store.
func NewService(logger *slog.Logger, repo Repository, store Store, songRepo song.Repository, songService song.Service) Service {
	return &service{logger, repo, store, songRepo, songService}
}

func (s *service) ProcessUpload(ctx context.Context, id uuid.UUID) error {
	// TODO: Testing!!!
	// FIXME:
	// This should probably be completely rewritten...
	upload, err := s.repo.GetUpload(ctx, id)
	if err != nil {
		return err
	}
	s.logger.InfoContext(ctx, "Beginning to process upload.", "uuid", id)
	upload.State = model.UploadStateProcessing
	upload.SongsTotal = -1
	upload.SongsProcessed = 0
	// TODO: Logging
	if err = s.repo.UpdateUpload(ctx, &upload); err != nil {
		return err
	}
	if _, err = s.repo.ClearErrors(ctx, &upload); err != nil {
		return err
	}
	if _, err = s.repo.ClearSongs(ctx, &upload); err != nil {
		return err
	}
	if _, err = s.repo.ClearFiles(ctx, &upload); err != nil {
		return err
	}

	var songFiles []string

	uploadFiles := s.store.FS(ctx, upload.UUID)
	err = fs.WalkDir(uploadFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return s.repo.CreateError(ctx, &upload, model.UploadProcessingError{File: path, Message: fmt.Sprintf("could not list files: %s", err)})
		}
		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(d.Name()))
		if ext == ".txt" || ext == ".txd" {
			songFiles = append(songFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	upload.SongsTotal = len(songFiles)
	upload.SongsProcessed = 0
	if err = s.repo.UpdateUpload(ctx, &upload); err != nil {
		return err
	}

	for _, path := range songFiles {
		ok, err := s.processFile(ctx, &upload, path)
		if err != nil {
			return err
		}
		if !ok {
			upload.Errors++
		}
		upload.SongsProcessed++
		if err = s.repo.UpdateUpload(ctx, &upload); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) processFile(ctx context.Context, upload *model.Upload, path string) (_ bool, err error) {
	f, err := s.store.Open(ctx, upload.UUID, path)
	if err != nil {
		err = s.repo.CreateError(ctx, upload, model.UploadProcessingError{File: path, Message: "could not open file"})
		if err != nil {
			return false, err
		}
	}
	defer func() {
		if cErr := f.Close(); cErr != nil {
			cErr = s.repo.CreateError(ctx, upload, model.UploadProcessingError{File: path, Message: fmt.Sprintf("could not close file: %s", err)})
			if err == nil {
				err = cErr
			}
		}
	}()
	rawSong, err := txt.NewReader(f).ReadSong()
	if err != nil {
		return false, s.repo.CreateError(ctx, upload, model.UploadProcessingError{File: path, Message: fmt.Sprintf("could not parse song: %s", err)})
	}
	sng := model.Song{
		Song:        rawSong,
		InUpload:    true,
		TxtFileName: filepath.Base(path),
	}
	s.songService.ParseArtists(ctx, &sng)
	if err = s.songRepo.CreateSong(ctx, &sng); err != nil {
		return false, s.repo.CreateError(ctx, upload, model.UploadProcessingError{File: path, Message: fmt.Sprintf("could not save song to database: %s", err)})
	}
	// TODO: Save media files
	return true, nil
}

// DeleteUpload deletes an upload from the database and file storage.
func (s *service) DeleteUpload(ctx context.Context, id uuid.UUID) error {
	err := s.store.Delete(ctx, id, ".")
	if err != nil {
		return err
	}
	// The DB schema takes care of deleting associated songs and files.
	_, err = s.repo.DeleteUpload(ctx, id)
	return err
}
