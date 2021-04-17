package sentry_test

import (
	"log"
	"time"

	sen "github.com/getsentry/sentry-go"
	"github.com/vearutop/sentry-go-exporter-opencensus"
	"go.opencensus.io/trace"
)

func ExampleNewExporter() {
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
}
