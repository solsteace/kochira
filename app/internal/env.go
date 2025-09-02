package internal

import (
	"log"
	"os"
	"strconv"
)

var (
	EnvPort int

	EnvCacheUrl string
	EnvDbUrl    string

	EnvTokenIssuer          string
	EnvTokenSecret          string
	EnvAccessTokenLifetime  int
	EnvRefreshTokenLifetime int
)

func loadEnv() {
	switch port, err := strconv.ParseInt(os.Getenv("PORT"), 10, 32); {
	case err != nil:
		log.Fatalf("`PORT`: %s", err)
	case EnvPort < 0 || EnvPort > 65535:
		log.Fatal("`PORT: port should be between 0 - 65535 (get: %d)`", port)
	default:
		EnvPort = int(port)
	}

	EnvCacheUrl = os.Getenv("CACHE_URL")
	EnvDbUrl = os.Getenv("DB_URL")

	EnvTokenIssuer = os.Getenv("TOKEN_ISSUER")
	EnvTokenSecret = os.Getenv("TOKEN_SECRET")
	switch tokenLifetime, err := strconv.ParseInt(os.Getenv("TOKEN_LIFETIME"), 10, 32); {
	case err != nil:
		log.Fatalf("`TOKEN_LIFETIME`: %s", err)
	case int(tokenLifetime) < 0:
		log.Fatal("`TOKEN_LIFETIME: converted value cannot be negative (get: %d)`", tokenLifetime)
	default:
		EnvAccessTokenLifetime = int(tokenLifetime)
		EnvRefreshTokenLifetime = EnvAccessTokenLifetime * 3
	}
}
