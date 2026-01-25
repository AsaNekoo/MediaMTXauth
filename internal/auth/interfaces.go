package auth

import "errors"

type Request struct {
	IP       string `json:"ip"`
	Token    string `json:"token"`
	User     string `json:"user"`
	Password string `json:"password"`
	Path     string `json:"path"`
	Protocol string `json:"protocol"`
	ID       string `json:"id"`
	Action   string `json:"action"` // "read" or "publish"
	Query    string `json:"query"`
}

var ErrAuthError = errors.New("auth error")
