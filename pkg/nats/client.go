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
	Host          string
	Token         string
	SkipTLSVerify bool
}

type client struct {
	conn *nats.Conn
}

func NewClient(opts ConnectionOptions) (Client, error) {
	url := fmt.Sprintf("nats://admin:%s@%s:4222", opts.Token, opts.Host)
	natsOpts := []nats.Option{}

	if opts.SkipTLSVerify {
		natsOpts = append(natsOpts, nats.Secure(&tls.Config{InsecureSkipVerify: true}))
	}

	conn, err := nats.Connect(url, natsOpts...)
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
	return false
}

func (c *client) Dispose() {}
