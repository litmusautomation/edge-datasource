package plugin

import (
	"context"
	"fmt"
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
	// Subscribe to the topic
	err := ds.Client.Subscribe(req.Path)
	if err != nil {
		log.DefaultLogger.Warn("Failed to subscribe to topic", "path", req.Path, "error", err)
		return fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	log.DefaultLogger.Info("Started Streaming", "path", req.Path)

	// Unsubscribe from the topic when the context is canceled
	defer ds.Client.Unsubscribe(req.Path)

	// Create a ticker to send data frames at the specified interval
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			log.DefaultLogger.Warn("Stopped Streaming (context canceled)", "path", req.Path, "err", ctx.Err())
			ticker.Stop()
			return nil
		case <-ticker.C:
			// Get the topic
			topic, ok := ds.Client.GetTopic(req.Path)
			if !ok {
				log.DefaultLogger.Warn("Topic not found", "path", req.Path)
				break
			}

			// Convert the topic messages to a data frame
			frame, err := topic.ToDataFrame()
			if err != nil {
				log.DefaultLogger.Warn("Failed to convert topic to data frame", "path", req.Path, "error", err)
				break
			}

			// Clear the topic Map
			topic.Messages = []edge.Message{}

			// Send the frame
			if err := sender.SendFrame(frame, data.IncludeAll); err != nil {
				log.DefaultLogger.Warn("Failed to send the data frame", "path", req.Path, "error", err)
			}
		}
	}
}
