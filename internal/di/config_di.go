package di

import (
	"statio/config"

	"github.com/google/wire"
)

var ConfigSet = wire.NewSet(
	config.LoadAppConfig,
	config.LoadDatabaseConfig,
	config.LoadRedisConfig,
	config.LoadSchedulerConfig,
)
