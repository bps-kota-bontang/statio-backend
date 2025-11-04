package di

import (
	"statio/internal/services"

	"github.com/google/wire"
)

var ServiceSet = wire.NewSet(
	services.NewTableService,
	services.NewFactService,
	services.NewIndicatorService,
	services.NewDimensionService,
)
