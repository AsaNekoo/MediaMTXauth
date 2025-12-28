package views

import (
	"MediaMTXAuth/internal"
	"html/template"
)

type Page struct {
	UserService internal.UserService
	Template    *template.Template
}

type LoginData struct {
	Error   string
	Message string
}

type AdminData struct {
	Error      string
	Message    string
	User       internal.User
	Users      []internal.User
	Namespaces []internal.Namespace
}

type PanelData struct {
	Error   string
	Message string
	User    internal.User
}
