package api

import (
	alexaClient "github.com/ahimgit/navidrome-alexa/pkg/alexa/client"
	alexaModel "github.com/ahimgit/navidrome-alexa/pkg/alexa/client/model"
	apiModel "github.com/ahimgit/navidrome-alexa/pkg/server/api/model"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"github.com/gin-gonic/gin"
	"net/http"
)

type PlayerAPI struct {
	SkillName   string
	AlexaClient alexaClient.IAlexaClient
}

func NewPlayerAPI(alexaClient alexaClient.IAlexaClient, skillName string) *PlayerAPI {
	return &PlayerAPI{
		SkillName:   skillName,
		AlexaClient: alexaClient,
	}
}

func (playerAPI *PlayerAPI) PostPlay(c *gin.Context) {
	executeTextCommand(c, playerAPI, "play")
}

func (playerAPI *PlayerAPI) PostStop(c *gin.Context) {
	executeTextCommand(c, playerAPI, "stop")
}

func (playerAPI *PlayerAPI) PostNext(c *gin.Context) {
	executeTextCommand(c, playerAPI, "next")
}

func (playerAPI *PlayerAPI) PostPrev(c *gin.Context) {
	executeTextCommand(c, playerAPI, "previous")
}

func (playerAPI *PlayerAPI) GetDevices(c *gin.Context) {
	devices, err := playerAPI.AlexaClient.GetDevices()
	if err != nil {
		log.GetRequestContextLogger(c).Error("GetDevices failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if devices.Devices == nil || len(devices.Devices) == 0 {
		log.GetRequestContextLogger(c).Warn("GetDevices, no devices on the account")
		c.JSON(http.StatusNotFound, gin.H{"status": "error", "message": "No devices on the account"})
		return
	}
	c.JSON(http.StatusOK, mapDevicesResponse(devices))
}

func (playerAPI *PlayerAPI) GetVolume(c *gin.Context) {
	volume, err := playerAPI.AlexaClient.GetVolume()
	if err != nil {
		log.GetRequestContextLogger(c).Error("GetVolume failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mapVolumeResponse(volume))
}

func (playerAPI *PlayerAPI) PostVolume(c *gin.Context) {
	var volumeRequest apiModel.VolumeRequest
	if err := c.BindJSON(&volumeRequest); err != nil {
		log.GetRequestContextLogger(c).Error("PostVolume unable to parse request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if err := playerAPI.AlexaClient.PostSequenceCmd(alexaModel.BuildVolumeCmd(
		volumeRequest.Volume, "en-US",
		volumeRequest.Device.DeviceType,
		volumeRequest.Device.SerialNumber,
		volumeRequest.Device.DeviceOwnerCustomerId),
	); err != nil {
		log.GetRequestContextLogger(c).Error("PostVolume failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "volume updated"})
}

func executeTextCommand(c *gin.Context, playerAPI *PlayerAPI, command string) {
	var playerDevice apiModel.PlayerDevice // request model
	if err := c.BindJSON(&playerDevice); err != nil {
		log.GetRequestContextLogger(c).Error("TextCommand unable to parse request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if err := playerAPI.AlexaClient.PostSequenceCmd(alexaModel.BuildTextCommandCmd(
		"ask "+playerAPI.SkillName+" to "+command, "en-US",
		playerDevice.DeviceType,
		playerDevice.SerialNumber,
		playerDevice.DeviceOwnerCustomerId),
	); err != nil {
		log.GetRequestContextLogger(c).Error("TextCommand failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": command + " executed"})
}

func mapDevicesResponse(input alexaModel.DevicesResponse) (output apiModel.DevicesResponse) {
	for _, device := range input.Devices {
		for _, capability := range device.Capabilities {
			if capability == "AUDIO_PLAYER" && device.DeviceFamily != "WHA" { // WHA - multiroom devices don't seem to work with skills unfortunately
				playerDevice := apiModel.PlayerDevice{
					Name:                  device.AccountName,
					DeviceOwnerCustomerId: device.DeviceOwnerCustomerId,
					DeviceType:            device.DeviceType,
					SerialNumber:          device.SerialNumber,
				}
				output.Devices = append(output.Devices, playerDevice)
				break // to next device
			}
		}
	}
	return output
}

func mapVolumeResponse(input alexaModel.VolumeResponse) (output apiModel.VolumeResponse) {
	for _, volume := range input.Volumes {
		output.Volumes = append(output.Volumes, apiModel.DeviceVolume{
			Volume:             volume.SpeakerVolume,
			Muted:              volume.SpeakerMuted,
			DeviceSerialNumber: volume.Dsn,
		})
	}
	return output
}
