package subscription

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rabbitmq/amqp091-go"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/go-lib/temporary/messaging"
	"github.com/solsteace/kochira/subscription/internal"
	"github.com/solsteace/kochira/subscription/internal/domain/service"
	"github.com/solsteace/kochira/subscription/internal/middleware"
	"github.com/solsteace/kochira/subscription/internal/repository"
	"github.com/solsteace/kochira/subscription/internal/utility"
)

func RunApp() {
	upSince := time.Now().Unix()
	dbCfg, err := pgx.ParseConfig(envDbUrl)
	if err != nil {
		log.Fatalf("Error during connecting to DB: %v", err)
	}
	dbClient := sqlx.NewDb(stdlib.OpenDB(*dbCfg), "pgx")
	dbClient.SetMaxOpenConns(25) // Based on your db's connection limit
	dbClient.SetMaxIdleConns(10)
	dbClient.SetConnMaxLifetime(30 * time.Minute) // Replace connections periodically
	dbClient.SetConnMaxIdleTime(30 * time.Second) // Close connections that aren't being used
	if err := dbClient.Ping(); err != nil {
		log.Fatalf("Error during Ping attempt: %v", err)
	}
	defer dbClient.Close()

	// ================================
	// Layers
	// ================================
	userContext := middleware.NewUserContext("X-User-Id")
	subscriptionPerks := service.NewSubscriptionPerks(
		service.NewPerks(time.Hour*24*3, 10),
		service.NewPerks(time.Hour*24*30*12, 500),
		time.Second*5)

	subscriptionRepo := repository.NewPgSubscription(dbClient)
	subscriptionService := internal.NewSubscriptionService(subscriptionRepo, subscriptionPerks)
	subscriptionRoute := internal.NewSubscriptionRoute(subscriptionService, userContext)

	// ================================
	// Routes
	// ================================
	app := chi.NewRouter()
	app.Use(chiMiddleware.Logger)
	app.Use(chiMiddleware.Recoverer)

	v1 := chi.NewRouter()
	subscriptionRoute.Use(v1)

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
	// Side effects & subscriptions
	// ========================================
	mq := utility.NewAmqp()
	mqMonitorEnd := make(chan struct{})
	mqInitReady := make(chan struct{})
	go mq.Start(envMqUrl, mqInitReady)
	go mq.Monitor(mqMonitorEnd)

	<-mqInitReady
	if err := mq.AddChannel("default"); err != nil {
		log.Fatalf("Failed to open a channel: %+v", err)
	}

	err = mq.AddQueue("default", utility.NewDefaultAmqpQueueOpts("hello2"))
	if err != nil {
		log.Fatalf("Failed to declare a queue: %+v", err)
	}

	err = mq.AddConsumer(
		"default",
		func(msg amqp091.Delivery) error {
			payload, err := messaging.DeCreateSubscription(msg.Body)
			if err != nil {
				return fmt.Errorf("<consume callback>: %w", err)
			}

			if err := subscriptionService.Init(payload.Data.Users); err != nil {
				return fmt.Errorf("<consume callback>: %w", err)
			}
			return nil
		},
		utility.NewDefaultAmqpConsumeOpts("hello2", false))
	if err != nil {
		log.Fatalf("Failed to start consuming: %+v", err)
	}

	// ========================================
	// Init
	// ========================================
	fmt.Printf("Server's running at :%d\n", envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
