package container

import "github.com/hibiken/asynq"

type WorkerContainer struct {
	Server *asynq.Server
	Mux    *asynq.ServeMux
}
