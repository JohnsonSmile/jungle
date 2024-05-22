package accesslog

import (
	"fmt"
	"jungle/server"
)

type accessLog struct {
	Method string `json:"method"`
	Host   string `json:"host"`
	Route  string `json:"route"`
	Path   string `json:"path"`
	Msg    string `json:"msg"`
}

func AccessLog(logger func(str string)) server.HandleFunc {
	return func(ctx *server.Context) {
		defer func(ctx *server.Context) {
			msg := ""
			if r := recover(); r != nil {
				msg = fmt.Sprint(r)
			}
			l := accessLog{
				Method: ctx.Req.Method,
				Host:   ctx.Req.Host,
				Route:  ctx.MatchedPath,
				Path:   ctx.Req.URL.Path,
				Msg:    msg,
			}
			logger("[" + l.Method + "]\t" + l.Host + "\t" + l.Route + "\t" + l.Path + "\t" + l.Msg)
		}(ctx)
		ctx.Next()
	}
}
