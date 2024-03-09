package main

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"sync"
	"testing"
)

var mutexEnv sync.Mutex
var mutexArg sync.Mutex

func TestParseConfiguration(t *testing.T) {

	t.Run("parse basic env config with defaults ", func(t *testing.T) {
		withArgs([]string{"command"}, func() {
			withEnv(map[string]string{
				"NA_AMAZON_USER":     "amazonUserValue",
				"NA_AMAZON_PASSWORD": "amazonPasswordValue",
				"NA_ALEXA_SKILL_ID":  "alexaSkillIdValue",
				"NA_STREAM_DOMAIN":   "navidrome.example.com",
				"NA_API_KEY":         "apiKeyValue",
			}, func() {
				config := parseConfiguration()
				assert.Equal(t, "amazon.com", config.AmazonDomain)
				assert.Equal(t, "amazonUserValue", config.AmazonUser)
				assert.Equal(t, "amazonPasswordValue", config.AmazonPassword)
				assert.Equal(t, "cookies.data", config.AmazonCookiePath)
				assert.Equal(t, "alexaSkillIdValue", config.AlexaSkillId)
				assert.Equal(t, "navi stream", config.AlexaSkillName)
				assert.Equal(t, "navidrome.example.com", config.StreamDomain)
				assert.Equal(t, "apiKeyValue", config.ApiKey)
				assert.Equal(t, ":8080", config.ListenAddress)
				assert.Equal(t, false, config.LogIncomingRequests)
				assert.Equal(t, false, config.LogOutgoingRequests)
				assert.Equal(t, false, config.LogStructured)
			})
		})
	})

	t.Run("parse basic cmd config with defaults ", func(t *testing.T) {
		withArgs([]string{"command",
			"-amazonUser", "amazonUserValue",
			"-amazonPassword", "amazonPasswordValue",
			"-alexaSkillId", "alexaSkillIdValue",
			"-streamDomain", "navidrome.example.com",
			"-apiKey", "apiKeyValue",
		}, func() {
			config := parseConfiguration()
			assert.Equal(t, "amazon.com", config.AmazonDomain)
			assert.Equal(t, "amazonUserValue", config.AmazonUser)
			assert.Equal(t, "amazonPasswordValue", config.AmazonPassword)
			assert.Equal(t, "cookies.data", config.AmazonCookiePath)
			assert.Equal(t, "alexaSkillIdValue", config.AlexaSkillId)
			assert.Equal(t, "navi stream", config.AlexaSkillName)
			assert.Equal(t, "navidrome.example.com", config.StreamDomain)
			assert.Equal(t, "apiKeyValue", config.ApiKey)
			assert.Equal(t, ":8080", config.ListenAddress)
			assert.Equal(t, false, config.LogIncomingRequests)
			assert.Equal(t, false, config.LogOutgoingRequests)
			assert.Equal(t, false, config.LogStructured)

		})
	})

	t.Run("parse with all cmd line args", func(t *testing.T) {
		withArgs([]string{"command",
			"-amazonDomain", "amazon.example.com",
			"-amazonUser", "amazonUserValue",
			"-amazonPassword", "amazonPasswordValue",
			"-amazonCookiePath", "amazonCookiePathValue",
			"-alexaSkillId", "alexaSkillIdValue",
			"-alexaSkillName", "alexaSkillNameValue",
			"-streamDomain", "navidrome.example.com",
			"-apiKey", "apiKeyValue",
			"-listenAddress", "localhost:9090",
			"-logIncomingRequests",
			"-logOutgoingRequests",
			"-logStructured",
		}, func() {
			config := parseConfiguration()
			assert.Equal(t, "amazon.example.com", config.AmazonDomain)
			assert.Equal(t, "amazonUserValue", config.AmazonUser)
			assert.Equal(t, "amazonPasswordValue", config.AmazonPassword)
			assert.Equal(t, "amazonCookiePathValue", config.AmazonCookiePath)
			assert.Equal(t, "alexaSkillIdValue", config.AlexaSkillId)
			assert.Equal(t, "alexaSkillNameValue", config.AlexaSkillName)
			assert.Equal(t, "navidrome.example.com", config.StreamDomain)
			assert.Equal(t, "apiKeyValue", config.ApiKey)
			assert.Equal(t, "localhost:9090", config.ListenAddress)
			assert.Equal(t, true, config.LogIncomingRequests)
			assert.Equal(t, true, config.LogOutgoingRequests)
			assert.Equal(t, true, config.LogStructured)
		})
	})

	t.Run("parse with all env args", func(t *testing.T) {
		withArgs([]string{"command"}, func() {
			withEnv(map[string]string{
				"NA_AMAZON_DOMAIN":         "amazon.example.com",
				"NA_AMAZON_USER":           "amazonUserValue",
				"NA_AMAZON_PASSWORD":       "amazonPasswordValue",
				"NA_AMAZON_COOKIE_PATH":    "amazonCookiePathValue",
				"NA_ALEXA_SKILL_ID":        "alexaSkillIdValue",
				"NA_ALEXA_SKILL_NAME":      "alexaSkillNameValue",
				"NA_STREAM_DOMAIN":         "navidrome.example.com",
				"NA_API_KEY":               "apiKeyValue",
				"NA_LISTEN_ADDRESS":        "localhost:9090",
				"NA_LOG_INCOMING_REQUESTS": "true",
				"NA_LOG_OUTGOING_REQUESTS": "true",
				"NA_LOG_STRUCTURED":        "true",
			}, func() {
				config := parseConfiguration()
				assert.Equal(t, "amazon.example.com", config.AmazonDomain)
				assert.Equal(t, "amazonUserValue", config.AmazonUser)
				assert.Equal(t, "amazonPasswordValue", config.AmazonPassword)
				assert.Equal(t, "amazonCookiePathValue", config.AmazonCookiePath)
				assert.Equal(t, "alexaSkillIdValue", config.AlexaSkillId)
				assert.Equal(t, "alexaSkillNameValue", config.AlexaSkillName)
				assert.Equal(t, "navidrome.example.com", config.StreamDomain)
				assert.Equal(t, "apiKeyValue", config.ApiKey)
				assert.Equal(t, "localhost:9090", config.ListenAddress)
				assert.Equal(t, true, config.LogIncomingRequests)
				assert.Equal(t, true, config.LogOutgoingRequests)
				assert.Equal(t, true, config.LogStructured)
			})
		})
	})

	t.Run("parse invalid config", func(t *testing.T) {
		withArgs([]string{"command"}, func() {
			assert.Panics(t, func() {
				parseConfiguration()
			})
		})
	})

}

func TestGetStrPrecedence(t *testing.T) {

	t.Run("flag value should take precedence over default and env", func(t *testing.T) {
		var value string
		withEnv(map[string]string{"NA_VAR_TEST1": "valueFromEnv1"}, func() {
			withArgs([]string{"command", "-varTest1", "valueFromFlag1"}, func() {
				getStr(&value, "varTest1", "valueFromDefault1", "usage1")
				flag.Parse()
				assert.Equal(t, "valueFromFlag1", value)
			})
		})
	})

	t.Run("env value should take precedence over default", func(t *testing.T) {
		var value string
		withEnv(map[string]string{"NA_VAR_TEST2": "valueFromEnv2"}, func() {
			withArgs([]string{"command"}, func() {
				getStr(&value, "varTest2", "valueFromDefault2", "usage2")
				flag.Parse()
				assert.Equal(t, "valueFromEnv2", value)
			})
		})
	})

	t.Run("default value should be used", func(t *testing.T) {
		var value string
		withArgs([]string{"command"}, func() {
			getStr(&value, "varTest3", "valueFromDefault3", "usage3")
			flag.Parse()
			assert.Equal(t, "valueFromDefault3", value)
		})
	})

	t.Run("should be empty if no default", func(t *testing.T) {
		var value string
		withArgs([]string{"command"}, func() {
			getStr(&value, "varTest4", "", "usage4")
			flag.Parse()
			assert.Equal(t, "", value)
		})
	})
}

func TestGetBoolPrecedence(t *testing.T) {

	t.Run("flag value should take precedence over default and env", func(t *testing.T) {
		var value bool
		withEnv(map[string]string{"NA_VAR_TEST1": "true"}, func() {
			withArgs([]string{"command", "-varTest1", "false"}, func() {
				getBool(&value, "varTest1", true, "usage1")
				flag.Parse()
				assert.True(t, value)
			})
		})
	})

	t.Run("env value should take precedence over default", func(t *testing.T) {
		var value bool
		withEnv(map[string]string{"NA_VAR_TEST2": "false"}, func() {
			withArgs([]string{"command"}, func() {
				getBool(&value, "varTest2", true, "usage2")
				flag.Parse()
				assert.False(t, value)
			})
		})
	})

	t.Run("default value should be used", func(t *testing.T) {
		var value bool
		withArgs([]string{"command"}, func() {
			getBool(&value, "varTest3", true, "usage3")
			flag.Parse()
			assert.True(t, value)
		})
	})
}

func TestToUpperSnakeCase(t *testing.T) {
	assert.Equal(t, "NA_ALEXA_SKILL_ID", toEnvVarName("alexaSkillId"))
	assert.Equal(t, "NA_API_KEY", toEnvVarName("apiKey"))
}

func TestValidate(t *testing.T) {
	assert.NotPanics(t, func() { validate("testVar", "value") }, "Should not panic if value is set")
	assert.Panics(t, func() { validate("testVar", "") }, "Should panic if value is empty")
}

func withArgs(args []string, code func()) {
	mutexArg.Lock()
	savedArgs := os.Args
	defer func() {
		os.Args = savedArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		mutexArg.Unlock()
	}()
	os.Args = args
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	code()
}

func withEnv(envVars map[string]string, code func()) {
	mutexEnv.Lock()
	originalValues := make(map[string]string)
	for key, value := range envVars {
		originalValues[key] = os.Getenv(key)
		_ = os.Setenv(key, value)
	}
	defer func() {
		for key, originalValue := range originalValues {
			_ = os.Setenv(key, originalValue)
		}
		mutexEnv.Unlock()
	}()
	code()
}
