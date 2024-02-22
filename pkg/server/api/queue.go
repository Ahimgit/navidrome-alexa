package api

import (
	"github.com/ahimgit/navidrome-alexa/pkg/server/api/model"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"github.com/gin-gonic/gin"
	"net/http"
)

type QueueAPI struct {
	Queue *model.Queue
}

func NewQueueAPI(queue *model.Queue) *QueueAPI {
	return &QueueAPI{
		Queue: queue,
	}
}

func (api *QueueAPI) PostQueue(c *gin.Context) {
	if err := c.BindJSON(&api.Queue); err != nil {
		log.GetRequestContextLogger(c).Error("PostQueue unable to parse request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "queue updated"})
}

func (api *QueueAPI) GetNowPlaying(c *gin.Context) { // todo: sse
	if api.Queue.HasItems() {
		c.JSON(http.StatusOK, gin.H{
			"state": api.Queue.State,
			"song":  api.Queue.Current(),
		})
	} else {
		c.JSON(http.StatusOK, gin.H{"state": "IDLE"})
	}
}

func (api *QueueAPI) GetQueue(c *gin.Context) {
	c.JSON(http.StatusOK, api.Queue)
}
