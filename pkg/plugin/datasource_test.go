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
	ds := NewEdgeDatasource(newMockClient(true), "uid", nil)
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusOk, res.Status)
	assert.Contains(t, res.Message, "Connected to the Edge")
}

func TestCheckHealth_Disconnected(t *testing.T) {
	ds := NewEdgeDatasource(newMockClient(false), "uid", nil)
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Contains(t, res.Message, "NATS Proxy is enabled")
	assert.Equal(t, backend.HealthStatusError, res.Status)
}

func TestCheckHealth_DeviceHubOk(t *testing.T) {
	hub := &mockDeviceHub{topics: []string{"topic.a"}}
	ds := NewEdgeDatasource(newMockClient(true), "uid", hub)
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusOk, res.Status)
	assert.Contains(t, res.Message, "Topic autocomplete is working")
}

func TestCheckHealth_DeviceHubUnauthorized(t *testing.T) {
	hub := &mockDeviceHub{err: edge.ErrUnauthorized}
	ds := NewEdgeDatasource(newMockClient(true), "uid", hub)
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusError, res.Status)
	assert.Contains(t, res.Message, "API token is invalid or expired")
}

func TestCheckHealth_DeviceHubUnreachable(t *testing.T) {
	hub := &mockDeviceHub{err: assert.AnError}
	ds := NewEdgeDatasource(newMockClient(true), "uid", hub)
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusError, res.Status)
	assert.Contains(t, res.Message, "could not reach the Edge API")
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

func TestGetSettings_Validation(t *testing.T) {
	tests := []struct {
		name         string
		jsonData     string
		token        string
		apiToken     string
		wantErr      string
		wantAPIToken string
	}{
		{
			name:     "missing hostname",
			jsonData: `{"hostname": ""}`,
			token:    "valid-token",
			wantErr:  "hostname is required",
		},
		{
			name:     "missing token",
			jsonData: `{"hostname": "192.168.1.1"}`,
			token:    "",
			wantErr:  "Access Account token is required",
		},
		{
			name:     "valid settings without apiToken",
			jsonData: `{"hostname": "192.168.1.1"}`,
			token:    "valid-token",
		},
		{
			name:         "valid settings with apiToken",
			jsonData:     `{"hostname": "192.168.1.1"}`,
			token:        "valid-token",
			apiToken:     "my-api-token",
			wantAPIToken: "my-api-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secureData := map[string]string{"token": tt.token}
			if tt.apiToken != "" {
				secureData["apiToken"] = tt.apiToken
			}
			s := backend.DataSourceInstanceSettings{
				JSONData:                []byte(tt.jsonData),
				DecryptedSecureJSONData: secureData,
			}
			opts, apiToken, err := getSettings(s)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "192.168.1.1", opts.Hostname)
				assert.Equal(t, "valid-token", opts.Token)
				assert.Equal(t, tt.wantAPIToken, apiToken)
			}
		})
	}
}
