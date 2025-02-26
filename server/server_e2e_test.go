//go:build integration

package server

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// go test --tags=integration -v server/*.go -run TestServer
func TestServer(t *testing.T) {

	var h = New(":8081")
	// err := http.ListenAndServe(":8081", h)
	// require.NoError(t, err)

	h.Get("/user", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello world"))
	})

	h.Use(func(ctx *Context) {
		now := time.Now()
		log.Println("start exec...")
		defer func(ctx *Context) {
			elapsed := time.Since(now)
			log.Printf("exec elapesd: %+v\n", elapsed)
		}(ctx)
		ctx.Next()
	}, func(ctx *Context) {
		log.Println("===========1")
		ctx.Next()
		log.Println("===========11")
	}, func(ctx *Context) {
		log.Println("===========2")
		ctx.AbortJSON(http.StatusBadRequest, map[string]any{
			"name": "unknown",
		})
		log.Println("===========22")
	})

	// h.Get("user/*", func(ctx *Context) {
	// 	vals := ctx.Req.URL.Query()
	// 	name := vals.Get("name")
	// 	if name == "" {
	// 		ctx.Resp.Write([]byte("param error"))
	// 		return
	// 	}
	// 	ctx.Resp.Write([]byte(name))
	// })

	h.Put("user/:id", func(ctx *Context) {

		result := ctx.HeaderValue("x-name")

		var decodedMap map[string]any
		err := ctx.BindJSON(&decodedMap)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, map[string]any{
				"msg": "body is empty",
				"err": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusOK, map[string]any{
			"decodedMap": decodedMap,
			"name":       result.val,
		})
	})
	h.Get("user/:id/:id/hello/:address", func(ctx *Context) {
		ids := ctx.PathParams["id"]
		addresses := ctx.PathParams["address"]
		ctx.Resp.Write([]byte(addresses[0] + "-" + ids[0] + "-" + ids[1]))
	})

	err := h.Start()
	require.NoError(t, err)

	err = h.ShutDown()
	require.NoError(t, err)
}
