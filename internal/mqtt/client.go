package mqtt

import (
	"crypto/tls"
	"fmt"
	"matter2mqtt/internal/config"

	paho "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	config config.MQTTConfig
	client paho.Client
}

func NewClient(cfg config.MQTTConfig) (*Client, error) {
	opts := paho.NewClientOptions()
	switch cfg.Tls {
	case false:
    	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", cfg.Server, cfg.Port))
    case true:
    	opts.AddBroker(fmt.Sprintf("ssl://%s:%d", cfg.Server, cfg.Port))
    	tlsConfig := &tls.Config{}
    	opts.SetTLSConfig(tlsConfig)
    }
	opts.SetUsername(cfg.User)
	opts.SetPassword(cfg.Password)
	opts.SetClientID("matter2mqtt")
	opts.SetAutoReconnect(true)

	return &Client{
		config: cfg,
		client: paho.NewClient(opts),
	}, nil
}

func (c *Client) Connect() error {
	token := c.client.Connect()
	token.Wait()
	return token.Error()
}

func (c *Client) Publish(topic, payload string, retained bool) error {
	token := c.client.Publish(topic, 0, retained, payload)
	token.Wait()
	return token.Error()
}
