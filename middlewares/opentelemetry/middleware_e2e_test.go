//go:build integration

package opentelemetry

import (
	"jungle/middlewares/opentelemetry/zipkin"
	"jungle/server"
	"net/http"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
)

// go test --tags=integration -v middlewares/opentelemetry/*.go -run TestMiddleware_Opentelemetry
func TestMiddleware_Opentelemetry(t *testing.T) {
	tracer := otel.GetTracerProvider().Tracer(instrumentationName)
	tracerMiddleware := New(tracer)
	serv := server.New(":8081")
	serv.Use(tracerMiddleware.Build())

	serv.AddRoute(http.MethodGet, "/user", func(ctx *server.Context) {

		func() {
			_, span1 := tracer.Start(ctx.Req.Context(), "call goods inventory")
			defer span1.End()
			time.Sleep(time.Second)
		}()

		func() {
			_, span2 := tracer.Start(ctx.Req.Context(), "call goods info")
			defer span2.End()
			time.Sleep(time.Second)
		}()

		func() {
			_, span3 := tracer.Start(ctx.Req.Context(), "call user info")
			defer span3.End()
			time.Sleep(time.Second)
		}()

		ctx.JSON(http.StatusOK, map[string]any{
			"name": "zhangsan",
			"age":  18,
		})

	})
	collector := zipkin.New("http://localhost:9411/api/v2/spans", "e2e_test")
	collector.Start()
	err := serv.Start()
	if err != nil {
		t.Fatal(err)
	}
	err = serv.ShutDown()
	if err != nil {
		t.Fatal(err)
	}
	collector.Shutdown()
}
