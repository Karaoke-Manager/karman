package media

import (
	"context"
	"strings"
	"testing"

	"github.com/Karaoke-Manager/karman/pkg/mediatype"
)

func TestFakeService_StoreFile(t *testing.T) {
	t.Parallel()

	repo := NewFakeRepository()
	svc := NewFakeService(repo)
	file, err := svc.StoreFile(context.TODO(), mediatype.AudioMPEG, strings.NewReader("Hello World"))
	if err != nil {
		t.Errorf("StoreFile(...) returned an unexpected error: %s", err)
	}
	if !file.Type.Equals(mediatype.AudioMPEG) {
		t.Errorf("StoreFile(...) produced file.Type = %s, expected %s", file.Type, mediatype.AudioMPEG)
	}
}
