package nats

import (
	"crypto/tls"
	"fmt"

	"github.com/nats-io/nats.go"
)

type Client interface {
	Subscribe(topic string, handler func(msg []byte)) error
	Unsubscribe(topic string) error
	IsConnected() bool
	Dispose()
}

type ConnectionOptions struct {
	Host  string `json:"hostname"`
	Token string `json:"token"`
}

type client struct {
	conn *nats.Conn
}

func NewClient(opts ConnectionOptions) (Client, error) {
	url := fmt.Sprintf("nats://admin:%s@%s:4222", opts.Token, opts.Host)
	skipVerify := nats.Secure(&tls.Config{InsecureSkipVerify: true})
	conn, err := nats.Connect(url, skipVerify)
	if err != nil {
		return nil, err
	}

	return &client{conn: conn}, nil
}

func (c *client) Subscribe(topic string, handler func(msg []byte)) error {
	return nil
}

func (c *client) Unsubscribe(topic string) error {
	return nil
}

func (c *client) IsConnected() bool {
	return c.conn.IsConnected()
}

func (c *client) Dispose() {
	c.conn.Close()
}
