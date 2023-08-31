package upload

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService_GetErrors(t *testing.T) {
	ctx := context.Background()
	svc, data := setupService(t, true)

	errors, total, err := svc.GetErrors(ctx, data.UploadWithErrors, 25, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, errors, 2)
}
