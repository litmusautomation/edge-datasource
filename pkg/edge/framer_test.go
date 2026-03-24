package edge

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/grafana/grafana-plugin-sdk-go/experimental"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var updateGoldenFiles = false

var fixedTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// runFramerTest is a helper that JSON-encodes Go values into messages, runs the
// framer, and checks the output against a golden file in testdata/.
func runFramerTest(t *testing.T, name string, messages []Message) {
	t.Helper()
	f := newFramer()
	frame, err := f.toFrame(messages)
	require.NoError(t, err)
	experimental.CheckGoldenJSONFrame(t, "testdata", name, frame, updateGoldenFiles)
}

// msg builds a Message from a Go value by JSON-marshaling it.
func msg(fieldName string, value interface{}) Message {
	b, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return Message{
		FieldName: fieldName,
		Timestamp: fixedTime,
		Value:     b,
		Labels:    data.Labels{},
	}
}

// rawMsg builds a Message from raw bytes (no JSON encoding).
func rawMsg(fieldName string, raw []byte) Message {
	return Message{
		FieldName: fieldName,
		Timestamp: fixedTime,
		Value:     raw,
		Labels:    data.Labels{},
	}
}

func TestFramer_toFrame_ClearsAllFields(t *testing.T) {
	f := newFramer()

	frame1, err := f.toFrame(makeMessages(4))
	require.NoError(t, err)
	for _, field := range frame1.Fields {
		assert.Equal(t, 4, field.Len(), "expected 4 rows after first call")
	}

	frame2, err := f.toFrame(makeMessages(2))
	require.NoError(t, err)
	for _, field := range frame2.Fields {
		assert.Equal(t, 2, field.Len(), "expected exactly 2 rows, no residual data")
	}
}

func TestFramer_toFrame_EmptyMessages(t *testing.T) {
	f := newFramer()

	_, err := f.toFrame(makeMessages(3))
	require.NoError(t, err)

	frame, err := f.toFrame([]Message{})
	require.NoError(t, err)
	for _, field := range frame.Fields {
		assert.Equal(t, 0, field.Len(), "expected 0 rows after empty messages")
	}
}

func makeMessages(n int) []Message {
	msgs := make([]Message, n)
	for i := range msgs {
		msgs[i] = Message{
			FieldName: "sensor",
			Timestamp: fixedTime,
			Value:     []byte(`{"temperature": 22.5}`),
		}
	}
	return msgs
}

func TestFramer(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		runFramerTest(t, "string", []Message{
			msg("tag", map[string]interface{}{"status": "running"}),
		})
	})

	t.Run("number", func(t *testing.T) {
		runFramerTest(t, "number", []Message{
			msg("tag", map[string]interface{}{"temperature": 22.5}),
			msg("tag", map[string]interface{}{"temperature": 23.1}),
		})
	})

	t.Run("bool", func(t *testing.T) {
		runFramerTest(t, "bool", []Message{
			msg("tag", map[string]interface{}{"enabled": true}),
			msg("tag", map[string]interface{}{"enabled": false}),
		})
	})

	t.Run("null", func(t *testing.T) {
		// First message creates the field, second sends null for it.
		runFramerTest(t, "null", []Message{
			msg("tag", map[string]interface{}{"temperature": 22.5}),
			msg("tag", map[string]interface{}{"temperature": nil}),
		})
	})

	t.Run("array", func(t *testing.T) {
		runFramerTest(t, "array", []Message{
			msg("tag", map[string]interface{}{"values": []int{1, 2, 3}}),
		})
	})

	t.Run("nested object", func(t *testing.T) {
		runFramerTest(t, "nested-object", []Message{
			msg("tag", map[string]interface{}{
				"location": map[string]interface{}{"lat": 40.7, "lon": -74.0},
			}),
		})
	})

	t.Run("multiple fields", func(t *testing.T) {
		runFramerTest(t, "multiple-fields", []Message{
			msg("tag", map[string]interface{}{"temperature": 22.5, "humidity": 65.0}),
			msg("tag", map[string]interface{}{"temperature": 23.1, "humidity": 63.0}),
		})
	})

	t.Run("sparse fields", func(t *testing.T) {
		// First message has field A, second has field B — each should have
		// a nil in the row where it wasn't present.
		runFramerTest(t, "sparse-fields", []Message{
			msg("tag", map[string]interface{}{"temperature": 22.5}),
			msg("tag", map[string]interface{}{"humidity": 65.0}),
		})
	})

	t.Run("mixed types across fields", func(t *testing.T) {
		runFramerTest(t, "mixed-types", []Message{
			msg("tag", map[string]interface{}{
				"name":    "sensor-1",
				"value":   42.0,
				"active":  true,
				"tags":    []string{"a", "b"},
				"meta":    map[string]interface{}{"version": 2},
			}),
		})
	})

	t.Run("labels", func(t *testing.T) {
		runFramerTest(t, "labels", []Message{
			{
				FieldName: "tag",
				Timestamp: fixedTime,
				Value:     []byte(`{"temperature": 22.5}`),
				Labels:    data.Labels{"device": "dev-1", "site": "factory-a"},
			},
		})
	})

	t.Run("invalid JSON", func(t *testing.T) {
		f := newFramer()
		// Invalid JSON should be skipped (logged + continue), not panic.
		// Use '~' prefix which triggers jsoniter's InvalidValue path.
		frame, err := f.toFrame([]Message{
			rawMsg("tag", []byte(`~broken`)),
			msg("tag", map[string]interface{}{"temperature": 22.5}),
		})
		require.NoError(t, err)
		// Only the valid message should produce a row.
		assert.Equal(t, 1, frame.Fields[0].Len())
	})

	t.Run("changing field type", func(t *testing.T) {
		// Field starts as float, then receives a string — second value
		// should be silently dropped (type mismatch logged).
		f := newFramer()
		frame, err := f.toFrame([]Message{
			msg("tag", map[string]interface{}{"value": 42.0}),
			msg("tag", map[string]interface{}{"value": "not a number"}),
		})
		require.NoError(t, err)
		// Time field has 2 rows (both messages parsed), but the value field
		// only has 1 row (second was type-mismatched and dropped), so it gets
		// extended with a nil via extendFields.
		assert.Equal(t, 2, frame.Fields[0].Len(), "Time field should have 2 rows")
	})
}
