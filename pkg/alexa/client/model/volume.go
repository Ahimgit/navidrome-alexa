package model

type VolumeResponse struct {
	Volumes []Volume `json:"volumes"`
}

type Volume struct {
	AlertVolume   *int    `json:"alertVolume,omitempty"`
	DeviceType    string  `json:"deviceType"`
	Dsn           string  `json:"dsn"`
	Error         *string `json:"error,omitempty"`
	SpeakerMuted  bool    `json:"speakerMuted"`
	SpeakerVolume int     `json:"speakerVolume"`
}
