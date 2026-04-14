package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"

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

	natsProxyPort := settings.NATSProxyPort
	if natsProxyPort == "" {
		natsProxyPort = edge.DefaultNATSProxyPort
	}

	var deviceHub edge.DeviceHubClient
	if apiToken != "" {
		deviceHub = edge.NewDeviceHubClient(settings.Hostname, apiToken)
	}

	return NewEdgeDatasource(client, s.UID, deviceHub, bool(settings.ExternalEdge), natsProxyPort), nil
}

func getSettings(s backend.DataSourceInstanceSettings) (*edge.ConnectionOptions, string, error) {
	opts := &edge.ConnectionOptions{}

	if err := json.Unmarshal(s.JSONData, opts); err != nil {
		log.DefaultLogger.Error("Failed to parse datasource settings JSON", "error", err)
		return nil, "", fmt.Errorf("invalid datasource configuration — please re-enter your settings and save")
	}

	if token, ok := s.DecryptedSecureJSONData["token"]; ok {
		opts.Token = token
	}
	if strings.TrimSpace(opts.NATSProxyPort) == "" {
		opts.NATSProxyPort = edge.DefaultNATSProxyPort
	}

	if bool(opts.ExternalEdge) {
		if opts.Hostname == "" {
			return nil, "", fmt.Errorf("Litmus Edge address is required when connecting to an external Litmus Edge")
		}
		if opts.Token == "" {
			return nil, "", fmt.Errorf("Access Account API Key is required when connecting to an external Litmus Edge")
		}
	} else {
		gatewayIP := strings.TrimSpace(opts.GatewayIP)
		if gatewayIP == "" {
			gatewayIP = edge.DefaultDockerGatewayIP
		}
		opts.GatewayIP = gatewayIP
		opts.Hostname = gatewayIP
		opts.Token = "" // no auth needed from docker0 whitelist
	}

	apiToken := s.DecryptedSecureJSONData["apiToken"]

	return opts, apiToken, nil
}

type EdgeDatasource struct {
	Client        edge.Client
	channelPrefix string
	deviceHub     edge.DeviceHubClient
	externalEdge  bool
	natsProxyPort string
}

func NewEdgeDatasource(client edge.Client, uid string, deviceHub edge.DeviceHubClient, externalEdge bool, natsProxyPort string) *EdgeDatasource {
	return &EdgeDatasource{
		Client:        client,
		channelPrefix: path.Join("ds", uid),
		deviceHub:     deviceHub,
		externalEdge:  externalEdge,
		natsProxyPort: natsProxyPort,
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
