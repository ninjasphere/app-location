package main

import (
	"github.com/ninjasphere/app-location/calibration"
	"github.com/ninjasphere/app-location/ui"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/go-ninja/support"
)

var info = ninja.LoadModuleInfo("./package.json")

type AppConfig struct{}

type App struct {
	support.AppSupport
}

func (a *App) Start(cfg *AppConfig) error {

	calibrationService := calibration.NewService(a.Conn)

	a.Conn.MustExportService(ui.NewUI(calibrationService), "$app/"+info.ID+"/configure", &model.ServiceAnnouncement{
		Schema: "/protocol/configuration",
	})

	return nil
}

func (a *App) Stop() error {
	return nil
}
