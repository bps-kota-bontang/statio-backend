package providers

import (
	"log"
	"statio/config"

	"github.com/hibiken/asynq"
)

// NewAsyncClient creates and returns a new Asynq client
func NewAsyncClient(redisConfig *config.RedisConfig) (*asynq.Client, error) {
	// Construct the Redis connection string
	redisClientOpt := asynq.RedisClientOpt{
		Addr: redisConfig.RedisHost + ":" + redisConfig.RedisPort,
	}

	// Create the Asynq client
	asynqClient := asynq.NewClient(redisClientOpt)

	// Verify the client connection by pinging Redis
	if err := asynqClient.Ping(); err != nil {
		log.Printf("Error connecting to Redis: %v", err)
		return nil, err
	}

	log.Println("Asynq client created and connected to Redis successfully")
	return asynqClient, nil
}
