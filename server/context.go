package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const MAX_INDEX = math.MaxInt - 10

type Context struct {
	Resp         http.ResponseWriter
	Req          *http.Request
	MatchedPath  string
	PathParams   url.Values
	QueryParams  url.Values
	HeaderParams url.Values

	HandlerChain []HandleFunc
	Index        int // 当前执行 handler chain 的下标.

	mux    sync.RWMutex
	values map[string]any // 存储值

	tplEngine TemplateEngine
}

func NewContext(req *http.Request, resp http.ResponseWriter, tplEngine TemplateEngine) *Context {
	return &Context{
		Req:          req,
		Resp:         resp,
		values:       make(map[string]any),
		HandlerChain: make([]HandleFunc, 0),
		PathParams:   make(url.Values),
		QueryParams:  make(url.Values),
		HeaderParams: make(url.Values),
		tplEngine:    tplEngine,
	}
}

// req

func (ctx *Context) BindJSON(val any) error {
	if val == nil {
		return errors.New("val is nil")
	}
	if ctx.Req.Body == nil {
		return errors.New("body is nil")
	}
	decoder := json.NewDecoder(ctx.Req.Body)
	return decoder.Decode(val)
}

func (ctx *Context) BindForm(val any) error {
	err := ctx.Req.ParseForm()
	if err != nil {
		return err
	}
	var data map[string]any = map[string]any{}
	for k, v := range ctx.Req.Form {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(dataBytes, val)
}

func (ctx *Context) BindQuery(val any) error {
	if ctx.QueryParams == nil {
		ctx.QueryParams = ctx.Req.URL.Query()
	}
	var data map[string]any = map[string]any{}
	for k, v := range ctx.QueryParams {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(dataBytes, val)
}

func (ctx *Context) BindHeader(val any) error {
	if ctx.HeaderParams == nil {
		ctx.HeaderParams = url.Values(ctx.Req.Header)
	}
	var data map[string]any = map[string]any{}
	for k, v := range ctx.HeaderParams {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(dataBytes, val)
}

func (ctx *Context) FormValue(key string) (result *Result) {
	err := ctx.Req.ParseForm()
	if err != nil {
		return &Result{
			err: err,
		}
	}
	return &Result{
		val: ctx.Req.FormValue(key),
	}
}

func (ctx *Context) QueryValue(key string) (result *Result) {
	if ctx.QueryParams == nil {
		ctx.QueryParams = ctx.Req.URL.Query()
	}
	if ctx.QueryParams.Has(key) {
		return &Result{
			val: ctx.QueryParams.Get(key),
		}
	}
	return &Result{
		err: fmt.Errorf("key: %s not exists", key),
	}
}

func (ctx *Context) PathValue(key string) (result *Result) {
	if ctx.PathParams.Has(key) {
		return &Result{
			val: ctx.PathParams.Get(key),
		}
	}
	return &Result{
		err: fmt.Errorf("key: %s not exists", key),
	}
}

func (ctx *Context) HeaderValue(key string) (result *Result) {
	if ctx.HeaderParams == nil {
		ctx.HeaderParams = url.Values(ctx.Req.Header)
	}
	// header 会转换成大写开头
	key = textproto.CanonicalMIMEHeaderKey(key)
	if ctx.HeaderParams.Has(key) {
		return &Result{
			val: ctx.HeaderParams.Get(key),
		}
	}
	return &Result{
		err: fmt.Errorf("key: %s not exists", key),
	}
}

type Result struct {
	val string
	err error
}

func (result *Result) Int() (val int, err error) {
	if result.err != nil {
		return 0, err
	}
	return strconv.Atoi(result.val)
}

func (result *Result) Int64() (val int64, err error) {
	if result.err != nil {
		return 0, err
	}
	return strconv.ParseInt(result.val, 10, 64)
}

func (result *Result) UInt64() (val uint64, err error) {
	if result.err != nil {
		return 0, err
	}
	return strconv.ParseUint(result.val, 10, 64)
}

func (result *Result) Float64() (val float64, err error) {
	if result.err != nil {
		return 0, err
	}
	return strconv.ParseFloat(result.val, 64)
}

func (result *Result) Bool() (val bool, err error) {
	if result.err != nil {
		return false, err
	}
	return strconv.ParseBool(result.val)
}

func (result *Result) Time(layout string) (val time.Time, err error) {
	if result.err != nil {
		return time.Time{}, err
	}
	return time.Parse(layout, result.val)
}

func (result *Result) TimeInLocation(layout string, loc *time.Location) (val time.Time, err error) {
	if result.err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation(layout, result.val, loc)
}

func (result *Result) TimeFromUnix() (val time.Time, err error) {
	if result.err != nil {
		return time.Time{}, err
	}
	secs, err := result.Int64()
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(secs, 0), nil
}

func (result *Result) TimeFromUnixMilli() (val time.Time, err error) {
	if result.err != nil {
		return time.Time{}, err
	}
	msecs, err := result.Int64()
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(msecs), nil
}

func (result *Result) TimeFromUnixMicro() (val time.Time, err error) {
	if result.err != nil {
		return time.Time{}, err
	}
	usecs, err := result.Int64()
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMicro(usecs), nil
}

// response
func (ctx *Context) JSON(status int, val any) {
	data, err := json.Marshal(val)
	if err != nil {
		panic(err)
	}

	// status
	ctx.Resp.WriteHeader(status)

	// 用来做trace的时候用的
	ctx.Set("status", status)
	ctx.Set("data", string(data))

	// header
	contentLength := len(data)
	ctx.Resp.Header().Set("Content-Type", "application/json")
	ctx.Resp.Header().Set("Content-Length", strconv.Itoa(contentLength))

	// body
	writenLength, err := ctx.Resp.Write(data)
	if err != nil {
		panic(err)
	}
	if writenLength != contentLength {
		panic("data not write complete..")
	}
}

func (ctx *Context) WriteString(code int, msg []byte) {
	ctx.Resp.WriteHeader(code)
	_, _ = ctx.Resp.Write(msg)
}

func (ctx *Context) Next() {
	ctx.Index++
	for ctx.Index < len(ctx.HandlerChain) {
		ctx.HandlerChain[ctx.Index](ctx)
		ctx.Index++
	}
}

func (ctx *Context) Abort() {
	ctx.Index = MAX_INDEX
}

func (ctx *Context) AbortWithStatus(status int) {
	ctx.Index = MAX_INDEX
	ctx.Resp.WriteHeader(status)

	// 用来做trace的时候用的
	ctx.Set("status", status)
}

func (ctx *Context) AbortJSON(status int, val any) {

	// 避免++溢出的时候成负数了
	ctx.Index = MAX_INDEX
	ctx.JSON(status, val)
}

// render

func (ctx *Context) Render(tplName string, data any) {
	dataBytes, err := ctx.tplEngine.Render(ctx.Req.Context(), tplName, data)
	if err != nil {
		panic(err)
	}
	ctx.WriteString(http.StatusOK, dataBytes)
}

// cookie

func (ctx *Context) SetCookie(ck *http.Cookie) {
	http.SetCookie(ctx.Resp, ck)
}

// values

func (c *Context) Get(key string) (val any, exists bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	val, exists = c.values[key]
	return
}

func (c *Context) Set(key string, val any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.values[key] = val
}
