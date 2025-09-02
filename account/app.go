package account

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/solsteace/go-lib/reqres"
	"github.com/solsteace/go-lib/token"
	account "github.com/solsteace/kochira/account/internal"
	"github.com/solsteace/kochira/account/internal/cache"
	domainService "github.com/solsteace/kochira/account/internal/domain/service"
	"github.com/solsteace/kochira/account/internal/repository"
	"github.com/solsteace/kochira/account/internal/utility"
	"github.com/valkey-io/valkey-go"
)

func RunApp() {

	// ========================================
	// Utils
	// ========================================
	dbClient, err := sqlx.Connect("pgx", envDbUrl)
	if err != nil {
		log.Fatalf("Error during connecting to DB: %v", err)
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

	accountRepo := repository.NewPgAccount(dbClient)
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
	// Init
	// ========================================
	fmt.Printf("Server's running at :%d\n", envPort)
	http.ListenAndServe(fmt.Sprintf(":%d", envPort), app)
}
