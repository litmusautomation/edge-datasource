package edge

import (
	"path"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/nats-io/nats.go"
)

type Message struct {
	Timestamp time.Time
	Value     []byte
}

type Topic struct {
	TopicPath string
	Interval  time.Duration
	Messages  []Message
	framer    *framer
}

// Key returns the key for the topic.
// The key is a combination of the interval string and the path.
// For example, if the path is "device.tag" and the interval is 1s, the key will be "1s/device.tag".
func (t *Topic) Key() string {
	return path.Join(t.Interval.String(), t.TopicPath)
}

// ToDataFrame converts the topic to a data frame.
func (t *Topic) ToDataFrame() (*data.Frame, error) {
	if t.framer == nil {
		t.framer = newFramer()
	}
	return t.framer.toFrame(t.Messages)
}

// * * * TopicMap * * *
// ? path is the stream identifier of the form "interval/topic"
// ? interval is the time between messages received from the frontend
// ? topicPath is the topic name (NATS subject)

// TopicMap is a thread-safe map of topics
type TopicMap struct {
	sync.Map
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
	tm.Map.Store(topic.Key(), topic)
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

// HasSubscription returns true if the topic map has a subscription for the given path
func (tm *TopicMap) HasSubscription(topicPath string) bool {
	hasSubscription := false
	tm.Range(func(_ string, topic *Topic) bool {
		if topic.TopicPath == topicPath {
			hasSubscription = true
			return false
		}
		return true
	})
	return hasSubscription
}

// AddMessage adds a message to the topic for the given path
func (tm *TopicMap) AddMessage(path string, msg Message) {
	topic, _ := tm.Load(path)
	if topic == nil {
		topic = &Topic{
			TopicPath: path,
		}
		tm.Store(topic)
	}
	topic.Messages = append(topic.Messages, msg)
}

func (tm *TopicMap) AddSubscription(topicPath string, sub *nats.Subscription) {
	tm.subscriptions[topicPath] = sub
}

func (tm *TopicMap) RemoveSubscription(topicPath string) {
	delete(tm.subscriptions, topicPath)
}

func (tm *TopicMap) GetSubscription(topicPath string) *nats.Subscription {
	return tm.subscriptions[topicPath]
}
