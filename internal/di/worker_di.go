package di

import (
	"statio/app"

	"github.com/google/wire"
)

// Wire Set for Worker dependencies
var WorkerSet = wire.NewSet(
	app.NewAsynqWorker,
	ConfigSet,
	ProviderSet,
	RepositorySet,
	ServiceSet,
	TaskSet,
)
