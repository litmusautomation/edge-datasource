package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/litmus/edge/pkg/edge"
)

type topicResponse struct {
	Topics []string `json:"topics"`
	Error  string   `json:"error,omitempty"`
}

func (d *EdgeDatasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	switch req.Path {
	case "topics":
		return d.handleTopics(ctx, req, sender)
	default:
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusNotFound,
			Body:   []byte(`{"error":"not found"}`),
		})
	}
}

func (d *EdgeDatasource) handleTopics(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	if d.deviceHub == nil {
		return sendJSON(sender, topicResponse{
			Topics: []string{},
			Error:  "api_token_not_configured",
		})
	}

	query := ""
	if u, err := url.Parse(req.URL); err == nil {
		query = u.Query().Get("query")
	}

	topics, err := d.deviceHub.SearchTopics(ctx, query)
	if err != nil {
		log.DefaultLogger.Warn("DeviceHub topic search failed", "error", err)

		errKey := "unreachable"
		if errors.Is(err, edge.ErrUnauthorized) {
			errKey = "unauthorized"
		}
		return sendJSON(sender, topicResponse{
			Topics: []string{},
			Error:  errKey,
		})
	}

	return sendJSON(sender, topicResponse{Topics: topics})
}

func sendJSON(sender backend.CallResourceResponseSender, v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(`{"error":"internal error"}`),
		})
	}

	return sender.Send(&backend.CallResourceResponse{
		Status: http.StatusOK,
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: body,
	})
}
