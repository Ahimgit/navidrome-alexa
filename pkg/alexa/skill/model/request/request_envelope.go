package request

import (
	"encoding/json"
)

type RequestEnvelope struct {
	Version string `json:"version"`
	Session struct {
		New       bool   `json:"new"`
		SessionID string `json:"sessionId"`
		User      struct {
			UserID      string `json:"userId"`
			AccessToken string `json:"accessToken"`
			Permissions struct {
				ConsentToken string                 `json:"consentToken"`
				Scopes       map[string]interface{} `json:"scopes"`
			} `json:"permissions"`
		} `json:"user"`
		Attributes  map[string]interface{} `json:"attributes"`
		Application struct {
			ApplicationID string `json:"applicationId"`
		} `json:"application"`
	} `json:"session"`
	Context struct {
		System struct {
			Application struct {
				ApplicationID string `json:"applicationId"`
			} `json:"application"`
			User struct {
				UserID      string `json:"userId"`
				AccessToken string `json:"accessToken"`
				Permissions struct {
					ConsentToken string                 `json:"consentToken"`
					Scopes       map[string]interface{} `json:"scopes"`
				} `json:"permissions"`
			} `json:"user"`
			Device struct {
				DeviceID            string `json:"deviceId"`
				SupportedInterfaces struct {
					AlexaPresentationAPL struct {
						MaxVersion string `json:"maxVersion"`
					} `json:"Alexa.Presentation.APL"`
					AudioPlayer map[string]interface{} `json:"AudioPlayer"`
					Display     struct {
						TemplateVersion string `json:"templateVersion"`
						MarkupVersion   string `json:"markupVersion"`
					} `json:"Display"`
					VideoApp    map[string]interface{} `json:"VideoApp"`
					Geolocation map[string]interface{} `json:"Geolocation"`
				} `json:"supportedInterfaces"`
			} `json:"device"`
			ApiEndpoint    string `json:"apiEndpoint"`
			ApiAccessToken string `json:"apiAccessToken"`
		} `json:"System"`
		AudioPlayer struct {
			OffsetInMilliseconds int    `json:"offsetInMilliseconds"`
			Token                string `json:"token"`
			PlayerActivity       string `json:"playerActivity"` //  PAUSED/FINISHED/BUFFER_UNDERRUN/IDLE/STOPPED
		} `json:"AudioPlayer"`
		Automative map[string]interface{} `json:"Automative"`
		Display    struct {
			TemplateVersion string `json:"templateVersion"`
			MarkupVersion   string `json:"markupVersion"`
		} `json:"Display"`
		Geolocation struct {
			Timestamp  string `json:"timestamp"`
			Coordinate struct {
				LatitudeInDegrees  float64 `json:"latitudeInDegrees"`
				LongitudeInDegrees float64 `json:"longitudeInDegrees"`
				AccuracyInMeters   float64 `json:"accuracyInMeters"`
			} `json:"coordinate"`
			Altitude struct {
				AltitudeInMeters float64 `json:"altitudeInMeters"`
				AccuracyInMeters float64 `json:"accuracyInMeters"`
			} `json:"altitude"`
			Heading struct {
				DirectionInDegrees float64 `json:"directionInDegrees"`
				AccuracyInDegrees  float64 `json:"accuracyInDegrees"`
			} `json:"heading"`
			Speed struct {
				SpeedInMetersPerSecond    float64 `json:"speedInMetersPerSecond"`
				AccuracyInMetersPerSecond float64 `json:"accuracyInMetersPerSecond"`
			} `json:"speed"`
			LocationServices struct {
				Status string `json:"status"`
				Access string `json:"access"`
			} `json:"locationServices"`
		} `json:"Geolocation"`
		Viewport struct {
			Experiences []struct {
				ArcMinuteWidth  float64 `json:"arcMinuteWidth"`
				ArcMinuteHeight float64 `json:"arcMinuteHeight"`
				CanRotate       bool    `json:"canRotate"`
				CanResize       bool    `json:"canResize"`
			} `json:"experiences"`
			Shape              string   `json:"shape"`
			PixelWidth         float64  `json:"pixelWidth"`
			PixelHeight        float64  `json:"pixelHeight"`
			Dpi                float64  `json:"dpi"`
			CurrentPixelWidth  float64  `json:"currentPixelWidth"`
			CurrentPixelHeight float64  `json:"currentPixelHeight"`
			Touch              []string `json:"touch"`
			Keyboard           []string `json:"keyboard"`
			Video              struct {
				Codecs []string `json:"codecs"`
			} `json:"video"`
		} `json:"Viewport"`
	} `json:"context"`
	RawRequest  json.RawMessage `json:"request"`
	BaseRequest BaseRequest     `json:"-"`
	Request     interface{}     `json:"-"`
}

type BaseRequest struct {
	Type      string `json:"type"`
	RequestID string `json:"requestId"`
	Timestamp string `json:"timestamp"`
	Locale    string `json:"locale"`
}

func (requestEnvelope *RequestEnvelope) UnmarshalJSON(bytes []byte) error {
	type rqType RequestEnvelope
	if err := json.Unmarshal(bytes, (*rqType)(requestEnvelope)); err != nil {
		return err
	}
	var baseRequest BaseRequest
	if err := json.Unmarshal(requestEnvelope.RawRequest, &baseRequest); err != nil {
		return err
	}
	requestEnvelope.BaseRequest = baseRequest
	var typedRequest interface{}
	switch requestEnvelope.BaseRequest.Type {
	case "LaunchRequest":
		typedRequest = &LaunchRequest{}
	case "IntentRequest":
		typedRequest = &IntentRequest{}
	case "AudioPlayer.PlaybackStarted":
		typedRequest = &AudioPlayerPlaybackStartedRequest{}
	case "AudioPlayer.PlaybackStopped":
		typedRequest = &AudioPlayerPlaybackStoppedRequest{}
	case "AudioPlayer.PlaybackNearlyFinished":
		typedRequest = &AudioPlayerPlaybackNearlyFinished{}
	case "AudioPlayer.PlaybackFinished":
		typedRequest = &AudioPlayerPlaybackFinishedRequest{}
	case "AudioPlayer.PlaybackFailed":
		typedRequest = &AudioPlayerPlaybackFailedRequest{}
	case "PlaybackController.PlayCommandIssued":
		typedRequest = &PlaybackControllerPlayCommandIssuedRequest{}
	case "PlaybackController.PauseCommandIssued":
		typedRequest = &PlaybackControllerPauseCommandIssuedRequest{}
	case "PlaybackController.NextCommandIssued":
		typedRequest = &PlaybackControllerNextCommandIssuedRequest{}
	case "PlaybackController.PreviousCommandIssued":
		typedRequest = &PlaybackControllerPreviousCommandIssuedRequest{}
	}
	if err := json.Unmarshal(requestEnvelope.RawRequest, typedRequest); err != nil {
		return err
	}
	requestEnvelope.Request = typedRequest
	return nil
}
