package providers

import (
	"context"
	"fmt"
	"log"

	"statio/config"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates and returns a new Redis client with connection test
func NewRedisClient(redisConfig *config.RedisConfig) (*redis.Client, error) {
	addr := fmt.Sprintf("%s:%s", redisConfig.RedisHost, redisConfig.RedisPort)

	client := redis.NewClient(&redis.Options{
		Addr: addr,
		// You can add Password and DB here if needed
	})

	// Ping to test the connection
	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Printf("Error connecting to Redis: %v", err)
		return nil, err
	}

	log.Println("Redis client created and connected successfully")
	return client, nil
}
