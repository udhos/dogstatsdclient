// Package main implements the example.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"log/slog"

	"github.com/udhos/dogstatsdclient/dogstatsdclient"
)

func main() {

	var mock bool
	var sampleRate float64
	var namespace string
	var tags string

	flag.BoolVar(&mock, "mock", false, "enable mock")
	flag.Float64Var(&sampleRate, "sampleRate", 1, "sample rate")
	flag.StringVar(&namespace, "namespace", "namespace1", "namespace")
	flag.StringVar(&tags, "tags", "k1:v1 k2:v2", "space-delimited tags")
	flag.Parse()

	slog.Info("flag",
		"mock", mock,
		"sampleRate", sampleRate,
		"namespace", namespace,
		"tags", tags,
	)

	//
	// metrics exporter
	//

	var client dogstatsdClient

	if mock {
		client = &statsdMock{}
	} else {
		c, errClient := dogstatsdclient.New(dogstatsdclient.Options{
			Namespace: namespace,
			Debug:     true,
		})
		if errClient != nil {
			slog.Error(errClient.Error())
			os.Exit(1)
		}
		client = c
	}

	//
	// send metrics
	//

	const interval = 5 * time.Second

	t := strings.Fields(tags)

	for {
		send(client, "metric1", 3, t, sampleRate)
		time.Sleep(interval)
	}
}

func send(client dogstatsdClient, metric string, value int64, tags []string, sampleRate float64) {
	slog.Info(fmt.Sprintf("sending COUNT name=%s value=%d", metric, value))
	client.Count(metric, value, tags, sampleRate)
}

// dogstatsdClient is implemented by *statsd.Client.
// Simplified version of statsd.ClientInterface.
type dogstatsdClient interface {
	// Gauge measures the value of a metric at a particular time.
	Gauge(name string, value float64, tags []string, rate float64) error

	// Count tracks how many times something happened per second.
	Count(name string, value int64, tags []string, rate float64) error

	// Close the client connection.
	Close() error
}

type statsdMock struct {
}

// Gauge measures the value of a metric at a particular time.
func (s *statsdMock) Gauge(name string, value float64, tags []string, rate float64) error {
	slog.Info(
		"statsdMock.Gauge",
		"name", name,
		"value", value,
		"tags", tags,
		"rate", rate,
	)
	return nil
}

// Count tracks how many times something happened per second.
func (s *statsdMock) Count(name string, value int64, tags []string, rate float64) error {
	slog.Info(
		"statsdMock.Count",
		"name", name,
		"value", value,
		"tags", tags,
		"rate", rate,
	)
	return nil
}

// Close the client connection.
func (s *statsdMock) Close() error {
	return nil
}
