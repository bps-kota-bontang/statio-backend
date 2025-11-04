package app

import (
	"statio/config"
	"statio/container"
	"statio/internal/tasks"

	"github.com/hibiken/asynq"
)

// NewAsynqWorker initializes the Asynq worker with all necessary routes
func NewAsynqWorker(
	appConfig *config.AppConfig,
	redisConfig *config.RedisConfig,
	TableTask *tasks.TableTask,
) (*container.WorkerContainer, error) {
	redisClientOpt := asynq.RedisClientOpt{
		Addr: redisConfig.RedisHost + ":" + redisConfig.RedisPort,
	}

	srv := asynq.NewServer(redisClientOpt, asynq.Config{
		Concurrency: 20,
		Queues: map[string]int{
			"critical": 12,
			"default":  5,
			"low":      3,
		},
	})

	mux := asynq.NewServeMux()

	return &container.WorkerContainer{
		Server: srv,
		Mux:    mux,
	}, nil
}
