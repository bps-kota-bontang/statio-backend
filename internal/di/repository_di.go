package di

import (
	"statio/internal/repositories"

	"github.com/google/wire"
)

var RepositorySet = wire.NewSet(
	repositories.NewTableRepository,
	repositories.NewFactRepository,
	repositories.NewIndicatorRepository,
	repositories.NewDimensionRepository,
	repositories.NewOrganizationRepository,
	repositories.NewUserRepository,
	repositories.NewConfigurationRepository,
)
