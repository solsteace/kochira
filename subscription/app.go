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
	"github.com/solsteace/kochira/subscription/internal/messaging"
	"github.com/solsteace/kochira/subscription/internal/middleware"
	"github.com/solsteace/kochira/subscription/internal/persistence"
	"github.com/solsteace/kochira/subscription/internal/route"
	"github.com/solsteace/kochira/subscription/internal/service"
	"github.com/solsteace/kochira/subscription/internal/utility"

	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	subscriptionService "github.com/solsteace/kochira/subscription/internal/domain/subscription/service"
	"github.com/solsteace/kochira/subscription/internal/domain/subscription/value"
)

const moduleName = "kochira/subscription"

type publisher struct {
	interval time.Duration // In what interval the routine should be done?
	callback func() error  // What to do in the routine?
}

type listener struct {
	callback func(msg []byte) error
	queue    string
}

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

	mq := utility.NewAmqp()
	mqInitReady := make(chan struct{})
	go mq.Start(envMqUrl, mqInitReady)

	<-mqInitReady
	if err := mq.AddChannel("default"); err != nil {
		err2 := fmt.Errorf("subscription<RunApp>: channel init: %w", err)
		log.Fatalf("%s: %v", moduleName, err2)
	}

	exchanges := map[string]string{
		"example": "fanout"}
	for name, kind := range exchanges {
		mq.AddExchange("default", utility.NewDefaultAmqpExchangeOpts(name, kind))
	}

	queues := map[string][]string{
		"default": []string{
			service.CreateSubcriptionQueue,
			service.CheckSubscriptionQueue}}
	for c, queue := range queues {
		for _, q := range queue {
			err := mq.AddQueue(c, utility.NewDefaultAmqpQueueOpts(q))
			if err != nil {
				err2 := fmt.Errorf("subscription<RunApp>: queue init: %w", err)
				log.Fatalf("%s: %v", moduleName, err2)
			}
		}
	}

	// ================================
	// Layers
	// ================================

	createSubscription := messaging.CreateSubscriptionMessenger{Version: 1}
	checkSubscription := messaging.CheckSubscriptionMessenger{Version: 1}
	finishShortening := messaging.FinishShorteningMessenger{Version: 1}
	subscriptionExpired := messaging.SubscriptionExpiredMessenger{Version: 1}
	userContext := middleware.NewUserContext("X-User-Id")
	perkHandler := subscriptionService.NewPerkInferer(
		value.NewPerks(time.Hour*24*3, 10, false),
		value.NewPerks(time.Hour*24*30*12, 500, true),
		time.Second*5)

	subscriptionRepo := persistence.NewPgSubscription(dbClient)
	subscriptionService := service.NewSubscription(subscriptionRepo, perkHandler, &mq)
	subscriptionController := controller.NewSubscription(
		subscriptionService,
		createSubscription,
		checkSubscription)

	// ================================
	// Routes
	// ================================
	app := chi.NewRouter()
	v1 := chi.NewRouter()
	app.Use(chiMiddleware.RequestID)
	app.Use(chiMiddleware.Logger)
	app.Use(chiMiddleware.Recoverer)

	route.NewSubscription(subscriptionController, userContext).Use(v1)
	app.Mount("/api/v1", v1)
	route.NewApi(upSince).Use(app)

	// ========================================
	// Side effects & subscriptions
	// ========================================
	// TODO: start in its own process (or machine, if needed)
	go func() {
		t := time.NewTicker(5 * time.Second)
		for range t.C {
			if err := subscriptionService.WatchExpiringSubscription(500); err != nil {
				log.Printf("internal<RunApp>: %v\n", err)
			}
		}
	}()

	publishers := []publisher{
		publisher{
			interval: time.Second * 2,
			callback: func() error {
				return subscriptionService.PublishFinishShortening(
					20, finishShortening.FromSubscriptionChecked)
			}},
		publisher{
			interval: time.Second * 2,
			callback: func() error {
				return subscriptionService.PublishSubscriptionExpired(
					20, subscriptionExpired.FromSubscriptionExpired)
			}},
		publisher{
			interval: time.Second * 2,
			callback: func() error {
				return subscriptionService.PublishSubscriptionExpired(
					0, subscriptionExpired.FromSubscriptionExpired)
			},
		}}
	for _, p := range publishers {
		go func() {
			t := time.NewTicker(p.interval)
			for range t.C {
				if err := p.callback(); err != nil {
					log.Printf("internal<RunApp>: %v\n", err)
				}
			}
		}()
	}

	listeners := []listener{
		listener{
			subscriptionController.ListenCreateSubscription,
			service.CreateSubcriptionQueue},
		listener{
			subscriptionController.ListenCheckSubscription,
			service.CheckSubscriptionQueue}}
	for _, l := range listeners {
		opts := utility.NewDefaultAmqpConsumeOpts(l.queue, false)
		if err := mq.AddConsumer("default", l.callback, opts); err != nil {
			log.Fatalf("internal<RunApp>: failed to setup listener: %v", err)
		}
	}

	// ========================================
	// Init
	// ========================================
	fmt.Printf("%s: Server's running at :%d\n", moduleName, envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
