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
	"github.com/jmoiron/sqlx"
	"github.com/solsteace/kochira/account/internal/controller"
	"github.com/solsteace/kochira/account/internal/messaging"
	"github.com/solsteace/kochira/account/internal/persistence"
	"github.com/solsteace/kochira/account/internal/route"
	"github.com/solsteace/kochira/account/internal/service"
	"github.com/solsteace/kochira/account/internal/utility"
	"github.com/solsteace/kochira/account/internal/utility/hash"
	"github.com/solsteace/kochira/account/internal/utility/token"
	"github.com/valkey-io/valkey-go"

	_ "github.com/jackc/pgx/v5/stdlib"
	authService "github.com/solsteace/kochira/account/internal/domain/auth/service"
)

type publisher struct {
	interval time.Duration // In what interval the routine should be done?
	callback func() error  // What to do in the routine?
}

type listener struct {
	callback func(msg []byte) error
	queue    string
}

func RunApp() {
	// ========================================
	// Utils
	// ========================================
	upSince := time.Now().Unix()
	hasher := hash.NewBcrypt(10)
	accessTokenHandler := token.NewJwt[token.Auth](
		envTokenIssuer,
		envTokenSecret,
		time.Duration(envAccessTokenLifetime))
	refreshTokenHandler := token.NewJwt[token.Auth](
		envTokenIssuer,
		strings.Repeat(envTokenSecret, 2),
		time.Duration(envRefreshTokenLifetime))
	createSubscriptionMessenger := messaging.CreateSubscriptionMessenger{Version: 1}

	// Props to: https://medium.com/@lokeahnming/that-time-i-took-down-my-production-site-with-too-many-database-connections-8758406445e5
	dbCfg, err := pgx.ParseConfig(envDbUrl)
	if err != nil {
		log.Fatalf("kochira/account : %v",
			fmt.Errorf("account<RunApp>: DB init: %w", err))
	}
	dbClient := sqlx.NewDb(stdlib.OpenDB(*dbCfg), "pgx")
	dbClient.SetMaxOpenConns(25) // Based on your db's connection limit
	dbClient.SetMaxIdleConns(10)
	dbClient.SetConnMaxLifetime(30 * time.Minute) // Replace connections periodically
	dbClient.SetConnMaxIdleTime(30 * time.Second) // Close connections that aren't being used
	if err := dbClient.Ping(); err != nil {
		log.Fatalf("kochira/account: %v",
			fmt.Errorf("account<RunApp>: DB ping: %w", err))
	}
	defer dbClient.Close()

	cacheClient, err := valkey.NewClient(
		valkey.MustParseURL(envCacheUrl))
	if err != nil {
		log.Fatalf("kochira/account: %v",
			fmt.Errorf("account<RunApp>: cache init: %w", err))
	}
	defer cacheClient.Close()

	mq := utility.NewAmqp()
	mqInitReady := make(chan struct{})
	go mq.Start(envMqUrl, mqInitReady)
	<-mqInitReady
	if err := mq.AddChannel("default"); err != nil {
		log.Fatalf("kochira/account: %v",
			fmt.Errorf("account<RunApp>: channel init: %w", err))
	}

	// ========================================
	// Layers
	// ========================================
	authJailer := authService.NewJailer(
		3, 3, 120*time.Second, 10*time.Second)

	authStore := persistence.NewPgAuth(dbClient)
	accountStore := persistence.NewPgAccount(dbClient)
	authCache := persistence.NewValkeyAuth(
		cacheClient,
		authJailer.RetentionTime(15*time.Second),
		time.Duration(envRefreshTokenLifetime)*time.Second)

	accountService := service.NewAccount(accountStore, hasher, &mq)
	authService := service.NewAuth(
		authStore,
		authCache,
		authCache,
		hasher,
		accessTokenHandler,
		refreshTokenHandler,
		authJailer)
	authController := controller.NewAuth(authService)
	accountController := controller.NewAccount(accountService)

	// ========================================
	// Routings
	// ========================================
	app := chi.NewRouter()
	v1 := chi.NewRouter()
	app.Use(middleware.RequestID)
	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)

	route.NewAccount(accountController).Use(v1)
	route.NewAuth(authController).Use(v1)
	app.Mount("/api/v1", v1)
	route.NewApi(upSince).Use(app)

	// ========================================
	// Side effects, susbcriptions
	// ========================================
	publishers := []publisher{
		publisher{
			interval: time.Second * 2,
			callback: func() error {
				return accountService.HandleRegisteredUsers(
					20, createSubscriptionMessenger.FromManyUserRegistered)
			}}}
	for _, p := range publishers {
		go func() {
			t := time.NewTicker(p.interval)
			for range t.C {
				if err := p.callback(); err != nil {
					log.Printf("kochira/account: account<RunApp>: %v\n", err)
				}
			}
		}()
	}

	// ========================================
	// Init
	// ========================================
	fmt.Printf("kochira/account: Server's running at :%d\n", envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
