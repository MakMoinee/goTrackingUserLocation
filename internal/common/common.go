package common

import "goTrackingUserLocation/internal/models"

var (
	SERVER_PORT    string
	LOCATION_MAP   map[string]models.Location
	COUNTDOWN      = 1
	EMAIL_ADDRESS  string
	EMAIL_PORT     int
	EMAIL_APP_PASS string
	EMAIL_HOST     string
	EMAIL_SUBJECT  string
)

const (
	DEVICES_REF = "devices"
)
