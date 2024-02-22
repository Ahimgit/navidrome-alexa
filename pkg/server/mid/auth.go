package mid

import (
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func ApiKeyAuthMiddleware(pathPrefix string, apiKey string) gin.HandlerFunc {
	return func(context *gin.Context) {
		if strings.HasPrefix(context.Request.URL.Path, pathPrefix) {
			authParam := context.Query("apiKey")
			if authParam != "" {
				if authParam != apiKey {
					log.GetRequestContextLogger(context).Error("Auth incorrect api key, unauthorized")
					context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}
			} else {
				authHeader := context.GetHeader("Authorization")
				if authHeader == "" {
					log.GetRequestContextLogger(context).Error("Auth empty, unauthorized")
					context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}
				authHeaderParts := strings.SplitN(authHeader, " ", 2)
				if !(len(authHeaderParts) == 2 && authHeaderParts[0] == "Bearer") {
					log.GetRequestContextLogger(context).Error("Auth incorrect auth type, unauthorized")
					context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}
				if authHeaderParts[1] != apiKey {
					log.GetRequestContextLogger(context).Error("Auth incorrect api key, unauthorized")
					context.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"status": "error", "message": "Unauthorized"})
					return
				}
			}
		}
		context.Next()
	}

}
