package skill

import (
	"context"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/skill/model/request"
	"github.com/ahimgit/navidrome-alexa/pkg/alexa/skill/model/response"
	"github.com/ahimgit/navidrome-alexa/pkg/server/api/model"
	"github.com/ahimgit/navidrome-alexa/pkg/util/log"
)

type HandlerSelector struct {
	StreamDomain string
	Queue        *model.Queue
}

func NewHandlerSelector(queue *model.Queue, StreamDomain string) *HandlerSelector {
	return &HandlerSelector{Queue: queue, StreamDomain: StreamDomain}
}

func (handlerSelector *HandlerSelector) HandleRequest(rqe *request.RequestEnvelope, c context.Context) (rs *response.ResponseEnvelope) {
	switch rq := rqe.Request.(type) {
	case *request.IntentRequest:
		switch rq.Intent.Name {
		//todo?: AMAZON.LoopOffIntent AMAZON.LoopOnIntent AMAZON.ShuffleOffIntent AMAZON.ShuffleOnIntent AMAZON.RepeatIntent
		case "AMAZON.ResumeIntent":
			return handlerSelector.handlePlayResumeIntent(c)
		case "AMAZON.NextIntent":
			return handlerSelector.handleNextIntent(c)
		case "AMAZON.PreviousIntent":
			return handlerSelector.handlePrevIntent(c)
		case "AMAZON.StopIntent":
			return handlerSelector.handleStopIntent(rqe, c)
		case "AMAZON.CancelIntent":
			return handlerSelector.handleStopIntent(rqe, c)
		case "AMAZON.PauseIntent":
			return handlerSelector.handleStopIntent(rqe, c)
		default:
			return handlerSelector.handleDefaultResponse()
		}
	case *request.AudioPlayerPlaybackNearlyFinished:
		return handlerSelector.handlePlaybackNearlyFinishedEnqueue(rq, c)
	case *request.AudioPlayerPlaybackFinishedRequest:
		return handlerSelector.handlePlaybackFinishedAdvanceQueue(rq, c)
	case *request.AudioPlayerPlaybackStartedRequest:
		return handlerSelector.handlePlaybackStarted(c)
	case *request.AudioPlayerPlaybackStoppedRequest:
		return handlerSelector.handlePlaybackStopped(rq, c)
	case *request.AudioPlayerPlaybackFailedRequest:
		return handlerSelector.handlePlaybackFailed(rq, c)
	default:
		return handlerSelector.handleDefaultResponse()
	}
}

func (handlerSelector *HandlerSelector) handlePlaybackStarted(c context.Context) (rs *response.ResponseEnvelope) {
	if handlerSelector.Queue.HasItems() {
		handlerSelector.Queue.State = "PLAYING"
		log.GetContextLogger(c).Info("|> playback started",
			"id", handlerSelector.Queue.Current().Id,
			"name", handlerSelector.Queue.Current().Name)
	}
	return handlerSelector.handleDefaultResponse()
}

func (handlerSelector *HandlerSelector) handlePlaybackFinishedAdvanceQueue(rq *request.AudioPlayerPlaybackFinishedRequest, c context.Context) (rs *response.ResponseEnvelope) {
	if handlerSelector.Queue.HasNext() {
		if handlerSelector.Queue.Current().Id == rq.Token {
			handlerSelector.Queue.Next()
			log.GetContextLogger(c).Info("+ playback finished, advancing queue",
				"id", handlerSelector.Queue.Current().Id,
				"name", handlerSelector.Queue.Current().Name)
		} else {
			log.GetContextLogger(c).Info("? playback finished, not advancing queue due un-matching ids",
				"id_amz", rq.Token,
				"id", handlerSelector.Queue.Current().Id)
		}
	} else {
		handlerSelector.Queue.State = "IDLE"
		log.GetContextLogger(c).Info("|| playback finished, no more items in the queue")
	}
	return handlerSelector.handleDefaultResponse()
}

func (handlerSelector *HandlerSelector) handlePlaybackNearlyFinishedEnqueue(rq *request.AudioPlayerPlaybackNearlyFinished, c context.Context) (rs *response.ResponseEnvelope) {
	if handlerSelector.Queue.HasNext() {
		song := SongToAudioItem(
			handlerSelector.StreamDomain, 0,
			handlerSelector.Queue.PeekNext())
		song.Stream.ExpectedPreviousToken = handlerSelector.Queue.Current().Id // required for enq
		if handlerSelector.Queue.Current().Id == rq.AudioPlayerPlaybackBase.Token {
			log.GetContextLogger(c).Info("+ playback nearly finished, enqueueing next song to play",
				"id", handlerSelector.Queue.PeekNext().Id,
				"name", handlerSelector.Queue.PeekNext().Name)
		} else {
			log.GetContextLogger(c).Info("? playback nearly finished, enqueueing likely to be skipped due un-matching ids",
				"id_amz", rq.AudioPlayerPlaybackBase.Token,
				"id", handlerSelector.Queue.Current().Id)
		}
		return response.NewResponseBuilder().
			WithShouldEndSession(true).
			AddAudioPlayerPlayDirective(response.NewAudioPlayerPlayDirectiveBuilder().
				WithPlayBehaviorEnqueue().
				WithAudioItem(song).Build()).
			Build()
	} else {
		log.GetContextLogger(c).Info("|| playback nearly finished, no more items in the queue to enqueue")
		return handlerSelector.handleDefaultResponse()
	}
}

func (handlerSelector *HandlerSelector) handlePlaybackStopped(rq *request.AudioPlayerPlaybackStoppedRequest, c context.Context) (rs *response.ResponseEnvelope) {
	if handlerSelector.Queue.HasItems() && handlerSelector.Queue.Current().Id == rq.Token {
		handlerSelector.Queue.TrackPosition = rq.OffsetInMilliseconds // save position
		handlerSelector.Queue.State = "IDLE"
		log.GetContextLogger(c).Info("|| stopped",
			"id", handlerSelector.Queue.Current().Id,
			"name", handlerSelector.Queue.Current().Name,
			"time_offset", rq.OffsetInMilliseconds)
	}
	return handlerSelector.handleDefaultResponse()
}

func (handlerSelector *HandlerSelector) handlePlaybackFailed(rq *request.AudioPlayerPlaybackFailedRequest, c context.Context) (rs *response.ResponseEnvelope) {
	if handlerSelector.Queue.HasItems() {
		log.GetContextLogger(c).Info("X playback failed",
			"amz_id", rq.CurrentPlaybackState.Token,
			"id", handlerSelector.Queue.Current().Id,
			"name", handlerSelector.Queue.Current().Name)
	} else {
		log.GetContextLogger(c).Info("X playback failed, queue is empty", "id_amz", rq.CurrentPlaybackState.Token)
	}
	return handlerSelector.handleNextIntent(c) // try next one
}

func (handlerSelector *HandlerSelector) handlePlayResumeIntent(c context.Context) (rs *response.ResponseEnvelope) {
	if handlerSelector.Queue.HasItems() {
		log.GetContextLogger(c).Info("|> playing",
			"id", handlerSelector.Queue.Current().Id,
			"name", handlerSelector.Queue.Current().Name,
			"time", handlerSelector.Queue.TrackPosition)
		song := SongToAudioItem(
			handlerSelector.StreamDomain,
			handlerSelector.Queue.TrackPosition,
			handlerSelector.Queue.Current())
		return response.NewResponseBuilder().
			WithShouldEndSession(true).
			AddAudioPlayerPlayDirective(response.NewAudioPlayerPlayDirectiveBuilder().
				WithPlayBehaviorReplaceAll().
				WithAudioItem(song).Build()).
			WithCanFulfillIntentYES().
			Build()
	} else {
		return handlerSelector.handleDefaultResponse()
	}
}

func (handlerSelector *HandlerSelector) handleNextIntent(c context.Context) (rs *response.ResponseEnvelope) {
	if handlerSelector.Queue.HasNext() {
		song := SongToAudioItem(handlerSelector.StreamDomain, 0, handlerSelector.Queue.Next())
		log.GetContextLogger(c).Info(">> skipping to next", "id", song.Stream.Token, "name", song.Metadata.Title)
		return response.NewResponseBuilder().
			WithShouldEndSession(true).
			AddAudioPlayerPlayDirective(response.NewAudioPlayerPlayDirectiveBuilder().
				WithPlayBehaviorReplaceAll().
				WithAudioItem(song).Build()).
			WithCanFulfillIntentYES().
			Build()
	} else {
		return handlerSelector.handleDefaultResponse()
	}
}

func (handlerSelector *HandlerSelector) handlePrevIntent(c context.Context) (rs *response.ResponseEnvelope) {
	if handlerSelector.Queue.HasPrev() {
		log.GetContextLogger(c).Info("<< skipping back", "id", handlerSelector.Queue.Current().Id, "name", handlerSelector.Queue.Current().Name)
		song := SongToAudioItem(handlerSelector.StreamDomain, 0, handlerSelector.Queue.Prev())
		return response.NewResponseBuilder().
			WithShouldEndSession(true).
			AddAudioPlayerPlayDirective(response.NewAudioPlayerPlayDirectiveBuilder().
				WithPlayBehaviorReplaceAll().
				WithAudioItem(song).Build()).
			WithCanFulfillIntentYES().
			Build()
	} else {
		return handlerSelector.handleDefaultResponse()
	}
}

func (handlerSelector *HandlerSelector) handleStopIntent(rqe *request.RequestEnvelope, c context.Context) (rs *response.ResponseEnvelope) {
	if rqe.Context.AudioPlayer.PlayerActivity != "PAUSED" &&
		rqe.Context.AudioPlayer.PlayerActivity != "FINISHED" &&
		rqe.Context.AudioPlayer.PlayerActivity != "IDLE" &&
		rqe.Context.AudioPlayer.PlayerActivity != "STOPPED" {
		if handlerSelector.Queue.HasItems() {
			log.GetContextLogger(c).Info("|| stopping",
				"id", handlerSelector.Queue.Current().Id,
				"name", handlerSelector.Queue.Current().Name)
		} else {
			log.GetContextLogger(c).Info("|| stopping something we did not queue")
		}
		return response.NewResponseBuilder().
			WithShouldEndSession(true).
			AddAudioPlayerStopDirective().
			Build()
	}
	return handlerSelector.handleDefaultResponse()
}

func (handlerSelector *HandlerSelector) handleDefaultResponse() (rs *response.ResponseEnvelope) {
	return response.NewResponseBuilder().WithShouldEndSession(true).Build()
}

func SongToAudioItem(streamDomain string, offset int, song *model.Song) (ai *response.AudioItem) {
	return response.NewAudioItemBuilder().
		WithStream(response.NewStreamBuilder().
			WithToken(song.Id).
			WithURL(streamDomain + song.Stream).
			WithOffsetInMilliseconds(offset).
			Build()).
		WithMetadata(response.NewMetadataBuilder().
			WithTitle(song.Name).
			WithSubtitle(song.Album + " - " + song.Artist).
			Build()).
		Build()
}
