[![license](http://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/udhos/dogstatsdclient/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/udhos/dogstatsdclient)](https://goreportcard.com/report/github.com/udhos/dogstatsdclient)
[![Go Reference](https://pkg.go.dev/badge/github.com/udhos/dogstatsdclient.svg)](https://pkg.go.dev/github.com/udhos/dogstatsdclient)

# dogstatsdclient

[dogstatsdclient](https://github.com/udhos/dogstatsdclient) creates a client for Dogstatsd.

# Usage

See [./examples/dogstatsdclient/main.go](./examples/dogstatsdclient/main.go).

```go
client, errClient := dogstatsdclient.New(dogstatsdclient.Options{
    Namespace: "my-namespace",
    Debug:     true,
})

perMetricTags := []string{"key1:value1"}
const sampleRate = 1

client.Count("metric1", 3, perMetricTags, sampleRate)
```
