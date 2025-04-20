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
	"runtime"
	"strings"
	"sync"
)

// Payload represents the message data
type Payload struct {
	Data interface{} // Can hold any type of data
}

// Subscriber represents a client that subscribes to topics
type Subscriber struct {
	UUID     string                // Unique identifier for the subscriber
	Name     string                // Human-readable name for the subscriber
	Callback func(string, Payload) // Callback function to handle received messages (topic, payload)
}

// topicNode represents a node in the topic tree
type topicNode struct {
	name        string
	subscribers sync.Map              // Thread-safe map of subscribers by UUID
	children    map[string]*topicNode // Map of child nodes by name
	mutex       sync.RWMutex          // Mutex for thread safety
}

// Broker manages topics and subscribers
type Broker struct {
	root          *topicNode         // Root of the topic tree
	mutex         sync.RWMutex       // Global broker mutex
	capacity      int                // Maximum number of messages in each queue
	messageQueues []chan messageItem // Multiple queues for message distribution
	queueCount    int                // Number of message queues
	stopCh        chan struct{}      // Channel to signal workers to stop
	wg            sync.WaitGroup     // WaitGroup for graceful shutdown
}

// messageItem represents a message in the queue
type messageItem struct {
	subscriber *Subscriber
	topic      string
	payload    Payload
}

// NewBroker creates a new broker with the specified channel capacity
// and starts worker goroutines to process messages
func NewBroker(size int) *Broker {
	// Use number of CPU cores as default queue/worker count
	return NewBrokerWithWorkers(size, runtime.NumCPU())
}

// NewBrokerWithWorkers creates a new broker with the specified channel capacity
// and number of worker goroutines
func NewBrokerWithWorkers(size int, workerCount int) *Broker {
	if workerCount <= 0 {
		workerCount = runtime.NumCPU() // Use number of CPU cores
	}

	// Make queue capacity larger to prevent dropping messages
	queueCapacity := size * 2
	if queueCapacity < 10000 { // Ensure reasonable minimum size
		queueCapacity = 10000
	}

	// Create multiple message queues for better distribution
	messageQueues := make([]chan messageItem, workerCount)
	for i := 0; i < workerCount; i++ {
		messageQueues[i] = make(chan messageItem, queueCapacity)
	}

	broker := &Broker{
		root: &topicNode{
			name:        "",
			subscribers: sync.Map{},
			children:    make(map[string]*topicNode),
		},
		capacity:      queueCapacity,
		messageQueues: messageQueues,
		queueCount:    workerCount,
		stopCh:        make(chan struct{}),
	}

	// Start worker goroutines to process messages
	broker.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go broker.worker(i)
	}

	return broker
}

// worker processes messages from its designated queue
func (b *Broker) worker(queueID int) {
	defer b.wg.Done()

	for {
		select {
		case <-b.stopCh:
			// Stop signal received
			return
		case item := <-b.messageQueues[queueID]:
			// Process the message
			item.subscriber.Callback(item.topic, item.payload)
		}
	}
}

// getQueueForSubscriber selects a queue for a subscriber using consistent hashing
func (b *Broker) getQueueForSubscriber(subscriberUUID string) int {
	// Simple hash function to distribute subscribers across queues
	var hash uint32
	for i := 0; i < len(subscriberUUID); i++ {
		hash = hash*31 + uint32(subscriberUUID[i])
	}
	return int(hash % uint32(b.queueCount))
}

// Publish broadcasts a message to all subscribers of a topic and its parent topics
func (b *Broker) Publish(topic string, payload Payload) {
	if topic == "" {
		return
	}

	// Using read lock for better concurrency
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// Split the topic into segments
	segments := strings.Split(topic, ".")

	// Use a temporary list to collect all subscribers
	// This allows us to release locks before queueing messages
	var subscriberList []*Subscriber
	var subscriberMutex sync.Mutex

	var collectSubscribers func(node *topicNode, segments []string, index int)
	collectSubscribers = func(node *topicNode, segments []string, index int) {
		if node == nil {
			return
		}

		// Add all subscribers from current node
		node.subscribers.Range(func(_, value interface{}) bool {
			subscriberMutex.Lock()
			subscriberList = append(subscriberList, value.(*Subscriber))
			subscriberMutex.Unlock()
			return true
		})

		// If we've reached the end of the segments, we're done with this branch
		if index >= len(segments) {
			return
		}

		// Continue traversing the tree
		if segment := segments[index]; segment != "" {
			// Check for exact matches
			node.mutex.RLock()
			if child, exists := node.children[segment]; exists {
				childNode := child // Make a copy to avoid race conditions
				node.mutex.RUnlock()
				collectSubscribers(childNode, segments, index+1)
			} else {
				node.mutex.RUnlock()
			}

			// Check for wildcard matches
			node.mutex.RLock()
			if child, exists := node.children["#"]; exists {
				childNode := child // Make a copy to avoid race conditions
				node.mutex.RUnlock()
				collectSubscribers(childNode, segments, len(segments)) // Skip to the end
			} else {
				node.mutex.RUnlock()
			}
		}
	}

	// Collect all matching subscribers
	collectSubscribers(b.root, segments, 0)

	// Now queue messages for all collected subscribers
	for _, subscriber := range subscriberList {
		queueID := b.getQueueForSubscriber(subscriber.UUID)
		select {
		case b.messageQueues[queueID] <- messageItem{
			subscriber: subscriber,
			topic:      topic,
			payload:    payload,
		}:
			// Message successfully queued
		default:
			// Queue is full, log or handle overflow
			// In a production system, you might want to implement a backpressure mechanism
		}
	}
}

// Subscribe adds a subscriber to a topic
func (b *Broker) Subscribe(topic string, subscriber Subscriber) error {
	if topic == "" {
		return fmt.Errorf("topic cannot be empty")
	}
	if subscriber.UUID == "" {
		return fmt.Errorf("subscriber UUID cannot be empty")
	}
	if subscriber.Callback == nil {
		return fmt.Errorf("callback function cannot be nil")
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Split the topic into segments
	segments := strings.Split(topic, ".")

	// Start from the root and add topic nodes as needed
	currentNode := b.root

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		// Create the node if it doesn't exist
		currentNode.mutex.Lock()
		if _, exists := currentNode.children[segment]; !exists {
			currentNode.children[segment] = &topicNode{
				name:        segment,
				subscribers: sync.Map{},
				children:    make(map[string]*topicNode),
			}
		}
		childNode := currentNode.children[segment]
		currentNode.mutex.Unlock()

		currentNode = childNode
	}

	// Add the subscriber to the final node
	subscriberCopy := subscriber // Make a copy to avoid race conditions
	currentNode.subscribers.Store(subscriber.UUID, &subscriberCopy)

	return nil
}

// Unsubscribe removes a subscriber from all topics
func (b *Broker) Unsubscribe(uuid string) {
	if uuid == "" {
		return
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// Start from the root and recursively remove the subscriber
	b.unsubscribeFromNode(b.root, uuid)
}

// unsubscribeFromNode is a recursive helper function for removing a subscriber from a node and its children
func (b *Broker) unsubscribeFromNode(node *topicNode, uuid string) {
	if node == nil {
		return
	}

	// Remove from the current node
	node.subscribers.Delete(uuid)

	// Recursively remove from all child nodes
	node.mutex.RLock()
	childrenCopy := make([]*topicNode, 0, len(node.children))
	for _, child := range node.children {
		childrenCopy = append(childrenCopy, child)
	}
	node.mutex.RUnlock()

	for _, child := range childrenCopy {
		b.unsubscribeFromNode(child, uuid)
	}
}

// UnsubscribeFromTopic removes a subscriber from a specific topic
func (b *Broker) UnsubscribeFromTopic(topic string, uuid string) error {
	if topic == "" || uuid == "" {
		return fmt.Errorf("topic and UUID cannot be empty")
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// Split the topic into segments
	segments := strings.Split(topic, ".")

	// Start from the root and find the topic node
	currentNode := b.root

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		// If the segment doesn't exist, the subscriber isn't subscribed to this topic
		currentNode.mutex.RLock()
		child, exists := currentNode.children[segment]
		currentNode.mutex.RUnlock()

		if !exists {
			return fmt.Errorf("topic not found")
		}

		currentNode = child
	}

	// Remove the subscriber from the final node
	currentNode.subscribers.Delete(uuid)

	return nil
}

// GetSubscribers returns all subscribers for a topic
func (b *Broker) GetSubscribers(topic string) ([]*Subscriber, error) {
	if topic == "" {
		return nil, fmt.Errorf("topic cannot be empty")
	}

	b.mutex.RLock()
	defer b.mutex.RUnlock()

	// Split the topic into segments
	segments := strings.Split(topic, ".")

	// Start from the root and find the topic node
	currentNode := b.root

	for _, segment := range segments {
		if segment == "" {
			continue
		}

		// If the segment doesn't exist, there are no subscribers
		currentNode.mutex.RLock()
		child, exists := currentNode.children[segment]
		currentNode.mutex.RUnlock()

		if !exists {
			return nil, nil
		}

		currentNode = child
	}

	// Collect all subscribers from the final node
	subscribers := []*Subscriber{}
	currentNode.subscribers.Range(func(_, value interface{}) bool {
		subscribers = append(subscribers, value.(*Subscriber))
		return true
	})

	return subscribers, nil
}

// Close shuts down the broker and cleans up resources
func (b *Broker) Close() {
	// Signal all workers to stop
	close(b.stopCh)

	// Wait for all workers to finish
	b.wg.Wait()

	// Close all message queues
	for _, queue := range b.messageQueues {
		close(queue)
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Reset the root node
	b.root = &topicNode{
		name:        "",
		subscribers: sync.Map{},
		children:    make(map[string]*topicNode),
	}
}
