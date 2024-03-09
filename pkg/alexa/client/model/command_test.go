package model

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommandBuilders(t *testing.T) {

	t.Run("Test SpeakCmd builder", func(t *testing.T) {
		expectedSequence := `{
			"@type": "com.amazon.alexa.behaviors.model.Sequence",
			"startNode": {
				"@type": "com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode",
				"type": "Alexa.Speak",
				"operationPayload": {
					"deviceType": "dt",
					"deviceSerialNumber": "ds",
					"customerId": "cid",
					"locale": "en-US",
					"textToSpeak": "SSML, do you speak it?"
				}
			}
		}`

		cmd := BuildSpeakCmd("SSML, do you speak it?", "en-US", "dt", "ds", "cid")

		assert.NotNil(t, cmd)
		assert.Equal(t, "PREVIEW", cmd.BehaviorID)
		assert.Equal(t, "ENABLED", cmd.Status)
		assertSequenceEqual(t, expectedSequence, cmd.SequenceJSON)
	})

	t.Run("Test TextCommand builder", func(t *testing.T) {
		expectedSequence := `{
			"@type": "com.amazon.alexa.behaviors.model.Sequence",
			"startNode": {
				"@type": "com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode",
				"type": "Alexa.TextCommand",
				"skillId": "amzn1.ask.1p.tellalexa",
				"operationPayload": {
					"deviceType": "dt",
					"deviceSerialNumber": "ds",
					"customerId": "cid",
					"locale": "en-US",
					"text": "play next song"
				}
			}
		}`

		cmd := BuildTextCommandCmd("play next song", "en-US", "dt", "ds", "cid")

		assert.NotNil(t, cmd)
		assert.Equal(t, "PREVIEW", cmd.BehaviorID)
		assert.Equal(t, "ENABLED", cmd.Status)
		assertSequenceEqual(t, expectedSequence, cmd.SequenceJSON)
	})

	t.Run("Test VolumeCmd builder", func(t *testing.T) {
		expectedSequence := `{
			"@type": "com.amazon.alexa.behaviors.model.Sequence",
			"startNode": {
				"@type": "com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode",
				"type": "Alexa.DeviceControls.Volume",
				"operationPayload": {
					"deviceType": "dt",
					"deviceSerialNumber": "ds",
					"customerId": "cid",
					"locale": "en-US",
					"value": "41"
				}
			}
		}`

		cmd := BuildVolumeCmd(41, "en-US", "dt", "ds", "cid")

		assert.NotNil(t, cmd)
		assert.Equal(t, "PREVIEW", cmd.BehaviorID)
		assert.Equal(t, "ENABLED", cmd.Status)
		assertSequenceEqual(t, expectedSequence, cmd.SequenceJSON)
	})

}

func assertSequenceEqual(t *testing.T, expected string, actual string) {
	var expectedStruct Sequence
	var actualStruct Sequence
	err := json.Unmarshal([]byte(expected), &expectedStruct)
	assert.NoError(t, err)
	err = json.Unmarshal([]byte(actual), &actualStruct)
	assert.NoError(t, err)
	assert.Equal(t, expectedStruct, actualStruct)
}
