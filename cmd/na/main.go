package main

import (
	"bytes"
	"flag"
	"github.com/ahimgit/navidrome-alexa/pkg/server"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

func main() {
	config := parseConfiguration()
	server.StartRouter(config)
}

func parseConfiguration() *server.Config {
	config := new(server.Config)
	getStr(&config.AmazonDomain, "amazonDomain", "amazon.com", "Base domain to use for Alexa API calls.")
	getStr(&config.AmazonUser, "amazonUser", "", "Amazon account email with Alexa devices, can be left blank if auth cookies already exist.")
	getStr(&config.AmazonPassword, "amazonPassword", "", "Amazon account password, can be left blank if auth cookies already exist.")
	getStr(&config.AmazonCookiePath, "amazonCookiePath", "cookies.data", "Path to a writable file to store auth cookies.")
	getStr(&config.ApiKey, "apiKey", "", "Required. API key to authenticate /client calls.")
	getStr(&config.StreamDomain, "streamDomain", "", "Required. Navidrome public server domain URL.")
	getStr(&config.AlexaSkillId, "alexaSkillId", "", "Required. Skill id to authenticate calls from Alexa.")
	getStr(&config.AlexaSkillName, "alexaSkillName", "navi stream", "Skill invocation name.")
	getStr(&config.ListenAddress, "listenAddress", ":8080", "Listen address.")
	getBool(&config.LogIncomingRequests, "logIncomingRequests", false, "Log API and Skill requests/responses.")
	getBool(&config.LogOutgoingRequests, "logOutgoingRequests", false, "Log outgoing (to Alexa APIs) requests/responses. Will leak sensitive data into logs.")
	getBool(&config.LogStructured, "logStructured", false, "Structured logs. Much JSON, Wow!")
	flag.Parse()
	log.Init(config.LogStructured, slog.LevelDebug)
	validate("amazonDomain", config.AmazonDomain)
	validate("amazonCookiePath", config.AmazonCookiePath)
	validate("apiKey", config.ApiKey)
	validate("streamDomain", config.StreamDomain)
	validate("alexaSkillId", config.AlexaSkillId)
	validate("alexaSkillName", config.AlexaSkillName)
	validate("listenAddress", config.ListenAddress)
	return config
}

func getStr(flagPointer *string, flagName, defaultValue, usage string) {
	envVar := toEnvVarName(flagName)
	*flagPointer = defaultValue
	if value, exists := os.LookupEnv(envVar); exists {
		*flagPointer = value
	}
	flag.StringVar(flagPointer, flagName, *flagPointer, usage)
}

func getBool(flagPointer *bool, flagName string, defaultValue bool, usage string) {
	envVar := toEnvVarName(flagName)
	*flagPointer = defaultValue
	if value, exists := os.LookupEnv(envVar); exists {
		*flagPointer = value == "true" || value == "TRUE"
	}
	flag.BoolVar(flagPointer, flagName, *flagPointer, usage)
}

func toEnvVarName(s string) string {
	var re = regexp.MustCompile("([A-Z])")
	snake := re.ReplaceAllString(s, "_$1")
	return "NA_" + strings.ToUpper(strings.TrimLeft(snake, "_"))
}

func validate(name string, value string) {
	if value == "" {
		buf := new(bytes.Buffer)
		flag.CommandLine.SetOutput(buf)
		flag.PrintDefaults()
		log.Logger().Error("Error. Param " + name + " is required.")
		log.Logger().Info("Usage:\n" + buf.String())
		os.Exit(0)
	}
}
