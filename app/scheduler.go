package app

import (
	"fmt"
	"log"
	"statio/config"
	"statio/container"
	"time"

	"github.com/hibiken/asynq"
)

// NewAsynqScheduler creates a new asynq scheduler
func NewAsynqScheduler(
	appConfig *config.AppConfig,
	schedulerConfig *config.SchedulerConfig,
	redisConfig *config.RedisConfig,
	asynqClient *asynq.Client,
) (*container.SchedulerContainer, error) {
	redisClientOpt := asynq.RedisClientOpt{
		Addr: redisConfig.RedisHost + ":" + redisConfig.RedisPort,
	}

	defer asynqClient.Close()

	scheduler := asynq.NewScheduler(redisClientOpt, nil)

	if schedulerConfig.CronExpressionSyncMainAvatar == "" {
		return nil, fmt.Errorf("cron expression is empty")
	}

	if schedulerConfig.CronExpressionSyncNewAvatar == "" {
		return nil, fmt.Errorf("cron expression is empty")
	}

	if _, err := scheduler.Register(schedulerConfig.CronExpressionSyncMainAvatar, asynq.NewTask("sync_main_avatar", nil), asynq.Timeout(120*time.Minute)); err != nil {
		return nil, fmt.Errorf("could not register task: %w", err)
	}

	if _, err := scheduler.Register(schedulerConfig.CronExpressionSyncNewAvatar, asynq.NewTask("sync_new_avatar", nil), asynq.Timeout(120*time.Minute)); err != nil {
		return nil, fmt.Errorf("could not register task: %w", err)
	}

	log.Println("Task registered successfully")

	if err := scheduler.Start(); err != nil {
		return nil, fmt.Errorf("failed to start scheduler: %w", err)
	}

	return &container.SchedulerContainer{
		Scheduler: scheduler,
	}, nil
}
