package zipkin

import (
	"context"
	"log"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

type Collector struct {
	traceProvider *trace.TracerProvider
}

func New(url string, serviceName string) *Collector {
	exp, err := zipkin.New(
		url,
		zipkin.WithLogger(log.New(os.Stderr, serviceName, log.Ldate|log.Ltime|log.Llongfile)),
	)
	if err != nil {
		panic(err)
	}

	batcher := trace.NewBatchSpanProcessor(exp)
	tracerProvider := trace.NewTracerProvider(
		trace.WithSpanProcessor(batcher),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	return &Collector{
		traceProvider: tracerProvider,
	}
}

func (c *Collector) Start() {
	otel.SetTracerProvider(c.traceProvider)
}

func (c *Collector) Shutdown() {
	if err := c.traceProvider.Shutdown(context.Background()); err != nil {
		log.Printf("exporter shutdown failed: %+v\n", err)
	}
}
