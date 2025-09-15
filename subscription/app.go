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
	mqClient, err := amqp091.Dial(envMqUrl)
	if err != nil {
		log.Fatalf("Failed to connect to mq:%+v", err)
	}

	ch, err := mqClient.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel:%+v", err)
	}
	queue, err := ch.QueueDeclare(
		"hello", // name: what should we name the queue we're going to use?
		false,   // durable: [read function doc]
		false,   // delete [read function doc]
		false,   // exclusive: can this queue be accessed using other connection (channel, I guess)?
		false,   // no-wait: should we assume this queue had already been declared on the broker?
		nil,     // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue:%+v", queue)
	}

	msgs, err := ch.Consume(
		queue.Name, // queue: which queue we're receiving messages from?
		"",         // consumer: what name should I identify myself with?
		false,      // auto-ack: upon receiving message, should it be ACK-ed automatically?
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	go func() {
		for msg := range msgs {
			payload, err := messaging.DeCreateSubscription(msg.Body)
			if err != nil {
				log.Printf("Failed to deserialize message: %v\n", err)
				continue
			}

			if err := subscriptionService.Init(payload.Data.Users); err != nil {
				log.Printf("Failed to init users: %v\n", err)
				continue
			}

			if err := msg.Ack(false); err != nil {
				log.Printf("Failed to ACK message: %v\n", err)
			}
		}
	}()

	// ========================================
	// Init
	// ========================================
	fmt.Printf("Server's running at :%d\n", envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
