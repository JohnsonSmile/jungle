//go:build integration

package recovery

import (
	"jungle/server"
	"log"
	"net/http"
	"runtime/debug"
	"testing"
)

// go test --tags=integration -v middlewares/recovery/*.go -run TestMiddleware_recovery
func TestMiddleware_recovery(t *testing.T) {
	recoveryMiddleware := &Middleware{
		Code: http.StatusInternalServerError,
		Data: []byte("Something goes wrong"),
		Log: func(ctx *server.Context, err any) {
			log.Println(err)
			log.Println(string(debug.Stack()))
		},
	}
	serv := server.New(":8081")
	serv.Use(recoveryMiddleware.Build())

	serv.AddRoute(http.MethodGet, "/user", func(ctx *server.Context) {
		n := 0
		m := 1 / n
		log.Println(m)
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
