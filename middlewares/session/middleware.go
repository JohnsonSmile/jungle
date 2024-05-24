package session

import (
	"github.com/google/uuid"
	"jungle/server"
	"jungle/session"
	"jungle/session/cookie"
	"jungle/session/memory"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Middleware struct {
	manager *session.Manager
}

const kSessionManagerKey = "SessionManager"

func New(opts ...Option) *Middleware {
	m := &Middleware{}

	for _, opt := range opts {
		opt(m)
	}

	if m.manager == nil {
		s := memory.NewStore(time.Minute, time.Second)
		p := cookie.NewPropagator("sess_id", "*")
		manager := session.NewManager(p, s, "session", func(ctx *server.Context) string {
			uid, _ := uuid.NewV7()
			return uid.String()
		})
		m.manager = manager
	}

	return m
}

func (m *Middleware) Build(omitPaths ...string) func(ctx *server.Context) {
	omitPathMap := make(map[string]bool)
	omitStarRegs := make([]*regexp.Regexp, 0)
	omitPathRegs := make([]*regexp.Regexp, 0)
	for _, omitPath := range omitPaths {
		if strings.Contains(omitPath, "*") {
			omitPath = strings.ReplaceAll(omitPath, "*", "(.*?)")
			starReg := regexp.MustCompile(omitPath)
			omitStarRegs = append(omitStarRegs, starReg)
		} else if strings.Contains(omitPath, ":") {
			segReg := regexp.MustCompile("(:[^/]*)")
			omitPath = segReg.ReplaceAllString(omitPath, "(.*?)")
			pathReg := regexp.MustCompile(omitPath)
			omitPathRegs = append(omitPathRegs, pathReg)
		} else {
			omitPathMap[omitPath] = true
		}
	}
	return func(ctx *server.Context) {
		ctx.Set("SessionManager", m.manager)
		urlPath := ctx.Req.URL.Path
		// 静态路径
		if omitPathMap[urlPath] {
			ctx.Next()
			return
		}

		// *路径
		for _, starReg := range omitStarRegs {
			if starReg.MatchString(urlPath) {
				ctx.Next()
				return
			}
		}

		// 路径参数路径
		for _, pathReg := range omitPathRegs {
			if pathReg.MatchString(urlPath) {
				ctx.Next()
				return
			}
		}

		sess, err := m.manager.GetSession(ctx)
		if err != nil {
			ctx.AbortJSON(http.StatusUnauthorized, map[string]any{
				"code": -1,
				"msg":  "unauthorized",
			})
			return
		}

		// 刷新session
		err = m.manager.RefreshSession(ctx)
		if err != nil {
			log.Printf("refresh session id:[%s] failed: %+v\n", sess.ID(), err)
		}
		ctx.Next()
	}
}

func GetManager(ctx *server.Context) (manager *session.Manager) {
	val, exists := ctx.Get(kSessionManagerKey)
	if !exists {
		panic("session manager not exists")
	}
	manager, ok := val.(*session.Manager)
	if !ok {
		panic("session manager not exists")
	}
	return manager
}
