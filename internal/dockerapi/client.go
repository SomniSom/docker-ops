package dockerapi

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/client"
)

const defaultLocalSocket = "/var/run/docker.sock"

// Client wraps docker/docker client with helpers.
type Client struct {
	inner *client.Client
	host  string
}

// NewLocal connects to local Docker via DOCKER_HOST or default unix socket.
func NewLocal(ctx context.Context) (*Client, error) {
	host := strings.TrimSpace(os.Getenv("DOCKER_HOST"))
	if host == "" {
		host = "unix://" + defaultLocalSocket
	} else if p, ok := ParseDockerHost(host); ok && !strings.HasPrefix(host, "unix://") {
		host = "unix://" + p
	}
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithHost(host), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	if _, err := cli.Ping(ctx); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("docker ping: %w", err)
	}
	return &Client{inner: cli, host: host}, nil
}

// NewFromHost connects to Docker at host (e.g. tcp://127.0.0.1:PORT from tunnel).
func NewFromHost(ctx context.Context, host string) (*Client, error) {
	cli, err := client.NewClientWithOpts(client.WithHost(host), client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	if _, err := cli.Ping(ctx); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("docker ping: %w", err)
	}
	return &Client{inner: cli, host: host}, nil
}

// Moby returns the underlying docker client.
func (c *Client) Moby() *client.Client {
	if c == nil {
		return nil
	}
	return c.inner
}

// Host returns the configured API host string.
func (c *Client) Host() string {
	if c == nil {
		return ""
	}
	return c.host
}

// Close closes the client.
func (c *Client) Close() error {
	if c == nil || c.inner == nil {
		return nil
	}
	return c.inner.Close()
}
