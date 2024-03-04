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

func TestHandlerSelectorPlayDirectives(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		request   *request.RequestEnvelope
		audioItem *response.AudioItem
	}{
		{"ResumeIntent, non-empty queue", intent("AMAZON.ResumeIntent"), expectedAudioItem(2, 123)},
		{"NextIntent, non-empty queue", intent("AMAZON.NextIntent"), expectedAudioItem(3, 0)},
		{"PreviousIntent, non-empty queue", intent("AMAZON.PreviousIntent"), expectedAudioItem(1, 0)},
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
		{"ResumeIntent, empty queue", intent("AMAZON.ResumeIntent"), model.NewQueue()},
		{"NextIntent, empty queue", intent("AMAZON.NextIntent"), model.NewQueue()},
		{"NextIntent, no next item", intent("AMAZON.NextIntent"), queue(2)},
		{"PreviousIntent, empty queue", intent("AMAZON.PreviousIntent"), model.NewQueue()},
		{"PreviousIntent, no prev item", intent("AMAZON.PreviousIntent"), queue(0)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			handlerSelector := NewHandlerSelector(testCase.queue, "example.com")

			responseEnvelope := handlerSelector.HandleRequest(testCase.request, ctx())

			assert.NotNil(t, responseEnvelope)
			assert.True(t, responseEnvelope.Response.ShouldEndSession)
			assert.Len(t, responseEnvelope.Response.Directives, 0)
		})
	}

}

func TestHandlerSelectorStopDirectives(t *testing.T) {
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
