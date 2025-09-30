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
	Users   []internal.User
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
	case http.MethodPost:
		v.HandleAddUser(rw, r)
	default:
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (v *Admin) showAdminForm(rw http.ResponseWriter, r *http.Request) {
	if !v.requireAdminAuth(rw, r) {
		return
	}

	// Get all users from storage
	users, err := v.UserService.GetAllUsers()
	if err != nil {
		data := AdminData{Error: "Failed to load users"}
		v.renderTemplate(rw, data)
		return
	}

	data := AdminData{Users: users}
	v.renderTemplate(rw, data)
}

func (v *Admin) HandleAddUser(rw http.ResponseWriter, r *http.Request) {
	if !v.requireAdminAuth(rw, r) {
		return
	}

	username := r.FormValue("username")
	isAdminStr := r.FormValue("isAdmin")
	isAdmin := isAdminStr == "true"

	password := "changeme123" //later

	_, err := v.UserService.Create(username, password, isAdmin)
	if err != nil {
		users, _ := v.UserService.GetAllUsers()
		data := AdminData{Error: err.Error(), Users: users}
		v.renderTemplate(rw, data)
		return
	}

	http.Redirect(rw, r, "/admin", http.StatusSeeOther)
}

func (v *Admin) HandleRemoveUser(rw http.ResponseWriter, r *http.Request) {
	if !v.requireAdminAuth(rw, r) {
		return
	}

	username := r.FormValue("username")

	err := v.UserService.Delete(username)
	if err != nil {
		users, _ := v.UserService.GetAllUsers()
		data := AdminData{Error: err.Error(), Users: users}
		v.renderTemplate(rw, data)
		return
	}

	http.Redirect(rw, r, "/admin", http.StatusSeeOther)
}

func (v *Admin) renderTemplate(rw http.ResponseWriter, data AdminData) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := v.template.Execute(rw, data); err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
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

func (v *Admin) requireAdminAuth(w http.ResponseWriter, r *http.Request) bool {
	if !v.requireAuth(w, r) {
		return false
	}

	usernameCookie, _ := r.Cookie("username")
	user, _ := v.UserService.Get(usernameCookie.Value)

	if !user.IsAdmin {
		http.Redirect(w, r, "/panel", http.StatusFound)
		return false
	}

	return true
}
