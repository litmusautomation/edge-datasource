package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/litmus/edge/pkg/edge"
)

const natsConnectError = `Could not connect to Litmus Edge.

Please check:
• Litmus Edge is reachable at the configured hostname
• NATS Proxy is enabled (port 4222)
• The token has NATS Proxy read access`

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *EdgeDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if !d.Client.IsConnected() {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: natsConnectError,
		}, nil
	}

	if d.deviceHub == nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusOk,
			Message: "Connected to the Edge. Tip: enable topic autocomplete for search-as-you-type suggestions.",
		}, nil
	}

	// Validate the Edge API connection when an API token is configured.
	if _, err := d.deviceHub.SearchTopics(ctx, ""); err != nil {
		var message string
		if errors.Is(err, edge.ErrUnauthorized) {
			message = "Topic autocomplete: API token is invalid or expired"
		} else {
			message = "Topic autocomplete: could not reach the Edge API"
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
		Message: "Connected to the Edge. Topic autocomplete is working.",
	}, nil
}
