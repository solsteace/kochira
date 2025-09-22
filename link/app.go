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
	"github.com/solsteace/kochira/link/internal/persistence"
	"github.com/solsteace/kochira/link/internal/route"
	"github.com/solsteace/kochira/link/internal/service"
)

const moduleName = "kochira/link"

func RunApp() {
	// ========================================
	// Utils
	// ========================================
	dbClient, err := sqlx.Connect("pgx", envDbUrl)
	if err != nil {
		log.Fatalf("%s: DB connect: %v", moduleName, err)
	}
	upSince := time.Now().Unix()
	userContext := middleware.NewUserContext("X-User-Id")

	// ========================================
	// Layers
	// ========================================
	linkRepo := persistence.NewPgLink(dbClient)

	shorteningService := service.NewShortening(linkRepo)
	shorteningController := controller.NewShortening(shorteningService)
	shorteningRoute := route.NewShortening(shorteningController, userContext)

	redirectSerivce := service.NewRedirect(linkRepo)
	redirectController := controller.NewRedirect(redirectSerivce)
	redirectionRoute := route.NewRedirect(redirectController)

	// ========================================
	// Routings
	// ========================================
	app := chi.NewRouter()
	v1 := chi.NewRouter()
	app.Use(chiMiddleware.RequestID)
	app.Use(chiMiddleware.Logger)
	app.Use(chiMiddleware.Recoverer)

	shorteningRoute.Use(v1)
	redirectionRoute.Use(v1)
	app.Mount("/api/v1", v1)
	route.NewApi(upSince).Use(app)

	// ========================================
	// Init
	// ========================================
	fmt.Printf("%s: Server's running at :%d\n", moduleName, envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
