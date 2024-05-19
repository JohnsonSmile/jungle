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

	})

	err := h.Start()
	require.NoError(t, err)

	err = h.ShutDown()
	require.NoError(t, err)
}
