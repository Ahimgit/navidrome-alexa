package skill

import (
	"context"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/skill/model/response"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"strconv"

	"github.com/ahimgit/navidrome-alexa/pkg/alexa/skill/model/request"
	"github.com/ahimgit/navidrome-alexa/pkg/server/api/model"
	"testing"
)

func TestHandlerSelectorPlayIntents(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		request   *request.RequestEnvelope
		audioItem *response.AudioItem
	}{
		{"ResumeIntent, non-empty queue, should start from the current queue position", intent("AMAZON.ResumeIntent"), expectedAudioItem(2, 123)},
		{"PreviousIntent, non-empty queue, should play prev song", intent("AMAZON.PreviousIntent"), expectedAudioItem(1, 0)},
		{"NextIntent, non-empty queue, should try to play next song", intent("AMAZON.NextIntent"), expectedAudioItem(3, 0)},
		{"PlaybackFailed, non-empty queue, should try to play next song", playbackFailed("failtoken"), expectedAudioItem(3, 0)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			handlerSelector := NewHandlerSelector(queue(1), "example.com")
			responseEnvelope := handlerSelector.HandleRequest(testCase.request, ctx())

			assert.NotNil(t, responseEnvelope)
			assert.True(t, responseEnvelope.Response.ShouldEndSession)
			assert.Equal(t, "YES", responseEnvelope.Response.CanFulfillIntent.CanFulfill)
			assert.Len(t, responseEnvelope.Response.Directives, 1)
			assert.IsType(t, (*response.AudioPlayerPlayDirective)(nil), responseEnvelope.Response.Directives[0])
			dir := responseEnvelope.Response.Directives[0].(*response.AudioPlayerPlayDirective)
			assert.Equal(t, "AudioPlayer.Play", dir.Type)
			assert.Equal(t, "REPLACE_ALL", dir.PlayBehavior)
			assert.Equal(t, testCase.audioItem, dir.AudioItem)
		})
	}

	for _, testCase := range []struct {
		name    string
		request *request.RequestEnvelope
		queue   *model.Queue
	}{
		{"ResumeIntent, empty queue, should return default empty response", intent("AMAZON.ResumeIntent"), model.NewQueue()},
		{"NextIntent, empty queue, should return default empty response", intent("AMAZON.NextIntent"), model.NewQueue()},
		{"NextIntent, no next item, should return default empty response", intent("AMAZON.NextIntent"), queue(2)},
		{"PreviousIntent, empty queue, should return default empty response", intent("AMAZON.PreviousIntent"), model.NewQueue()},
		{"PreviousIntent, no prev item, should return default empty response", intent("AMAZON.PreviousIntent"), queue(0)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			handlerSelector := NewHandlerSelector(testCase.queue, "example.com")

			responseEnvelope := handlerSelector.HandleRequest(testCase.request, ctx())

			assertDefaultEmptyResponse(t, responseEnvelope)
		})
	}

}

func TestHandlerSelectorStopIntents(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		request *request.RequestEnvelope
	}{
		{"StopIntent, playing", intent("AMAZON.StopIntent")},
		{"CancelIntent, playing", intent("AMAZON.CancelIntent")},
		{"PauseIntent, playing", intent("AMAZON.PauseIntent")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			handlerSelector := NewHandlerSelector(queue(1), "example.com")

			responseEnvelope := handlerSelector.HandleRequest(testCase.request, ctx())

			assert.NotNil(t, responseEnvelope)
			assert.True(t, responseEnvelope.Response.ShouldEndSession)
			assert.Len(t, responseEnvelope.Response.Directives, 1)
			assert.IsType(t, (*response.AudioPlayerStopDirective)(nil), responseEnvelope.Response.Directives[0])
			dir := responseEnvelope.Response.Directives[0].(*response.AudioPlayerStopDirective)
			assert.Equal(t, "AudioPlayer.Stop", dir.Type)
		})
	}

	for _, testCase := range []struct {
		name    string
		request *request.RequestEnvelope
	}{
		{"StopIntent, should issue stop even if our queue is empty", intent("AMAZON.StopIntent")},
		{"CancelIntent, should issue stop even if our queue is empty", intent("AMAZON.CancelIntent")},
		{"PauseIntent, should issue stop even if our queue is empty", intent("AMAZON.PauseIntent")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			handlerSelector := NewHandlerSelector(model.NewQueue(), "example.com")

			responseEnvelope := handlerSelector.HandleRequest(testCase.request, ctx())

			assert.NotNil(t, responseEnvelope)
			assert.True(t, responseEnvelope.Response.ShouldEndSession)
			assert.Len(t, responseEnvelope.Response.Directives, 1)
			assert.IsType(t, (*response.AudioPlayerStopDirective)(nil), responseEnvelope.Response.Directives[0])
			dir := responseEnvelope.Response.Directives[0].(*response.AudioPlayerStopDirective)
			assert.Equal(t, "AudioPlayer.Stop", dir.Type)
		})
	}

	for _, testCase := range []struct {
		name    string
		request *request.RequestEnvelope
	}{
		{"StopIntent, should do noting for already idle player", intent("AMAZON.StopIntent")},
		{"CancelIntent, should do noting for already idle player", intent("AMAZON.CancelIntent")},
		{"PauseIntent, should do noting for already idle player", intent("AMAZON.PauseIntent")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			handlerSelector := NewHandlerSelector(queue(1), "example.com")
			testCase.request.Context.AudioPlayer.PlayerActivity = "STOPPED"
			responseEnvelope := handlerSelector.HandleRequest(testCase.request, ctx())
			assertDefaultEmptyResponse(t, responseEnvelope)
		})
	}
}

func TestHandlerSelectorPlaybackFailedCallback(t *testing.T) {
	for _, testCase := range []struct {
		name    string
		request *request.RequestEnvelope
		queue   *model.Queue
	}{
		{"PlaybackFailed, no next item, should return default empty response", playbackFailed("failtoken"), queue(2)},
		{"PlaybackFailed, empty queue, should return default empty response", playbackFailed("failtoken"), model.NewQueue()},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			handlerSelector := NewHandlerSelector(testCase.queue, "example.com")
			responseEnvelope := handlerSelector.HandleRequest(testCase.request, ctx())

			assertDefaultEmptyResponse(t, responseEnvelope)
		})
	}

	t.Run("PlaybackFailed, non-empty queue, should try to play next song", func(t *testing.T) {
		handlerSelector := NewHandlerSelector(queue(1), "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackFailed("failtoken"), ctx())

		assert.NotNil(t, responseEnvelope)
		assert.True(t, responseEnvelope.Response.ShouldEndSession)
		assert.Equal(t, "YES", responseEnvelope.Response.CanFulfillIntent.CanFulfill)
		assert.Len(t, responseEnvelope.Response.Directives, 1)
		assert.IsType(t, (*response.AudioPlayerPlayDirective)(nil), responseEnvelope.Response.Directives[0])
		dir := responseEnvelope.Response.Directives[0].(*response.AudioPlayerPlayDirective)
		assert.Equal(t, "AudioPlayer.Play", dir.Type)
		assert.Equal(t, "REPLACE_ALL", dir.PlayBehavior)
		assert.Equal(t, expectedAudioItem(3, 0), dir.AudioItem)
	})

}

func TestHandlerSelectorPlaybackStartedCallback(t *testing.T) {

	t.Run("PlaybackStarted callback should set queue state to playing", func(t *testing.T) {
		queue := queue(1)
		handlerSelector := NewHandlerSelector(queue, "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackStarted("does not matter"), ctx())

		assert.Equal(t, model.QueueStatePlaying, queue.State)
		assert.Equal(t, 1, queue.QueuePosition)
		assertDefaultEmptyResponse(t, responseEnvelope)
	})

	t.Run("PlaybackStarted callback for empty queue should do nothing", func(t *testing.T) {
		queue := model.NewQueue()
		handlerSelector := NewHandlerSelector(queue, "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackStopped("does not matter", 134), ctx())

		assert.Equal(t, model.QueueStateIdle, queue.State)
		assertDefaultEmptyResponse(t, responseEnvelope)
	})
}

func TestHandlerSelectorPlaybackStoppedCallback(t *testing.T) {

	t.Run("StoppedRequest callback should remember queue and track position and set state to idle", func(t *testing.T) {
		queue := queue(1)
		handlerSelector := NewHandlerSelector(queue, "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackStopped("Id2", 134), ctx())

		assert.Equal(t, model.QueueStateIdle, queue.State)
		assert.Equal(t, 1, queue.QueuePosition)
		assert.Equal(t, 134, queue.TrackPosition)
		assertDefaultEmptyResponse(t, responseEnvelope)
	})

	t.Run("StoppedRequest callback with unknown id should still set idle", func(t *testing.T) {
		queue := queue(1)
		handlerSelector := NewHandlerSelector(queue, "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackStopped("UNKNOWN", 134), ctx())

		assert.Equal(t, model.QueueStateIdle, queue.State)
		assert.Equal(t, 1, queue.QueuePosition)
		assert.Equal(t, 0, queue.TrackPosition)
		assertDefaultEmptyResponse(t, responseEnvelope)
	})
}

func TestHandlerSelectorPlaybackNearlyFinishedCallback(t *testing.T) {

	t.Run("PlaybackNearlyFinished should enqueue next song without advancing queue (that happens in finished)", func(t *testing.T) {
		queue := queue(1)
		handlerSelector := NewHandlerSelector(queue, "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackNearlyFinished("Id2"), ctx())

		assert.Equal(t, 1, queue.QueuePosition) // q stays where it is until playback is really finished

		assert.NotNil(t, responseEnvelope)
		assert.True(t, responseEnvelope.Response.ShouldEndSession)

		assert.Len(t, responseEnvelope.Response.Directives, 1)
		dir := responseEnvelope.Response.Directives[0].(*response.AudioPlayerPlayDirective)
		assert.Equal(t, "AudioPlayer.Play", dir.Type)
		assert.Equal(t, "ENQUEUE", dir.PlayBehavior)

		expectedAudioItem := expectedAudioItem(3, 0)
		expectedAudioItem.Stream.ExpectedPreviousToken = "Id2"
		assert.Equal(t, expectedAudioItem, dir.AudioItem)
	})

	t.Run("PlaybackNearlyFinished should enqueue even with un-matching tokens", func(t *testing.T) {
		queue := queue(1)
		handlerSelector := NewHandlerSelector(queue, "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackNearlyFinished("some unexpected token"), ctx())

		assert.Equal(t, 1, queue.QueuePosition) // q stays where it is until playback is really finished

		assert.NotNil(t, responseEnvelope)
		assert.True(t, responseEnvelope.Response.ShouldEndSession)

		assert.Len(t, responseEnvelope.Response.Directives, 1)
		dir := responseEnvelope.Response.Directives[0].(*response.AudioPlayerPlayDirective)
		assert.Equal(t, "AudioPlayer.Play", dir.Type)
		assert.Equal(t, "ENQUEUE", dir.PlayBehavior)

		expectedAudioItem := expectedAudioItem(3, 0)
		expectedAudioItem.Stream.ExpectedPreviousToken = "Id2"
		assert.Equal(t, expectedAudioItem, dir.AudioItem)
	})

	t.Run("PlaybackNearlyFinished should not do anything if nothing left in the queue", func(t *testing.T) {
		queue := queue(2)
		handlerSelector := NewHandlerSelector(queue, "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackNearlyFinished("Id3"), ctx())

		assert.Equal(t, 2, queue.QueuePosition)
		assertDefaultEmptyResponse(t, responseEnvelope)
	})

}

func TestHandlerSelectorPlaybackFinishedCallback(t *testing.T) {

	t.Run("PlaybackFinished should advance queue forward if token matches current song", func(t *testing.T) {
		queue := queue(1)
		handlerSelector := NewHandlerSelector(queue, "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackFinished("Id2"), ctx())

		assert.Equal(t, 2, queue.QueuePosition) // 1 -> 2
		assertDefaultEmptyResponse(t, responseEnvelope)
	})

	t.Run("PlaybackFinished should do nothing if token does not match queue", func(t *testing.T) {
		queue := queue(1)
		handlerSelector := NewHandlerSelector(queue, "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackFinished("wrong token"), ctx())

		assert.Equal(t, 1, queue.QueuePosition)
		assertDefaultEmptyResponse(t, responseEnvelope)
	})

	t.Run("PlaybackFinished should set queue state to IDLE if noting in the queue", func(t *testing.T) {
		queue := model.NewQueue()
		handlerSelector := NewHandlerSelector(queue, "example.com")
		responseEnvelope := handlerSelector.HandleRequest(playbackFinished("does not matter"), ctx())

		assert.Equal(t, model.QueueStateIdle, queue.State)
		assertDefaultEmptyResponse(t, responseEnvelope)
	})
}

func TestUnknownRequest(t *testing.T) {

	t.Run("Unknown intents should respond with empty response", func(t *testing.T) {
		handlerSelector := NewHandlerSelector(queue(0), "example.com")
		responseEnvelope := handlerSelector.HandleRequest(intent("?"), ctx())
		assertDefaultEmptyResponse(t, responseEnvelope)
	})

	t.Run("Unknown requests should also respond with empty response", func(t *testing.T) {
		handlerSelector := NewHandlerSelector(queue(0), "example.com")
		responseEnvelope := handlerSelector.HandleRequest(&request.RequestEnvelope{}, ctx())
		assertDefaultEmptyResponse(t, responseEnvelope)
	})
}

func assertDefaultEmptyResponse(t *testing.T, responseEnvelope *response.ResponseEnvelope) {
	assert.NotNil(t, responseEnvelope)
	assert.True(t, responseEnvelope.Response.ShouldEndSession)
	assert.Len(t, responseEnvelope.Response.Directives, 0)
}

func expectedAudioItem(num int, expectedOffset int) *response.AudioItem {
	suffix := strconv.Itoa(num)
	return &response.AudioItem{
		Stream: &response.Stream{
			ExpectedPreviousToken: "",
			Token:                 "Id" + suffix,
			URL:                   "example.com/Stream" + suffix,
			OffsetInMilliseconds:  expectedOffset,
		},
		Metadata: &response.Metadata{
			Title:    "Name" + suffix,
			Subtitle: "Album" + suffix + " - Artist" + suffix,
		},
	}
}

func ctx() context.Context {
	ctx := context.Background()
	ctx = log.SetContextLogger(ctx, slog.Default())
	return ctx
}

func queue(position int) *model.Queue {
	queue := model.NewQueue()
	queue.Songs = append(queue.Songs, song(1))
	queue.Songs = append(queue.Songs, song(2))
	queue.Songs = append(queue.Songs, song(3))
	queue.QueuePosition = position
	queue.TrackPosition = 123
	return queue
}

func song(num int) model.Song {
	suffix := strconv.Itoa(num)
	return model.Song{
		Id:       "Id" + suffix,
		Name:     "Name" + suffix,
		Album:    "Album" + suffix,
		Artist:   "Artist" + suffix,
		Duration: num * 10,
		Cover:    "/Cover" + suffix,
		Stream:   "/Stream" + suffix,
	}
}

func intent(name string) *request.RequestEnvelope {
	return &request.RequestEnvelope{
		Request: &request.IntentRequest{
			Intent: request.Intent{
				Name: name,
			},
		},
	}
}

func playbackStarted(token string) *request.RequestEnvelope {
	return &request.RequestEnvelope{
		Request: &request.AudioPlayerPlaybackStartedRequest{
			AudioPlayerPlaybackBase: request.AudioPlayerPlaybackBase{
				Token: token,
			},
		}}
}

func playbackFailed(token string) *request.RequestEnvelope {
	return &request.RequestEnvelope{
		Request: &request.AudioPlayerPlaybackFailedRequest{
			CurrentPlaybackState: struct {
				OffsetInMilliseconds int    `json:"offsetInMilliseconds"`
				PlayerActivity       string `json:"playerActivity"`
				Token                string `json:"token"`
			}{
				OffsetInMilliseconds: 5000,
				PlayerActivity:       "PLAYING",
				Token:                token,
			},
			Error: struct {
				Message string `json:"message"`
				Type    string `json:"type"`
			}{
				Message: "An error occurred",
				Type:    "ERROR_TYPE",
			},
			Token: token,
		},
	}
}

func playbackStopped(token string, offset int) *request.RequestEnvelope {
	return &request.RequestEnvelope{
		Request: &request.AudioPlayerPlaybackStoppedRequest{
			AudioPlayerPlaybackBase: request.AudioPlayerPlaybackBase{
				Token:                token,
				OffsetInMilliseconds: offset,
			},
		},
	}
}

func playbackFinished(token string) *request.RequestEnvelope {
	return &request.RequestEnvelope{
		Request: &request.AudioPlayerPlaybackFinishedRequest{
			AudioPlayerPlaybackBase: request.AudioPlayerPlaybackBase{
				Token: token,
			},
		},
	}
}

func playbackNearlyFinished(token string) *request.RequestEnvelope {
	return &request.RequestEnvelope{
		Request: &request.AudioPlayerPlaybackNearlyFinished{
			AudioPlayerPlaybackBase: request.AudioPlayerPlaybackBase{
				Token: token,
			},
		},
	}
}
