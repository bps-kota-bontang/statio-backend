package di

import (
	"statio/internal/middlewares"

	"github.com/google/wire"
)

var MiddlewareSet = wire.NewSet(
	middlewares.NewJWTMiddleware,
)
