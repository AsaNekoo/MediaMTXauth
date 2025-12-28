package main

import (
	"MediaMTXAuth/internal/api"
	"MediaMTXAuth/internal/services"
	"MediaMTXAuth/internal/storage/bolt"
	"MediaMTXAuth/internal/views/pages"
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
	requirePost := func(handler func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			handler(w, r)
		}
	}

	loginView := pages.NewLogin(userService)
	adminView := pages.NewAdmin(userService)
	panelView := pages.NewPanel(userService)
	api := api.NewApi(userService)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/views/pages/html/static"))))

	// API
	mux.HandleFunc("/api/auth", api.AuthHandler)

	// Views
	mux.Handle("/login", loginView)
	mux.Handle("/admin", adminView)
	mux.Handle("/panel", panelView)

	// POST
	mux.HandleFunc("/admin/add", requirePost(adminView.HandleAddUser))
	mux.HandleFunc("/admin/remove", requirePost(adminView.HandleRemoveUser))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
