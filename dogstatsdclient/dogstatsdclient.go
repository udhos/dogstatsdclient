// Package dogstatsdclient creates Dogstatsd client.
package dogstatsdclient

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

// Options define options for datadog client.
type Options struct {
	// Host defaults to env var DD_AGENT_HOST. Undefined DD_AGENT_HOST defaults to localhost.
	Host string

	// Port defaults to env var DD_AGENT_PORT. Undefined DD_AGENT_PORT defaults to 8125.
	Port string

	// Namespace sets the namespace.
	Namespace string

	// Service is used to define default Tags. If undefined, defaults to DD_SERVICE.
	Service string

	// Tags defaults to env var DD_TAGS.
	Tags []string

	// TagHostnameKey defaults to "pod_name".
	TagHostnameKey string

	// DisableTagHostnameKey prevents adding tag $TagHostnameKey:$hosname.
	DisableTagHostnameKey bool

	// Debug enables debugging logs.
	Debug bool

	// TTL defines maximum lifetime for internal Dogstatsd client created by method New.
	// Method NewUnsafe ignores TTL.
	// The internal client is renewed every TTL period in order to withstand DNS changes.
	// If unspecified, defaults to 1 minute.
	TTL time.Duration
}

// Client holds Dogstatsd client.
// Client implements the interface DogstatsdClient.
type Client struct {
	options        Options
	client         *statsd.Client
	clientCreation time.Time
	lock           sync.Mutex
}

const defaultTTL = time.Minute

// New creates Dogstatsd client.
func New(options Options) (*Client, error) {
	const me = "dogstatsdclient.New"
	if options.TTL < 1 {
		if options.Debug {
			slog.Info(me,
				"newTTL", defaultTTL,
				"oldTTL", options.TTL,
			)
		}
		options.TTL = defaultTTL
	}
	client := &Client{
		options: options,
	}
	err := client.renewIfExpired()
	return client, err
}

func (c *Client) isAlive() bool {
	return time.Since(c.clientCreation) < c.options.TTL
}

// renewIfExpired is unsafe for concurrency and must ge guarded by mutex.
func (c *Client) renewIfExpired() error {
	const me = "dogstatsdclient.renewIfExpired"
	if c.isAlive() {
		return nil
	}
	if c.options.Debug {
		slog.Info(me,
			"renewing", "client has expired, renewing",
			"ttl", c.options.TTL,
		)
	}
	// client has expired
	client, err := NewUnsafe(c.options)
	if err != nil {
		return err
	}
	if c.client != nil {
		c.client.Close()
	}
	c.client = client
	c.clientCreation = time.Now()
	return nil
}

// Close the client connection.
func (c *Client) Close() error {
	return c.client.Close()
}

// Count tracks how many times something happened per second.
func (c *Client) Count(name string, value int64, tags []string, rate float64) error {
	const me = "dogstatsdclient.Count"
	c.lock.Lock()
	defer c.lock.Unlock()
	if err := c.renewIfExpired(); err != nil {
		return err
	}
	c.debug(me, name, value, tags, rate)
	return c.client.Count(name, value, tags, rate)
}

// Gauge measures the value of a metric at a particular time.
func (c *Client) Gauge(name string, value float64, tags []string, rate float64) error {
	const me = "dogstatsdclient.Gauge"
	c.lock.Lock()
	defer c.lock.Unlock()
	if err := c.renewIfExpired(); err != nil {
		return err
	}
	c.debug(me, name, value, tags, rate)
	return c.client.Gauge(name, value, tags, rate)
}

// TimeInMilliseconds sends timing information in milliseconds.
func (c *Client) TimeInMilliseconds(name string, value float64, tags []string, rate float64) error {
	const me = "dogstatsdclient.TimeInMilliseconds"
	c.lock.Lock()
	defer c.lock.Unlock()
	if err := c.renewIfExpired(); err != nil {
		return err
	}
	c.debug(me, name, value, tags, rate)
	return c.client.TimeInMilliseconds(name, value, tags, rate)
}

func (c *Client) debug(caller string, name string, value any, tags []string, rate float64) {
	if c.options.Debug {
		slog.Info(caller,
			"name", name,
			"value", value,
			"tags", tags,
			"rate", rate,
		)
	}
}

// NewUnsafe creates UNSAFE Dogstatsd client. See New for a safe version.
//
// Dogstatds is unsafe for DNS changes.
//
// See: https://github.com/DataDog/datadog-go/pull/280
func NewUnsafe(options Options) (*statsd.Client, error) {

	const me = "dogstatsdclient.NewUnsafe"

	if options.Host == "" {
		options.Host = envString("DD_AGENT_HOST", "localhost")
	}

	if options.Port == "" {
		options.Port = envString("DD_AGENT_PORT", "8125")
	}

	if options.Service == "" {
		options.Service = envString("DD_SERVICE", "service-unknown")
	}

	if len(options.Tags) == 0 {
		options.Tags = strings.Fields(envString("DD_TAGS", ""))
	}

	if !options.DisableTagHostnameKey {
		if options.TagHostnameKey == "" {
			options.TagHostnameKey = "pod_name"
		}
		hostname, err := os.Hostname()
		if err != nil {
			return nil, err
		}
		// add tag pod_name:hostname
		options.Tags = append(options.Tags, fmt.Sprintf("%s:%s", options.TagHostnameKey, hostname))
	}

	// add service to tags
	options.Tags = append(options.Tags, fmt.Sprintf("service:%s", options.Service))

	// compact tags
	slices.Sort(options.Tags)
	options.Tags = slices.Compact(options.Tags)

	host := fmt.Sprintf("%s:%s", options.Host, options.Port)

	if options.Debug {
		slog.Info(
			me,
			"host", host,
			"namespace", options.Namespace,
			"service", options.Service,
			"tags", options.Tags,
		)
	}

	c, err := statsd.New(host,
		statsd.WithNamespace(options.Namespace),
		statsd.WithTags(options.Tags))

	return c, err
}

// envString extracts string from env var.
// It returns the provided defaultValue if the env var is empty.
// The string returned is also recorded in logs.
func envString(name string, defaultValue string) string {
	str := os.Getenv(name)
	if str != "" {
		slog.Info(fmt.Sprintf("%s=[%s] using %s=%s default=%s",
			name, str, name, str, defaultValue))
		return str
	}
	slog.Info(fmt.Sprintf("%s=[%s] using %s=%s default=%s",
		name, str, name, defaultValue, defaultValue))
	return defaultValue
}

// DogstatsdClient is simplified version of statsd.ClientInterface.
// DogstatsdClient is implemented by *dogstatsd.Client, created by New.
// DogstatsdClient is implemented by *statsd.Client, created by NewUnsafe.
type DogstatsdClient interface {
	// Gauge measures the value of a metric at a particular time.
	Gauge(name string, value float64, tags []string, rate float64) error

	// Count tracks how many times something happened per second.
	Count(name string, value int64, tags []string, rate float64) error

	// TimeInMilliseconds sends timing information in milliseconds.
	TimeInMilliseconds(name string, value float64, tags []string, rate float64) error

	// Close the client connection.
	Close() error
}
