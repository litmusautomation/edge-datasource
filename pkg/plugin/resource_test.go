package plugin

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/litmus/edge/pkg/edge"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDeviceHub struct {
	topics []string
	err    error
}

func (m *mockDeviceHub) SearchTopics(_ context.Context, _ string) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.topics, nil
}

type capturedResponse struct {
	status int
	body   []byte
}

func callResource(t *testing.T, ds *EdgeDatasource, rawURL string) capturedResponse {
	t.Helper()
	var captured capturedResponse
	sender := backend.CallResourceResponseSenderFunc(func(resp *backend.CallResourceResponse) error {
		captured.status = resp.Status
		captured.body = resp.Body
		return nil
	})

	// Split path from query string, matching Grafana SDK behavior.
	parsedPath := rawURL
	if qIdx := strings.IndexByte(rawURL, '?'); qIdx >= 0 {
		parsedPath = rawURL[:qIdx]
	}

	err := ds.CallResource(context.Background(), &backend.CallResourceRequest{
		Path: parsedPath,
		URL:  rawURL,
	}, sender)
	require.NoError(t, err)
	return captured
}


func TestCallResource_UnknownPath(t *testing.T) {
	ds := NewEdgeDatasource(newMockClient(true), "uid", nil, false)
	resp := callResource(t, ds, "unknown")
	assert.Equal(t, 404, resp.status)
}

func TestHandleTopics_NoToken(t *testing.T) {
	ds := NewEdgeDatasource(newMockClient(true), "uid", nil, false)
	resp := callResource(t, ds, "topics")

	assert.Equal(t, 200, resp.status)
	var body topicResponse
	require.NoError(t, json.Unmarshal(resp.body, &body))
	assert.Equal(t, "api_token_not_configured", body.Error)
	assert.Equal(t, []string{}, body.Topics)
}

func TestHandleTopics_HappyPath(t *testing.T) {
	hub := &mockDeviceHub{topics: []string{"topic.a", "topic.b"}}
	ds := NewEdgeDatasource(newMockClient(true), "uid", hub, false)
	resp := callResource(t, ds, "topics?query=test")

	assert.Equal(t, 200, resp.status)
	var body topicResponse
	require.NoError(t, json.Unmarshal(resp.body, &body))
	assert.Empty(t, body.Error)
	assert.Equal(t, []string{"topic.a", "topic.b"}, body.Topics)
}

func TestHandleTopics_EmptyQuery(t *testing.T) {
	hub := &mockDeviceHub{topics: []string{"all.topics"}}
	ds := NewEdgeDatasource(newMockClient(true), "uid", hub, false)
	resp := callResource(t, ds, "topics")

	assert.Equal(t, 200, resp.status)
	var body topicResponse
	require.NoError(t, json.Unmarshal(resp.body, &body))
	assert.Equal(t, []string{"all.topics"}, body.Topics)
}

func TestHandleTopics_Unauthorized(t *testing.T) {
	hub := &mockDeviceHub{err: edge.ErrUnauthorized}
	ds := NewEdgeDatasource(newMockClient(true), "uid", hub, false)
	resp := callResource(t, ds, "topics?query=test")

	assert.Equal(t, 200, resp.status)
	var body topicResponse
	require.NoError(t, json.Unmarshal(resp.body, &body))
	assert.Equal(t, "unauthorized", body.Error)
	assert.Equal(t, []string{}, body.Topics)
}

func TestHandleTopics_Unreachable(t *testing.T) {
	hub := &mockDeviceHub{err: assert.AnError}
	ds := NewEdgeDatasource(newMockClient(true), "uid", hub, false)
	resp := callResource(t, ds, "topics?query=test")

	assert.Equal(t, 200, resp.status)
	var body topicResponse
	require.NoError(t, json.Unmarshal(resp.body, &body))
	assert.Equal(t, "unreachable", body.Error)
	assert.Equal(t, []string{}, body.Topics)
}
