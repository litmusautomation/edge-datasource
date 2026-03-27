package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/litmus/edge/pkg/edge"
)

const natsConnectErrorExternal = `Could not connect to Litmus Edge.

Please check:
• Litmus Edge is reachable at the configured address
• NATS Proxy is enabled and reachable on the configured NATS Proxy port
• The token has NATS Proxy read access`

const natsConnectErrorInternal = `Could not connect to Litmus Edge.

Please check:
• Edge Docker Gateway IP is correct for this Litmus Edge instance
• NATS Proxy is enabled and reachable on the configured NATS Proxy port
If the problem persists, turn on Remote Connection and provide the Litmus Edge address manually.`

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *EdgeDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if !d.Client.IsConnected() {
		msg := natsConnectErrorInternal
		if d.externalEdge {
			msg = natsConnectErrorExternal
		}
		msg = fmt.Sprintf("%s\n• Configured NATS Proxy port: %s", msg, d.natsProxyPort)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: msg,
		}, nil
	}

	if d.deviceHub == nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusOk,
			Message: "Connected to the Edge. Tip: add an API Token to enable topic discovery.",
		}, nil
	}

	// Validate the Edge API connection when an API Token is configured.
	if _, err := d.deviceHub.SearchTopics(ctx, ""); err != nil {
		var message string
		if errors.Is(err, edge.ErrUnauthorized) {
			message = "Topic discovery: API Token is invalid or expired"
		} else {
			message = "Topic discovery: could not reach the Edge API"
		}

		details, _ := json.Marshal(map[string]string{
			"nats":  "ok",
			"error": fmt.Sprintf("%v", err),
		})

		return &backend.CheckHealthResult{
			Status:      backend.HealthStatusError,
			Message:     message,
			JSONDetails: details,
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Connected to the Edge. Topic discovery is working.",
	}, nil
}
