package auth

import (
	"MediaMTXAuth/internal/services"
	"MediaMTXAuth/internal/storage/memory"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuth_ServeHTTP(t *testing.T) {
	storage := &memory.Storage{}
	nsService := services.NewNamespaceService(storage)
	userService := services.NewUserService(storage)
	auth := New(userService, nsService)
	ns, err := nsService.Create("test_namespace")

	if err != nil {
		t.Fatal(err)
	}

	u, err := userService.Create("test", "testtest", false, ns.Name)

	if err != nil {
		t.Fatal(err)
	}

	t.Run("ok", func(t *testing.T) {
		r := Request{
			IP:     "127.0.0.1",
			Path:   fmt.Sprintf("/%s/%s", ns.Name, u.Name),
			Query:  fmt.Sprintf("key=%s", u.StreamKey),
			Action: "publish",
		}
		buf := &bytes.Buffer{}
		err := json.NewEncoder(buf).Encode(&r)

		if err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest("POST", "/api/auth", buf)
		w := httptest.NewRecorder()

		auth.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
	})

	// TODO: cover other cases
}
