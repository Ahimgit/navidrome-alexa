package server

import (
	alexa "github.com/ahimgit/navidrome-alexa/pkg/alexa/client"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/client/httpclient"
	server "github.com/ahimgit/navidrome-alexa/pkg/server/api"
	"github.com/ahimgit/navidrome-alexa/pkg/server/api/model"
	"github.com/ahimgit/navidrome-alexa/pkg/server/mid"
	"github.com/ahimgit/navidrome-alexa/pkg/server/skill"
	"github.com/ahimgit/navidrome-alexa/pkg/server/ui"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/gin"
	"time"
)

type Config struct {
	AmazonDomain        string
	AmazonUser          string
	AmazonPassword      string
	AmazonCookiePath    string
	AlexaSkillId        string
	AlexaSkillName      string
	StreamDomain        string
	ApiKey              string
	ListenAddress       string
	LogIncomingRequests bool
	LogOutgoingRequests bool
}

func StartRouter(config *Config) {
	store := persistence.NewInMemoryStore(time.Minute)
	queue := model.NewQueue()
	alexaClient := initAlexaClient(
		config.AmazonDomain,
		config.AmazonUser,
		config.AmazonPassword,
		config.AmazonCookiePath,
		config.LogOutgoingRequests,
	)
	healthCheck := mid.NewHealth(alexaClient)
	queueAPI := server.NewQueueAPI(queue)
	playerAPI := server.NewPlayerAPI(alexaClient, config.AlexaSkillName)
	skillHandler := skill.NewHandlerSelector(queue, config.StreamDomain)
	skillAPI := skill.NewSkillAPI(skillHandler, config.AlexaSkillId)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(gin.Recovery())
	engine.Use(mid.CorsMiddleware())
	engine.Use(mid.RequestLogsMiddleware(config.LogIncomingRequests))
	engine.Use(mid.ApiKeyAuthMiddleware("/api/", config.ApiKey))
	engine.Use(mid.MetricsMiddleware("/metrics", engine))
	engine.GET("/health", cached(healthCheck.GetHealth, store))

	engine.GET("/api/playing", queueAPI.GetNowPlaying) // player api
	engine.GET("/api/queue", queueAPI.GetQueue)
	engine.POST("/api/queue", queueAPI.PostQueue)
	engine.POST("/api/play", playerAPI.PostPlay)
	engine.POST("/api/stop", playerAPI.PostStop)
	engine.POST("/api/next", playerAPI.PostNext)
	engine.POST("/api/prev", playerAPI.PostPrev)
	engine.POST("/api/volume", playerAPI.PostVolume)
	engine.GET("/api/volume", playerAPI.GetVolume)
	engine.GET("/api/devices", cached(playerAPI.GetDevices, store))

	engine.POST("/skill", skillAPI.Post) // alexa skill api

	engine.GET("/proxy", ui.GetWidget) // ui widget
	engine.StaticFS("/static", ui.NewEmbedFileSystem())

	log.Logger().Info("Starting NA", "listenAddress", config.ListenAddress)
	log.Logger().Error("Error starting server", engine.Run(config.ListenAddress))
}

func initAlexaClient(amazonDomain string, amazonUser string, amazonPassword string, amazonCookiePath string, logRequests bool) alexa.IAlexaClient {
	var client alexa.IAlexaClient
	if logRequests {
		http := httpclient.NewHttpClient().WithResponseLogger(mid.RequestLogsForClients())
		cookie := httpclient.NewCookieHelper(amazonCookiePath)
		client = alexa.NewAlexaClientWithHttpClient(amazonDomain, amazonUser, amazonPassword, cookie, http)
	} else {
		client = alexa.NewAlexaClient(amazonDomain, amazonUser, amazonPassword, amazonCookiePath)
	}
	if err := client.LogIn(); err != nil {
		log.Logger().Error("Unable to log in to Alexa account", err)
	}
	return client
}

func cached(handler gin.HandlerFunc, store *persistence.InMemoryStore) gin.HandlerFunc {
	return cache.CachePageWithoutQuery(store, time.Minute, handler)
}
