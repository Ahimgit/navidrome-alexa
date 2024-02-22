package model

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"testing"
)

func TestModel(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Model Suite")
}

var _ = Describe("Queue", func() {
	var queue *Queue

	BeforeEach(func() {
		queue = NewQueue()
	})

	Describe("test constructor", func() {
		It("creates an empty queue", func() {
			Expect(queue.Songs).To(BeEmpty())
			Expect(queue.QueuePosition).To(Equal(0))
			Expect(queue.TrackPosition).To(Equal(0))
			Expect(queue.Repeat).To(Equal(false))
			Expect(queue.Shuffle).To(Equal(false))
			Expect(queue.HasItems()).To(Equal(false))
			Expect(queue.HasNext()).To(Equal(false))
			Expect(queue.HasPrev()).To(Equal(false))
		})
	})

	Describe("test simple queue", func() {
		It("navigating queue", func() {
			queue.Songs = append(queue.Songs, Song{Id: "1"})
			queue.Songs = append(queue.Songs, Song{Id: "2"})
			queue.Songs = append(queue.Songs, Song{Id: "3"})
			Expect(queue.HasItems()).To(Equal(true))

			// 0 element
			Expect(queue.HasNext()).To(Equal(true))
			Expect(queue.HasPrev()).To(Equal(false))
			Expect(queue.Prev()).To(BeNil())
			Expect(queue.Current()).To(Equal(&queue.Songs[0]))
			Expect(queue.PeekNext()).To(Equal(&queue.Songs[1]))
			Expect(queue.Next()).To(Equal(&queue.Songs[1])) //advance

			// 1 element
			Expect(queue.HasNext()).To(Equal(true))
			Expect(queue.HasPrev()).To(Equal(true))
			Expect(queue.Current()).To(Equal(&queue.Songs[1]))
			Expect(queue.PeekNext()).To(Equal(&queue.Songs[2]))
			Expect(queue.Next()).To(Equal(&queue.Songs[2])) //advance

			// 2 element
			Expect(queue.HasNext()).To(Equal(false))
			Expect(queue.HasPrev()).To(Equal(true))
			Expect(queue.Current()).To(Equal(&queue.Songs[2]))
			Expect(queue.PeekNext()).To(BeNil())
			Expect(queue.Next()).To(BeNil()) //end

			// back to 1 element
			Expect(queue.Prev()).To(Equal(&queue.Songs[1])) // go back
			Expect(queue.Current()).To(Equal(&queue.Songs[1]))
		})
	})
})
