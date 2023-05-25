package routes

import (
	"encoding/json"
	"goTrackingUserLocation/internal/common"
	"goTrackingUserLocation/internal/localfirebase"
	"goTrackingUserLocation/internal/models"
	"io"
	"log"
	"net/http"

	"github.com/MakMoinee/go-mith/pkg/goserve"
	"github.com/MakMoinee/go-mith/pkg/response"
	"github.com/go-chi/cors"
)

type service struct {
	FirebaseService localfirebase.FirebaseIntf
}

type RoutesIntf interface {
	GetLocation(w http.ResponseWriter, r *http.Request)
	PostLocation(w http.ResponseWriter, r *http.Request)
}

func newRoutes() RoutesIntf {
	svc := service{}
	svc.FirebaseService = localfirebase.NewFirebaseApp()
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

	if val, exist := common.LOCATION_MAP[location.SerialNumber]; exist {
		if location.Status != val.Status {
			s.updateFirebase(w, location)
		}
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
		log.Println("Successfully Updated Firebase")
	}

}
