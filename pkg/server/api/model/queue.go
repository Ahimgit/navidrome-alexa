package model

type Song struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Album    string `json:"album"`
	Artist   string `json:"artist"`
	Duration int    `json:"duration"`
	Cover    string `json:"cover"`
	Stream   string `json:"stream"`
}

type queueState string

const (
	QueueStatePlaying queueState = "PLAYING"
	QueueStateIdle    queueState = "IDLE"
)

type Queue struct {
	State         queueState `json:"state"`
	QueuePosition int        `json:"queuePosition"`
	TrackPosition int        `json:"trackPosition"`
	Songs         []Song     `json:"queue"`
	Shuffle       bool       `json:"shuffle"`
	Repeat        bool       `json:"repeat"`
}

func NewQueue() *Queue {
	return &Queue{
		Shuffle:       false,
		Repeat:        false,
		QueuePosition: 0,
		TrackPosition: 0,
		Songs:         make([]Song, 0),
	}
}

func (q *Queue) HasItems() bool {
	return len(q.Songs) > 0
}

func (q *Queue) HasNext() bool {
	return q.QueuePosition < len(q.Songs)-1
}

func (q *Queue) HasPrev() bool {
	return q.QueuePosition > 0
}

func (q *Queue) Prev() *Song {
	if q.HasPrev() {
		q.QueuePosition--
		return q.Current()
	}
	return nil
}

func (q *Queue) Next() *Song {
	if q.HasNext() {
		q.QueuePosition++
		return q.Current()
	}
	return nil
}

func (q *Queue) PeekNext() *Song {
	if q.HasNext() {
		return &q.Songs[q.QueuePosition+1]
	}
	return nil
}

func (q *Queue) Current() *Song {
	return &q.Songs[q.QueuePosition]
}
