package main

import (
	"bytes"
	"flag"
	"github.com/ahimgit/navidrome-alexa/pkg/server"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"log/slog"
	"os"
)

func main() {
	config := new(server.Config)
	//todo secure way of passing conf
	flag.StringVar(&config.AmazonDomain, "amazonDomain", "amazon.com", "Base domain to use for Alexa API calls.")
	flag.StringVar(&config.AmazonUser, "amazonUser", "", "Amazon account email with Alexa devices, can be left blank if auth cookies already exist.")
	flag.StringVar(&config.AmazonPassword, "amazonPassword", "", "Amazon account password, can be left blank if auth cookies already exist.")
	flag.StringVar(&config.AmazonCookiePath, "amazonCookiePath", "cookies.data", "Path to a writable file to store auth cookies.")
	flag.StringVar(&config.ApiKey, "apiKey", "", "Required. API key to authenticate /client calls.")
	flag.StringVar(&config.StreamDomain, "streamDomain", "", "Required. Navidrome public server domain URL.")
	flag.StringVar(&config.AlexaSkillId, "alexaSkillId", "", "Required. Skill id to authenticate calls from Alexa.")
	flag.StringVar(&config.AlexaSkillName, "alexaSkillName", "navi stream", "Skill invocation name.")
	flag.StringVar(&config.ListenAddress, "listenAddress", ":8080", "Listen address.")
	flag.BoolVar(&config.LogIncomingRequests, "logIncomingRequests", false, "Log API and Skill requests/responses.")
	flag.BoolVar(&config.LogOutgoingRequests, "logOutgoingRequests", false, "Log outgoing (to Alexa APIs) requests/responses. Will leak sensitive data into logs. ")
	flag.Parse()
	log.Init(false, slog.LevelDebug)
	validate("amazonDomain", config.AmazonDomain)
	validate("amazonCookiePath", config.AmazonCookiePath)
	validate("apiKey", config.ApiKey)
	validate("streamDomain", config.StreamDomain)
	validate("alexaSkillId", config.AlexaSkillId)
	validate("alexaSkillName", config.AlexaSkillName)
	validate("listenAddress", config.ListenAddress)
	server.StartRouter(config)
}

func validate(name string, value string) {
	if value == "" {
		buf := new(bytes.Buffer)
		flag.CommandLine.SetOutput(buf)
		flag.PrintDefaults()
		log.Logger().Error("Error. Param " + name + " required.")
		log.Logger().Info("Usage:\n" + buf.String())
		os.Exit(1)
	}
}
