package accesslog

import (
	"fmt"
	"jungle/server"
	"net/http"
	"net/http/httptest"
	"testing"
)

// go test -v middlewares/accesslog/*.go -run TestMiddleware
func TestMiddleware(t *testing.T) {

	var srv server.Server = server.New(":8081")
	accesslogMiddleware := AccessLog(func(str string) {
		fmt.Println(str)
	})
	srv.Use(accesslogMiddleware)
	srv.AddRoute(http.MethodPost, "/a/b/:id", GetUser)
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8081/a/b/1", nil)
	srv.ServeHTTP(resp, req)

}

func GetUser(ctx *server.Context) {
	_ = ctx.JSON(http.StatusOK, map[string]any{
		"name": "zhangsan",
	})
}
