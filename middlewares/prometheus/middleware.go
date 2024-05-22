package prometheus

import (
	"jungle/server"
	"net/http"
	"strconv"
	"time"

	prome "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Middleware struct {
	CounterName   string
	HistogramName string
	SummaryName   string
	Subsystem     string
	Namespace     string
	CounterHelp   string
	HistogramHelp string
	SummaryHelp   string
	Server        string
	Env           string
	App           string
}

func (m *Middleware) Build() func(ctx *server.Context) {

	// counter 可以统计pv - gauge 可以统计uv
	counterVec := prome.NewCounterVec(prome.CounterOpts{
		Name:      m.CounterName,
		Subsystem: m.Subsystem,
		Namespace: m.Namespace,
		Help:      m.CounterHelp,
		ConstLabels: prome.Labels{
			"server": m.Server,
			"env":    m.Env,
			"app":    m.App,
		},
	}, []string{
		"pattern",
		"method",
		"status",
	})

	prome.MustRegister(counterVec)

	histogramVec := prome.NewHistogramVec(prome.HistogramOpts{
		Name:      m.HistogramName,
		Subsystem: m.Subsystem,
		Namespace: m.Namespace,
		Help:      m.HistogramHelp,
		ConstLabels: map[string]string{
			"server": m.Server,
			"env":    m.Env,
			"app":    m.App,
		},
		Buckets: []float64{ // 分桶
			10,
			50,
			100,
			200,
			500,
			1000,
			10000,
		},
		// NativeHistogramBucketFactor:     0,
		// NativeHistogramZeroThreshold:    0,
		// NativeHistogramMaxBucketNumber:  0,
		// NativeHistogramMinResetDuration: 0,
		// NativeHistogramMaxZeroThreshold: 0,
	}, []string{
		"pattern",
		"method",
		"status",
	})
	prome.MustRegister(histogramVec)

	summaryVec := prome.NewSummaryVec(prome.SummaryOpts{
		Name:      m.SummaryName,
		Subsystem: m.Subsystem,
		Namespace: m.Namespace,
		Help:      m.SummaryHelp,
		ConstLabels: prome.Labels{
			"server": m.Server,
			"env":    m.Env,
			"app":    m.App,
		},
		Objectives: map[float64]float64{
			0.5:   0.01,
			0.75:  0.01,
			0.90:  0.005,
			0.98:  0.002,
			0.99:  0.001,
			0.999: 0.0001, // 百分比:误差,误差越小,对prometheus服务器性能要求越高.
		},
		MaxAge: prome.DefMaxAge, // 10 * time.Minute 默认10分钟
	}, []string{
		"pattern",
		"method",
		"status",
	})
	prome.MustRegister(summaryVec)

	return func(ctx *server.Context) {

		startTime := time.Now()
		defer func() {
			// TODO: 并发执行,不要影响正常业务的SLA.
			pattern := ctx.MatchedPath
			if pattern == "" {
				pattern = "unknown"
			}
			elapsed := float64(time.Since(startTime).Milliseconds())

			var statusCode string
			status, exists := ctx.Get("status")
			if sts, ok := status.(int); exists && ok {
				statusCode = strconv.Itoa(sts)
			} else {
				statusCode = "unknown"
			}

			// counterVec
			counterVec.WithLabelValues(
				pattern,
				ctx.Req.Method,
				statusCode,
			).Add(1)

			// histogramVec
			histogramVec.WithLabelValues(pattern,
				ctx.Req.Method,
				statusCode,
			).Observe(elapsed)

			// summaryVec
			summaryVec.WithLabelValues(
				pattern,
				ctx.Req.Method,
				statusCode,
			).Observe(elapsed)
		}()
		ctx.Next()
	}
}

func (m *Middleware) Start(addr string) {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			panic(err)
		}
	}()
}
