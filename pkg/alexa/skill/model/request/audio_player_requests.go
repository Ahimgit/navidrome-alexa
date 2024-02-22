package request

type AudioPlayerPlaybackBase struct {
	BaseRequest
	OffsetInMilliseconds int    `json:"offsetInMilliseconds"`
	Token                string `json:"token"`
}

type AudioPlayerPlaybackStartedRequest struct {
	AudioPlayerPlaybackBase
}

type AudioPlayerPlaybackStoppedRequest struct {
	AudioPlayerPlaybackBase
}

type AudioPlayerPlaybackNearlyFinished struct {
	AudioPlayerPlaybackBase
}

type AudioPlayerPlaybackFinishedRequest struct {
	AudioPlayerPlaybackBase
}

type AudioPlayerPlaybackFailedRequest struct {
	BaseRequest
	CurrentPlaybackState struct {
		OffsetInMilliseconds int    `json:"offsetInMilliseconds"`
		PlayerActivity       string `json:"playerActivity"`
		Token                string `json:"token"`
	} `json:"currentPlaybackState"`
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
	Token string `json:"token"`
}
