package main

import (
	"fmt"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/client"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/client/model"
	"os"
	"time"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Please provide domain user and password e.g.: meow amazon.com your_amazon_user@email.com your_amazon_password")
		os.Exit(1)
	}
	alexaClient := client.NewAlexaClient(os.Args[1], os.Args[2], os.Args[3], "cookies.data")
	err := alexaClient.LogIn(false)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	devices, err := alexaClient.GetDevices()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if devices.Devices == nil || len(devices.Devices) == 0 {
		fmt.Println("No devices on the account")
	}
	for i, device := range devices.Devices {
		fmt.Println(i, "Device: ",
			device.DeviceType,
			device.SerialNumber,
			device.AccountName,
			device.DeviceOwnerCustomerId,
		)
		fmt.Println("Meow!")
		time.Sleep(3 * time.Second)
		err = alexaClient.PostSequenceCmd(model.BuildSpeakCmd(
			`<audio src="soundbank://soundlibrary/animals/amzn_sfx_cat_angry_meow_1x_02"/>`, "en-US",
			device.DeviceType,
			device.SerialNumber,
			device.DeviceOwnerCustomerId),
		)
		if err != nil {
			fmt.Println(err)
		}
	}
}
