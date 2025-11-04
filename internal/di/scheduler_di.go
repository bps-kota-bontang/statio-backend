package di

import (
	"statio/app"

	"github.com/google/wire"
)

// Wire Set for Scheduler dependencies
var SchedulerSet = wire.NewSet(
	app.NewAsynqScheduler,
	ConfigSet,
	ProviderSet,
)
