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

package rhilex

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// BenchmarkBroker 进行消息代理系统的性能基准测试
func BenchmarkBroker(b *testing.B) {
	tests := []struct {
		name         string
		topics       int  // 主题数量
		subscribers  int  // 每个主题的订阅者数量
		publishers   int  // 发布者数量
		messages     int  // 每个发布者发送的消息数量
		queueSize    int  // 消息队列大小
		workerCount  int  // 工作协程数量
		hierarchical bool // 是否使用层次化主题
	}{
		{
			name:         "小规模_平面结构",
			topics:       10,
			subscribers:  5,
			publishers:   5,
			messages:     1000,
			queueSize:    1000,
			workerCount:  4,
			hierarchical: false,
		},
		{
			name:         "中规模_平面结构",
			topics:       50,
			subscribers:  10,
			publishers:   20,
			messages:     1000,
			queueSize:    5000,
			workerCount:  8,
			hierarchical: false,
		},
		{
			name:         "大规模_平面结构",
			topics:       100,
			subscribers:  20,
			publishers:   50,
			messages:     1000,
			queueSize:    10000,
			workerCount:  16,
			hierarchical: false,
		},
		{
			name:         "小规模_层次结构",
			topics:       10,
			subscribers:  5,
			publishers:   5,
			messages:     1000,
			queueSize:    1000,
			workerCount:  4,
			hierarchical: true,
		},
		{
			name:         "中规模_层次结构",
			topics:       50,
			subscribers:  10,
			publishers:   20,
			messages:     1000,
			queueSize:    5000,
			workerCount:  8,
			hierarchical: true,
		},
		{
			name:         "大规模_层次结构",
			topics:       100,
			subscribers:  20,
			publishers:   50,
			messages:     1000,
			queueSize:    10000,
			workerCount:  16,
			hierarchical: true,
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// 跳过默认的 b.N 循环，我们将使用自定义的循环计数
			b.StopTimer()
			b.ResetTimer()

			// 创建并启动计时器
			startTime := time.Now()

			// 创建代理
			broker := NewBrokerWithWorkers(tt.queueSize, tt.workerCount)
			defer broker.Close()

			// 用于等待所有消息处理完成
			var wg sync.WaitGroup

			// 用于记录接收到的消息数量
			var receivedCount int64

			// 记录消息延迟
			var totalLatency int64
			var maxLatency int64
			var latencyCount int64

			// 准备主题名称
			topics := make([]string, tt.topics)
			for i := 0; i < tt.topics; i++ {
				if tt.hierarchical {
					// 创建层次化主题，如 "level1.level2.level3"
					depth := i%3 + 1 // 1-3层深度
					topic := "topic"
					for d := 0; d < depth; d++ {
						topic = fmt.Sprintf("%s.level%d", topic, d+1)
					}
					topics[i] = topic
				} else {
					// 简单平面主题
					topics[i] = fmt.Sprintf("topic%d", i)
				}
			}

			// 为每个主题添加订阅者
			for _, topic := range topics {
				for j := 0; j < tt.subscribers; j++ {
					subscriberID := fmt.Sprintf("sub_%s_%d", topic, j)

					// 创建订阅者并设置回调
					subscriber := Subscriber{
						UUID: subscriberID,
						Name: fmt.Sprintf("Subscriber %s-%d", topic, j),
						Callback: func(topic string, payload Payload) {
							// 记录接收时间
							if timestamp, ok := payload.Data.(int64); ok {
								latency := time.Now().UnixNano() - timestamp
								atomic.AddInt64(&totalLatency, latency)

								// 更新最大延迟
								for {
									current := atomic.LoadInt64(&maxLatency)
									if latency <= current {
										break
									}
									if atomic.CompareAndSwapInt64(&maxLatency, current, latency) {
										break
									}
								}

								atomic.AddInt64(&latencyCount, 1)
							}

							atomic.AddInt64(&receivedCount, 1)
						},
					}

					if err := broker.Subscribe(topic, subscriber); err != nil {
						b.Fatalf("Failed to subscribe: %v", err)
					}
				}
			}

			// 创建通道用于同步发布者启动
			startCh := make(chan struct{})

			// 启动发布者 goroutine
			wg.Add(tt.publishers)
			for i := 0; i < tt.publishers; i++ {
				go func(pubID int) {
					defer wg.Done()

					// 等待启动信号
					<-startCh

					// 每个发布者向随机主题发送消息
					for m := 0; m < tt.messages; m++ {
						topicIndex := (pubID + m) % len(topics)
						topic := topics[topicIndex]

						// 使用当前时间作为负载，用于计算延迟
						payload := Payload{
							Data: time.Now().UnixNano(),
						}

						broker.Publish(topic, payload)
					}
				}(i)
			}

			// 开始计时并发送启动信号
			b.StartTimer()
			close(startCh)

			// 等待所有发布者完成
			wg.Wait()

			// 给一些时间让消息处理完成
			time.Sleep(2 * time.Second)

			// 停止计时
			b.StopTimer()
			duration := time.Since(startTime)

			// 计算并报告指标
			totalMessages := tt.publishers * tt.messages
			expectedDeliveries := totalMessages * tt.subscribers
			actualDeliveries := atomic.LoadInt64(&receivedCount)
			deliveryRate := float64(actualDeliveries) / float64(expectedDeliveries) * 100

			messagesPerSecond := float64(actualDeliveries) / duration.Seconds()

			var avgLatency float64
			if latencyCount > 0 {
				avgLatency = float64(atomic.LoadInt64(&totalLatency)) / float64(latencyCount) / float64(time.Millisecond)
			}
			maxLatencyMs := float64(atomic.LoadInt64(&maxLatency)) / float64(time.Millisecond)

			// 输出结果
			b.Logf("测试配置: %s", tt.name)
			b.Logf("总持续时间: %v", duration)
			b.Logf("预期消息投递数: %d", expectedDeliveries)
			b.Logf("实际消息投递数: %d (%.2f%%)", actualDeliveries, deliveryRate)
			b.Logf("吞吐量: %.2f 消息/秒", messagesPerSecond)
			b.Logf("平均延迟: %.2f ms", avgLatency)
			b.Logf("最大延迟: %.2f ms", maxLatencyMs)

			// 验证消息投递率不低于98%（允许少量丢失）
			if deliveryRate < 98.0 {
				b.Errorf("消息投递率过低: %.2f%% < 98%%", deliveryRate)
			}
		})
	}
}

// TestBrokerConcurrency 测试代理在并发条件下的正确性
func TestBrokerConcurrency(t *testing.T) {
	// 测试参数
	topicCount := 10
	subscriberCount := 5
	publisherCount := 10
	messagesPerPublisher := 100
	queueSize := 1000
	workerCount := 4

	// 创建代理
	broker := NewBrokerWithWorkers(queueSize, workerCount)
	defer broker.Close()

	// 用于等待所有消息处理完成
	var wg sync.WaitGroup

	// 用于记录每个主题收到的消息数量
	receivedCounts := make(map[string]int64)
	var countMutex sync.Mutex

	// 创建主题和订阅者
	topics := make([]string, topicCount)
	for i := 0; i < topicCount; i++ {
		topics[i] = fmt.Sprintf("test.topic.%d", i)

		// 为每个主题注册多个订阅者
		for j := 0; j < subscriberCount; j++ {
			subID := fmt.Sprintf("sub_%d_%d", i, j)

			subscriber := Subscriber{
				UUID: subID,
				Name: fmt.Sprintf("Subscriber %d-%d", i, j),
				Callback: func(topic string, payload Payload) {
					countMutex.Lock()
					receivedCounts[topic]++
					countMutex.Unlock()
				},
			}

			if err := broker.Subscribe(topics[i], subscriber); err != nil {
				t.Fatalf("Failed to subscribe: %v", err)
			}
		}
	}

	// 启动发布者 goroutines
	wg.Add(publisherCount)
	for i := 0; i < publisherCount; i++ {
		go func(pubID int) {
			defer wg.Done()

			// 每个发布者向多个主题发送消息
			for m := 0; m < messagesPerPublisher; m++ {
				topicIndex := (pubID + m) % len(topics)
				topic := topics[topicIndex]

				payload := Payload{
					Data: fmt.Sprintf("Message %d from publisher %d", m, pubID),
				}

				broker.Publish(topic, payload)

				// 添加少量随机性，模拟真实负载
				if m%10 == 0 {
					time.Sleep(time.Millisecond)
				}
			}
		}(i)
	}

	// 等待所有发布者完成
	wg.Wait()

	// 给一些时间让消息处理完成
	time.Sleep(2 * time.Second)

	// 验证每个主题都收到了预期数量的消息
	countMutex.Lock()
	defer countMutex.Unlock()

	// 计算每个主题应该收到的消息总数
	expectedMessagesPerTopic := make(map[string]int)
	for i := 0; i < publisherCount; i++ {
		for m := 0; m < messagesPerPublisher; m++ {
			topicIndex := (i + m) % len(topics)
			topic := topics[topicIndex]
			expectedMessagesPerTopic[topic]++
		}
	}

	// 验证每个主题的消息数
	for topic, expected := range expectedMessagesPerTopic {
		// 计算预期总投递数（消息数 * 订阅者数）
		expectedTotal := expected * subscriberCount

		// 获取实际投递数
		actual := receivedCounts[topic]

		// 允许有少量丢失（不超过2%）
		minAcceptable := int64(float64(expectedTotal) * 0.98)

		if actual < minAcceptable {
			t.Errorf("主题 %s: 实际消息数 %d 低于最小可接受值 %d (预期: %d)",
				topic, actual, minAcceptable, expectedTotal)
		} else {
			t.Logf("主题 %s: 实际消息数 %d / 预期 %d (%.2f%%)",
				topic, actual, expectedTotal, float64(actual)/float64(expectedTotal)*100)
		}
	}
}

// TestBrokerWildcardSubscription 测试通配符订阅功能
func TestBrokerWildcardSubscription(t *testing.T) {
	// 创建代理
	broker := NewBroker(100)
	defer broker.Close()

	// 为了测试通配符，我们需要设置层次化主题
	// 并使用通配符订阅它们
	topics := []string{
		"sensors.temperature.room1",
		"sensors.temperature.room2",
		"sensors.humidity.room1",
		"sensors.pressure.outside",
	}

	// 用于记录收到的消息
	var receivedMessages sync.Map
	var wg sync.WaitGroup

	// 创建通配符订阅 "sensors.#"
	wildcard := Subscriber{
		UUID: "wildcard_sub",
		Name: "Wildcard Subscriber",
		Callback: func(topic string, payload Payload) {
			if msg, ok := payload.Data.(string); ok {
				receivedMessages.Store(topic, msg)
				wg.Done()
			}
		},
	}

	if err := broker.Subscribe("sensors.#", wildcard); err != nil {
		t.Fatalf("Failed to subscribe with wildcard: %v", err)
	}

	// 发布消息到每个主题
	wg.Add(len(topics))
	for i, topic := range topics {
		msg := fmt.Sprintf("Message %d for %s", i, topic)
		broker.Publish(topic, Payload{Data: msg})
	}

	// 等待所有消息处理完成
	wg.Wait()

	// 验证每个主题的消息都被通配符订阅接收
	for _, topic := range topics {
		if msg, ok := receivedMessages.Load(topic); !ok {
			t.Errorf("通配符订阅没有接收到主题 %s 的消息", topic)
		} else {
			t.Logf("主题 %s 成功接收: %s", topic, msg)
		}
	}
}

// TestBrokerUnsubscribe 测试取消订阅功能
func TestBrokerUnsubscribe(t *testing.T) {
	// 创建代理
	broker := NewBroker(100)
	defer broker.Close()

	// 测试主题
	topic := "test.unsubscribe"

	// 记录每个订阅者收到的消息数
	counts := make(map[string]int)
	var countMutex sync.Mutex

	// 创建多个订阅者
	subscribers := make([]Subscriber, 3)
	for i := range subscribers {
		uuid := fmt.Sprintf("sub_%d", i)
		subscribers[i] = Subscriber{
			UUID: uuid,
			Name: fmt.Sprintf("Subscriber %d", i),
			Callback: func(subID string) func(string, Payload) {
				return func(topic string, payload Payload) {
					countMutex.Lock()
					counts[subID]++
					countMutex.Unlock()
				}
			}(uuid),
		}

		if err := broker.Subscribe(topic, subscribers[i]); err != nil {
			t.Fatalf("Failed to subscribe: %v", err)
		}
	}

	// 发布第一条消息，所有订阅者都应该接收到
	broker.Publish(topic, Payload{Data: "Message 1"})

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 取消订阅第一个订阅者
	broker.Unsubscribe(subscribers[0].UUID)

	// 发布第二条消息，只有剩余的订阅者应该接收到
	broker.Publish(topic, Payload{Data: "Message 2"})

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 取消特定主题的第二个订阅者
	broker.UnsubscribeFromTopic(topic, subscribers[1].UUID)

	// 发布第三条消息，只有最后一个订阅者应该接收到
	broker.Publish(topic, Payload{Data: "Message 3"})

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证结果
	countMutex.Lock()
	defer countMutex.Unlock()

	// 第一个订阅者应该只收到1条消息
	if count := counts[subscribers[0].UUID]; count != 1 {
		t.Errorf("订阅者 0 应该收到 1 条消息，但实际收到 %d 条", count)
	}

	// 第二个订阅者应该收到2条消息
	if count := counts[subscribers[1].UUID]; count != 2 {
		t.Errorf("订阅者 1 应该收到 2 条消息，但实际收到 %d 条", count)
	}

	// 第三个订阅者应该收到3条消息
	if count := counts[subscribers[2].UUID]; count != 3 {
		t.Errorf("订阅者 2 应该收到 3 条消息，但实际收到 %d 条", count)
	}
}
