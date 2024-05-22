package prometheus

import (
	"jungle/server"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

// go test --tags=integration -v middlewares/prometheus/*.go -run TestMiddleware_Prometheus
func TestMiddleware_Prometheus(t *testing.T) {

	promeMiddleware := &Middleware{
		CounterName:   "test_e2e_count",
		HistogramName: "test_e2e_history",
		SummaryName:   "test_e2e_summary",
		Subsystem:     "web",
		Namespace:     "shop",
		CounterHelp:   "接口访问次数",
		HistogramHelp: "历史访问次数",
		SummaryHelp:   "百分比访问次数",
		Server:        "localhost:8081",
		Env:           "test",
		App:           "shop",
	}

	serv := server.New(":8081")
	serv.Use(promeMiddleware.Build())
	promeMiddleware.Start(":8082")

	serv.AddRoute(http.MethodGet, "/user", func(ctx *server.Context) {
		val := rand.Intn(1000) + 1
		time.Sleep(time.Duration(val) * time.Millisecond)
		ctx.JSON(http.StatusOK, map[string]any{
			"name": "zhangsan",
			"age":  19,
		})
	})
	err := serv.Start()
	if err != nil {
		t.Fatal(err)
	}
	err = serv.ShutDown()
	if err != nil {
		t.Fatal(err)
	}
}
