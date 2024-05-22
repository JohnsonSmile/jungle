package server

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HandleFunc func(ctx *Context)

type Server interface {
	http.Handler
	Start() error
	ShutDown() error
	// AddRoute 添加路由
	//
	// - method 方法
	// - path 路由
	// - handlers 路由业务回调
	AddRoute(method string, path string, handler HandleFunc, middlewares ...HandleFunc)

	// Use 添加中间件
	//
	// handlers 中间件
	Use(middlewares ...HandleFunc)
}

type HTTPServer struct {
	srv             *http.Server
	addr            string
	shutDownTimeout time.Duration
	router          *router
	middlewares     []HandleFunc
}

func New(addr string) *HTTPServer {
	return &HTTPServer{
		addr:            addr,
		shutDownTimeout: time.Second * 15,
		router:          newRouter(),
		middlewares:     make([]HandleFunc, 0),
	}
}

// AddRoute 添加路由
func (serv *HTTPServer) AddRoute(method string, path string, handler HandleFunc, middlewares ...HandleFunc) {
	serv.router.AddRoute(method, path, serv.middlewares, handler, middlewares...)
}

// ServeHTTP implements Server.
func (serv *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &Context{
		Req:  r,
		Resp: w,
	}
	// 查找路由,并实现命中的路由
	serv.serve(ctx)
}

// TODO:
func (h *HTTPServer) serve(ctx *Context) {
	matchInfo, ok := h.router.FindRoute(ctx.Req.Method, ctx.Req.URL.Path)
	if !ok {
		// 路由没有找到
		ctx.Resp.WriteHeader(http.StatusNotFound)
		_, _ = ctx.Resp.Write([]byte("NOT FOUND"))
		return
	}
	ctx.HandlerChain = matchInfo.node.handlerChains
	ctx.PathParams = matchInfo.pathParams
	ctx.MatchedPath = matchInfo.node.matchedPath
	// ctx.FuncName = matchInfo.FuncName
	// 执行 handler chain
	for ctx.Index < len(ctx.HandlerChain) {
		ctx.HandlerChain[ctx.Index](ctx)
		ctx.Index++
	}
}

// Start implements Server.
func (h *HTTPServer) Start() error {

	l, err := net.Listen("tcp", h.addr)
	if err != nil {
		return err
	}
	startCh := make(chan struct{}, 1)
	go func() {
		h.srv = &http.Server{
			Addr:    h.addr,
			Handler: h,
		}
		startCh <- struct{}{}
		if err = h.srv.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()
	<-startCh
	log.Printf("server already started at address: [%s]\n", h.addr)
	return nil
}

// ShutDown implements Server.
func (h *HTTPServer) ShutDown() error {
	// 优雅退出
	// kill -9 是捕捉不到的
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), h.shutDownTimeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			break
		default:
			return h.srv.Shutdown(ctx)
		}
	}

}

// Use implements Server.
func (serv *HTTPServer) Use(middlewares ...HandleFunc) {
	serv.middlewares = append(serv.middlewares, middlewares...)
}

var _ Server = (*HTTPServer)(nil)
