package account

import (
	"fmt"
	"os"
	"strconv"
)

var (
	envPort int

	envCacheUrl string
	envDbUrl    string
	envMqUrl    string

	envTokenIssuer          string
	envTokenSecret          string
	envAccessTokenLifetime  int
	envRefreshTokenLifetime int
)

func LoadEnv() error {
	switch port, err := strconv.ParseInt(os.Getenv("ACCOUNT_PORT"), 10, 32); {
	case err != nil:
		return fmt.Errorf("account<LoadEnv>: `ACCOUNT_PORT`: %w", err)
	case envPort < 0 || envPort > 65535:
		err := fmt.Errorf("port should be between 0 - 65535 (get: %d)", port)
		return fmt.Errorf("account<LoadEnv>: `ACCOUNT_PORT`: %w", err)
	default:
		envPort = int(port)
	}

	envCacheUrl = os.Getenv("ACCOUNT_CACHE_URL")
	envDbUrl = os.Getenv("ACCOUNT_DB_URL")
	envMqUrl = os.Getenv("ACCOUNT_MQ_URL")

	envTokenIssuer = os.Getenv("ACCOUNT_TOKEN_ISSUER")
	envTokenSecret = os.Getenv("ACCOUNT_TOKEN_SECRET")
	switch tokenLifetime, err := strconv.ParseInt(os.Getenv("ACCOUNT_TOKEN_LIFETIME"), 10, 32); {
	case err != nil:
		return fmt.Errorf("account<LoadEnv>: `ACCOUNT_TOKEN_LIFETIME`: %w", err)
	case int(tokenLifetime) < 0:
		err := fmt.Errorf("`ACCOUNT_TOKEN_LIFETIME: converted value cannot be negative (get: %d)`", tokenLifetime)
		return fmt.Errorf("account<LoadEnv>: `ACCOUNT_TOKEN_LIFETIME`: %w", err)
	default:
		envAccessTokenLifetime = int(tokenLifetime)
		envRefreshTokenLifetime = envAccessTokenLifetime * 3
	}

	return nil
}
