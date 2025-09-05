package link

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5"
	"github.com/jmoiron/sqlx"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/kochira/link/internal"
	"github.com/solsteace/kochira/link/internal/middleware"
	"github.com/solsteace/kochira/link/internal/repository"
)

func RunApp() {
	// ========================================
	// Utils
	// ========================================
	dbClient, err := sqlx.Connect("pgx", envDbUrl)
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}
	upSince := time.Now().Unix()
	userContext := middleware.NewUserContext("X-User-Id")

	// ========================================
	// Layers
	// ========================================
	linkRepo := repository.NewPgLink(dbClient)
	linkService := internal.NewLinkService(linkRepo)
	linkController := internal.NewLinkController(linkService)

	// ========================================
	// Routings
	// ========================================
	app := chi.NewRouter()
	app.Use(chiMiddleware.Logger)
	app.Use(chiMiddleware.Recoverer)

	v1 := chi.NewRouter()
	internal.UseLink(v1, linkController, userContext)

	app.Mount("/api/v1", v1)
	app.Get("/health", reqres.HttpHandlerWithError(
		func(w http.ResponseWriter, r *http.Request) error {
			return reqres.HttpOk(
				w,
				http.StatusOK,
				map[string]any{
					"msg": "Server is healthy",
					"data": map[string]any{
						"uptime": time.Now().Unix() - upSince,
					}})
		}))
	app.NotFound(reqres.HttpHandlerWithError(
		func(w http.ResponseWriter, r *http.Request) error {
			return reqres.HttpOk(
				w,
				http.StatusNotFound,
				map[string]any{
					"msg": "The endpoint you're reaching wasn't found"})
		}))

	// ========================================
	// Init
	// ========================================
	fmt.Printf("Server's running at :%d\n", envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
