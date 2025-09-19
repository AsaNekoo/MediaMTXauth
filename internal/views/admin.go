package views

import (
	"MediaMTXAuth/internal"
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed pages/admin.html
var AdminPageHTML string

type Admin struct {
	UserService internal.UserService
	template    *template.Template
}

type AdminData struct {
	Error   string
	Message string
}

func NewAdmin(userService internal.UserService) *Admin {
	tmpl := template.Must(template.New("pages").Parse(AdminPageHTML))
	return &Admin{
		UserService: userService,
		template:    tmpl,
	}
}

func (v *Admin) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v.showAdminForm(rw, r)
	default:
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (v *Admin) showAdminForm(rw http.ResponseWriter, r *http.Request) {
	if !v.requireAuth(rw, r) {
		return
	}

	data := AdminData{}
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := v.template.Execute(rw, data); err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (v *Admin) requireAuth(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("session_id")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return false
	}

	usernameCookie, err := r.Cookie("username")
	if err != nil || usernameCookie.Value == "" {
		http.Redirect(w, r, "/login", http.StatusFound)
		return false
	}

	valid, err := v.UserService.VerifySession(usernameCookie.Value, cookie.Value)
	if err != nil || !valid {
		http.Redirect(w, r, "/login", http.StatusFound)
		return false
	}

	return true
}
