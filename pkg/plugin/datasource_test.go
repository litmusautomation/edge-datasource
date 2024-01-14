package plugin

import (
	"context"
	"fmt"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	edge "github.com/litmus/edge/pkg/nats"
	"github.com/stretchr/testify/require"
)

var (
	HOSTNAME        = "127.0.0.1"
	TOKEN           = "s3cr3t"
	SKIP_TLS_VERIFY = true

	CLIENT_SETTINGS = edge.ConnectionOptions{
		Token: TOKEN,
		Host:  HOSTNAME,
	}

	SERVER_SETTINGS = backend.DataSourceInstanceSettings{
		JSONData: []byte(fmt.Sprintf(`{"host": "%s"}`, HOSTNAME)),
		DecryptedSecureJSONData: map[string]string{
			"token": TOKEN,
		},
	}
)

func TestQueryData(t *testing.T) {
	ds := EdgeDatasource{}

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			Queries: []backend.DataQuery{
				{RefID: "A"},
			},
		},
	)
	if err != nil {
		t.Error(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}
}

func TestNewEdgeInstance(t *testing.T) {
	t.Run("should return a new instance of EdgeDatasource", func(t *testing.T) {
		ctx := context.Background()
		settings := SERVER_SETTINGS
		instance, err := NewEdgeInstance(ctx, settings)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if instance == nil {
			t.Error("Expected non-nil instance")
		}
	})

	t.Run("should return an error if settings are invalid", func(t *testing.T) {
		ctx := context.Background()
		settings := backend.DataSourceInstanceSettings{}
		_, err := NewEdgeInstance(ctx, settings)
		if err == nil {
			t.Error("Expected error")
		}
	})
}

func TestCheckHealth(t *testing.T) {
	t.Run("should return HealthStatusOk", func(t *testing.T) {
		client, err := edge.NewClient(CLIENT_SETTINGS)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		ds := NewEdgeDatasource(client)
		res, _ := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
		require.Equal(t, res.Status, backend.HealthStatusOk)
		require.Equal(t, res.Message, "Connected to the Edge")
	})

	t.Run("should return HealthStatusError", func(t *testing.T) {
		client, err := edge.NewClient(CLIENT_SETTINGS)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		ds := NewEdgeDatasource(client)
		ds.Dispose()
		res, _ := ds.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
		require.Equal(t, res.Status, backend.HealthStatusError)
		require.Equal(t, res.Message, "Not connected to the Edge")
	})
}
