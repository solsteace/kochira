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
	"github.com/solsteace/go-lib/token"
	"github.com/solsteace/kochira/account/internal/cache"
	"github.com/solsteace/kochira/account/internal/controller"
	domainService "github.com/solsteace/kochira/account/internal/domain/service"
	"github.com/solsteace/kochira/account/internal/repository"
	"github.com/solsteace/kochira/account/internal/route"
	"github.com/solsteace/kochira/account/internal/service"
	"github.com/solsteace/kochira/account/internal/utility"
	"github.com/valkey-io/valkey-go"
)

const moduleName = "kochira/account"

func RunApp() {
	// ========================================
	// Utils
	// ========================================
	// Props to: https://medium.com/@lokeahnming/that-time-i-took-down-my-production-site-with-too-many-database-connections-8758406445e5
	dbCfg, err := pgx.ParseConfig(envDbUrl)
	if err != nil {
		err2 := fmt.Errorf("account<RunApp>: DB init: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
	}
	dbClient := sqlx.NewDb(stdlib.OpenDB(*dbCfg), "pgx")
	dbClient.SetMaxOpenConns(25) // Based on your db's connection limit
	dbClient.SetMaxIdleConns(10)
	dbClient.SetConnMaxLifetime(30 * time.Minute) // Replace connections periodically
	dbClient.SetConnMaxIdleTime(30 * time.Second) // Close connections that aren't being used
	if err := dbClient.Ping(); err != nil {
		err2 := fmt.Errorf("account<RunApp>: DB ping: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
	}
	defer dbClient.Close()

	cacheClient, err := valkey.NewClient(
		valkey.MustParseURL(envCacheUrl))
	if err != nil {
		err2 := fmt.Errorf("account<RunApp>: cache init: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
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

	authService := service.NewAuth(
		accountRepo,
		authAttemptCache,
		tokenCache,
		secretHandler,
		accessTokenHandler,
		refreshTokenHandler,
		authAttemptDomainService)
	controller := controller.NewAuth(authService)
	authRoute := route.NewAuth(controller)
	apiRoute := route.NewApi(upSince)

	// ========================================
	// Routings
	// ========================================
	app := chi.NewRouter()
	v1 := chi.NewRouter()
	app.Use(middleware.RequestID)
	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)

	authRoute.Use(v1)
	app.Mount("/api/v1", v1)
	apiRoute.Use(app)

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
		err2 := fmt.Errorf("account<RunApp>: channel init: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
	}

	err = mq.AddQueue("default", utility.NewDefaultAmqpQueueOpts("hello2"))
	if err != nil {
		err2 := fmt.Errorf("account<RunApp>: queue init: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
	}

	go func() {
		opts := utility.NewDefaultAmqpPublishOpts("", "hello2", "application/json")
		send := func(body []byte) error {
			return mq.Publish("default", body, opts)
		}

		t := time.NewTicker(time.Second * 2)
		for range t.C {
			if err := controller.PublishNewUser(20, send); err != nil {
				log.Printf("%s: %v\n", moduleName, err)
			}
		}
	}()

	// ========================================
	// Init
	// ========================================
	fmt.Printf("%s: Server's running at :%d\n", moduleName, envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
