package jaeger

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
)

type Collector struct {
	traceProvider *trace.TracerProvider
}

func New(url string, serviceName string) *Collector {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		panic(err)
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(exp),
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
