package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func (a *Auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	parts := strings.Split(strings.TrimPrefix(req.Path, "/"), "/")

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
	err = a.Validate(namespaceName, userName, streamKey, req.Action)

	if errors.Is(err, ErrAuthError) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Printf("Failed to validate user: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
