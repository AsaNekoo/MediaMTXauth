package main

import (
	"MediaMTXAuth/internal/api"
	"MediaMTXAuth/internal/services"
	"MediaMTXAuth/internal/storage/bolt"
	"flag"
	"log"
	"net/http"
)

var dbPath string

func init() {
	flag.StringVar(&dbPath, "db", "auth.db", "path to database file")
}

func main() {
	flag.Parse()

	store, err := bolt.New(dbPath)

	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}

	err = store.Init()

	if err != nil {
		log.Fatalf("failed to init DB: %v", err)
	}

	userService := services.NewUserService(store)

	adminPassword, err := userService.CreateDefaultAdminUser()

	if err != nil {
		log.Fatalf("failed to create default admin user: %v", err)
	}

	if adminPassword != "" {
		log.Println("admin password:", adminPassword)
	}

	defer store.Close()

	//web
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
