package upload

import (
	"testing"

	"github.com/Karaoke-Manager/karman/test"
)

func setupService(t *testing.T, withData bool) (svc Service, data *test.Dataset) {
	db := test.NewDB(t)
	if withData {
		data = test.NewDataset(db)
	}
	store, _ := fileStore(t)
	svc = NewService(db, store)
	return
}
