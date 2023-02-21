//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"
	"github.com/gozelle/gin"
	injector2 "github.com/koyeo/nest/hub/injector"
	"github.com/koyeo/nest/hub/internal/config"
)

func wireApp(config *config.Config) (*gin.Engine, func(), error) {
	panic(wire.Build(
		injector2.GormProviderSet,
		injector2.ApiProviderSet,
		injector2.PublisherSet,
		newApp,
	))
}
