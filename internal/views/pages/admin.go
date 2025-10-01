package pages

import (
	"MediaMTXAuth/internal"
	"MediaMTXAuth/internal/views"
	"MediaMTXAuth/internal/views/handlers"
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed html/admin.html
var AdminPageHTML string

type AdminPage struct {
	*views.Page
}

func NewAdmin(userService internal.UserService) *AdminPage {
	tmpl := template.Must(template.New("pages").Parse(AdminPageHTML))
	return &AdminPage{
		Page: &views.Page{
			UserService: userService,
			Template:    tmpl,
		},
	}
}

func (v *AdminPage) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v.showAdminForm(rw, r)
	default:
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (v *AdminPage) showAdminForm(rw http.ResponseWriter, r *http.Request) {
	if !handlers.RequireAdminAuth(v.Page, rw, r) {
		return
	}

	// Get all users from storage
	users, err := v.UserService.GetAllUsers()
	if err != nil {
		data := views.AdminData{Error: "Failed to load users"}
		v.renderTemplate(rw, data)
		return
	}

	data := views.AdminData{Users: users}
	v.renderTemplate(rw, data)
}

func (v *AdminPage) HandleAddUser(rw http.ResponseWriter, r *http.Request) {
	if !handlers.RequireAdminAuth(v.Page, rw, r) {
		return
	}

	username := r.FormValue("username")
	isAdminStr := r.FormValue("isAdmin")
	isAdmin := isAdminStr == "true"

	password := "changeme123" //later

	_, err := v.UserService.Create(username, password, isAdmin)
	if err != nil {
		users, _ := v.UserService.GetAllUsers()
		data := views.AdminData{Error: err.Error(), Users: users}
		v.renderTemplate(rw, data)
		return
	}

	http.Redirect(rw, r, "/admin", http.StatusSeeOther)
}

func (v *AdminPage) HandleRemoveUser(rw http.ResponseWriter, r *http.Request) {
	if !handlers.RequireAdminAuth(v.Page, rw, r) {
		return
	}

	username := r.FormValue("username")

	err := v.UserService.Delete(username)
	if err != nil {
		users, _ := v.UserService.GetAllUsers()
		data := views.AdminData{Error: err.Error(), Users: users}
		v.renderTemplate(rw, data)
		return
	}

	http.Redirect(rw, r, "/admin", http.StatusSeeOther)
}

func (v *AdminPage) renderTemplate(rw http.ResponseWriter, data views.AdminData) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := v.Template.Execute(rw, data); err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
	}
}
