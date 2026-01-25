package pages

import (
	"MediaMTXAuth/internal"
	"MediaMTXAuth/internal/views"
	"MediaMTXAuth/internal/views/handlers"
	"crypto/rand"
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed html/admin.html
var AdminPageHTML string

type AdminPage struct {
	*views.Page
	NamespaceService internal.NamespaceService
}

func NewAdmin(userService internal.UserService, namespaceService internal.NamespaceService) *AdminPage {
	tmpl := template.Must(template.New("pages").Parse(AdminPageHTML))
	return &AdminPage{
		Page: &views.Page{
			UserService: userService,
			Template:    tmpl,
		},
		NamespaceService: namespaceService,
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
	username, authenticated := handlers.RequireAdminAuth(v.Page, rw, r)
	if !authenticated {
		return
	}

	currentUser, err := v.UserService.Get(username)
	if err != nil {
		http.Redirect(rw, r, "/login", http.StatusFound)
		return
	}

	// Get all users from storage
	users, err := v.UserService.GetAllUsers()
	if err != nil {
		data := views.AdminData{Error: "Failed to load users", User: *currentUser}
		v.renderTemplate(rw, data)
		return
	}

	// Get all namespaces from storage
	namespaces, err := v.NamespaceService.GetAllNamespaces()
	if err != nil {
		data := views.AdminData{Error: "Failed to load namespaces", Users: users, User: *currentUser}
		v.renderTemplate(rw, data)
		return
	}

	data := views.AdminData{Users: users, Namespaces: namespaces, User: *currentUser}
	v.renderTemplate(rw, data)
}

func (v *AdminPage) HandleAddUser(rw http.ResponseWriter, r *http.Request) {
	var err error
	var currentUser *internal.User
	var data views.AdminData

	usernameAuth, authenticated := handlers.RequireAdminAuth(v.Page, rw, r)
	if !authenticated {
		return
	}

	username := r.FormValue("username")
	namespace := r.FormValue("namespace")
	isAdminStr := r.FormValue("isAdmin")
	isAdmin := isAdminStr == "true"
	password := rand.Text()
	currentUser, err = v.UserService.Get(usernameAuth)

	if currentUser != nil {
		data.User = *currentUser
	}

	if err != nil {
		goto end
	}

	data.Namespaces, err = v.NamespaceService.GetAllNamespaces()

	if err != nil {
		goto end
	}

	_, err = v.UserService.Create(username, password, isAdmin, namespace)

	if err != nil {
		goto end
	}

	data.Users, err = v.UserService.GetAllUsers()

end:
	if err == nil {
		data.TempPassword = password
	} else {
		data.Error = err.Error()
	}

	v.renderTemplate(rw, data)

	return
}

func (v *AdminPage) HandleRemoveUser(rw http.ResponseWriter, r *http.Request) {
	usernameAuth, authenticated := handlers.RequireAdminAuth(v.Page, rw, r)
	if !authenticated {
		return
	}

	username := r.FormValue("username")

	err := v.UserService.Delete(username)
	if err != nil {
		currentUser, _ := v.UserService.Get(usernameAuth)
		users, _ := v.UserService.GetAllUsers()
		namespaces, _ := v.NamespaceService.GetAllNamespaces()
		data := views.AdminData{Error: err.Error(), Users: users, Namespaces: namespaces, User: *currentUser}
		v.renderTemplate(rw, data)
		return
	}

	http.Redirect(rw, r, "/admin", http.StatusSeeOther)
}

func (v *AdminPage) HandleAddNamespace(rw http.ResponseWriter, r *http.Request) {
	usernameAuth, authenticated := handlers.RequireAdminAuth(v.Page, rw, r)
	if !authenticated {
		return
	}

	name := r.FormValue("name")

	_, err := v.NamespaceService.Create(name)
	if err != nil {
		currentUser, _ := v.UserService.Get(usernameAuth)
		users, _ := v.UserService.GetAllUsers()
		namespaces, _ := v.NamespaceService.GetAllNamespaces()
		data := views.AdminData{Error: err.Error(), Users: users, Namespaces: namespaces, User: *currentUser}
		v.renderTemplate(rw, data)
		return
	}

	http.Redirect(rw, r, "/admin", http.StatusSeeOther)
}

func (v *AdminPage) HandleRemoveNamespace(rw http.ResponseWriter, r *http.Request) {
	usernameAuth, authenticated := handlers.RequireAdminAuth(v.Page, rw, r)
	if !authenticated {
		return
	}

	name := r.FormValue("name")

	err := v.NamespaceService.Delete(name)
	if err != nil {
		currentUser, _ := v.UserService.Get(usernameAuth)
		users, _ := v.UserService.GetAllUsers()
		namespaces, _ := v.NamespaceService.GetAllNamespaces()
		data := views.AdminData{Error: err.Error(), Users: users, Namespaces: namespaces, User: *currentUser}
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
