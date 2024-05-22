package accesslog

import (
	"fmt"
	"jungle/server"
	"log"
)

type AccessLogBuilder struct {
	logger func(str string)
}

func New(logger func(str string)) *AccessLogBuilder {
	return &AccessLogBuilder{
		logger: logger,
	}
}

func (builder *AccessLogBuilder) Build() func(ctx *server.Context) {
	if builder.logger == nil {
		builder.logger = func(str string) {
			log.Println(str)
		}
	}
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
			builder.logger("[" + l.Method + "]\t" + l.Host + "\t" + l.Route + "\t" + l.Path + "\t" + l.Msg)
		}(ctx)
		ctx.Next()
	}
}

type accessLog struct {
	Method string `json:"method"`
	Host   string `json:"host"`
	Route  string `json:"route"`
	Path   string `json:"path"`
	Msg    string `json:"msg"`
}
