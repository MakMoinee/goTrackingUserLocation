package main

import (
	"goTrackingUserLocation/cmd/webapp/config"
	"goTrackingUserLocation/cmd/webapp/routes"
	"goTrackingUserLocation/internal/common"
	"goTrackingUserLocation/internal/models"
	"log"

	"github.com/MakMoinee/go-mith/pkg/goserve"
)

func main() {
	config.Set()
	common.LOCATION_MAP = make(map[string]models.Location)
	httpService := goserve.NewService(common.SERVER_PORT)
	routes.Set(httpService)
	log.Println("Service Started at: http://localhost", common.SERVER_PORT)
	if err := httpService.Start(); err != nil {
		panic(err)
	}
}
