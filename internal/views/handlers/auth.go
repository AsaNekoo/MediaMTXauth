package handlers

import (
	"MediaMTXAuth/internal/views"
	"net/http"
)

func RequireAuth(page *views.Page, w http.ResponseWriter, r *http.Request) (string, bool) {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return "", false
	}

	usernameCookie, err := r.Cookie("username")
	if err != nil || usernameCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return "", false
	}

	valid, err := page.UserService.VerifySession(usernameCookie.Value, cookie.Value)
	if err != nil || !valid {
		http.Redirect(w, r, "/login", http.StatusFound)
		return "", false
	}

	return usernameCookie.Value, true
}

func RequireAdminAuth(page *views.Page, w http.ResponseWriter, r *http.Request) bool {
	username, authenticated := RequireAuth(page, w, r)
	if !authenticated {
		return false
	}

	user, err := page.UserService.Get(username)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return false
	}

	if !user.IsAdmin {
		http.Redirect(w, r, "/panel", http.StatusFound)
		return false
	}

	return true
}
