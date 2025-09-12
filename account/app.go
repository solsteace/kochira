package account

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rabbitmq/amqp091-go"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/go-lib/temporary/messaging"
	"github.com/solsteace/go-lib/token"
	account "github.com/solsteace/kochira/account/internal"
	"github.com/solsteace/kochira/account/internal/cache"
	"github.com/solsteace/kochira/account/internal/domain/outbox"
	domainService "github.com/solsteace/kochira/account/internal/domain/service"
	"github.com/solsteace/kochira/account/internal/repository"
	"github.com/solsteace/kochira/account/internal/utility"
	"github.com/valkey-io/valkey-go"
)

func RunApp() {
	// ========================================
	// Utils
	// ========================================
	// Props to: https://medium.com/@lokeahnming/that-time-i-took-down-my-production-site-with-too-many-database-connections-8758406445e5
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

	cacheClient, err := valkey.NewClient(
		valkey.MustParseURL(envCacheUrl))
	if err != nil {
		log.Fatalf("Error during connecting to cache: %v", err)
	}
	defer cacheClient.Close()

	upSince := time.Now().Unix()
	secretHandler := utility.NewBcrypt(10)
	accessTokenHandler := utility.NewJwt[token.Auth](
		envTokenIssuer,
		envTokenSecret,
		time.Duration(envAccessTokenLifetime))
	refreshTokenHandler := utility.NewJwt[token.Auth](
		envTokenIssuer,
		strings.Repeat(envTokenSecret, 2),
		time.Duration(envRefreshTokenLifetime))

	// ========================================
	// Layers
	// ========================================
	authAttemptDomainService := domainService.NewAuthAttempt(
		3,
		3,
		120*time.Second,
		10*time.Second)

	accountRepo := repository.NewPgUser(dbClient)
	authAttemptCache := cache.NewValkeyAuthAttempt(
		cacheClient,
		authAttemptDomainService.RetentionTime(15*time.Second))
	tokenCache := cache.NewValkeyToken(
		cacheClient,
		time.Duration(envRefreshTokenLifetime)*time.Second)

	authService := account.NewAuthService(
		accountRepo,
		authAttemptCache,
		tokenCache,
		secretHandler,
		accessTokenHandler,
		refreshTokenHandler,
		authAttemptDomainService)
	authController := account.NewAuthController(authService)

	// ========================================
	// Routings
	// ========================================
	app := chi.NewRouter()
	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)

	v1 := chi.NewRouter()
	account.UseAuth(v1, authController)

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
	// Side effects, susbcriptions
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
		"hello", // name: what's the name of the queue we're going to use?
		false,   // durable: [read function doc]
		false,   // delete [read function doc]
		false,   // exclusive: can this queue be accessed using other connection (channel, I guess)?
		false,   // no-wait: should we assume this queue had already been declared on the broker?
		nil,     // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue:%+v", queue)
	}

	go func() {
		handle := func(outboxes []outbox.Register) ([]uint64, error) {
			userId := []uint64{}
			for _, o := range outboxes {
				userId = append(userId, o.UserId())
			}

			body, err := messaging.SerCreateSubscription(userId)
			if err != nil {
				return []uint64{}, err
			}

			ctx := context.Background() // Use the default one, for now
			err = ch.PublishWithContext(
				ctx,
				"",
				queue.Name,
				false,
				false,
				amqp091.Publishing{
					ContentType: "application/json",
					Body:        body})
			if err != nil {
				return []uint64{}, err
			}
			return userId, nil
		}

		t := time.NewTicker(time.Second * 2)
		for {
			select {
			case <-t.C:
				if err := authService.HandleNewUsers(20, handle); err != nil {
					fmt.Println(err)
				}
			}
		}
	}()

	// ========================================
	// Init
	// ========================================
	fmt.Printf("Server's running at :%d\n", envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
