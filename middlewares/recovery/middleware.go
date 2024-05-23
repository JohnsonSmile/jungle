package recovery

import "jungle/server"

type Middleware struct {
	Code int
	Data []byte
	Log  func(ctx *server.Context, err any)
}

func (m *Middleware) Build() func(ctx *server.Context) {
	return func(ctx *server.Context) {

		defer func() {
			if err := recover(); err != nil {
				ctx.WriteString(m.Code, m.Data)
				ctx.Abort()
				m.Log(ctx,err)
			}
		}()
		ctx.Next()
	}
}
