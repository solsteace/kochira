package link

import (
	"fmt"
	"os"
	"strconv"
)

var (
	envPort int

	envMqUrl string
	envDbUrl string
)

func LoadEnv() error {
	switch port, err := strconv.ParseInt(os.Getenv("LINK_PORT"), 10, 32); {
	case err != nil:
		err := fmt.Errorf("`LINK_PORT`: %s", err)
		return fmt.Errorf("internal<LoadEnv>: %w", err)
	case port < 0 || port > 65535:
		err := fmt.Errorf("`LINK_PORT`: port should be between 0 - 65535 (get: %d)", port)
		return fmt.Errorf("internal<LoadEnv>: %w", err)
	default:
		envPort = int(port)
	}

	envMqUrl = os.Getenv("LINK_MQ_URL")
	envDbUrl = os.Getenv("LINK_DB_URL")
	return nil
}
