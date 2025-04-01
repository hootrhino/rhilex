package xmanager

import (
	"fmt"
	"sync"
)

// Message represents a message in the queue
type Message struct {
	Id      uint
	Payload any
}

func (m *Message) String() string {
	return fmt.Sprintf("Message ID: %d, Payload: %v", m.Id, m.Payload)
}

// Callback represents a subscriber's callback
type Callback struct {
	Name    string
	Handler func(Message)
}

// MessageQueue represents a topic-based message queue
type MessageQueue struct {
	mu       sync.RWMutex
	topics   map[string][]Callback // Map of topics to their subscribers
	queue    chan Message          // Internal message queue
	stopChan chan struct{}         // Channel to signal queue destruction
}

// NewMessageQueue creates a new MessageQueue with a specified buffer size
func NewMessageQueue(size int) *MessageQueue {
	mq := &MessageQueue{
		topics:   make(map[string][]Callback),
		queue:    make(chan Message, size),
		stopChan: make(chan struct{}),
	}

	// Start the message dispatcher
	go mq.dispatchMessages()

	return mq
}

// Publish publishes a message to a specific topic
func (mq *MessageQueue) Publish(topic string, message Message) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	// Check if the topic exists
	if _, exists := mq.topics[topic]; !exists {
		fmt.Printf("Topic %s does not exist, message dropped\n", topic)
		return
	}

	// Send the message to the internal queue
	select {
	case mq.queue <- message:
		// Message successfully queued
	default:
		// Drop the message if the queue is full
		fmt.Printf("Message queue is full, message dropped for topic %s\n", topic)
	}
}

// Subscribe subscribes a callback to a specific topic
func (mq *MessageQueue) Subscribe(topic string, callback Callback) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	// Add the callback to the topic
	mq.topics[topic] = append(mq.topics[topic], callback)
}

// UnSubscribe removes all subscribers from a specific topic
func (mq *MessageQueue) UnSubscribe(topic string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	// Remove the topic and its subscribers
	delete(mq.topics, topic)
}

// Destroy stops the message queue and cleans up resources
func (mq *MessageQueue) Destroy() {
	close(mq.stopChan) // Signal the dispatcher to stop
	close(mq.queue)    // Close the internal queue
}

// dispatchMessages dispatches messages to the appropriate topic subscribers
func (mq *MessageQueue) dispatchMessages() {
	for {
		select {
		case message, ok := <-mq.queue:
			if !ok {
				// Queue has been closed, stop the dispatcher
				return
			}

			// Dispatch the message to the appropriate topic subscribers
			mq.mu.RLock()
			for _, callbacks := range mq.topics {
				for _, callback := range callbacks {
					go callback.Handler(message) // Call the handler asynchronously
				}
			}
			mq.mu.RUnlock()

		case <-mq.stopChan:
			// Stop signal received, exit the dispatcher
			return
		}
	}
}
