package accesslog

import (
	"fmt"
	"jungle/server"
	"net/http"
	"net/http/httptest"
	"testing"
)

// go test -v middlewares/accesslog/*.go -run TestMiddleware_AccessLog
func TestMiddleware_AccessLog(t *testing.T) {

	var srv = server.New(":8081")
	accesslogBuilder := New(func(str string) {
		fmt.Println(str)
	})
	srv.Use(accesslogBuilder.Build())
	srv.Post("/a/b/:id", GetUser)
	resp := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "http://localhost:8081/a/b/1", nil)
	srv.ServeHTTP(resp, req)

}

func GetUser(ctx *server.Context) {
	ctx.JSON(http.StatusOK, map[string]any{
		"name": "zhangsan",
	})
}
