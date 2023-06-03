package localfirebase

import (
	"context"
	"fmt"
	"goTrackingUserLocation/internal/common"
	"goTrackingUserLocation/internal/models"
	"log"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

type FirebaseIntf interface {
	Setup()
	GetUpdatedDependents() []models.Dependents
	WriteToDb(models.Location) error
	SendMessage(models.Location) error
	WriteHistoryToDB(location models.Location) error
	RetrieveAndDeleteHistory(sn string) error
}

type service struct {
	App             *firebase.App
	DBClient        *db.Client
	MessageClient   *messaging.Client
	FirestoreClient *firestore.Client
	Dependents      []models.Dependents
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
	s.Dependents = []models.Dependents{}

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

func (s *service) WriteHistoryToDB(location models.Location) error {
	s.openFirestore()
	ref := s.FirestoreClient.Collection(common.HISTORY_REF)

	_, _, err := ref.Add(context.Background(), location)
	defer s.FirestoreClient.Close()
	return err
}

func (s *service) RetrieveAndDeleteHistory(sn string) error {
	var err error
	ids := s.retrieveHistoryIds(sn)
	for _, id := range ids {
		err = s.deleteHistory(id)
		if err != nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

func (s *service) retrieveHistoryIds(sn string) []string {
	s.openFirestore()
	defer s.FirestoreClient.Close()
	ids := []string{}
	docRef := s.FirestoreClient.Collection("history").Where("SerialNumber", "==", sn)
	documents, err := docRef.Documents(context.Background()).GetAll()
	if err != nil {
		log.Printf("error retrieving history: %s", err.Error())
		return ids
	}

	for _, d := range documents {
		ids = append(ids, d.Ref.ID)
	}
	return ids
}

func (s *service) deleteHistory(historyDocID string) error {
	s.openFirestore()
	docRef := s.FirestoreClient.Collection("history").Doc(historyDocID)
	_, err := docRef.Delete(context.Background())
	if err != nil {
		log.Printf("error deleting history: %s", err.Error())
	}
	defer s.FirestoreClient.Close()
	return err
}

func (s *service) retrieveDependents(userID string) []models.Dependents {
	s.openFirestore()
	listOfDependents := []models.Dependents{}
	docRef := s.FirestoreClient.Collection("dependents").Where("userID", "==", userID)
	defer s.FirestoreClient.Close()

	documents, err := docRef.Documents(context.Background()).GetAll()
	if err != nil {
		log.Println(fmt.Sprintf("(7) failed to get firestore data: %v", err))

		return listOfDependents
	}

	// Iterate over the retrieved documents
	for _, doc := range documents {
		// fmt.Printf("Document ID: %s\n", doc.Ref.ID)
		dependent := models.Dependents{}
		if doc.Data()["dependentEmail"] != nil {
			dependent.DependentEmail = doc.Data()["dependentEmail"].(string)
		}

		if doc.Data()["dependentName"] != nil {
			dependent.DependentName = doc.Data()["dependentName"].(string)
		}

		if doc.Data()["dependentPhoneNumber"] != nil {
			dependent.DependentPhoneNumber = doc.Data()["dependentPhoneNumber"].(string)
		}

		if dependent.DependentEmail == "" {
			continue
		}

		listOfDependents = append(listOfDependents, dependent)
	}

	return listOfDependents
}

func (s *service) retrieveDeviceToken(deviceID string) ([]string, []string) {
	deviceToken := []string{}
	users := []string{}
	deviceUsers := make(map[string]interface{})
	s.openFirestore()
	docRef := s.FirestoreClient.Collection("deviceTokens").Doc(deviceID)
	tmpMap := make(map[string]interface{})
	defer s.FirestoreClient.Close()

	snap, err := docRef.Get(context.Background())
	if err != nil {
		log.Println(fmt.Sprintf("(5) failed to get firestore data: %v", err))
		return deviceToken, users
	}

	if err != nil {
		log.Println(fmt.Sprintf("(6) failed to get firestore data: %v", err))
		return deviceToken, users
	}

	// Unmarshal the snapshot data into the user struct
	if err := snap.DataTo(&tmpMap); err != nil {
		log.Printf("Failed to unmarshal document data: %v", err)
		return deviceToken, users
	}

	if tmpMap["deviceTokens"] != nil {
		for _, rawData := range tmpMap["deviceTokens"].([]interface{}) {
			if rawData.(map[string]interface{})["deviceToken"] != nil {
				tmpStr := rawData.(map[string]interface{})["deviceToken"].(string)
				deviceToken = append(deviceToken, tmpStr)
			}

			if rawData.(map[string]interface{})["userID"] != nil {
				tmpStr := rawData.(map[string]interface{})["userID"].(string)
				deviceUsers[tmpStr] = nil
			}

		}
	}

	for k := range deviceUsers {
		users = append(users, k)
	}

	return deviceToken, users
}

func (s *service) SendMessage(location models.Location) error {
	log.Println("SendMessage() invoked ...")
	deviceToken, deviceUsers := s.retrieveDeviceToken(location.SerialNumber)
	s.getDependents(deviceUsers)
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
			log.Printf("(4) Failed to send message: %v", err)
			errs = err
			break
		}

		log.Println("Successfully Sent Message: ", data)
	}

	return errs
}

func (s *service) getDependents(userIDs []string) {
	allDependents := []models.Dependents{}
	for _, id := range userIDs {
		dependents := s.retrieveDependents(id)
		if len(dependents) > 0 {
			allDependents = append(allDependents, dependents...)
		}
	}

	s.Dependents = allDependents
}

func (s *service) GetUpdatedDependents() []models.Dependents {
	return s.Dependents
}
