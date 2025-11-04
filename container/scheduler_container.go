package container

import "github.com/hibiken/asynq"

type SchedulerContainer struct {
	Scheduler *asynq.Scheduler
}
