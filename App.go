package main

import (
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/support"
)

var info = ninja.LoadModuleInfo("./package.json")

type AppConfig struct{}

type App struct {
	support.AppSupport
}

func (a *App) Start(cfg *AppConfig) error {
	return nil
}

func (a *App) Stop() error {
	return nil
}
