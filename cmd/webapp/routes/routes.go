package routes

import (
	"encoding/json"
	"goTrackingUserLocation/internal/common"
	"goTrackingUserLocation/internal/models"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/MakMoinee/go-mith/pkg/goserve"
	"github.com/MakMoinee/go-mith/pkg/response"
	"github.com/go-chi/cors"
)

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
	httpService.Router.Use(cors.Handler)
	initiateRoutes(httpService)

}

// initiateRoutes - initiates controllers
func initiateRoutes(httpService *goserve.Service) {
	httpService.Router.Get("/", GetLocation)
	httpService.Router.Post("/location", PostLocation)
}

func GetLocation(w http.ResponseWriter, r *http.Request) {
	log.Println("GetLocation() invoked ...")
	w.Write([]byte("Sample word"))
}

func PostLocation(w http.ResponseWriter, r *http.Request) {
	log.Println("PostLocation() invoked ...")
	body, err := ioutil.ReadAll(r.Body)
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

	common.LOCATION_MAP[location.SerialNumber] = location
	successMsg := response.NewSuccessBuilder("Successful")
	response.Success(w, successMsg)
}
