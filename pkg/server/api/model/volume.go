package model

type VolumeResponse struct {
	Volumes []DeviceVolume `json:"volumes"`
}

type VolumeRequest struct {
	Device PlayerDevice `json:"device"`
	Volume int          `json:"volume"`
}

type DeviceVolume struct {
	DeviceSerialNumber string `json:"deviceSerialNumber"`
	Muted              bool   `json:"muted"`
	Volume             int    `json:"volume"`
}
