package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"path"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/litmus/edge/pkg/edge"
)

// Make sure EdgeDatasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.CheckHealthHandler    = (*EdgeDatasource)(nil)
	_ backend.CallResourceHandler   = (*EdgeDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*EdgeDatasource)(nil)
	_ backend.StreamHandler         = (*EdgeDatasource)(nil) // Streaming data source needs to implement this
)

// NewEdgeInstance creates a new datasource instance.
func NewEdgeInstance(_ context.Context, s backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, apiToken, err := getSettings(s)
	if err != nil {
		return nil, err
	}

	client, err := edge.NewClient(*settings)
	if err != nil {
		log.DefaultLogger.Error("Error creating the client", "error", err)
		return nil, err
	}

	var deviceHub edge.DeviceHubClient
	if apiToken != "" {
		deviceHub = edge.NewDeviceHubClient(settings.Hostname, apiToken)
	}

	return NewEdgeDatasource(client, s.UID, deviceHub), nil
}

func getSettings(s backend.DataSourceInstanceSettings) (*edge.ConnectionOptions, string, error) {
	opts := &edge.ConnectionOptions{}

	if err := json.Unmarshal(s.JSONData, opts); err != nil {
		return nil, "", fmt.Errorf("error reading settings: %w", err)
	}

	if token, ok := s.DecryptedSecureJSONData["token"]; ok {
		opts.Token = token
	}

	if opts.Hostname == "" {
		return nil, "", fmt.Errorf("hostname is required")
	}
	if opts.Token == "" {
		return nil, "", fmt.Errorf("Access Account token is required")
	}

	apiToken := s.DecryptedSecureJSONData["apiToken"]

	return opts, apiToken, nil
}

type EdgeDatasource struct {
	Client        edge.Client
	channelPrefix string
	deviceHub     edge.DeviceHubClient
}

func NewEdgeDatasource(client edge.Client, uid string, deviceHub edge.DeviceHubClient) *EdgeDatasource {
	return &EdgeDatasource{
		Client:        client,
		channelPrefix: path.Join("ds", uid),
		deviceHub:     deviceHub,
	}
}

// * HeathCheck implements backend.CheckHealthHandler interface. See ./health.go
// * Streaming implements backend.StreamHandler interface. See ./stream.go

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *EdgeDatasource) Dispose() {
	// Clean up datasource instance resources.
	d.Client.Dispose()
}
