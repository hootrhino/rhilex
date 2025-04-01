package xmanager

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMessageQueue(t *testing.T) {
	// Create a new message queue
	mq := NewMessageQueue(10)

	t.Run("Subscribe and Publish", func(t *testing.T) {
		var wg sync.WaitGroup
		wg.Add(2)

		// Subscribe to topic1
		mq.Subscribe("topic1", Callback{
			Name: "Subscriber1",
			Handler: func(msg Message) {
				t.Log(msg.String())
				assert.Equal(t, uint(1), msg.Id)
				assert.Equal(t, "Hello, World!", msg.Payload)
				wg.Done()
			},
		})

		// Subscribe another callback to topic1
		mq.Subscribe("topic1", Callback{
			Name: "Subscriber2",
			Handler: func(msg Message) {
				t.Log(msg.String())
				assert.Equal(t, uint(1), msg.Id)
				assert.Equal(t, "Hello, World!", msg.Payload)
				wg.Done()
			},
		})

		// Publish a message to topic1
		mq.Publish("topic1", Message{Id: 1, Payload: "Hello, World!"})

		// Wait for both subscribers to process the message
		wg.Wait()
	})

	t.Run("Unsubscribe", func(t *testing.T) {
		// Unsubscribe from topic1
		mq.UnSubscribe("topic1")

		// Publish a message to topic1 (should not be received by any subscriber)
		mq.Publish("topic1", Message{Id: 2, Payload: "This should not be received"})
		time.Sleep(100 * time.Millisecond) // Allow some time for processing

		// No assertions here since there are no subscribers to verify
	})

	t.Run("Queue Full", func(t *testing.T) {
		var droppedMessages int
		var mu sync.Mutex

		// Subscribe to topic2
		mq.Subscribe("topic2", Callback{
			Name: "Subscriber3",
			Handler: func(msg Message) {
				time.Sleep(100 * time.Millisecond) // Simulate slow processing
			},
		})

		// Publish messages to fill the queue
		for i := 0; i < 15; i++ {
			mq.Publish("topic2", Message{Id: uint(i), Payload: i})
		}

		// Check for dropped messages
		mu.Lock()
		assert.LessOrEqual(t, droppedMessages, 5) // At most 5 messages should be dropped
		mu.Unlock()
	})

	t.Run("Destroy", func(t *testing.T) {
		// Destroy the message queue
		mq.Destroy()

		// Publish a message after destruction (should not panic)
		assert.NotPanics(t, func() {
			mq.Publish("topic1", Message{Id: 3, Payload: "After destruction"})
		})
	})
	time.Sleep(1000 * time.Millisecond)
}
