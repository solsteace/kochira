package account

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/solsteace/kochira/account/repository"
	"github.com/solsteace/kochira/internal/persistence"
)

func NewApp() chi.Router {
	dbConn, err := persistence.NewPgConnection(os.Getenv("DB_URL"))
	if err != nil {
		log.Fatalln(err.Error())
	}

	accountRepo := repository.PgAccount{Conn: dbConn.Conn}
	app := chi.NewRouter()

	upSince := time.Now().Unix()
	app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("Up for %d seconds", time.Now().Unix()-upSince)
		w.Write([]byte(msg))
	})

	authService := AuthService{accountRepo: accountRepo}
	authController := AuthController{service: authService}
	app.Mount("/auth", newAuthRouter(authController))
	return app
}
