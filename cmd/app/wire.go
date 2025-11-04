//go:build wireinject
// +build wireinject

package main

import (
	"statio/container"
	"statio/internal/di"

	"github.com/google/wire"
)

// Initialize App
func InitializeApp() (*container.AppContainer, error) {
	wire.Build(di.AppSet)
	return &container.AppContainer{}, nil
}
