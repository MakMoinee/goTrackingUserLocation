package localfirebase

import (
	"context"
	"goTrackingUserLocation/internal/common"
	"goTrackingUserLocation/internal/models"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"google.golang.org/api/option"
)

type FirebaseIntf interface {
	Setup()
	WriteToDb(models.Location) error
}

type service struct {
	App      *firebase.App
	DBClient *db.Client
}

func NewFirebaseApp() FirebaseIntf {
	svc := service{}
	svc.Setup()
	return &svc
}

func (s *service) Setup() {
	opt := option.WithCredentialsFile("../../firebase.json")
	config := &firebase.Config{ProjectID: "trackinguserapp", DatabaseURL: "https://trackinguserapp-default-rtdb.asia-southeast1.firebasedatabase.app"}
	app, err := firebase.NewApp(context.Background(), config, opt)
	if err != nil {
		log.Fatalf(" (1) Failed to initialize Firebase app: %v", err)
	}
	s.App = app

	ctx := context.Background()

	db, err := app.Database(ctx)
	if err != nil {
		log.Fatalf("(2) Failed to access database: %v", err)
	}
	s.DBClient = db
}

func (s *service) WriteToDb(location models.Location) error {
	ref := s.DBClient.NewRef(common.DEVICES_REF)

	err := ref.Child(location.SerialNumber).Set(context.Background(), map[string]interface{}{
		"latitude":          location.Latitude,
		"longitude":         location.Longitude,
		"status":            location.Status,
		"serialNumber":      location.SerialNumber,
		"lastCommunication": location.LastCommunication,
	})

	return err
}
