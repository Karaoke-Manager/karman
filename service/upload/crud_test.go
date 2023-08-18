package upload

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Karaoke-Manager/karman/model"
	"github.com/Karaoke-Manager/karman/service/common"
)

func TestService_CreateUpload(t *testing.T) {
	ctx := context.Background()
	svc, _ := setupService(t, false)

	upload, err := svc.CreateUpload(ctx)
	if assert.NoError(t, err) {
		assert.NotEmpty(t, upload.UUID)
		assert.Equal(t, model.UploadStateOpen, upload.State)
	}
}

func TestService_GetUpload(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	t.Run("empty", func(t *testing.T) {
		_, err := svc.GetUpload(ctx, data.AbsentUploadUUID)
		assert.ErrorIs(t, err, common.ErrNotFound)
	})

	cases := map[string]struct {
		upload *model.Upload
		state  model.UploadState
	}{
		"open":       {data.OpenUpload, model.UploadStateOpen},
		"pending":    {data.PendingUpload, model.UploadStatePending},
		"processing": {data.ProcessingUpload, model.UploadStateProcessing},
		"done":       {data.UploadWithSongs, model.UploadStateDone},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			upload, err := svc.GetUpload(ctx, c.upload.UUID)
			if assert.NoError(t, err) {
				assert.Equal(t, c.state, upload.State)
			}
		})
	}

	t.Run("errors", func(t *testing.T) {
		upload, err := svc.GetUpload(ctx, data.UploadWithErrors.UUID)
		if assert.NoError(t, err) {
			assert.Equal(t, 2, upload.Errors)
		}
	})
}

func TestService_FindUploads(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	uploads, total, err := svc.FindUploads(ctx, int(data.TotalUploads), 0)
	if assert.NoError(t, err) {
		assert.Equal(t, data.TotalUploads, total)
		assert.Len(t, uploads, int(data.TotalUploads))
	}
}

func TestService_DeleteUpload(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	t.Run("success", func(t *testing.T) {
		err := svc.DeleteUpload(ctx, data.OpenUpload.UUID)
		assert.NoError(t, err)
		_, err = svc.GetUpload(ctx, data.OpenUpload.UUID)
		require.ErrorIs(t, err, common.ErrNotFound)
	})

	t.Run("already absent", func(t *testing.T) {
		err := svc.DeleteUpload(ctx, data.AbsentUploadUUID)
		assert.NoError(t, err)
	})
}
