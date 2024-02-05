package plugin

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/litmus/edge/pkg/edge"
)

func (ds *EdgeDatasource) QueryData(_ context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	response := backend.NewQueryDataResponse()

	for _, q := range req.Queries {
		res := ds.query(q)
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (ds *EdgeDatasource) query(query backend.DataQuery) backend.DataResponse {
	var (
		t        edge.Topic
		response backend.DataResponse
	)

	if err := json.Unmarshal(query.JSON, &t); err != nil {
		response.Error = err
		return response
	}

	if t.TopicName == "" {
		response.Error = fmt.Errorf("topic name is required")
		return response
	}

	frame := data.NewFrame("")
	frame.SetMeta(&data.FrameMeta{
		Channel: t.TopicName,
	})

	response.Frames = append(response.Frames, frame)
	return response
}
