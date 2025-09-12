package views

import (
	"MediaMTXAuth/internal"
	_ "embed"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

//go:embed pages/login.html
var LoginPageHTML string

type Login struct {
	UserService internal.UserService
	template    *template.Template
}

type LoginData struct {
	Error   string
	Message string
}

func NewLogin(userService internal.UserService) *Login {
	tmpl := template.Must(template.New("pages").Parse(LoginPageHTML))
	return &Login{
		UserService: userService,
		template:    tmpl,
	}
}

func (v *Login) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v.showLoginForm(rw, r)
	case http.MethodPost:
		v.handleLogin(rw, r)
	default:
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (v *Login) showLoginForm(rw http.ResponseWriter, r *http.Request) {
	data := LoginData{}

	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := v.template.Execute(rw, data); err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (v *Login) renderWithError(rw http.ResponseWriter, r *http.Request, errorMsg string) {
	data := LoginData{Error: errorMsg}
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	rw.WriteHeader(http.StatusUnauthorized)
	if err := v.template.Execute(rw, data); err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
	}
}

func (v *Login) handleLogin(rw http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		v.renderWithError(rw, r, "Username and password are required")
		return
	}

	user, err := v.UserService.Login(username, password)
	if err != nil {
		v.renderWithError(rw, r, "Invalid credentials")
		return
	}

	sessionCookie := &http.Cookie{
		Name:     "session_id",
		Value:    strconv.FormatUint(user.Session.ID, 10),
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(15 * time.Minute.Seconds()),
	}
	http.SetCookie(rw, sessionCookie)

	usernameCookie := &http.Cookie{
		Name:     "username",
		Value:    user.Name,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(15 * time.Minute.Seconds()),
	}
	http.SetCookie(rw, usernameCookie)

	http.Redirect(rw, r, "/admin", http.StatusSeeOther)
}
