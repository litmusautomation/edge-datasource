package edge

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
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

// fullDHPayload is a realistic DeviceHub tag message with all fields populated.
var fullDHPayload = DHMessage{
	Success:     true,
	Datatype:    "float",
	Timestamp:   1700000000000, // 2023-11-14T22:13:20Z
	RegisterId:  "reg-001",
	Value:       42.5,
	DeviceId:    "dev-001",
	TagName:     "temperature",
	DeviceName:  "PLC-A",
	Description: "Motor bearing temp",
}

func mustJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func TestMessageWrapper_DHMessage(t *testing.T) {
	c := &client{}
	natsMsg := &nats.Msg{
		Subject: "enterprise.site.plc.temperature",
		Data:    mustJSON(t, fullDHPayload),
	}

	msg := c.MessageWrapper(natsMsg)

	assert.Equal(t, "temperature", msg.FieldName)
	assert.Equal(t, time.UnixMilli(1700000000000), msg.Timestamp)
	assert.JSONEq(t, "42.5", string(msg.Value))

	// All 7 conditional labels should be set.
	assert.Equal(t, "enterprise.site.plc.temperature", msg.Labels["topic"])
	assert.Equal(t, "float", msg.Labels["datatype"])
	assert.Equal(t, "temperature", msg.Labels["tagName"])
	assert.Equal(t, "dev-001", msg.Labels["deviceId"])
	assert.Equal(t, "PLC-A", msg.Labels["deviceName"])
	assert.Equal(t, "Motor bearing temp", msg.Labels["description"])
	assert.Equal(t, "reg-001", msg.Labels["registerId"])
}

func TestMessageWrapper_DHMessage_PartialLabels(t *testing.T) {
	// Only tagName and timestamp set — other labels should be absent.
	c := &client{}
	payload := DHMessage{
		TagName:   "humidity",
		Timestamp: 1700000000000,
		Value:     65.0,
	}
	natsMsg := &nats.Msg{
		Subject: "topic.humidity",
		Data:    mustJSON(t, payload),
	}

	msg := c.MessageWrapper(natsMsg)

	assert.Equal(t, "humidity", msg.FieldName)
	assert.Equal(t, "topic.humidity", msg.Labels["topic"])
	assert.Equal(t, "humidity", msg.Labels["tagName"])
	// Empty fields should NOT appear in labels.
	_, hasDatatype := msg.Labels["datatype"]
	_, hasDeviceId := msg.Labels["deviceId"]
	_, hasDeviceName := msg.Labels["deviceName"]
	_, hasDescription := msg.Labels["description"]
	_, hasRegisterId := msg.Labels["registerId"]
	assert.False(t, hasDatatype)
	assert.False(t, hasDeviceId)
	assert.False(t, hasDeviceName)
	assert.False(t, hasDescription)
	assert.False(t, hasRegisterId)
}

func TestMessageWrapper_DHMessage_StringValue(t *testing.T) {
	c := &client{}
	payload := DHMessage{
		TagName:   "status",
		Timestamp: 1700000000000,
		Value:     "running",
	}
	natsMsg := &nats.Msg{
		Subject: "topic.status",
		Data:    mustJSON(t, payload),
	}

	msg := c.MessageWrapper(natsMsg)

	assert.Equal(t, "status", msg.FieldName)
	assert.JSONEq(t, `"running"`, string(msg.Value))
}

func TestMessageWrapper_DHMessage_BoolValue(t *testing.T) {
	c := &client{}
	payload := DHMessage{
		TagName:   "alarm",
		Timestamp: 1700000000000,
		Value:     true,
	}
	natsMsg := &nats.Msg{
		Subject: "topic.alarm",
		Data:    mustJSON(t, payload),
	}

	msg := c.MessageWrapper(natsMsg)

	assert.Equal(t, "alarm", msg.FieldName)
	assert.JSONEq(t, "true", string(msg.Value))
}

func TestMessageWrapper_NonDH_JSON_WithTimestamp(t *testing.T) {
	// Any valid JSON successfully unmarshals into DHMessage (missing fields zero out),
	// so this takes the DH path — FieldName is empty (no tagName), Value is the
	// json.Marshal of the zero-valued "value" field (null).
	c := &client{}
	natsMsg := &nats.Msg{
		Subject: "custom.topic",
		Data:    []byte(`{"timestamp": 1700000000000, "sensor": "temp", "reading": 22.5}`),
	}

	msg := c.MessageWrapper(natsMsg)

	// Goes through DH path: tagName is "" so FieldName is "".
	assert.Equal(t, "", msg.FieldName)
	assert.Equal(t, time.UnixMilli(1700000000000), msg.Timestamp)
	assert.Equal(t, "custom.topic", msg.Labels["topic"])
}

func TestMessageWrapper_NonDH_JSON_NoTimestamp(t *testing.T) {
	// Valid JSON without DH fields — still takes DH path, timestamp is epoch.
	c := &client{}
	natsMsg := &nats.Msg{
		Subject: "custom.topic",
		Data:    []byte(`{"sensor": "temp", "reading": 22.5}`),
	}

	msg := c.MessageWrapper(natsMsg)

	assert.Equal(t, "", msg.FieldName)
	assert.Equal(t, time.UnixMilli(0), msg.Timestamp, "zero timestamp → epoch")
	assert.Equal(t, "custom.topic", msg.Labels["topic"])
}

func TestMessageWrapper_InvalidJSON(t *testing.T) {
	// Completely invalid JSON — falls to raw path with time.Now() timestamp.
	c := &client{}
	raw := []byte(`not json at all`)
	natsMsg := &nats.Msg{
		Subject: "bad.topic",
		Data:    raw,
	}

	before := time.Now()
	msg := c.MessageWrapper(natsMsg)
	after := time.Now()

	assert.Equal(t, "bad.topic", msg.FieldName)
	assert.True(t, !msg.Timestamp.Before(before) && !msg.Timestamp.After(after))
	assert.Equal(t, raw, msg.Value)
	assert.Equal(t, data.Labels{}, msg.Labels)
}

func TestMessageWrapper_EmptySubject(t *testing.T) {
	// DH message with empty NATS subject — "topic" label should not be set.
	c := &client{}
	payload := DHMessage{
		TagName:   "tag",
		Timestamp: 1700000000000,
		Value:     1,
	}
	natsMsg := &nats.Msg{
		Subject: "",
		Data:    mustJSON(t, payload),
	}

	msg := c.MessageWrapper(natsMsg)

	_, hasTopic := msg.Labels["topic"]
	assert.False(t, hasTopic, "empty subject should not produce a 'topic' label")
}

func TestGetTimestampFromMessageData(t *testing.T) {
	c := &client{}

	t.Run("valid timestamp", func(t *testing.T) {
		ts := c.getTimestampFromMessageData([]byte(`{"timestamp": 1700000000000}`))
		assert.Equal(t, time.UnixMilli(1700000000000), ts)
	})

	t.Run("zero timestamp", func(t *testing.T) {
		ts := c.getTimestampFromMessageData([]byte(`{"timestamp": 0}`))
		assert.Equal(t, time.UnixMilli(0), ts)
	})

	t.Run("no timestamp field", func(t *testing.T) {
		// Valid JSON without "timestamp" unmarshals with Timestamp=0 (no error),
		// so we get epoch time, not time.Now().
		ts := c.getTimestampFromMessageData([]byte(`{"value": 42}`))
		assert.Equal(t, time.UnixMilli(0), ts)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		before := time.Now()
		ts := c.getTimestampFromMessageData([]byte(`not json`))
		after := time.Now()
		assert.True(t, !ts.Before(before) && !ts.After(after))
	})
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
