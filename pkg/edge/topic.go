package edge

import (
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/nats-io/nats.go"
)

// maxMessages caps the per-topic message buffer to prevent unbounded growth.
const maxMessages = 10_000

type Message struct {
	FieldName string
	Labels    data.Labels
	Timestamp time.Time
	Value     []byte
}

type Topic struct {
	TopicName string `json:"topic"`
	mu        sync.Mutex
	messages  []Message
	framer    *framer
}

// AddMessage appends a message to the topic's buffer, dropping it if the cap is reached.
func (t *Topic) AddMessage(msg Message) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.messages) >= maxMessages {
		return
	}
	t.messages = append(t.messages, msg)
}

// DrainMessages returns the buffered messages and resets the buffer.
func (t *Topic) DrainMessages() []Message {
	t.mu.Lock()
	defer t.mu.Unlock()
	msgs := t.messages
	t.messages = nil
	return msgs
}

// ToDataFrame converts the given messages to a data frame.
func (t *Topic) ToDataFrame(messages []Message) (*data.Frame, error) {
	if t.framer == nil {
		t.framer = newFramer()
	}
	return t.framer.toFrame(messages)
}

// * * * TopicMap * * *
// TopicMap is a thread-safe map of topics
// * The key is the Topic name
// * The value is a pointer to the topic
// * subscriptions is a map of topic name and subscription pointer
type TopicMap struct {
	sync.Map
	mu            sync.RWMutex
	subscriptions map[string]*nats.Subscription
}

// Load returns the topic for the given topic key
func (tm *TopicMap) Load(key string) (*Topic, bool) {
	v, ok := tm.Map.Load(key)
	if !ok {
		return nil, false
	}
	return v.(*Topic), true
}

// Store stores the topic for the given topic key
func (tm *TopicMap) Store(topic *Topic) {
	tm.Map.Store(topic.TopicName, topic)
}

// Delete deletes the topic for the given topic key
func (tm *TopicMap) Delete(key string) {
	tm.Map.Delete(key)
}

// Range calls f sequentially for each key and topic present in the map.
// If f returns false, range stops the iteration.
func (tm *TopicMap) Range(f func(key string, topic *Topic) bool) {
	tm.Map.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*Topic))
	})
}

// AddMessage adds a message to the topic for the given path
func (tm *TopicMap) AddMessage(topicName string, msg Message) {
	tm.Range(func(_ string, topic *Topic) bool {
		if topic.TopicName == topicName {
			topic.AddMessage(msg)
			return false
		}
		return true
	})
}

func (tm *TopicMap) AddSubscription(topicName string, sub *nats.Subscription) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.subscriptions[topicName] = sub
}

func (tm *TopicMap) RemoveSubscription(topicName string) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	delete(tm.subscriptions, topicName)
}

func (tm *TopicMap) GetSubscription(topicName string) *nats.Subscription {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.subscriptions[topicName]
}
