// Package sentry provides OpenCensus traces exporter for Sentry.
package sentry

import (
	"context"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opencensus.io/trace"
)

// NewExporter creates Sentry Trace Exporter.
func NewExporter() *Exporter {
	return &Exporter{}
}

// Exporter is an implementation of trace.Exporter that uploads spans to Jaeger.
type Exporter struct {
	mu            sync.Mutex
	parentedSpans map[trace.SpanID][]*trace.SpanData
	lastCleanup   time.Time
}

var _ trace.Exporter = (*Exporter)(nil)

// ExportSpan exports SpanData to Sentry.
func (e *Exporter) ExportSpan(data *trace.SpanData) {
	if !data.IsSampled() {
		return
	}

	var (
		zero trace.SpanID
		name = data.Name
	)

	if data.ParentSpanID != zero && !data.HasRemoteParent {
		e.mu.Lock()
		if e.parentedSpans == nil {
			e.parentedSpans = make(map[trace.SpanID][]*trace.SpanData)
		}

		e.parentedSpans[data.ParentSpanID] = append(e.parentedSpans[data.ParentSpanID], data)
		if time.Since(e.lastCleanup) > time.Minute {
			e.lastCleanup = time.Now()
			go e.cleanup()
		}
		e.mu.Unlock()

		return
	}

	span := sentry.StartSpan(context.Background(), name, sentry.TransactionName(name))
	prepareSpan(span, data)

	e.addChildren(span, data.SpanID)

	span.Finish()
}

func (e *Exporter) cleanup() {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now()

	for id, spans := range e.parentedSpans {
		rm := false

		for _, span := range spans {
			if now.Sub(span.EndTime) > time.Minute {
				rm = true

				break
			}
		}

		if rm || len(spans) == 0 {
			delete(e.parentedSpans, id)
		}
	}
}

func (e *Exporter) addChildren(span *sentry.Span, spanID trace.SpanID) {
	e.mu.Lock()

	children := e.parentedSpans[spanID]
	if len(children) > 0 {
		delete(e.parentedSpans, spanID)
	}

	e.mu.Unlock()

	for _, c := range children {
		cs := span.StartChild(c.Name)
		prepareSpan(cs, c)
		e.addChildren(cs, c.SpanID)
	}
}

func prepareSpan(span *sentry.Span, data *trace.SpanData) {
	span.TraceID = sentry.TraceID(data.TraceID)
	span.SpanID = sentry.SpanID(data.SpanID)
	span.ParentSpanID = sentry.SpanID(data.ParentSpanID)
	span.Op = data.Name
	span.Status = sentryStatus(data.Status.Code)
	span.StartTime = data.StartTime
	span.EndTime = data.EndTime
	span.Data = data.Attributes

	if data.IsSampled() {
		span.Sampled = sentry.SampledTrue
	} else {
		span.Sampled = sentry.SampledFalse
	}
}

func sentryStatus(code int32) sentry.SpanStatus {
	return sentry.SpanStatus(code + 1)
}
