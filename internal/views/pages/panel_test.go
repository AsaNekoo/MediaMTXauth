package pages

import (
	"MediaMTXAuth/internal/services"
	"MediaMTXAuth/internal/storage/memory"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestPanelPage(t *testing.T) {
	storage := &memory.Storage{}
	_ = storage.Init()
	userService := services.NewUserService(storage)
	page := NewPanel(userService)

	t.Run("GET panel unauthenticated", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		req := httptest.NewRequest("GET", "/panel", nil)
		rec := httptest.NewRecorder()

		page.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusFound {
			t.Fatalf("expected redirect for unauthenticated user, got %d", resp.StatusCode)
		}

		loc := resp.Header.Get("Location")
		if loc != "/login" {
			t.Fatalf("expected redirect to /login, got %s", loc)
		}
	})

	t.Run("GET panel with generated password", func(t *testing.T) {
		t.Cleanup(storage.Clear)

		// Create user with generated password (default behavior of Create)
		// Wait, Create sets IsGenerated=true.
		_, _ = userService.Create("user1", "password", false, "")

		// Login to get session
		loggedInUser, _ := userService.Login("user1", "password")

		req := httptest.NewRequest("GET", "/panel", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: fmt.Sprintf("%d", loggedInUser.Session.ID)})
		req.AddCookie(&http.Cookie{Name: "username", Value: loggedInUser.Name})
		rec := httptest.NewRecorder()

		page.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "Change Password") {
			t.Fatalf("expected change password form, got body: %s", body)
		}
		if strings.Contains(body, "Your StreamKey") {
			t.Fatalf("should not show stream key when password is generated")
		}
	})

	t.Run("GET panel with normal password", func(t *testing.T) {
		t.Cleanup(storage.Clear)

		_, _ = userService.Create("user1", "password", false, "")
		userService.ChangePassword("user1", "newpassword")

		loggedInUser, _ := userService.Login("user1", "newpassword")

		req := httptest.NewRequest("GET", "/panel", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: fmt.Sprintf("%d", loggedInUser.Session.ID)})
		req.AddCookie(&http.Cookie{Name: "username", Value: loggedInUser.Name})
		rec := httptest.NewRecorder()

		page.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		body := rec.Body.String()
		if strings.Contains(body, "Change Password") {
			t.Fatalf("should not show change password form")
		}
		if !strings.Contains(body, "Your StreamKey") {
			t.Fatalf("expected stream key, got body: %s", body)
		}
	})

	t.Run("POST change password", func(t *testing.T) {
		t.Cleanup(storage.Clear)

		_, _ = userService.Create("user1", "password", false, "")
		loggedInUser, _ := userService.Login("user1", "password")

		form := url.Values{}
		form.Set("password", "newpassword")
		req := httptest.NewRequest("POST", "/panel/change_password", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: fmt.Sprintf("%d", loggedInUser.Session.ID)})
		req.AddCookie(&http.Cookie{Name: "username", Value: loggedInUser.Name})
		rec := httptest.NewRecorder()

		page.HandleChangePassword(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("expected redirect after password change, got %d", resp.StatusCode)
		}

		// Verify password changed
		updatedUser, _ := userService.Get("user1")
		if updatedUser.Password.IsGenerated {
			t.Fatalf("password should not be marked as generated anymore")
		}
	})
}
