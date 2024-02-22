package mid

import (
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/client"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type HealthResponse struct {
	StatusCode int
	Body       gin.H
	UpdatedAt  time.Time
}

type Health struct {
	AlexaClient client.IAlexaClient
}

func NewHealth(alexaClient client.IAlexaClient) *Health {
	return &Health{
		AlexaClient: alexaClient,
	}
}

func (api *Health) GetHealth(context *gin.Context) {
	devices, err := api.AlexaClient.GetDevices()
	var response *HealthResponse
	if err != nil {
		response = &HealthResponse{
			UpdatedAt:  time.Now(),
			StatusCode: http.StatusInternalServerError,
			Body: gin.H{
				"status": "dead",
				"error":  err.Error(),
			},
		}
	} else {
		response = &HealthResponse{
			UpdatedAt:  time.Now(),
			StatusCode: http.StatusOK,
			Body: gin.H{
				"status":  "ok",
				"devices": len(devices.Devices),
			},
		}
	}
	context.JSON(response.StatusCode, response.Body)
}
