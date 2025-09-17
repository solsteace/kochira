package subscription

import (
	"fmt"
	"os"
	"strconv"
)

var (
	envPort int

	envDbUrl string
	envMqUrl string
)

func LoadEnv() error {
	switch port, err := strconv.ParseInt(os.Getenv("SUBSCRIPTION_PORT"), 10, 64); {
	case err != nil:
		return fmt.Errorf("subscription<LoadEnv>: `SUBSCRIPTION_PORT`: %w", err)
	case port < 0 || port > 65535:
		err := fmt.Errorf("port should be between 0 - 65535 (get: %d)", port)
		return fmt.Errorf("subscription<LoadEnv>: `SUBSCRIPTION_PORT`: %w", err)
	default:
		envPort = int(port)
	}

	envDbUrl = os.Getenv("SUBSCRIPTION_DB_URL")
	envMqUrl = os.Getenv("SUBSCRIPTION_MQ_URL")
	return nil
}
