package bolt

import (
	"MediaMTXAuth/internal/storage"
	"path"
	"testing"
)

func TestStorage(t *testing.T) {
	s, err := New(path.Join(t.TempDir(), "test.db"))

	if err != nil {
		t.Fatal(err)
	}

	defer s.Close()

	storage.XTestStorage(t, s)
}
