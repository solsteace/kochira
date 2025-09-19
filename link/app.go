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
	"github.com/solsteace/kochira/link/internal/controller"
	"github.com/solsteace/kochira/link/internal/middleware"
	"github.com/solsteace/kochira/link/internal/repository"
	"github.com/solsteace/kochira/link/internal/route"
	"github.com/solsteace/kochira/link/internal/service"
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
	linkService := service.NewLink(linkRepo)
	redirectionController := controller.NewRedirection(linkService)
	linkController := controller.NewLink(linkService)
	linkRoute := route.NewLink(linkController, userContext)
	redirectionRoute := route.NewRedirection(redirectionController)
	apiRoute := route.NewApi(upSince)

	// ========================================
	// Routings
	// ========================================
	app := chi.NewRouter()
	v1 := chi.NewRouter()
	app.Use(chiMiddleware.Logger)
	app.Use(chiMiddleware.Recoverer)

	linkRoute.Use(v1)
	redirectionRoute.Use(v1)
	app.Mount("/api/v1", v1)
	apiRoute.Use(app)

	// ========================================
	// Init
	// ========================================
	fmt.Printf("Server's running at :%d\n", envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
