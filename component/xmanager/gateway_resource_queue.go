// Copyright (C) 2025 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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

// GenericMessageQueue represents a topic-based message queue
type GenericMessageQueue struct {
	mu       sync.RWMutex
	topics   map[string][]Callback // Map of topics to their subscribers
	queue    chan Message          // Internal message queue
	stopChan chan struct{}         // Channel to signal queue destruction
}

// NewGenericMessageQueue creates a new GenericMessageQueue with a specified buffer size
func NewGenericMessageQueue(size int) *GenericMessageQueue {
	mq := &GenericMessageQueue{
		topics:   make(map[string][]Callback),
		queue:    make(chan Message, size),
		stopChan: make(chan struct{}),
	}

	// Start the message dispatcher
	go mq.dispatchMessages()

	return mq
}

// Publish publishes a message to a specific topic
func (mq *GenericMessageQueue) Publish(topic string, message Message) {
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
func (mq *GenericMessageQueue) Subscribe(topic string, callback Callback) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	// Add the callback to the topic
	mq.topics[topic] = append(mq.topics[topic], callback)
}

// UnSubscribe removes all subscribers from a specific topic
func (mq *GenericMessageQueue) UnSubscribe(topic string) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	// Remove the topic and its subscribers
	delete(mq.topics, topic)
}

// Destroy stops the message queue and cleans up resources
func (mq *GenericMessageQueue) Destroy() {
	close(mq.stopChan) // Signal the dispatcher to stop
	close(mq.queue)    // Close the internal queue
}

// dispatchMessages dispatches messages to the appropriate topic subscribers
func (mq *GenericMessageQueue) dispatchMessages() {
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
