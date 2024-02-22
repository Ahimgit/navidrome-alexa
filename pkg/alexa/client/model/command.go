package model

import (
	"encoding/json"
)

type OperationPayload struct {
	DeviceType         string `json:"deviceType"`
	DeviceSerialNumber string `json:"deviceSerialNumber"`
	CustomerID         string `json:"customerId"`
	Locale             string `json:"locale,omitempty"`
	Text               string `json:"text,omitempty"`
	TextToSpeak        string `json:"textToSpeak,omitempty"`
	SoundStringID      string `json:"soundStringId,omitempty"`
}

type StartNode struct {
	Type             string           `json:"@type"`
	NodeType         string           `json:"type"`
	SkillID          string           `json:"skillId,omitempty"`
	OperationPayload OperationPayload `json:"operationPayload"`
}

type Sequence struct {
	Type      string    `json:"@type"`
	StartNode StartNode `json:"startNode"`
}

type AlexaCmd struct {
	BehaviorID   string `json:"behaviorId"`
	SequenceJSON string `json:"sequenceJson"`
	Status       string `json:"status"`
}

func BuildTextCommandCmd(
	text string,
	locale string,
	deviceType string,
	deviceSerialNumber string,
	mediaOwnerCustomerID string) AlexaCmd {
	return buildAlexaCmd("Alexa.TextCommand", "amzn1.ask.1p.tellalexa", OperationPayload{
		DeviceType:         deviceType,
		DeviceSerialNumber: deviceSerialNumber,
		CustomerID:         mediaOwnerCustomerID,
		Locale:             locale,
		Text:               text,
	})
}

func BuildSpeakCmd(
	text string,
	locale string,
	deviceType string,
	deviceSerialNumber string,
	mediaOwnerCustomerID string) AlexaCmd {
	return buildAlexaCmd("Alexa.Speak", "", OperationPayload{
		DeviceType:         deviceType,
		DeviceSerialNumber: deviceSerialNumber,
		CustomerID:         mediaOwnerCustomerID,
		Locale:             locale,
		TextToSpeak:        text,
	})
}

func buildAlexaCmd(commandType string, skillID string, commandPayload OperationPayload) AlexaCmd {
	seq := Sequence{
		Type: "com.amazon.alexa.behaviors.model.Sequence",
		StartNode: StartNode{
			Type:             "com.amazon.alexa.behaviors.model.OpaquePayloadOperationNode",
			NodeType:         commandType,
			SkillID:          skillID,
			OperationPayload: commandPayload,
		},
	}
	seqJSON, _ := json.Marshal(seq)
	return AlexaCmd{
		BehaviorID:   "PREVIEW",
		Status:       "ENABLED",
		SequenceJSON: string(seqJSON),
	}
}
