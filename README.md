# sentry-go-exporter-opencensus

[![Build Status](https://github.com/vearutop/sentry-go-exporter-opencensus/workflows/test-unit/badge.svg)](https://github.com/vearutop/sentry-go-exporter-opencensus/actions?query=branch%3Amaster+workflow%3Atest-unit)
[![Coverage Status](https://codecov.io/gh/vearutop/sentry-go-exporter-opencensus/branch/master/graph/badge.svg)](https://codecov.io/gh/vearutop/sentry-go-exporter-opencensus)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/vearutop/sentry-go-exporter-opencensus)
[![Time Tracker](https://wakatime.com/badge/github/vearutop/sentry-go-exporter-opencensus.svg)](https://wakatime.com/badge/github/vearutop/sentry-go-exporter-opencensus)
![Code lines](https://sloc.xyz/github/vearutop/sentry-go-exporter-opencensus/?category=code)
![Comments](https://sloc.xyz/github/vearutop/sentry-go-exporter-opencensus/?category=comments)

Provides [OpenCensus](https://github.com/opencensus-integrations) exporter support for [Sentry](https://sentry.io/).

![Sentry Trace](./sentry-trace.png)

OpenCensus has tracing instrumentations for a variety of technologies (databases, services, caches, etc...), this library
enables those instrumentations for Sentry performance tools with zero effort.

## Usage

```go
package main

import (
	"log"
	"time"

	sen "github.com/getsentry/sentry-go"
	"github.com/vearutop/sentry-go-exporter-opencensus"
	"go.opencensus.io/trace"
)

func main() {
	// Initialize Sentry.
	err := sen.Init(sen.ClientOptions{
		Dsn:        "https://abc123abc123abc123abc123@o123456.ingest.sentry.io/1234567",
		ServerName: "my-service",
		Release:    "v1.2.3",
	})
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		sen.Flush(time.Second)
	}()

	// Setup OC sampling.
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.ProbabilitySampler(0.01),
	})

	// Enable Sentry exporter.
	trace.RegisterExporter(sentry.NewExporter())

	// Use OpenCensus integrations.
}

```