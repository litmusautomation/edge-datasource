package edge

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/nats-io/nats.go"
)

// topicTokenPattern validates that each dot-separated token contains only non-whitespace, non-dot characters.
var topicTokenPattern = regexp.MustCompile(`^[^\s.]+$`)

const DefaultNATSProxyPort = "4222"
const DefaultDockerGatewayIP = "10.30.50.1"

type Client interface {
	Subscribe(string) error
	Unsubscribe(string) error
	GetTopic(string) (*Topic, bool)
	IsConnected() bool
	Dispose()
}

type ConnectionOptions struct {
	Hostname      string     `json:"hostname"`
	GatewayIP     string     `json:"gatewayIp"`
	NATSProxyPort string     `json:"natsProxyPort"`
	Token         string     `json:"token"`
	ExternalEdge  StringBool `json:"externalEdge"`
}

// StringBool handles JSON values that may be a bool or a string ("true"/"false").
// Grafana provisioning passes env-var defaults as strings.
type StringBool bool

func (b *StringBool) UnmarshalJSON(data []byte) error {
	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch v := raw.(type) {
	case bool:
		*b = StringBool(v)
	case string:
		*b = StringBool(strings.EqualFold(v, "true") || v == "1")
	default:
		*b = false
	}
	return nil
}

type client struct {
	conn     *nats.Conn
	topicMap *TopicMap
}

func NewClient(opts ConnectionOptions) (Client, error) {
	host := stripPort(opts.Hostname)
	port, err := normalizeNATSProxyPort(opts.NATSProxyPort)
	if err != nil {
		return nil, backend.PluginErrorf("invalid NATS Proxy Port: %v", err)
	}
	natsURL := &url.URL{
		Scheme: "nats",
		Host:   net.JoinHostPort(host, port),
	}

	natsOpts := []nats.Option{
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			log.DefaultLogger.Warn("NATS disconnected", "hostname", opts.Hostname, "err", err)
		}),
		nats.ReconnectHandler(func(_ *nats.Conn) {
			log.DefaultLogger.Info("NATS reconnected", "hostname", opts.Hostname)
		}),
	}

	if bool(opts.ExternalEdge) {
		// External mode: authenticate and use TLS (LE uses self-signed certs)
		natsURL.User = url.UserPassword("admin", opts.Token)
		natsOpts = append(natsOpts, nats.Secure(&tls.Config{InsecureSkipVerify: true})) //nolint:gosec
	}
	// Inside LE: no auth, no TLS — loopedge-access whitelists docker0 connections

	conn, err := nats.Connect(natsURL.String(), natsOpts...)
	if err != nil {
		log.DefaultLogger.Error("NATS connection failed", "url", natsURL.Redacted(), "error", err)
		return nil, backend.DownstreamErrorf("could not connect to the NATS server — check that Litmus Edge is reachable and the NATS Proxy is enabled on port %s", port)
	}

	log.DefaultLogger.Info("Connected to NATS Server", "hostname", opts.Hostname)
	return &client{
		conn: conn,
		topicMap: &TopicMap{
			Map:           sync.Map{},
			subscriptions: make(map[string]*nats.Subscription),
		},
	}, nil
}

func (c *client) Subscribe(topicName string) error {
	_, span := tracing.DefaultTracer().Start(context.Background(), "edge.Subscribe")
	defer span.End()

	if topicName == "" {
		err := fmt.Errorf("empty topic")
		tracing.Error(span, err)
		return err
	}

	// Validate the topic
	if err := c.validateTopic(topicName); err != nil {
		wrapped := fmt.Errorf("invalid topic: %w", err)
		tracing.Error(span, wrapped)
		return wrapped
	}

	// Idempotent: if already subscribed, return without error
	if _, ok := c.topicMap.Load(topicName); ok {
		return nil
	}

	topic := &Topic{
		TopicName: topicName,
	}

	log.DefaultLogger.Debug("Subscribing to NATS Topic", "topic", topicName)
	sub, err := c.conn.Subscribe(topicName, c.MessageHandler)
	if err != nil {
		wrapped := backend.DownstreamErrorf("failed to subscribe to NATS topic: %s", err)
		tracing.Error(span, wrapped)
		return wrapped
	}

	c.topicMap.AddSubscription(topicName, sub)
	c.topicMap.Store(topic)
	return nil
}

func (c *client) Unsubscribe(topicName string) error {
	t, ok := c.GetTopic(topicName)
	if !ok {
		return nil // No error if topic doesn't exist
	}

	// Get the subscription
	sub := c.topicMap.GetSubscription(t.TopicName)
	if sub == nil {
		log.DefaultLogger.Debug("Subscription not found", "topic", topicName)
		return nil
	}

	// Unsubscribe from the topic
	log.DefaultLogger.Debug("Unsubscribing from NATS Topic", "topic", topicName)
	if err := sub.Unsubscribe(); err != nil {
		return backend.DownstreamErrorf("failed to unsubscribe from NATS topic %s: %s", topicName, err)
	}

	// Delete the topic
	c.topicMap.Delete(t.TopicName)

	// Remove the subscription
	c.topicMap.RemoveSubscription(t.TopicName)
	log.DefaultLogger.Debug("Unsubscribed from NATS Topic", "topic", topicName)
	return nil
}

func (c *client) GetTopic(topicName string) (*Topic, bool) {
	return c.topicMap.Load(topicName)
}

func (c *client) IsConnected() bool {
	return c.conn.IsConnected()
}

func (c *client) Dispose() {
	log.DefaultLogger.Debug("Disconnecting from NATS Topic", "url", c.conn.Opts.Url)
	c.conn.Close()
}

func (c *client) MessageHandler(msg *nats.Msg) {
	log.DefaultLogger.Debug("Received message", "topic", msg.Subject)
	c.topicMap.AddMessage(msg.Subject, c.MessageWrapper(msg))
}

// validateTopic validates the given topic string according to the following rules:
// - Wildcards ">" and "*" are not allowed.
// - Each token in the topic should consist of non-whitespace characters and should not contain any dots.
// Returns an error if the topic is invalid.
func (c *client) validateTopic(topic string) error {
	tokens := strings.Split(topic, ".")
	for _, token := range tokens {
		if token == ">" || token == "*" {
			return fmt.Errorf("wildcards are not allowed")
		}
		if !topicTokenPattern.MatchString(token) {
			return fmt.Errorf("invalid token: %s", token)
		}
	}

	return nil
}

// DH Tag Message type:
type DHMessage struct {
	Success     bool            `json:"success"`
	Datatype    string          `json:"datatype"`
	Timestamp   int64           `json:"timestamp"`
	RegisterId  string          `json:"registerId"`
	Value       interface{}     `json:"value"`
	DeviceId    string          `json:"deviceId"`
	TagName     string          `json:"tagName"`
	DeviceName  string          `json:"deviceName"`
	Description string          `json:"description"`
	Metadata    json.RawMessage `json:"metadata"`
}

const unixSecondsThreshold = int64(1_000_000_000)
const unixMillisecondsThreshold = int64(1_000_000_000_000)

// isDHMessage returns true if the parsed message has the required DeviceHub fields.
func isDHMessage(dh DHMessage) bool {
	return dh.TagName != "" && dh.Timestamp != 0 && dh.DeviceId != ""
}

// MessageWrapper is a wrapper for the NATS message
func (c *client) MessageWrapper(msg *nats.Msg) Message {
	var dhMessage DHMessage
	if err := json.Unmarshal(msg.Data, &dhMessage); err != nil || !isDHMessage(dhMessage) {
		return c.createMessageFromRawData(msg)
	}

	return c.createMessageFromDHMessage(msg, dhMessage)
}

func (c *client) createMessageFromRawData(msg *nats.Msg) Message {
	timestamp := c.getTimestampFromMessageData(msg.Data)

	return Message{
		FieldName: msg.Subject,
		Labels:    data.Labels{},
		Timestamp: timestamp,
		Value:     msg.Data,
	}
}

func (c *client) getTimestampFromMessageData(data []byte) time.Time {
	type hasTime struct {
		Timestamp int64 `json:"timestamp"`
	}
	var v hasTime
	err := json.Unmarshal(data, &v)
	if err == nil {
		if timestamp, ok := parseTimestamp(v.Timestamp); ok {
			return timestamp
		}
	}

	return time.Now()
}

func (c *client) createMessageFromDHMessage(msg *nats.Msg, dhMessage DHMessage) Message {
	fieldName := dhMessage.TagName
	timestamp, ok := parseTimestamp(dhMessage.Timestamp)
	if !ok {
		timestamp = time.Now()
	}
	labels := make(data.Labels)

	if msg.Subject != "" {
		labels["topic"] = msg.Subject
	}
	if dhMessage.Datatype != "" {
		labels["datatype"] = dhMessage.Datatype
	}
	if dhMessage.TagName != "" {
		labels["tagName"] = dhMessage.TagName
	}
	if dhMessage.DeviceId != "" {
		labels["deviceId"] = dhMessage.DeviceId
	}
	if dhMessage.DeviceName != "" {
		labels["deviceName"] = dhMessage.DeviceName
	}
	if dhMessage.Description != "" {
		labels["description"] = dhMessage.Description
	}
	if dhMessage.RegisterId != "" {
		labels["registerId"] = dhMessage.RegisterId
	}

	valueBytes, err := json.Marshal(dhMessage.Value)
	if err != nil {
		log.DefaultLogger.Warn("Failed to marshal DH value, falling back to raw data", "topic", msg.Subject, "error", err)
		return c.createMessageFromRawData(msg)
	}

	return Message{
		FieldName: fieldName,
		Labels:    labels,
		Timestamp: timestamp,
		Value:     valueBytes,
		Metadata:  dhMessage.Metadata,
	}
}

func parseTimestamp(value int64) (time.Time, bool) {
	if value == 0 {
		return time.Time{}, false
	}

	abs := value
	if abs < 0 {
		abs = -abs
	}

	if abs >= unixMillisecondsThreshold {
		return time.UnixMilli(value), true
	}
	if abs >= unixSecondsThreshold {
		return time.Unix(value, 0), true
	}

	return time.Time{}, false
}

// stripPort removes the port from a host:port string.
// If there is no port, the hostname is returned unchanged.
func stripPort(hostname string) string {
	host, _, err := net.SplitHostPort(hostname)
	if err != nil {
		return hostname // no port present
	}
	return host
}

func normalizeNATSProxyPort(port string) (string, error) {
	trimmed := strings.TrimSpace(port)
	if trimmed == "" {
		return DefaultNATSProxyPort, nil
	}

	value, err := strconv.Atoi(trimmed)
	if err != nil {
		return "", fmt.Errorf("must be a number between 1 and 65535")
	}
	if value < 1 || value > 65535 {
		return "", fmt.Errorf("must be a number between 1 and 65535")
	}

	return strconv.Itoa(value), nil
}
