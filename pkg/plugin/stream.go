package plugin

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/litmus/edge/pkg/edge"
)

// SubscribeStream just returns an ok in this case, since we will always allow the user to successfully connect.
// Permissions verifications could be done here. Check backend.StreamHandler docs for more details.
func (ds *EdgeDatasource) SubscribeStream(_ context.Context, req *backend.SubscribeStreamRequest) (*backend.SubscribeStreamResponse, error) {
	return &backend.SubscribeStreamResponse{
		Status: backend.SubscribeStreamStatusOK,
	}, nil
}

// PublishStream just returns permission denied in this case, since in this example we don't want the user to send stream data.
// Permissions verifications could be done here. Check backend.StreamHandler docs for more details.
func (ds *EdgeDatasource) PublishStream(context.Context, *backend.PublishStreamRequest) (*backend.PublishStreamResponse, error) {
	return &backend.PublishStreamResponse{
		Status: backend.PublishStreamStatusPermissionDenied,
	}, nil
}

func (ds *EdgeDatasource) RunStream(ctx context.Context, req *backend.RunStreamRequest, sender *backend.StreamSender) error {
	// * req.Path must be in the format of "interval/topic"
	// * interval is the time between messages received from the frontend
	// * topic is the topic name (NATS subject)
	// * example: "1s/device.tag"
	log.DefaultLogger.Debug("Starting Streaming", "path", req.Path)

	// Split the path into chunks
	chunks := strings.Split(req.Path, "/")
	if len(chunks) != 2 {
		return fmt.Errorf("invalid path: %s", req.Path)
	}

	// Parse the interval
	interval, err := time.ParseDuration(chunks[0])
	if err != nil {
		return fmt.Errorf("failed to parse interval: %w", err)
	}

	// Subscribe to the topic
	err = ds.Client.Subscribe(chunks[1], interval)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	// Unsubscribe from the topic when the context is canceled
	defer ds.Client.Unsubscribe(req.Path)

	// Create a ticker
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Debug("Stopped Streaming (context canceled)", "path", req.Path, "err", ctx.Err())
			ticker.Stop()
			return nil
		case <-ticker.C:
			// Get the topic
			topic, ok := ds.Client.GetTopic(req.Path)
			if !ok {
				log.DefaultLogger.Debug("Topic not found", "path", req.Path)
				break
			}

			// Convert the topic messages to a data frame
			frame, err := topic.ToDataFrame()
			if err != nil {
				log.DefaultLogger.Debug("Failed to convert topic to data frame", "path", req.Path, "error", err)
				break
			}

			// Clear the topic Map
			topic.Messages = []edge.Message{}

			// Send the frame
			if err := sender.SendFrame(frame, data.IncludeAll); err != nil {
				log.DefaultLogger.Debug("Failed to send the data frame", "path", req.Path, "error", err)
			}
		}
	}
}
