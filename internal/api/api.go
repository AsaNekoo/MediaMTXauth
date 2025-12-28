package api

import (
	"MediaMTXAuth/internal"
	"net/http"
)

type AuthRequest struct {
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

type UserSession struct {
	internal.UserService
}

func NewApi(userService internal.UserService) *UserSession {
	return &UserSession{
		UserService: userService,
	}
}

func (a UserSession) AuthHandler(w http.ResponseWriter, r *http.Request) {

	//	if r.Method != http.MethodPost {
	//		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	//		return
	//	}
	//
	//	var req AuthRequest
	//
	//	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	//		http.Error(w, "Bad request", http.StatusBadRequest)
	//		return
	//	}
	//	log.Println("Auth request for user:", req.Path, "token:", req.Token, "Query:", req.Query)
	//
	//	_, err := a.Get()
	//	if err != nil {
	//		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	//		return
	//	}

	w.WriteHeader(http.StatusOK)

	// Example static check
	// if req.User == "asahi" && req.Password == "asahi" && req.Action == "publish" {
	// 	w.WriteHeader(http.StatusOK)
	// 	return
	// }
}
