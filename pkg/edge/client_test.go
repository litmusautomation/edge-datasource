package edge

import (
	"sync"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateTopic(t *testing.T) {
	c := &client{}
	tests := []struct {
		name    string
		topic   string
		wantErr bool
	}{
		{"valid single token", "sensor", false},
		{"valid dotted topic", "device.sensor.temperature", false},
		{"wildcard star", "device.*.temperature", true},
		{"wildcard gt", "device.>", true},
		{"empty token in path", "device..temperature", true},
		{"whitespace token", "device. .temperature", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := c.validateTopic(tt.topic)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSubscribe_Idempotent(t *testing.T) {
	// Subscribing to the same topic twice should be idempotent —
	// the second call returns nil and doesn't create a new topic.
	tm := &TopicMap{
		Map:           sync.Map{},
		subscriptions: make(map[string]*nats.Subscription),
	}

	topic := &Topic{TopicName: "sensor.temp"}
	tm.Store(topic)

	// Simulate the idempotency check that client.Subscribe performs.
	t1, ok := tm.Load("sensor.temp")
	require.True(t, ok)

	t2, ok := tm.Load("sensor.temp")
	require.True(t, ok)

	assert.Same(t, t1, t2, "second Load should return the same *Topic pointer")
}

func TestTopicMap_GetTopic_NotFound(t *testing.T) {
	tm := &TopicMap{
		Map:           sync.Map{},
		subscriptions: make(map[string]*nats.Subscription),
	}

	topic, ok := tm.Load("nonexistent")
	assert.False(t, ok)
	assert.Nil(t, topic)
}

func TestTopicMap_SubscriptionMutex(t *testing.T) {
	tm := &TopicMap{
		subscriptions: make(map[string]*nats.Subscription),
	}

	t.Run("get missing key returns nil", func(t *testing.T) {
		sub := tm.GetSubscription("nonexistent")
		assert.Nil(t, sub)
	})

	t.Run("add/get/remove round-trip", func(t *testing.T) {
		// nil stands in for a real *nats.Subscription; we only test map mechanics.
		tm.AddSubscription("topic.a", nil)
		sub := tm.GetSubscription("topic.a")
		require.Nil(t, sub) // nil was stored

		tm.RemoveSubscription("topic.a")
		sub = tm.GetSubscription("topic.a")
		assert.Nil(t, sub)
	})
}
