package plugin

import (
	"context"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"github.com/grafana/grafana-plugin-sdk-go/data"
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
	ctx, span := tracing.DefaultTracer().Start(ctx, "RunStream")
	defer span.End()

	logger := log.DefaultLogger.FromContext(ctx)

	// Subscribe to the topic
	err := ds.Client.Subscribe(req.Path)
	if err != nil {
		tracing.Error(span, err)
		return err
	}

	logger.Debug("Started Streaming", "path", req.Path)

	// Unsubscribe from the topic when the context is canceled
	defer func() {
		if err := ds.Client.Unsubscribe(req.Path); err != nil {
			logger.Error("Failed to unsubscribe from NATS topic", "path", req.Path, "error", err)
		}
	}()

	// Create a ticker to send data frames at the specified interval
	// TODO: Make the interval configurable
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Stopped streaming (context canceled)", "path", req.Path)
			return nil
		case <-ticker.C:
			topic, ok := ds.Client.GetTopic(req.Path)
			if !ok {
				logger.Debug("Topic not found", "path", req.Path)
				break
			}

			msgs := topic.DrainMessages()
			frame, err := topic.ToDataFrame(msgs)
			if err != nil {
				logger.Error("Failed to convert topic to data frame", "path", req.Path, "error", backend.DownstreamError(err))
				break
			}

			if err := sender.SendFrame(frame, data.IncludeAll); err != nil {
				logger.Error("Failed to send data frame", "path", req.Path, "error", backend.DownstreamError(err))
				return nil
			}
		}
	}
}
