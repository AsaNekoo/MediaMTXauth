package pages

import (
	"MediaMTXAuth/internal/services"
	"MediaMTXAuth/internal/storage/memory"
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const username = "admin"
const password = "admin"

func TestLoginView(t *testing.T) {
	storage := &memory.Storage{}
	_ = storage.Init()
	userService := services.NewUserService(storage)
	view := NewLogin(userService)

	t.Run("GET login page", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		req := httptest.NewRequest("GET", "/login", nil)
		rec := httptest.NewRecorder()

		view.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "<form") {
			t.Errorf("Expected login form in response")
		}
	})

	t.Run("POST invalid credentials", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, _ = userService.CreateDefaultAdminUser()

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("username", username)
		_ = writer.WriteField("password", "wrongpass")
		writer.Close()

		req := httptest.NewRequest("POST", "/login", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		view.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", resp.StatusCode)
		}

		if !strings.Contains(rec.Body.String(), "error") {
			t.Errorf("Expected error message in response")
		}
	})

	t.Run("POST valid credentials", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, _ = userService.CreateDefaultAdminUser()

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("username", username)
		_ = writer.WriteField("password", password)
		writer.Close()

		req := httptest.NewRequest("POST", "/login", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		view.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("Expected redirect (303), got %d", resp.StatusCode)
		}
		location := resp.Header.Get("Location")
		if location != "/admin" {
			t.Errorf("Expected redirect to /admin, got %s", location)
		}
	})

	t.Run("POST non-existent user", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		_, _ = userService.CreateDefaultAdminUser()

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("username", "nouser")
		_ = writer.WriteField("password", "any")
		writer.Close()

		req := httptest.NewRequest("POST", "/login", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		view.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("Expected status 401 for non-existent user, got %d", resp.StatusCode)
		}
		if !strings.Contains(rec.Body.String(), "error") {
			t.Errorf("Expected error message for non-existent user")
		}
	})
}
