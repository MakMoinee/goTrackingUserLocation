package routes

import (
	"encoding/json"
	"fmt"
	"goTrackingUserLocation/internal/common"
	"goTrackingUserLocation/internal/email"
	"goTrackingUserLocation/internal/localfirebase"
	"goTrackingUserLocation/internal/models"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/MakMoinee/go-mith/pkg/goserve"
	"github.com/MakMoinee/go-mith/pkg/response"
	"github.com/go-chi/cors"
)

type service struct {
	FirebaseService localfirebase.FirebaseIntf
	EmailService    email.EmailIntf
}

type RoutesIntf interface {
	GetLocation(w http.ResponseWriter, r *http.Request)
	PostLocation(w http.ResponseWriter, r *http.Request)
}

func newRoutes() RoutesIntf {
	svc := service{}
	svc.FirebaseService = localfirebase.NewFirebaseApp()
	svc.EmailService = email.NewEmailService(common.EMAIL_PORT, common.EMAIL_HOST, common.EMAIL_ADDRESS, common.EMAIL_APP_PASS)
	return &svc
}

func Set(httpService *goserve.Service) {
	// setting cors
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "PUT", "DELETE", "POST"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-TOKEN"},
		ExposedHeaders:   []string{"Link", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	route := newRoutes()
	httpService.Router.Use(cors.Handler)
	initiateRoutes(httpService, route)

}

// initiateRoutes - initiates controllers
func initiateRoutes(httpService *goserve.Service, controllers RoutesIntf) {
	httpService.Router.Get("/", controllers.GetLocation)
	httpService.Router.Post("/location", controllers.PostLocation)
}

func (s *service) GetLocation(w http.ResponseWriter, r *http.Request) {
	log.Println("GetLocation() invoked ...")
	locationArr := []models.OutgoingResponse{}
	for _, v := range common.LOCATION_MAP {
		outgoingResponse := models.OutgoingResponse{}
		outgoingResponse.Latitude = v.Latitude
		outgoingResponse.Longitude = v.Longitude
		outgoingResponse.Status = v.Status
		outgoingResponse.SerialNumber = v.SerialNumber
		outgoingResponse.LastCommunication = v.LastCommunication
		locationArr = append(locationArr, outgoingResponse)
	}
	response.Success(w, locationArr)
}

func (s *service) PostLocation(w http.ResponseWriter, r *http.Request) {
	log.Println("PostLocation() invoked ...")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		errorBuilder := response.ErrorResponse{}
		errorBuilder.ErrorMessage = err.Error()
		errorBuilder.ErrorStatus = http.StatusInternalServerError
		response.Error(w, errorBuilder)
	}

	defer r.Body.Close()

	location := models.Location{}

	err = json.Unmarshal(body, &location)
	if err != nil {
		errorBuilder := response.ErrorResponse{}
		errorBuilder.ErrorMessage = err.Error()
		errorBuilder.ErrorStatus = http.StatusInternalServerError
		response.Error(w, errorBuilder)
	}

	if strings.EqualFold(location.Status, "Stop") {
		common.COUNTDOWN--
		if common.COUNTDOWN < 0 {
			common.COUNTDOWN = 1
			err := s.FirebaseService.SendMessage(location)
			if err == nil {
				dependents := s.FirebaseService.GetUpdatedDependents()
				if len(dependents) > 0 {
					for _, dependent := range dependents {
						go func(dep models.Dependents) {
							isSend, err := s.EmailService.SendEmail(dep.DependentEmail, fmt.Sprintf(common.ALARM_SUBJ, location.SerialNumber), fmt.Sprintf(common.ALARM_MSG, dep.DependentName, location.SerialNumber, location.SerialNumber, location.SerialNumber, common.EMAIL_ADDRESS, fmt.Sprintf(common.GOOGLE_MAP, location.Latitude, location.Longitude)))
							if err != nil {
								log.Println("Deoendent email: ", dep.DependentEmail)
								log.Println("Email error: ", err)
							}

							log.Println("Email Sent: ", isSend)
						}(dependent)
					}
				}
			}
			s.updateFirebase(w, location)
		}
	} else {
		s.updateFirebase(w, location)
	}
	common.LOCATION_MAP[location.SerialNumber] = location
	successMsg := response.NewSuccessBuilder("Successful")
	response.Success(w, successMsg)
}

func (s *service) updateFirebase(w http.ResponseWriter, location models.Location) {
	log.Println("updateFirebase() invoked ...")
	err := s.FirebaseService.WriteToDb(location)
	if err != nil {
		errorBuilder := response.ErrorResponse{}
		errorBuilder.ErrorMessage = err.Error()
		errorBuilder.ErrorStatus = http.StatusInternalServerError
		response.Error(w, errorBuilder)
		return
	} else {
		if common.HISTORY_COUNT == 0 {
			common.HISTORY_COUNT = 10
			common.HISTORY_DELETE_COUNT++
			if common.HISTORY_DELETE_COUNT == 10 {
				err := s.FirebaseService.RetrieveAndDeleteHistory(location.SerialNumber)
				if err != nil {
					log.Printf("error in retrieving and deleting history: %s", err.Error())
				} else {
					err := s.FirebaseService.WriteHistoryToDB(location)
					if err != nil {
						log.Printf("error in writing history: %s", err.Error())
					}
				}
				common.HISTORY_DELETE_COUNT = 0
			} else {
				err := s.FirebaseService.WriteHistoryToDB(location)
				if err != nil {
					log.Printf("error in writing history: %s", err.Error())
				}
			}

		} else {
			common.HISTORY_COUNT--
		}
		log.Println("Successfully Updated Firebase")
	}

}
