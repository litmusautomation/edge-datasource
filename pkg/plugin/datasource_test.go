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
	connected    bool
	subscribed   map[string]bool
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
	ds := NewEdgeDatasource(newMockClient(true), "uid", nil, false, edge.DefaultNATSProxyPort)
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusOk, res.Status)
	assert.Contains(t, res.Message, "Connected to the Edge")
}

func TestCheckHealth_DisconnectedInsideLE(t *testing.T) {
	ds := NewEdgeDatasource(newMockClient(false), "uid", nil, false, edge.DefaultNATSProxyPort)
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusError, res.Status)
	assert.Contains(t, res.Message, "Docker bridge network")
	assert.Contains(t, res.Message, "switch to External mode")
}

func TestCheckHealth_DisconnectedExternal(t *testing.T) {
	ds := NewEdgeDatasource(newMockClient(false), "uid", nil, true, "5222")
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusError, res.Status)
	assert.Contains(t, res.Message, "configured address")
	assert.Contains(t, res.Message, "token has NATS Proxy read access")
	assert.Contains(t, res.Message, "Configured NATS Proxy port: 5222")
}

func TestCheckHealth_DeviceHubOk(t *testing.T) {
	hub := &mockDeviceHub{topics: []string{"topic.a"}}
	ds := NewEdgeDatasource(newMockClient(true), "uid", hub, false, edge.DefaultNATSProxyPort)
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusOk, res.Status)
	assert.Contains(t, res.Message, "Topic autocomplete is working")
}

func TestCheckHealth_DeviceHubUnauthorized(t *testing.T) {
	hub := &mockDeviceHub{err: edge.ErrUnauthorized}
	ds := NewEdgeDatasource(newMockClient(true), "uid", hub, false, edge.DefaultNATSProxyPort)
	res, err := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err)
	assert.Equal(t, backend.HealthStatusError, res.Status)
	assert.Contains(t, res.Message, "EDGE Token is invalid or expired")
}

func TestCheckHealth_DeviceHubUnreachable(t *testing.T) {
	hub := &mockDeviceHub{err: assert.AnError}
	ds := NewEdgeDatasource(newMockClient(true), "uid", hub, false, edge.DefaultNATSProxyPort)
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

func TestNewEdgeInstance_ExternalEmptySettings(t *testing.T) {
	// External mode with no hostname/token → should fail validation.
	data, err := json.Marshal(map[string]interface{}{"externalEdge": true, "hostname": ""})
	require.NoError(t, err)
	settings := backend.DataSourceInstanceSettings{
		JSONData:                data,
		DecryptedSecureJSONData: map[string]string{"token": ""},
	}
	_, err = NewEdgeInstance(context.Background(), settings)
	require.Error(t, err, "expected error for empty external NATS settings")
}

func TestGetSettings_ExternalValidation(t *testing.T) {
	tests := []struct {
		name         string
		jsonData     string
		token        string
		apiToken     string
		wantErr      string
		wantAPIToken string
		wantHost     string
		wantNATSPort string
	}{
		{
			name:     "external: missing hostname",
			jsonData: `{"externalEdge": true, "hostname": ""}`,
			token:    "valid-token",
			wantErr:  "Litmus Edge address is required when connecting to an external Litmus Edge",
		},
		{
			name:     "external: missing token",
			jsonData: `{"externalEdge": true, "hostname": "192.168.1.1"}`,
			token:    "",
			wantErr:  "Access Account token is required when connecting to an external Litmus Edge",
		},
		{
			name:         "external: valid settings without apiToken",
			jsonData:     `{"externalEdge": true, "hostname": "192.168.1.1"}`,
			token:        "valid-token",
			wantHost:     "192.168.1.1",
			wantNATSPort: edge.DefaultNATSProxyPort,
		},
		{
			name:         "external: valid settings with apiToken",
			jsonData:     `{"externalEdge": true, "hostname": "192.168.1.1"}`,
			token:        "valid-token",
			apiToken:     "my-api-token",
			wantAPIToken: "my-api-token",
			wantHost:     "192.168.1.1",
			wantNATSPort: edge.DefaultNATSProxyPort,
		},
		{
			name:         "external: custom nats proxy port",
			jsonData:     `{"externalEdge": true, "hostname": "192.168.1.1:8443", "natsProxyPort": "5222"}`,
			token:        "valid-token",
			wantHost:     "192.168.1.1:8443",
			wantNATSPort: "5222",
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
				assert.Equal(t, tt.wantHost, opts.Hostname)
				assert.Equal(t, "valid-token", opts.Token)
				assert.Equal(t, tt.wantAPIToken, apiToken)
				if tt.wantNATSPort == "" {
					assert.Equal(t, edge.DefaultNATSProxyPort, opts.NATSProxyPort)
				} else {
					assert.Equal(t, tt.wantNATSPort, opts.NATSProxyPort)
				}
			}
		})
	}
}

func TestGetSettings_ExternalEdgeStringBool(t *testing.T) {
	// Grafana provisioning passes env-var defaults as strings, not bools.
	s := backend.DataSourceInstanceSettings{
		JSONData:                []byte(`{"externalEdge": "true", "hostname": "192.168.1.1"}`),
		DecryptedSecureJSONData: map[string]string{"token": "my-token"},
	}
	opts, _, err := getSettings(s)
	require.NoError(t, err)
	assert.True(t, bool(opts.ExternalEdge))
	assert.Equal(t, "192.168.1.1", opts.Hostname)

	// "false" as string
	s.JSONData = []byte(`{"externalEdge": "false"}`)
	s.DecryptedSecureJSONData = map[string]string{}
	opts, _, err = getSettings(s)
	if err != nil {
		// Gateway detection may fail in test environments — that's fine
		assert.Contains(t, err.Error(), "could not auto-detect")
	} else {
		assert.False(t, bool(opts.ExternalEdge))
	}
}

func TestGetSettings_InsideLE(t *testing.T) {
	// Inside-LE mode calls ResolveGatewayHost() which reads /proc/net/route.
	// In CI/test environments this may fail — that's expected; we just verify
	// that hostname and token are NOT required for inside-LE mode.
	s := backend.DataSourceInstanceSettings{
		JSONData:                []byte(`{"externalEdge": false}`),
		DecryptedSecureJSONData: map[string]string{},
	}
	opts, _, err := getSettings(s)
	if err != nil {
		assert.Contains(t, err.Error(), "could not auto-detect the Litmus Edge address")
	} else {
		assert.NotEmpty(t, opts.Hostname, "hostname should be resolved from gateway")
		assert.Empty(t, opts.Token, "token should be empty in inside-LE mode")
	}
}
