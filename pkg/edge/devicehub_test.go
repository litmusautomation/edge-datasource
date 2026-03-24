package edge

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

// stubDeviceHubResponse builds a minimal GraphQL response body.
func stubDeviceHubResponse(registers []register) []byte {
	resp := graphQLResponse{
		Data: graphQLData{
			ListRegistersFromAllDevices: listRegistersResult{
				Registers: registers,
			},
		},
	}
	b, _ := json.Marshal(resp)
	return b
}

func TestSearchTopics_HappyPath(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Contains(t, r.Header.Get("Authorization"), "Basic ")

		w.WriteHeader(http.StatusOK)
		w.Write(stubDeviceHubResponse([]register{
			{Topics: []topicInfo{
				{Format: "Raw", Topic: "devicehub.alias.demo.temp"},
				{Format: "Value", Topic: "devicehub.write.demo.temp"},
			}},
		}))
	})

	// Use the TLS server's host (strip scheme)
	client := newTestClient(srv)
	topics, err := client.SearchTopics(context.Background(), "temp")
	require.NoError(t, err)
	assert.Equal(t, []string{"devicehub.alias.demo.temp"}, topics)
}

func TestSearchTopics_FiltersOnlyRawFormat(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(stubDeviceHubResponse([]register{
			{Topics: []topicInfo{
				{Format: "Raw", Topic: "topic.raw"},
				{Format: "Value", Topic: "topic.value"},
				{Format: "WriteResponse", Topic: "topic.wr"},
				{Format: "Command", Topic: "topic.cmd"},
			}},
		}))
	})

	client := newTestClient(srv)
	topics, err := client.SearchTopics(context.Background(), "topic")
	require.NoError(t, err)
	assert.Equal(t, []string{"topic.raw"}, topics)
}

func TestSearchTopics_Deduplication(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(stubDeviceHubResponse([]register{
			{Topics: []topicInfo{{Format: "Raw", Topic: "dup.topic"}}},
			{Topics: []topicInfo{{Format: "Raw", Topic: "dup.topic"}}},
			{Topics: []topicInfo{{Format: "Raw", Topic: "unique.topic"}}},
		}))
	})

	client := newTestClient(srv)
	topics, err := client.SearchTopics(context.Background(), "")
	require.NoError(t, err)
	assert.Equal(t, []string{"dup.topic", "unique.topic"}, topics)
}

func TestSearchTopics_EmptyResults(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(stubDeviceHubResponse([]register{}))
	})

	client := newTestClient(srv)
	topics, err := client.SearchTopics(context.Background(), "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, []string{}, topics)
}

func TestSearchTopics_Unauthorized(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	})

	client := newTestClient(srv)
	_, err := client.SearchTopics(context.Background(), "temp")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnauthorized))
}

func TestSearchTopics_ServerError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	})

	client := newTestClient(srv)
	_, err := client.SearchTopics(context.Background(), "temp")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestSearchTopics_Unreachable(t *testing.T) {
	// Point to a closed server
	srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {})
	srv.Close()

	client := newTestClient(srv)
	_, err := client.SearchTopics(context.Background(), "temp")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "calling DeviceHub API")
}

func TestSearchTopics_MalformedJSON(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json"))
	})

	client := newTestClient(srv)
	_, err := client.SearchTopics(context.Background(), "temp")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing DeviceHub response")
}

func TestSearchTopics_GraphQLError(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		resp := graphQLResponse{Errors: []graphQLError{{Message: "field not found"}}}
		b, _ := json.Marshal(resp)
		w.Write(b)
	})

	client := newTestClient(srv)
	_, err := client.SearchTopics(context.Background(), "temp")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "field not found")
}

func TestSearchTopics_SendsCorrectPayload(t *testing.T) {
	srv := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var req graphQLRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		assert.Equal(t, "ListRegistersFromAllDevices", req.OperationName)
		assert.Contains(t, req.Query, "ListRegistersFromAllDevices")

		vars, ok := req.Variables.(map[string]interface{})
		require.True(t, ok)
		input, ok := vars["input"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "my_search", input["TagPattern"])
		assert.Equal(t, "CONTAINS", input["TagPatternSearchOption"])
		assert.Equal(t, float64(15), input["Limit"])

		w.WriteHeader(http.StatusOK)
		w.Write(stubDeviceHubResponse([]register{}))
	})

	client := newTestClient(srv)
	_, err := client.SearchTopics(context.Background(), "my_search")
	require.NoError(t, err)
}

// newTestClient creates a DeviceHubClient that points to the test server,
// reusing its TLS-configured HTTP client.
func newTestClient(srv *httptest.Server) DeviceHubClient {
	c := &deviceHubClient{
		httpClient: srv.Client(),
		endpoint:   srv.URL,
		authHeader: "Basic dGVzdA==",
	}
	return c
}
