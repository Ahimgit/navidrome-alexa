package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQueueConstructor(t *testing.T) {
	queue := NewQueue()
	assert.Empty(t, queue.Songs)
	assert.Equal(t, 0, queue.QueuePosition)
	assert.Equal(t, 0, queue.TrackPosition)
	assert.False(t, queue.Repeat)
	assert.False(t, queue.Shuffle)
	assert.False(t, queue.HasItems())
	assert.False(t, queue.HasNext())
	assert.False(t, queue.HasPrev())
}

func TestQueueNavigation(t *testing.T) {
	queue := NewQueue()
	queue.Songs = append(queue.Songs, Song{Id: "1"})
	queue.Songs = append(queue.Songs, Song{Id: "2"})
	queue.Songs = append(queue.Songs, Song{Id: "3"})
	assert.True(t, queue.HasItems())

	assert.True(t, queue.HasNext())
	assert.False(t, queue.HasPrev())
	assert.Nil(t, queue.Prev())
	assert.Equal(t, &queue.Songs[0], queue.Current())
	assert.Equal(t, &queue.Songs[1], queue.PeekNext())
	assert.Equal(t, &queue.Songs[1], queue.Next()) // advance to 1

	assert.True(t, queue.HasNext())
	assert.True(t, queue.HasPrev())
	assert.Equal(t, &queue.Songs[1], queue.Current())
	assert.Equal(t, &queue.Songs[2], queue.PeekNext())
	assert.Equal(t, &queue.Songs[2], queue.Next()) // advance to 2

	assert.False(t, queue.HasNext())
	assert.True(t, queue.HasPrev())
	assert.Equal(t, &queue.Songs[2], queue.Current())
	assert.Nil(t, queue.PeekNext())
	assert.Nil(t, queue.Next()) // no more elements

	assert.Equal(t, &queue.Songs[1], queue.Prev()) // go back one element
	assert.Equal(t, &queue.Songs[1], queue.Current())
}
