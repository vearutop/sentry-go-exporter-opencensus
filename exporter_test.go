package sentry_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	sen "github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/swaggest/assertjson"
	"github.com/vearutop/sentry-go-exporter-opencensus"
	"go.opencensus.io/trace"
)

func TestNewExporter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/123123/envelope/", r.URL.String())
		assert.Equal(t, http.MethodPost, r.Method)

		b, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)

		lines := bytes.Split(b, []byte("\n"))
		assert.Len(t, lines, 4)

		assertjson.Equal(t, []byte(`{"event_id":"<ignore-diff>","sent_at":"<ignore-diff>"}`), lines[0])
		assertjson.Equal(t, []byte(`{"type":"transaction","length":"<ignore-diff>"}`), lines[1])
		assertjson.Equal(t, []byte(`{
  "contexts": {
    "device": "<ignore-diff>",
    "os": "<ignore-diff>",
    "runtime": "<ignore-diff>",
    "trace": {
      "trace_id": "<ignore-diff>",
      "span_id": "<ignore-diff>",
      "op": "s1",
      "status": "ok"
    }
  },
  "event_id": "<ignore-diff>",
  "level": "info",
  "platform": "go",
  "sdk": "<ignore-diff>",
  "server_name": "<ignore-diff>",
  "transaction": "s1",
  "user": {},
  "type": "transaction",
  "spans": [
    {
      "trace_id": "<ignore-diff>",
      "span_id": "<ignore-diff>",
      "op": "s2",
      "status": "ok",
      "start_timestamp": "<ignore-diff>",
      "timestamp": "<ignore-diff>",
      "parent_span_id": "<ignore-diff>"
    }
  ],
  "start_timestamp": "<ignore-diff>",
  "timestamp": "<ignore-diff>"
}`), lines[2])
	}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	require.NoError(t, err)

	u.User = url.UserPassword("foo", "")
	u.Path = "123123"

	assert.NoError(t, sen.Init(sen.ClientOptions{
		Dsn: u.String(),
	}))

	defer func() {
		sen.Flush(time.Second)
	}()

	e := sentry.NewExporter()
	trace.RegisterExporter(e)

	defer func() { trace.UnregisterExporter(e) }()
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	ctx, s1 := trace.StartSpan(context.Background(), "s1")

	time.Sleep(10 * time.Millisecond)

	_, s2 := trace.StartSpan(ctx, "s2")

	time.Sleep(10 * time.Millisecond)

	s2.End()

	time.Sleep(10 * time.Millisecond)

	s1.End()
}

func BenchmarkExporter_ExportSpan(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()

	u, err := url.Parse(srv.URL)
	require.NoError(b, err)

	u.User = url.UserPassword("foo", "")
	u.Path = "123123"

	e := sentry.NewExporter()
	trace.RegisterExporter(e)

	defer func() { trace.UnregisterExporter(e) }()
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.AlwaysSample(),
	})

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i <= b.N; i++ {
		ctx, s1 := trace.StartSpan(context.Background(), "s1")
		_, s2 := trace.StartSpan(ctx, "s2")
		s2.End()
		s1.End()
	}
}
