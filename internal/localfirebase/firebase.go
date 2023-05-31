package localfirebase

import (
	"context"
	"fmt"
	"goTrackingUserLocation/internal/common"
	"goTrackingUserLocation/internal/models"
	"log"
	"net/smtp"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FirebaseIntf interface {
	Setup()
	WriteToDb(models.Location) error
	SendMessage(models.Location) error
}

type service struct {
	App             *firebase.App
	DBClient        *db.Client
	MessageClient   *messaging.Client
	FirestoreClient *firestore.Client
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

	messenger, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("(3) Failed to access messaging: %v", err)
	}
	s.MessageClient = messenger

}

func (s *service) openFirestore() {
	firestore, err := s.App.Firestore(context.Background())
	if err != nil {
		log.Fatalf("(4) Failed to access firestore: %v", err)
	}

	s.FirestoreClient = firestore
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

// func (s *service) retrieveUser(deviceID string) []string {
// 	deviceToken := []string{}
// 	s.openFirestore()
// 	docRef := s.FirestoreClient.Collection("users").WhereEq
// 	tmpMap := make(map[string]interface{})
// 	defer s.FirestoreClient.Close()

// 	snap, err := docRef.Get(context.Background())
// 	if err != nil {
// 		log.Fatalf("(5) Failed to get firestore data: %v", err)
// 	}

// 	if err != nil {
// 		log.Fatalf("Failed to get document: %v", err)
// 	}

// 	// Unmarshal the snapshot data into the user struct
// 	if err := snap.DataTo(&tmpMap); err != nil {
// 		log.Fatalf("Failed to unmarshal document data: %v", err)
// 	}

// 	if tmpMap["deviceTokens"] != nil {
// 		for _, rawData := range tmpMap["deviceTokens"].([]interface{}) {
// 			if rawData.(map[string]interface{})["deviceToken"] != nil {
// 				tmpStr := rawData.(map[string]interface{})["deviceToken"].(string)
// 				deviceToken = append(deviceToken, tmpStr)
// 			}

// 		}
// 	}

// 	return deviceToken
// }

func (s *service) retrieveDeviceToken(deviceID string) []string {
	deviceToken := []string{}
	s.openFirestore()
	docRef := s.FirestoreClient.Collection("deviceTokens").Doc(deviceID)
	tmpMap := make(map[string]interface{})
	defer s.FirestoreClient.Close()

	snap, err := docRef.Get(context.Background())
	if err != nil {
		log.Println(fmt.Sprintf("(5) Failed to get firestore data: %v", err))
		return deviceToken
	}

	if err != nil {
		log.Println(fmt.Sprintf("(6) Failed to get firestore data: %v", err))
		return deviceToken
	}

	// Unmarshal the snapshot data into the user struct
	if err := snap.DataTo(&tmpMap); err != nil {
		log.Fatalf("Failed to unmarshal document data: %v", err)
	}

	if tmpMap["deviceTokens"] != nil {
		for _, rawData := range tmpMap["deviceTokens"].([]interface{}) {
			if rawData.(map[string]interface{})["deviceToken"] != nil {
				tmpStr := rawData.(map[string]interface{})["deviceToken"].(string)
				deviceToken = append(deviceToken, tmpStr)
			}

		}
	}

	return deviceToken
}

func (s *service) SendMessage(location models.Location) error {
	log.Println("SendMessage() invoked ...")
	deviceToken := s.retrieveDeviceToken(location.SerialNumber)
	var errs error
	for _, token := range deviceToken {
		if token == "" {
			continue
		}
		message := &messaging.Message{
			Notification: &messaging.Notification{
				Title: "Device Notification",
				Body:  fmt.Sprintf("Person in Device: %s might be in danger, Please open this notification to track", location.SerialNumber),
			},
			Token: token, // Replace with the device token of the target Android device
		}

		data, err := s.MessageClient.Send(context.Background(), message)
		if err != nil {
			log.Fatalf("(4) Failed to send message: %v", err)
			errs = err
			break
		}

		log.Println("Successfully Sent Message: ", data)
	}

	return errs
}

// createEmail creates a new email message.
func createEmail(to, subject, body string) []byte {
	message := fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", to, subject, body)

	return []byte(message)
}

// sendEmail sends the email using SMTP and the provided app password.
func sendEmail(senderEmail, appPassword string, email []byte) error {
	auth := smtp.PlainAuth("", senderEmail, appPassword, "smtp.gmail.com")

	err := smtp.SendMail("smtp.gmail.com:587", auth, senderEmail, []string{senderEmail}, email)
	if err != nil {
		return err
	}

	return nil
}
