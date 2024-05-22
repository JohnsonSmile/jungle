package opentelemetry

import (
	"jungle/server"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "jungle/middlewares/opentelemetry"

type Middleware struct {
	tracer trace.Tracer
}

func New(tracer trace.Tracer) *Middleware {
	return &Middleware{
		tracer: tracer,
	}
}

func (m *Middleware) Build() func(ctx *server.Context) {
	if m.tracer == nil {
		m.tracer = otel.GetTracerProvider().Tracer(instrumentationName)
	}
	return func(ctx *server.Context) {

		reqCtx := ctx.Req.Context()

		// 尝试从上游获取trace
		reqCtx = otel.GetTextMapPropagator().Extract(reqCtx, propagation.HeaderCarrier(ctx.Req.Header))
		// 先设置spanName为 unknown, 在执行完逻辑之后, 在ctx中会有 MatchedPath, 再修改name
		spanCtx, span := m.tracer.Start(reqCtx, "unknown")
		defer func() {
			span.SetName(ctx.MatchedPath)
			status, exists := ctx.Get("status")
			if statusCode, ok := status.(int); exists && ok {
				span.SetAttributes(attribute.Int("http.status", statusCode))
			}
			data, exists := ctx.Get("data")
			if dataStr, ok := data.(string); exists && ok {
				span.SetAttributes(attribute.String("resp.data", dataStr))
			}
			span.End()
		}()

		span.SetAttributes(attribute.String("http.method", ctx.Req.Method))
		span.SetAttributes(attribute.String("http.url", ctx.Req.URL.String()))
		span.SetAttributes(attribute.String("http.host", ctx.Req.Host))

		// 传递context信息给下游
		ctx.Req = ctx.Req.WithContext(spanCtx)
		ctx.Next()
	}
}
