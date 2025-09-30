package subscription

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/solsteace/kochira/subscription/internal/controller"
	"github.com/solsteace/kochira/subscription/internal/middleware"
	"github.com/solsteace/kochira/subscription/internal/repository"
	"github.com/solsteace/kochira/subscription/internal/route"
	"github.com/solsteace/kochira/subscription/internal/service"
	"github.com/solsteace/kochira/subscription/internal/utility"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	domainService "github.com/solsteace/kochira/subscription/internal/domain/service"
)

const moduleName = "kochira/subscription"

func RunApp() {
	upSince := time.Now().Unix()
	dbCfg, err := pgx.ParseConfig(envDbUrl)
	if err != nil {
		err2 := fmt.Errorf("subscription<RunApp>: DB init: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
	}
	dbClient := sqlx.NewDb(stdlib.OpenDB(*dbCfg), "pgx")
	dbClient.SetMaxOpenConns(25) // Based on your db's connection limit
	dbClient.SetMaxIdleConns(10)
	dbClient.SetConnMaxLifetime(30 * time.Minute) // Replace connections periodically
	dbClient.SetConnMaxIdleTime(30 * time.Second) // Close connections that aren't being used
	if err := dbClient.Ping(); err != nil {
		err2 := fmt.Errorf("subscription<RunApp>: DB ping: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
	}
	defer dbClient.Close()

	// ================================
	// Layers
	// ================================
	userContext := middleware.NewUserContext("X-User-Id")
	subscriptionPerks := domainService.NewSubscriptionPerks(
		domainService.NewPerks(time.Hour*24*3, 10),
		domainService.NewPerks(time.Hour*24*30*12, 500),
		time.Second*5)

	subscriptionRepo := repository.NewPgSubscription(dbClient)
	statusService := service.NewStatus(subscriptionRepo, subscriptionPerks)
	statusController := controller.NewStatus(statusService)
	statusRoute := route.NewStatus(statusController, userContext)
	apiRoute := route.NewApi(upSince)

	// ================================
	// Routes
	// ================================
	app := chi.NewRouter()
	v1 := chi.NewRouter()
	app.Use(chiMiddleware.RequestID)
	app.Use(chiMiddleware.Logger)
	app.Use(chiMiddleware.Recoverer)

	statusRoute.Use(v1)
	app.Mount("/api/v1", v1)
	apiRoute.Use(app)

	// ========================================
	// Side effects & subscriptions
	// ========================================
	mq := utility.NewAmqp()
	mqInitReady := make(chan struct{})
	go mq.Start(envMqUrl, mqInitReady)

	<-mqInitReady
	if err := mq.AddChannel("default"); err != nil {
		err2 := fmt.Errorf("subscription<RunApp>: channel init: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
	}
	if err := mq.AddQueue("default", utility.NewDefaultAmqpQueueOpts("hello2")); err != nil {
		err2 := fmt.Errorf("subscription<RunApp>: queue init: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
	}
	err = mq.AddConsumer(
		"default",
		statusController.InitSubscription,
		utility.NewDefaultAmqpConsumeOpts("hello2", false))
	if err != nil {
		err2 := fmt.Errorf("subscription<RunApp>: consumer init: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
	}

	// ========================================
	// Init
	// ========================================
	fmt.Printf("%s: Server's running at :%d\n", moduleName, envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
