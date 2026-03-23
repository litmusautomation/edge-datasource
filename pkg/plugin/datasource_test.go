package plugin

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/litmus/edge/pkg/edge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient implements edge.Client for tests without a real NATS connection.
type mockClient struct {
	connected   bool
	subscribed  map[string]bool
	subscribeErr error
}

func newMockClient(connected bool) *mockClient {
	return &mockClient{connected: connected, subscribed: make(map[string]bool)}
}

func (m *mockClient) Subscribe(topic string) error {
	if m.subscribeErr != nil {
		return m.subscribeErr
	}
	m.subscribed[topic] = true
	return nil
}

func (m *mockClient) Unsubscribe(topic string) error {
	delete(m.subscribed, topic)
	return nil
}

func (m *mockClient) GetTopic(topic string) (*edge.Topic, bool) {
	return nil, false
}

func (m *mockClient) IsConnected() bool {
	return m.connected
}

func (m *mockClient) Dispose() {
	m.connected = false
}

func TestCheckHealth_Connected(t *testing.T) {
	ds := NewEdgeDatasource(newMockClient(true), "uid")
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusOk, res.Status)
}

func TestCheckHealth_Disconnected(t *testing.T) {
	ds := NewEdgeDatasource(newMockClient(false), "uid")
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusError, res.Status)
}

func TestNewEdgeInstance_InvalidSettings(t *testing.T) {
	settings := backend.DataSourceInstanceSettings{
		JSONData: []byte(`not-valid-json`),
	}
	_, err := NewEdgeInstance(context.Background(), settings)
	require.Error(t, err)
}

func TestNewEdgeInstance_EmptySettings(t *testing.T) {
	// Valid JSON but no hostname/token → NATS connect should fail.
	data, err := json.Marshal(map[string]string{"hostname": ""})
	require.NoError(t, err)
	settings := backend.DataSourceInstanceSettings{
		JSONData:                data,
		DecryptedSecureJSONData: map[string]string{"token": ""},
	}
	_, err = NewEdgeInstance(context.Background(), settings)
	require.Error(t, err, "expected error for empty NATS settings")
}
