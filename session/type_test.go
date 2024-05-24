//go:build integration

package session_test

import (
	"github.com/google/uuid"
	"jungle/server"
	"jungle/session"
	"jungle/session/cookie"
	"jungle/session/memory"
	"log"
	"net/http"
	"testing"
	"time"
)

// go test --tags=integration -v session/*.go -run TestSession
func TestSession(t *testing.T) {

	s := memory.NewStore(time.Minute, time.Second)
	p := cookie.NewPropagator("sess_id", "http://localhost:8081")
	var m = session.NewManager(p, s, "session", func(ctx *server.Context) string {
		uid, _ := uuid.NewV7()
		return uid.String()
	})

	serv := server.New(":8081")
	serv.Use(func(ctx *server.Context) {
		if ctx.Req.URL.Path == "/login" {
			ctx.Next()
			return
		}
		sess, err := m.GetSession(ctx)
		if err != nil {
			ctx.AbortJSON(http.StatusUnauthorized, map[string]any{
				"code": -1,
				"msg":  "unauthorized",
			})
			return
		}

		// 刷新session
		err = m.RefreshSession(ctx)
		if err != nil {
			log.Printf("refresh session id:[%s] failed: %+v\n", sess.ID(), err)
		}
		ctx.Next()
	})

	serv.Post("/login", func(ctx *server.Context) {
		sess, err := m.GenerateSession(ctx)
		if err != nil {
			ctx.AbortJSON(http.StatusInternalServerError, map[string]any{
				"code": -1,
				"msg":  "login failed",
			})
			return
		}
		err = sess.Set(ctx.Req.Context(), "nickname", "ZhangSan")
		if err != nil {
			ctx.AbortJSON(http.StatusInternalServerError, map[string]any{
				"code": -1,
				"msg":  "login failed",
			})
			return
		}
		ctx.JSON(http.StatusOK, map[string]any{
			"code": http.StatusOK,
			"msg":  "success",
		})
	})

	serv.Post("/logout", func(ctx *server.Context) {
		err := m.RemoveSession(ctx)
		if err != nil {
			ctx.AbortJSON(http.StatusInternalServerError, map[string]any{
				"code": -1,
				"msg":  "logout failed",
			})
			return
		}
		ctx.JSON(http.StatusOK, map[string]any{
			"code": http.StatusOK,
			"msg":  "success",
		})
	})

	serv.Get("/user", func(ctx *server.Context) {
		sess, err := m.GetSession(ctx)
		if err != nil {
			ctx.AbortJSON(http.StatusUnauthorized, map[string]any{
				"code": -1,
				"msg":  "unauthorized",
			})
			return
		}
		nickname, err := sess.Get(ctx.Req.Context(), "nickname")
		if err != nil {
			ctx.AbortJSON(http.StatusUnauthorized, map[string]any{
				"code": -1,
				"msg":  "unauthorized",
			})
			return
		}
		ctx.JSON(http.StatusOK, map[string]any{
			"nickname": nickname,
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
