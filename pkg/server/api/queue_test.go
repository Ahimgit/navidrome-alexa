package api

import (
	"github.com/ahimgit/navidrome-alexa/pkg/server/api/model"
	"github.com/ahimgit/navidrome-alexa/pkg/util/tests"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueueAPIGetQueue(t *testing.T) {

	t.Run("GetQueue, with non-empty queue", func(t *testing.T) {
		rs := `{
			"state": "IDLE",
			"queuePosition": 0,
			"trackPosition": 123,
			"queue": [{
				"id": "Id1",
				"name": "Name1",
				"album": "Album1",
				"artist": "Artist1",
				"duration": 321,
				"cover": "/Cover1",
				"stream": "/Stream1"
			}],
			"shuffle": false,
			"repeat": false
		}`
		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONGet("/"))

		queueAPI := NewQueueAPI(queue())
		queueAPI.GetQueue(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 200, responseRecorder.Code)
	})

}

func TestQueueAPIGetNowPlaying(t *testing.T) {

	t.Run("GetNowPlaying, with non-empty queue", func(t *testing.T) {
		rs := `{
			"song": {
				"id": "Id1",
				"name": "Name1",
				"album": "Album1",
				"artist": "Artist1",
				"duration": 321,
				"cover": "/Cover1",
				"stream": "/Stream1"
			},
			"state": "IDLE"
		}`
		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONGet("/"))

		queueAPI := NewQueueAPI(queue())
		queueAPI.GetNowPlaying(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 200, responseRecorder.Code)
	})

	t.Run("GetNowPlaying, with empty queue", func(t *testing.T) {
		rs := `{ "state": "IDLE" }`
		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONGet("/"))

		queueAPI := NewQueueAPI(model.NewQueue())
		queueAPI.GetNowPlaying(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 200, responseRecorder.Code)
	})

}

func TestQueueAPIPostQueue(t *testing.T) {

	t.Run("PostQueue, with non-empty queue", func(t *testing.T) {
		rq := `{
			"state": "IDLE",
			"trackPosition": 123, 
			"queue": [{
				"id": "Id1",
				"name": "Name1",
				"album": "Album1",
				"artist": "Artist1",
				"duration": 321,
				"cover": "/Cover1",
				"stream": "/Stream1"
			}],
			"shuffle": false,
			"repeat": false
		}`
		rs := `{"message":"queue updated", "status":"success"}`

		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONPost(rq))
		queueUpdated := model.NewQueue()

		queueAPI := NewQueueAPI(queueUpdated)
		queueAPI.PostQueue(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 200, responseRecorder.Code)
		assert.Equal(t, queue(), queueUpdated)
	})

	t.Run("PostQueue, invalid request", func(t *testing.T) {
		rq := `{`
		rs := `{"message":"unexpected EOF", "status":"error"}`

		mockGinContext, responseRecorder := tests.MockGin(tests.MockJSONPost(rq))

		queueAPI := NewQueueAPI(model.NewQueue())
		queueAPI.PostQueue(mockGinContext)

		assert.JSONEq(t, rs, responseRecorder.Body.String())
		assert.Equal(t, 400, responseRecorder.Code)
	})

}

func queue() *model.Queue {
	queue := model.NewQueue()
	queue.Songs = append(queue.Songs, model.Song{
		Id:       "Id1",
		Name:     "Name1",
		Album:    "Album1",
		Artist:   "Artist1",
		Duration: 321,
		Cover:    "/Cover1",
		Stream:   "/Stream1",
	})
	queue.QueuePosition = 0
	queue.TrackPosition = 123
	queue.State = "IDLE"
	return queue
}
