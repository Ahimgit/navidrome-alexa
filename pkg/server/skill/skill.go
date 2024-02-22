package skill

import (
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/skill/model/request"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"github.com/gin-gonic/gin"
	"net/http"
)

type SkillAPI struct {
	HandlerSelector *HandlerSelector
	AlexaSkillId    string
}

func NewSkillAPI(handlerSelector *HandlerSelector, alexaSkillId string) *SkillAPI {
	return &SkillAPI{
		HandlerSelector: handlerSelector,
		AlexaSkillId:    alexaSkillId,
	}
}

func (api *SkillAPI) Post(c *gin.Context) {
	var requestEnvelope request.RequestEnvelope
	if err := c.BindJSON(&requestEnvelope); err != nil {
		log.GetRequestContextLogger(c).Error("SkillAPI unable to parse request", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	if requestEnvelope.Context.System.Application.ApplicationID != api.AlexaSkillId {
		log.GetRequestContextLogger(c).Error("SkillAPI incorrect skill id in the request, unauthorized",
			"skillId", requestEnvelope.Context.System.Application.ApplicationID)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
		return
	}
	c.JSON(http.StatusOK, api.HandlerSelector.HandleRequest(&requestEnvelope, log.CreateLoggerContext(c)))
}
