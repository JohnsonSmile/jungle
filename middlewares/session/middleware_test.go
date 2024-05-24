package session_test

import (
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	sessionmiddleware "jungle/middlewares/session"
	"jungle/server"
	"jungle/session"
	"jungle/session/cookie"
	"jungle/session/memory"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

// go test -v middlewares/session/*.go -run TestPathRegexp
func TestPathRegexp(t *testing.T) {
	starPath := "/user/*"
	starPath = strings.ReplaceAll(starPath, "*", "(.*?)")
	reg := regexp.MustCompile(starPath)
	matched := reg.MatchString("/user/id")
	require.Equal(t, matched, true)
	matched = reg.MatchString("/user/friends")
	require.Equal(t, matched, true)
	matched = reg.MatchString("/user/settings/privacy")
	require.Equal(t, matched, true)
	matched = reg.MatchString("/user1/settings/privacy")
	require.Equal(t, matched, false)

	pathPath := "/user/:address/fav/:id"
	segReg := regexp.MustCompile("(:[^/]*)")
	pathPath = segReg.ReplaceAllString(pathPath, "(.*?)")
	pathReg := regexp.MustCompile(pathPath)
	matched = pathReg.MatchString("/user/xxxx/fav/123")
	require.Equal(t, matched, true)
	matched = pathReg.MatchString("/user/settings/fav")
	require.Equal(t, matched, false)
}

// go test -v middlewares/session/*.go -run TestMiddleware_Session
func TestMiddleware_Session(t *testing.T) {

	serv := server.New(":8081")

	// session 中间件
	sessionMiddleware := sessionmiddleware.New(sessionmiddleware.WithSessionManager(func() *session.Manager {
		s := memory.NewStore(time.Minute, time.Second)
		p := cookie.NewPropagator("sess_id", "http://localhost:8081")
		var m = session.NewManager(p, s, "session", func(ctx *server.Context) string {
			uid, _ := uuid.NewV7()
			return uid.String()
		})
		return m
	}))

	serv.Use(sessionMiddleware.Build("/login"))

	serv.Post("/login", func(ctx *server.Context) {
		m := sessionmiddleware.GetManager(ctx)
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
		m := sessionmiddleware.GetManager(ctx)
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
		m := sessionmiddleware.GetManager(ctx)
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
