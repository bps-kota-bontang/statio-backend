package di

import (
	"statio/internal/services"

	"github.com/google/wire"
)

var ServiceSet = wire.NewSet(
	services.NewExcelService,
	services.NewTableService,
	services.NewFactService,
	services.NewIndicatorService,
	services.NewDimensionService,
	services.NewOrganizationService,
	services.NewUserService,
	services.NewAuthService,
	services.NewBPSService,
	services.NewDashboardService,
	services.NewAggregationService,
)
