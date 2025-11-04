//go:build wireinject
// +build wireinject

package main

import (
	"statio/container"
	"statio/internal/di"

	"github.com/google/wire"
)

// InitializeScheduler builds SchedulerContainer
func InitializeScheduler() (*container.SchedulerContainer, error) {
	wire.Build(di.SchedulerSet)
	return &container.SchedulerContainer{}, nil
}
