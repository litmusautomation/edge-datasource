package edge

import (
	"crypto/tls"
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/nats-io/nats.go"
)

type Client interface {
	Subscribe(string) error
	Unsubscribe(string)
	GetTopic(string) (*Topic, bool)
	IsConnected() bool
	Dispose()
}

type ConnectionOptions struct {
	Hostname string `json:"hostname"`
	Token    string `json:"token"`
}

type client struct {
	conn     *nats.Conn
	topicMap *TopicMap
}

func NewClient(opts ConnectionOptions) (Client, error) {
	url := fmt.Sprintf("nats://admin:%s@%s:4222", opts.Token, opts.Hostname)
	skipVerify := nats.Secure(&tls.Config{InsecureSkipVerify: true})
	conn, err := nats.Connect(url, skipVerify)
	if err != nil {
		return nil, err
	}

	log.DefaultLogger.Debug("Connected to NATS", "hostname", opts.Hostname)
	return &client{
		conn: conn,
		topicMap: &TopicMap{
			Map:           sync.Map{},
			subscriptions: make(map[string]*nats.Subscription),
		},
	}, nil
}

func (c *client) Subscribe(reqPath string) error {
	chunks := strings.Split(reqPath, "/")
	if len(chunks) != 3 {
		return fmt.Errorf("invalid topic path: %s", reqPath)
	}

	topicPath := chunks[2]
	// Validate the topic
	if err := c.validateTopic(topicPath); err != nil {
		return fmt.Errorf("invalid topic: %w", err)
	}

	addrPrefix := path.Join(chunks[0], chunks[1])
	topic := &Topic{
		TopicPath:  topicPath,
		AddrPrefix: addrPrefix,
	}

	if _, ok := c.topicMap.Load(reqPath); ok {
		return fmt.Errorf("already subscribed to topic: [%s]", topicPath)
	}

	log.DefaultLogger.Debug("Subscribing to NATS Topic", "topic", topicPath)
	sub, err := c.conn.Subscribe(topicPath, c.MessageHandler)
	if err != nil {
		return fmt.Errorf("failed to subscribe to NATS Topic: %w", err)
	}

	c.topicMap.AddSubscription(topicPath, sub)
	c.topicMap.Store(topic)
	return nil
}

func (c *client) Unsubscribe(reqPath string) {
	t, ok := c.GetTopic(reqPath)
	if !ok {
		log.DefaultLogger.Debug("Topic not found", "topic", reqPath)
		return
	}

	// Delete the topic
	c.topicMap.Delete(t.Key())

	// Check if the topic still has subscriptions
	// ? A Topic can be subscribed to by multiple streams (e.g. 1s/topic1, 5s/topic1)
	// ? If the topic still has subscriptions, we don't want to unsubscribe from it
	if exists := c.topicMap.HasSubscription(t.TopicPath); exists {
		log.DefaultLogger.Debug("Topic still has subscriptions", "topic", reqPath)
		return
	}

	// Get the subscription
	sub := c.topicMap.GetSubscription(t.TopicPath)
	if sub == nil {
		log.DefaultLogger.Debug("Subscription not found", "topic", reqPath)
		return
	}

	// Unsubscribe from the topic
	log.DefaultLogger.Debug("Unsubscribing from NATS Topic", "topic", reqPath)
	if err := sub.Unsubscribe(); err != nil {
		log.DefaultLogger.Debug("Failed to unsubscribe from NATS Topic", "topic", reqPath, "err", err)
		return
	}

	// Remove the subscription
	c.topicMap.RemoveSubscription(t.TopicPath)
	log.DefaultLogger.Debug("Unsubscribed from NATS Topic", "topic", reqPath)
}

func (c *client) GetTopic(reqPath string) (*Topic, bool) {
	return c.topicMap.Load(reqPath)
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

	message := Message{
		Timestamp: time.Now(),
		Value:     msg.Data,
	}

	c.topicMap.AddMessage(msg.Subject, message)
}

// validateTopic validates the given topic string according to the following rules:
// - Wildcards ">" and "*" are not allowed.
// - Each token in the topic should consist of non-whitespace characters and should not contain any dots.
// Returns an error if the topic is invalid.
func (c *client) validateTopic(topic string) error {
	// Compile the regex pattern once and reuse it
	pattern := regexp.MustCompile(`^[^\s.]+$`)

	tokens := strings.Split(topic, ".")
	for _, token := range tokens {
		if token == ">" || token == "*" {
			return fmt.Errorf("wildcards are not allowed")
		}
		if !pattern.MatchString(token) {
			return fmt.Errorf("invalid token: %s", token)
		}
	}

	return nil
}
