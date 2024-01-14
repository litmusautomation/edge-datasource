package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	edge "github.com/litmus/edge/pkg/nats"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*EdgeDatasource)(nil)
	_ backend.CheckHealthHandler    = (*EdgeDatasource)(nil)
	_ instancemgmt.InstanceDisposer = (*EdgeDatasource)(nil)
)

// NewEdgeInstance creates a new datasource instance.
func NewEdgeInstance(_ context.Context, s backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	settings, err := getSettings(s)
	if err != nil {
		return nil, err
	}

	client, err := edge.NewClient(*settings)
	if err != nil {
		return nil, err
	}

	return NewEdgeDatasource(client), nil
}

type EdgeDatasource struct {
	Client edge.Client
}

func NewEdgeDatasource(client edge.Client) *EdgeDatasource {
	return &EdgeDatasource{
		Client: client,
	}
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *EdgeDatasource) Dispose() {
	// Clean up datasource instance resources.
	d.Client.Dispose()
}

func getSettings(s backend.DataSourceInstanceSettings) (*edge.ConnectionOptions, error) {
	opts := &edge.ConnectionOptions{}

	if err := json.Unmarshal(s.JSONData, opts); err != nil {
		return nil, fmt.Errorf("error reading settings: %w", err)
	}

	if token, ok := s.DecryptedSecureJSONData["token"]; ok {
		opts.Token = token
	}

	return opts, nil
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *EdgeDatasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

type queryModel struct{}

func (d *EdgeDatasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	// https://grafana.com/developers/plugin-tools/introduction/data-frames
	frame := data.NewFrame("response")

	// add fields.
	frame.Fields = append(frame.Fields,
		data.NewField("time", nil, []time.Time{query.TimeRange.From, query.TimeRange.To}),
		data.NewField("values", nil, []int64{10, 20}),
	)

	// add the frames to the response.
	response.Frames = append(response.Frames, frame)

	return response
}
