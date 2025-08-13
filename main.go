package main

import (
	"MediaMTXAuth/internal/api"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth", api.AuthHandler)
	mux.HandleFunc("/login", api.Login)

	server := &http.Server{
		Addr:    ":8080", // MediaMTX will call http://host:8080/auth
		Handler: mux,
	}

	log.Println("Auth server listening on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
