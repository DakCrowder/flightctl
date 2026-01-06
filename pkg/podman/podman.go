// Package podman provides a client for interacting with the Podman CLI.
// This package contains no agent-specific dependencies and can be used by any component.
package podman

import (
	"time"

	"github.com/flightctl/flightctl/pkg/executer"
	"github.com/flightctl/flightctl/pkg/fileio"
	"github.com/flightctl/flightctl/pkg/log"
	"github.com/flightctl/flightctl/pkg/poll"
)

const (
	podmanCmd              = "podman"
	defaultPodmanTimeout   = 10 * time.Minute
	defaultPullLogInterval = 30 * time.Second
)

// Client provides methods for interacting with Podman.
type Client struct {
	exec       executer.Executer
	log        *log.PrefixLogger
	timeout    time.Duration
	readWriter fileio.ReadWriter
	backoff    *poll.Config
}

// Option is a functional option for configuring the Client.
type Option func(*Client)

// NewClient creates a new Podman client.
func NewClient(log *log.PrefixLogger, exec executer.Executer, rw fileio.ReadWriter, opts ...Option) *Client {
	c := &Client{
		log:        log,
		exec:       exec,
		timeout:    defaultPodmanTimeout,
		readWriter: rw,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithBackoff configures retry backoff settings for operations like pull.
func WithBackoff(config poll.Config) Option {
	return func(c *Client) {
		c.backoff = &config
	}
}

// WithTimeout sets a custom default timeout for operations.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
	}
}

// CallOption is a functional option for individual method calls.
type CallOption func(*callOptions)

type callOptions struct {
	pullSecretPath string
	timeout        time.Duration
}

// WithPullSecret sets the path to the authentication file for registry operations.
func WithPullSecret(path string) CallOption {
	return func(opts *callOptions) {
		opts.pullSecretPath = path
	}
}

// Timeout sets a custom timeout for this specific operation.
func Timeout(timeout time.Duration) CallOption {
	return func(opts *callOptions) {
		opts.timeout = timeout
	}
}
