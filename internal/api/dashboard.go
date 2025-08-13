package api

import (
	"fmt"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "external/login.html")
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		login := r.FormValue("login")
		password := r.FormValue("password")
		hashed, err := HashPassword(password)
		fmt.Println(login, password, hashed, err)
	}
}
