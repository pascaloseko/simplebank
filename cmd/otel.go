package cmd

import (
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/GoogleCloudPlatform/opentelemetry-operations-go/propagator"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// DefaultPooledTransport This function can be used in place of http.DefaultTransport which is the default
// http.Transport when calling otelhttp.NewTransport() and having the first arg as nil
// ref: https://github.com/hashicorp/go-cleanhttp/blob/6d9e2ac5d828e5f8594b97f88c4bde14a67bb6d2/cleanhttp.go#L23-L39
func DefaultPooledTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
}

func InitOtel() (trace.TracerProvider, propagation.TextMapPropagator) {
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.01))),
	)
	prop := propagator.CloudTraceFormatPropagator{}
	return tp, prop
}

// Instrument returns an instrumented http client
func Instrument(base http.RoundTripper, tracerProvider trace.TracerProvider, propagator propagation.TextMapPropagator) http.RoundTripper {
	return otelhttp.NewTransport(base, otelhttp.WithTracerProvider(tracerProvider), otelhttp.WithPropagators(propagator))
}
