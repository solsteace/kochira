package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/solsteace/kochira/account"
)

func main() {
	app := chi.NewRouter()
	app.Use(middleware.Logger)

	app.Mount("/api/account", account.NewApp())

	http.ListenAndServe(":10000", app)
	fmt.Println("Up and running at :10000")
}
