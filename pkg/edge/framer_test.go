package edge

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeMessages(n int) []Message {
	msgs := make([]Message, n)
	for i := range msgs {
		msgs[i] = Message{
			FieldName: "sensor",
			Timestamp: time.Now(),
			Value:     []byte(`{"temperature": 22.5}`),
		}
	}
	return msgs
}

func TestFramer_toFrame_ClearsAllFields(t *testing.T) {
	f := newFramer()

	// First call: populate with 4 rows.
	frame1, err := f.toFrame(makeMessages(4))
	require.NoError(t, err)
	for _, field := range frame1.Fields {
		assert.Equal(t, 4, field.Len(), "expected 4 rows after first call")
	}

	// Second call: populate with 2 rows — old data must be fully cleared.
	frame2, err := f.toFrame(makeMessages(2))
	require.NoError(t, err)
	for _, field := range frame2.Fields {
		assert.Equal(t, 2, field.Len(), "expected exactly 2 rows, no residual data")
	}
}

func TestFramer_toFrame_EmptyMessages(t *testing.T) {
	f := newFramer()

	// Populate with data first.
	_, err := f.toFrame(makeMessages(3))
	require.NoError(t, err)

	// Call with empty slice — all fields should have 0 rows.
	frame, err := f.toFrame([]Message{})
	require.NoError(t, err)
	for _, field := range frame.Fields {
		assert.Equal(t, 0, field.Len(), "expected 0 rows after empty messages")
	}
}
