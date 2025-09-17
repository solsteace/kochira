package account

import (
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
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/go-lib/temporary/messaging"
	"github.com/solsteace/go-lib/token"
	"github.com/solsteace/kochira/account/internal"
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
	authRoute := internal.NewAuthRoute(authService)

	// ========================================
	// Routings
	// ========================================
	app := chi.NewRouter()
	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)

	v1 := chi.NewRouter()
	authRoute.Use(v1)

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

	go func() {
		sendFx := func(body []byte) error {
			return mq.Publish(
				"default",
				body,
				utility.NewDefaultAmqpPublishOpts("", "hello2", "application/json"))
		}

		handle := func(outboxes []outbox.Register) ([]uint64, error) {
			userId := []uint64{}
			for _, o := range outboxes {
				userId = append(userId, o.UserId())
			}

			body, err := messaging.SerCreateSubscription(userId)
			if err != nil {
				return []uint64{}, err
			}

			if err := sendFx(body); err != nil {
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
