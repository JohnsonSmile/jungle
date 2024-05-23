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
func (s *HTTPServer) AddRoute(method string, path string, handler HandleFunc, middlewares ...HandleFunc) {
	s.router.AddRoute(method, path, s.middlewares, handler, middlewares...)
}

// ServeHTTP implements Server.
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w)
	// 查找路由,并实现命中的路由
	s.serve(ctx)
}

// TODO:
func (s *HTTPServer) serve(ctx *Context) {
	matchInfo, ok := s.router.FindRoute(ctx.Req.Method, ctx.Req.URL.Path)
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
func (s *HTTPServer) Start() error {

	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	startCh := make(chan struct{}, 1)
	go func() {
		s.srv = &http.Server{
			Addr:    s.addr,
			Handler: s,
		}
		startCh <- struct{}{}
		if err = s.srv.Serve(l); err != nil {
			log.Fatal(err)
		}
	}()
	<-startCh
	log.Printf("server already started at address: [%s]\n", s.addr)
	return nil
}

// ShutDown implements Server.
func (s *HTTPServer) ShutDown() error {
	// 优雅退出
	// kill -9 是捕捉不到的
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), s.shutDownTimeout)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			break
		default:
			return s.srv.Shutdown(ctx)
		}
	}
}

// Use implements Server.
func (s *HTTPServer) Use(middlewares ...HandleFunc) {
	s.middlewares = append(s.middlewares, middlewares...)
}

// Extension methods

func (s *HTTPServer) Get(path string, handler HandleFunc, middlewares ...HandleFunc) {
	s.AddRoute(http.MethodGet, path, handler, middlewares...)
}

func (s *HTTPServer) Head(path string, handler HandleFunc, middlewares ...HandleFunc) {
	s.AddRoute(http.MethodHead, path, handler, middlewares...)
}

func (s *HTTPServer) Post(path string, handler HandleFunc, middlewares ...HandleFunc) {
	s.AddRoute(http.MethodPost, path, handler, middlewares...)
}

func (s *HTTPServer) Put(path string, handler HandleFunc, middlewares ...HandleFunc) {
	s.AddRoute(http.MethodPut, path, handler, middlewares...)
}

func (s *HTTPServer) Patch(path string, handler HandleFunc, middlewares ...HandleFunc) {
	s.AddRoute(http.MethodPatch, path, handler, middlewares...)
}

func (s *HTTPServer) Delete(path string, handler HandleFunc, middlewares ...HandleFunc) {
	s.AddRoute(http.MethodDelete, path, handler, middlewares...)
}

func (s *HTTPServer) Connect(path string, handler HandleFunc, middlewares ...HandleFunc) {
	s.AddRoute(http.MethodConnect, path, handler, middlewares...)
}

func (s *HTTPServer) Options(path string, handler HandleFunc, middlewares ...HandleFunc) {
	s.AddRoute(http.MethodOptions, path, handler, middlewares...)
}

func (s *HTTPServer) Trace(path string, handler HandleFunc, middlewares ...HandleFunc) {
	s.AddRoute(http.MethodTrace, path, handler, middlewares...)
}

var _ Server = (*HTTPServer)(nil)
