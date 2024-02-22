package model

type DevicesResponse struct {
	Devices []Device `json:"devices"`
}

type Device struct {
	AccountName string `json:"accountName"`
	//AppDeviceList                 `json:"appDeviceList"`
	AssociatedUnitIds      *string  `json:"associatedUnitIds"`
	Capabilities           []string `json:"capabilities"`
	Charging               *bool    `json:"charging"`
	ClusterMembers         []string `json:"clusterMembers"`
	DeviceAccountId        string   `json:"deviceAccountId"`
	DeviceFamily           string   `json:"deviceFamily"`
	DeviceOwnerCustomerId  string   `json:"deviceOwnerCustomerId"`
	DeviceType             string   `json:"deviceType"`
	DeviceTypeFriendlyName *string  `json:"deviceTypeFriendlyName"`
	Essid                  *string  `json:"essid"`
	Language               *string  `json:"language"`
	MacAddress             *string  `json:"macAddress"`
	Online                 bool     `json:"online"`
	ParentClusters         []string `json:"parentClusters"`
	PostalCode             *string  `json:"postalCode"`
	RegistrationId         *string  `json:"registrationId"`
	RemainingBatteryLevel  *int     `json:"remainingBatteryLevel"`
	SerialNumber           string   `json:"serialNumber"`
	SoftwareVersion        string   `json:"softwareVersion"`
}
