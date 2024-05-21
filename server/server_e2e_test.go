//go:build integration

package server

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// go test --tags=integration -v server/*.go -run TestServer
func TestServer(t *testing.T) {

	var h Server = New(":8081")
	// err := http.ListenAndServe(":8081", h)
	// require.NoError(t, err)

	h.AddRoute(http.MethodGet, "/user", func(ctx *Context) {
		ctx.Resp.Write([]byte("hello world"))
	})

	// h.AddRoute(http.MethodGet, "user/*", func(ctx *Context) {
	// 	vals := ctx.Req.URL.Query()
	// 	name := vals.Get("name")
	// 	if name == "" {
	// 		ctx.Resp.Write([]byte("param error"))
	// 		return
	// 	}
	// 	ctx.Resp.Write([]byte(name))
	// })

	h.AddRoute(http.MethodPut, "user/:id", func(ctx *Context) {

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
	h.AddRoute(http.MethodGet, "user/:id/:id/hello/:address", func(ctx *Context) {
		ids := ctx.PathParams["id"]
		addresses := ctx.PathParams["address"]
		ctx.Resp.Write([]byte(addresses[0] + "-" + ids[0] + "-" + ids[1]))
	})

	err := h.Start()
	require.NoError(t, err)

	err = h.ShutDown()
	require.NoError(t, err)
}
