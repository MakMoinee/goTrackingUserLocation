package common

import "goTrackingUserLocation/internal/models"

var (
	SERVER_PORT  string
	LOCATION_MAP map[string]models.Location
	COUNTDOWN    = 1
)

const (
	DEVICES_REF = "devices"
)
