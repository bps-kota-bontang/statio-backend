package di

import (
	"statio/app"

	"github.com/google/wire"
)

var AppSet = wire.NewSet(
	app.NewFiberApp,
	ConfigSet,
	ProviderSet,
	RepositorySet,
	ServiceSet,
	HandlerSet,
)
