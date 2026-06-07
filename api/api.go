package api

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
	fmt.Printf("[api] BootstrapWithString called, toml length: %d\n", len(tomlStr))

	internalOverrides := config.ClientConfigOverrides{}
	if len(overrides.Resolvers) > 0 {
		fmt.Printf("[api] Parsing %d resolver overrides\n", len(overrides.Resolvers))
		var resAddrs []config.ResolverAddress
		for _, r := range overrides.Resolvers {
			ip := r
			port := 53
			if strings.Contains(r, ":") {
				parts := strings.Split(r, ":")
				ip = parts[0]
				if p, err := strconv.Atoi(parts[1]); err == nil {
					port = p
				}
			}
			resAddrs = append(resAddrs, config.ResolverAddress{IP: ip, Port: port})
		}
		internalOverrides.Resolvers = resAddrs
	}

	cfg, err := config.LoadClientConfigWithString(tomlStr, internalOverrides)
	if err != nil {
		fmt.Printf("[api] LoadClientConfigWithString failed: %v\n", err)
		return nil, err
	}

	fmt.Printf("[api] Calling client.BootstrapLoadedConfig...\n")
	c, err := client.BootstrapLoadedConfig(cfg)
	if err != nil {
		fmt.Printf("[api] client.BootstrapLoadedConfig failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("[api] Bootstrap successful\n")
	return &Client{inner: c}, nil
}

func RunClient(ctx context.Context, c *Client) error {
	fmt.Printf("[api] RunClient started\n")
	err := c.inner.Run(ctx)
	if err != nil {
		fmt.Printf("[api] RunClient exited with error: %v\n", err)
	} else {
		fmt.Printf("[api] RunClient exited cleanly\n")
	}
	return err
}

func (c *Client) SetMinValidResolvers(n int) {
	c.inner.SetMinValidResolvers(n)
}

func (c *Client) GetStats() map[string]uint64 {
	stats := c.inner.GetTrafficStats()
	// Add MTU stats
	t, com, v, rej := c.inner.GetMtuStats()
	stats["mtuTotal"] = uint64(t)
	stats["mtuCompleted"] = uint64(com)
	stats["mtuValid"] = uint64(v)
	stats["mtuRejected"] = uint64(rej)
	return stats
}

func (c *Client) GetValidResolvers() []string {
	return c.inner.GetValidResolvers()
}

func (c *Client) GetRejectedResolvers() []string {
	return c.inner.GetRejectedResolvers()
}

func ParseClientResolvers(raw string) ([]string, int, error) {
	endpoints, _, err := config.ParseClientResolversString(raw)
	if err != nil {
		return nil, 0, err
	}
	var res []string
	for _, e := range endpoints {
		res = append(res, fmt.Sprintf("%s:%d", e.IP, e.Port))
	}
	return res, len(res), nil
}
