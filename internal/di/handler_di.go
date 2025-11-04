package di

import (
	"statio/internal/handlers"

	"github.com/google/wire"
)

var HandlerSet = wire.NewSet(
	handlers.NewTableHandler,
	handlers.NewIndicatorHandler,
	handlers.NewDimensionHandler,
)
