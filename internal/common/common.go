package common

import "goTrackingUserLocation/internal/models"

var (
	SERVER_PORT  string
	LOCATION_MAP map[string]models.Location
)

const (
	DEVICES_REF = "devices"
)
