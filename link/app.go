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
	"github.com/solsteace/kochira/link/internal/messaging"
	"github.com/solsteace/kochira/link/internal/middleware"
	"github.com/solsteace/kochira/link/internal/persistence"
	"github.com/solsteace/kochira/link/internal/route"
	"github.com/solsteace/kochira/link/internal/service"
	"github.com/solsteace/kochira/link/internal/utility"
)

type publisher struct {
	interval time.Duration // In what interval the routine should be done?
	callback func() error  // What to do in the routine?
}

type listener struct {
	callback func(msg []byte) error
	queue    string
}

const moduleName = "kochira/link"

func RunApp() {
	// ========================================
	// Utils
	// ========================================
	upSince := time.Now().Unix()
	userContext := middleware.NewUserContext("X-User-Id")

	dbClient, err := sqlx.Connect("pgx", envDbUrl)
	if err != nil {
		log.Fatalf("%s: DB connect: %v", moduleName, err)
	}

	mq := utility.NewAmqp()
	mqInitReady := make(chan struct{})
	go mq.Start(envMqUrl, mqInitReady)

	<-mqInitReady
	if err := mq.AddChannel("default"); err != nil {
		log.Fatalf("%s: channel init: %v", moduleName, err)
	}
	queues := map[string][]string{
		"default": []string{
			service.FinishShorteningQueue,
			service.SubscriptionExpiredQueue}}
	for c, queue := range queues {
		for _, q := range queue {
			err := mq.AddQueue(c, utility.NewDefaultAmqpQueueOpts(q))
			if err != nil {
				log.Fatalf("%s: queue init: %v", moduleName, err)
			}
		}
	}

	err = mq.BindQueue(
		"default",
		utility.NewDefaultAmqpQueueBindOpts(
			service.SubscriptionExpiredQueue,
			"#",
			service.SubscriptionExpiredExchange))
	if err != nil {
		log.Fatalf("%s: queue binding: %v", moduleName, err)
	}

	// ========================================
	// Layers
	// ========================================
	linkRepo := persistence.NewPgLink(dbClient)

	shorteningService := service.NewShortening(linkRepo, &mq)
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
	// Subscriptions, messaging, side-effects
	// ========================================
	checkSubscriptionMsg := messaging.CheckSubscriptionMessenger{Version: 1}
	publishers := []publisher{
		publisher{
			interval: time.Second * 2,
			callback: func() error {
				return shorteningService.PublishLinkShortened(
					20, checkSubscriptionMsg.FromLinkShortened)
			}},
		publisher{
			interval: time.Second * 2,
			callback: func() error {
				return shorteningService.PublishShortConfigured(
					20, checkSubscriptionMsg.FromShortConfigured)
			}}}
	for _, p := range publishers {
		go func() {
			t := time.NewTicker(p.interval)
			for range t.C {
				if err := p.callback(); err != nil {
					log.Fatalf("%s: publisher callback: %v", moduleName, err)
				}
			}
		}()
	}

	listeners := []listener{
		listener{
			shorteningController.ListenFinishShortening,
			service.FinishShorteningQueue},
		listener{
			shorteningController.ListenSubscriptionExpired,
			service.SubscriptionExpiredQueue},
	}
	for _, l := range listeners {
		opts := utility.NewDefaultAmqpConsumeOpts(l.queue, false)
		if err := mq.AddConsumer("default", l.callback, opts); err != nil {
			log.Fatalf("%s: listener setup: %v", moduleName, err)
		}
	}

	// ========================================
	// Init
	// ========================================
	fmt.Printf("%s: Server's running at :%d\n", moduleName, envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
