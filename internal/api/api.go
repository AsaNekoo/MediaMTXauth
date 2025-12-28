package api

import (
	"MediaMTXAuth/internal"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
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

type Api struct {
	UserService      internal.UserService
	NamespaceService internal.NamespaceService
}

func NewApi(userService internal.UserService, namespaceService internal.NamespaceService) *Api {
	return &Api{
		UserService:      userService,
		NamespaceService: namespaceService,
	}
}

func (a *Api) AuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	parts := strings.Split(req.Path, "/")
	if len(parts) != 2 {
		log.Printf("Invalid path format: %s", req.Path)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	namespaceName := parts[0]
	userName := parts[1]

	q, err := url.ParseQuery(req.Query)
	if err != nil {
		log.Printf("Invalid query format: %s", req.Query)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	streamKey := q.Get("key")
	if streamKey == "" {
		log.Printf("Missing stream key")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate Namespace
	_, err = a.NamespaceService.Get(namespaceName)
	if err != nil {
		log.Printf("Namespace not found: %s", namespaceName)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate User
	user, err := a.UserService.Get(userName)
	if err != nil {
		log.Printf("User not found: %s", userName)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if user.Namespace != "" && user.Namespace != namespaceName {
		log.Printf("User %s is not allowed in namespace %s", userName, namespaceName)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Validate Stream Key
	if user.StreamKey != streamKey {
		log.Printf("Invalid stream key for user: %s", userName)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
}
