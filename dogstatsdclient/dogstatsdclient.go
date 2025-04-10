// Package dogstatsdclient creates Dogstatsd client.
package dogstatsdclient

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

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

	Debug bool
}

// New creates datadog client.
func New(options Options) (*statsd.Client, error) {

	const me = "dogstatsdclient.New"

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

	// add service to tags
	options.Tags = append(options.Tags, fmt.Sprintf("service:%s", options.Service))

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
