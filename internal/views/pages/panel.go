package pages

import (
	"MediaMTXAuth/internal"
	"MediaMTXAuth/internal/views"
	"MediaMTXAuth/internal/views/handlers"
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed html/panel.html
var PanelPageHTML string

type PanelPage struct {
	*views.Page
}

func NewPanel(userService internal.UserService) *PanelPage {
	tmpl := template.Must(template.New("pages").Parse(PanelPageHTML))
	return &PanelPage{
		Page: &views.Page{
			UserService: userService,
			Template:    tmpl,
		},
	}
}

func (v *PanelPage) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		v.showPanelForm(rw, r)
	default:
		http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (v *PanelPage) showPanelForm(rw http.ResponseWriter, r *http.Request) {
	username, authenticated := handlers.RequireAuth(v.Page, rw, r)
	if !authenticated {
		return
	}

	user, err := v.UserService.Get(username)
	if err != nil {
		data := views.PanelData{Error: "Failed to load user"}
		v.renderTemplate(rw, data)
		return
	}

	data := views.PanelData{User: *user}
	v.renderTemplate(rw, data)
}

func (v *PanelPage) HandleChangePassword(rw http.ResponseWriter, r *http.Request) {
	username, authenticated := handlers.RequireAuth(v.Page, rw, r)
	if !authenticated {
		return
	}

	password := r.FormValue("password")
	if password == "" {
		user, _ := v.UserService.Get(username)
		data := views.PanelData{Error: "Password cannot be empty", User: *user}
		v.renderTemplate(rw, data)
		return
	}

	err := v.UserService.ChangePassword(username, password)
	if err != nil {
		user, _ := v.UserService.Get(username)
		data := views.PanelData{Error: err.Error(), User: *user}
		v.renderTemplate(rw, data)
		return
	}

	http.Redirect(rw, r, "/panel", http.StatusSeeOther)
}

func (v *PanelPage) renderTemplate(rw http.ResponseWriter, data views.PanelData) {
	rw.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := v.Template.Execute(rw, data); err != nil {
		http.Error(rw, "Internal server error", http.StatusInternalServerError)
	}
}
