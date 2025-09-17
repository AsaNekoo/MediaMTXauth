package main

import (
	"MediaMTXAuth/internal/services"
	"MediaMTXAuth/internal/storage/bolt"
	"MediaMTXAuth/internal/views"
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

	loginView := views.NewLogin(userService)
	adminView := views.NewAdmin(userService)

	// Setup routes
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/views/pages/static"))))
	mux.Handle("/login", loginView)
	mux.Handle("/admin", adminView)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
