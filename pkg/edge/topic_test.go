package edge

import (
	"sync"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestMessage(name string) Message {
	return Message{
		FieldName: name,
		Labels:    data.Labels{},
		Timestamp: time.Now(),
		Value:     []byte(`{"value": 1}`),
	}
}

func TestTopic_AddMessage_And_DrainMessages(t *testing.T) {
	topic := &Topic{TopicName: "test"}

	topic.AddMessage(newTestMessage("a"))
	topic.AddMessage(newTestMessage("b"))

	msgs := topic.DrainMessages()
	assert.Len(t, msgs, 2)

	// Second drain returns empty.
	msgs = topic.DrainMessages()
	assert.Len(t, msgs, 0)
}

func TestTopic_AddMessage_BoundsCheck(t *testing.T) {
	topic := &Topic{TopicName: "test"}

	for i := 0; i < maxMessages+100; i++ {
		topic.AddMessage(newTestMessage("x"))
	}

	msgs := topic.DrainMessages()
	assert.Len(t, msgs, maxMessages)
}

func TestTopic_AddMessage_TracksDrops(t *testing.T) {
	topic := &Topic{TopicName: "test"}

	for i := 0; i < maxMessages+50; i++ {
		topic.AddMessage(newTestMessage("x"))
	}

	// dropped counter should reflect the 50 excess messages
	assert.Equal(t, int64(50), topic.dropped.Load())

	// DrainMessages resets the counter
	topic.DrainMessages()
	assert.Equal(t, int64(0), topic.dropped.Load())
}

func TestTopic_ConcurrentAccess(t *testing.T) {
	topic := &Topic{TopicName: "test"}
	var wg sync.WaitGroup

	// NATS callback goroutine — writes messages
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			topic.AddMessage(newTestMessage("w"))
		}
	}()

	// Two RunStream goroutines — drain and convert to frames concurrently
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				msgs := topic.DrainMessages()
				_, _ = topic.ToDataFrame(msgs)
			}
		}()
	}

	wg.Wait()
	// If the race detector doesn't fire, the test passes.
}

func TestTopic_ToDataFrame(t *testing.T) {
	topic := &Topic{TopicName: "test"}

	msgs := []Message{
		{
			FieldName: "sensor",
			Labels:    data.Labels{},
			Timestamp: time.Now(),
			Value:     []byte(`{"temperature": 22.5}`),
		},
	}

	frame, err := topic.ToDataFrame(msgs)
	require.NoError(t, err)
	assert.Equal(t, "edge", frame.Name)
	// Time field + temperature field = 2 fields
	assert.Len(t, frame.Fields, 2)
	assert.Equal(t, 1, frame.Fields[0].Len())
}

func TestTopicMap_AddMessage_DirectLookup(t *testing.T) {
	tm := &TopicMap{
		subscriptions: make(map[string]*nats.Subscription),
	}
	topic := &Topic{TopicName: "sensor.temp"}
	tm.Store(topic)

	tm.AddMessage("sensor.temp", newTestMessage("val"))

	msgs := topic.DrainMessages()
	assert.Len(t, msgs, 1)
}

func TestTopicMap_AddMessage_UnknownTopic(t *testing.T) {
	tm := &TopicMap{
		subscriptions: make(map[string]*nats.Subscription),
	}
	topic := &Topic{TopicName: "sensor.temp"}
	tm.Store(topic)

	// Message for unknown topic is silently ignored, existing topic unaffected.
	tm.AddMessage("unknown.topic", newTestMessage("val"))

	msgs := topic.DrainMessages()
	assert.Len(t, msgs, 0)
}
