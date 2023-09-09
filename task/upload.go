package task

import (
	"context"

	"github.com/Karaoke-Manager/karman/model"
)

type UploadQueue struct{}

func (q *UploadQueue) ProcessUpload(ctx context.Context, upload model.Upload) error {
	return nil
}

func (q *UploadQueue) CancelUploadProcessing() {

}
