package model

type DevicesResponse struct {
	Devices []PlayerDevice `json:"devices"`
}

type PlayerDevice struct {
	Name                  string `json:"name,omitempty"`
	DeviceOwnerCustomerId string `json:"deviceOwnerCustomerId"`
	DeviceType            string `json:"deviceType"`
	SerialNumber          string `json:"serialNumber"`
}
