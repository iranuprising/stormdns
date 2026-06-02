package api

import (
	"context"
	"fmt"
	"stormdns-go/internal/client"
	"stormdns-go/internal/config"
	"sync"
)

type Client struct {
	inner *client.Client
	mu    sync.Mutex
}

type ClientConfigOverrides struct {
	Resolvers []string
}

func BootstrapWithString(tomlStr string, overrides ClientConfigOverrides) (*Client, error) {
	cfg, err := config.ParseClientString(tomlStr)
	if err != nil {
		return nil, err
	}
	if len(overrides.Resolvers) > 0 {
		cfg.Resolvers = overrides.Resolvers
	}
	c, err := client.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{inner: c}, nil
}

func RunClient(ctx context.Context, c *Client) error {
	return c.inner.Run(ctx)
}

func (c *Client) SetMinValidResolvers(n int) {
	c.inner.SetMinValidResolvers(n)
}

func (c *Client) GetStats() map[string]uint64 {
	return c.inner.GetTrafficStats()
}

func (c *Client) GetValidResolvers() []string {
	return c.inner.GetValidResolvers()
}

func (c *Client) GetRejectedResolvers() []string {
	return c.inner.GetRejectedResolvers()
}

func ParseClientResolvers(raw string) ([]string, int, error) {
	return config.ParseResolvers(raw)
}
