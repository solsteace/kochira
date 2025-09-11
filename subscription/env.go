package subscription

import (
	"log"
	"os"
	"strconv"
)

var (
	envPort int

	envDbUrl string
	envMqUrl string
)

func LoadEnv() {
	switch port, err := strconv.ParseInt(os.Getenv("SUBSCRIPTION_PORT"), 10, 64); {
	case err != nil:
		log.Fatalf("`LINK_PORT`: %s", err)
	case port < 0 || port > 65535:
		log.Fatal("`LINK_PORT: port should be between 0 - 65535 (get: %d)`", port)
	default:
		envPort = int(port)
	}

	envDbUrl = os.Getenv("SUBSCRIPTION_DB_URL")
	envMqUrl = os.Getenv("SUBSCRIPTION_MQ_URL")
}
