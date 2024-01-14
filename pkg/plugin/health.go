package plugin

import (
	"context"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *EdgeDatasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	var status = backend.HealthStatusOk
	var message = "Connected to the Edge"

	if !d.Client.IsConnected() {
		status = backend.HealthStatusError
		message = "Not connected to the Edge"
	}

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}
