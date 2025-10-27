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

func TestAdminPage(t *testing.T) {
	storage := &memory.Storage{}
	_ = storage.Init()
	userService := services.NewUserService(storage)
	page := NewAdmin(userService)

	t.Run("GET admin page unauthenticated", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		req := httptest.NewRequest("GET", "/admin", nil)
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

	t.Run("GET admin page non-admin user", func(t *testing.T) {
		t.Cleanup(storage.Clear)

		_, _ = userService.Create("user1", "password1", false)
		user, err := userService.Login("user1", "password1")
		if err != nil {
			t.Fatalf("login failed: %v", err)
		}

		req := httptest.NewRequest("GET", "/admin", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: fmt.Sprintf("%d", user.Session.ID)})
		req.AddCookie(&http.Cookie{Name: "username", Value: user.Name})
		rec := httptest.NewRecorder()

		page.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusFound {
			t.Fatalf("expected redirect for non-admin user, got %d", resp.StatusCode)
		}

		loc := resp.Header.Get("Location")
		if loc != "/panel" {
			t.Fatalf("expected redirect to /panel for non-admin, got %s", loc)
		}
	})

	t.Run("GET admin page as admin", func(t *testing.T) {
		t.Cleanup(storage.Clear)

		adminPass, err := userService.CreateDefaultAdminUser()
		if err != nil {
			t.Fatalf("failed to create default admin: %v", err)
		}
		if adminPass == "" {
			adminPass = password
		}

		adminUser, err := userService.Login(username, adminPass)
		if err != nil {
			t.Fatalf("admin login failed: %v", err)
		}

		req := httptest.NewRequest("GET", "/admin", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: fmt.Sprintf("%d", adminUser.Session.ID)})
		req.AddCookie(&http.Cookie{Name: "username", Value: adminUser.Name})
		rec := httptest.NewRecorder()

		page.ServeHTTP(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK for admin, got %d", resp.StatusCode)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "Admin Panel") {
			t.Fatalf("expected admin page content, got body: %s", body)
		}
	})

	t.Run("POST add user as admin", func(t *testing.T) {
		t.Cleanup(storage.Clear)

		adminPass, _ := userService.CreateDefaultAdminUser()
		adminUser, _ := userService.Login(username, adminPass)

		form := url.Values{}
		form.Set("username", "newuser")
		form.Set("isAdmin", "true")

		req := httptest.NewRequest("POST", "/admin/add", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: fmt.Sprintf("%d", adminUser.Session.ID)})
		req.AddCookie(&http.Cookie{Name: "username", Value: adminUser.Name})
		rec := httptest.NewRecorder()

		page.HandleAddUser(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("expected redirect after add, got %d", resp.StatusCode)
		}

		loc := resp.Header.Get("Location")
		if loc != "/admin" {
			t.Fatalf("expected redirect to /admin after add, got %s", loc)
		}

		created, _ := userService.Get("newuser")
		if created == nil {
			t.Fatalf("expected newuser to be created")
		}

		if !created.IsAdmin {
			t.Fatalf("expected newuser to be admin")
		}
	})

	t.Run("POST remove user as admin", func(t *testing.T) {
		t.Cleanup(storage.Clear)
		adminPass, _ := userService.CreateDefaultAdminUser()
		adminUser, _ := userService.Login(username, adminPass)

		_, _ = userService.Create("toremove", "password", false)

		form := url.Values{}
		form.Set("username", "toremove")

		req := httptest.NewRequest("POST", "/admin/remove", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: fmt.Sprintf("%d", adminUser.Session.ID)})
		req.AddCookie(&http.Cookie{Name: "username", Value: adminUser.Name})
		rec := httptest.NewRecorder()

		page.HandleRemoveUser(rec, req)
		resp := rec.Result()

		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("expected redirect after remove, got %d", resp.StatusCode)
		}

		loc := resp.Header.Get("Location")
		if loc != "/admin" {
			t.Fatalf("expected redirect to /admin after remove, got %s", loc)
		}

		removed, _ := userService.Get("toremove")
		if removed != nil {
			t.Fatalf("expected toremove to be deleted")
		}
	})
}
