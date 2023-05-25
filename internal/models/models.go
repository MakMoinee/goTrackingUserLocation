package models

// Location Data Model
type Location struct {
	Latitude          float64 `json:"Latitude"`
	Longitude         float64 `json:"Longitude"`
	Status            string  `json:"Status"`
	SerialNumber      string  `json:"SN"`
	LastCommunication string  `json:"LastCommunication"`
}

// OutgoingResponse Data Model
type OutgoingResponse struct {
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	Status            string  `json:"status"`
	SerialNumber      string  `json:"serialNumber"`
	LastCommunication string  `json:"lastCommunication"`
}
