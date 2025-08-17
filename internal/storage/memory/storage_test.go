package memory

import (
	"MediaMTXAuth/internal/storage"
	"testing"
)

func TestStorage(t *testing.T) {
	var s storage.Storage = &Storage{}

	defer s.Close()

	storage.XTestStorage(t, s)
}
