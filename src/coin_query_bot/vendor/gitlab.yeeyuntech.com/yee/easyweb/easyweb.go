/**
 * Created by angelina on 2017/8/24.
 */

package easyweb

import (
	"sync"
	"net/http"
	"gitlab.yeeyuntech.com/yee/easyweb/render"
	"log"
	"os/signal"
	"syscall"
	"os"
	"time"
	"net/http/pprof"
)

type (
	// 核心的处理函数
	HandlerFunc func(*Context)

	// 某个请求对应的函数处理链
	HandlersChain []HandlerFunc

	// 定义Map为map[string]interface{}
	Map map[string]interface{}

	// easyweb主体
	EasyWeb struct {
		RouterGroup                         // 路由
		noMethodHandler         HandlerFunc // 方法不存在处理函数
		allNoMethodHandlerChain HandlersChain
		noRouteHandler          HandlerFunc // 路由不存在处理函数
		allNoRouteHandlerChain  HandlersChain
		pool                    sync.Pool   // 缓冲池
		trees                   methodTrees // 路由树
		Server                  *http.Server
		TLSServer               *http.Server
	}

	// 路由信息
	RouteInfo struct {
		Method  string
		Path    string
		Handler string
	}

	// 全部的路由集合
	RoutesInfo []RouteInfo
)

var _ IRouter = &EasyWeb{}

// 获取函数链中最后一个handlerFunc，是要进行操作的处理函数，前面的都是中间件
func (c HandlersChain) Last() HandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

func New() *EasyWeb {
	debugPrintWARNINGNew()
	initSomething()
	easyWeb := &EasyWeb{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		noMethodHandler: defaultNoMethodHandler(),
		noRouteHandler:  defaultNoRouteHandler(),
		trees:           make(methodTrees, 0, 9),
		Server:          new(http.Server),
		TLSServer:       new(http.Server),
	}
	easyWeb.Server.Handler = easyWeb
	easyWeb.TLSServer.Handler = easyWeb
	easyWeb.RouterGroup.easyWeb = easyWeb
	easyWeb.pool.New = func() interface{} {
		return easyWeb.allocateContext()
	}
	// 启动pprof
	if Config.Pprof {
		easyWeb.pprof()
	}
	return easyWeb
}

func initSomething() {
	defaultRecoveryWriter()
	defaultLogger()
	// 启动session
	if err := StartSession(); err != nil {
		log.Printf("[EasyWeb] Start Session Error : %s\n", err.Error())
	}
	// 设置模板文件路径
	render.SetTemplateDir(Config.TemplateDir)
	render.Build()
}

// 获取context
func (easyWeb *EasyWeb) allocateContext() *Context {
	return &Context{easyWeb: easyWeb, request: nil, response: &Response{}}
}

/************************************/
/************路由相关*****************/
/************************************/

// 注册中间件
func (easyWeb *EasyWeb) Use(middleware ...HandlerFunc) IRoutes {
	easyWeb.RouterGroup.Use(middleware...)
	easyWeb.rebuild404Handler()
	easyWeb.rebuild405Handler()
	return easyWeb
}

// 设置默认的404handler
func (easyWeb *EasyWeb) SetDefault404Handler(handler HandlerFunc) {
	easyWeb.noRouteHandler = handler
	easyWeb.rebuild404Handler()
}

// 设置默认的405
func (easyWeb *EasyWeb) SetDefault405Handler(handler HandlerFunc) {
	easyWeb.noMethodHandler = handler
	easyWeb.rebuild405Handler()
}

func (easyWeb *EasyWeb) rebuild404Handler() {
	easyWeb.allNoRouteHandlerChain = easyWeb.combineHandlers(HandlersChain{easyWeb.noRouteHandler})
}

func (easyWeb *EasyWeb) rebuild405Handler() {
	easyWeb.allNoMethodHandlerChain = easyWeb.combineHandlers(HandlersChain{easyWeb.noMethodHandler})
}

func (easyWeb *EasyWeb) addRoute(method, path string, handlers HandlersChain) {
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(len(method) > 0, "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")

	debugPrintRoute(method, path, handlers)
	root := easyWeb.trees.get(method)
	if root == nil {
		root = new(node)
		easyWeb.trees = append(easyWeb.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)
}

// 返回已注册的全部路由信息
func (easyWeb *EasyWeb) Routes() (routes RoutesInfo) {
	for _, tree := range easyWeb.trees {
		routes = iterate("", tree.method, routes, tree.root)
	}
	return routes
}

func iterate(path, method string, routes RoutesInfo, root *node) RoutesInfo {
	path += root.path
	if len(root.handlers) > 0 {
		routes = append(routes, RouteInfo{
			Method:  method,
			Path:    path,
			Handler: nameOfFunction(root.handlers.Last()),
		})
	}
	for _, child := range root.children {
		routes = iterate(path, method, routes, child)
	}
	return routes
}

/************************************/
/************提供一些设置方法***********/
/************************************/

// 添加模板函数
func AddTplFunc(key string, fn interface{}) {
	render.AddFuncMap(key, fn)
}

var beforeQuitFunc func()

// 设置结束前执行的操作，一般为清理内存，只有10s时间
func SetQuitFunc(f func()) {
	beforeQuitFunc = f
}

// 开启pprof
func (easyWeb *EasyWeb) pprof() {
	easyWeb.Any("/debug/pprof/", fromHandlerFunc(pprof.Index).Handle)
	easyWeb.Any("/debug/pprof/heap", fromHTTPHandler(pprof.Handler("heap")).Handle)
	easyWeb.Any("/debug/pprof/goroutine", fromHTTPHandler(pprof.Handler("goroutine")).Handle)
	easyWeb.Any("/debug/pprof/block", fromHTTPHandler(pprof.Handler("block")).Handle)
	easyWeb.Any("/debug/pprof/threadcreate", fromHTTPHandler(pprof.Handler("threadcreate")).Handle)
	easyWeb.Any("/debug/pprof/cmdline", fromHandlerFunc(pprof.Cmdline).Handle)
	easyWeb.Any("/debug/pprof/profile", fromHandlerFunc(pprof.Profile).Handle)
	easyWeb.Any("/debug/pprof/symbol", fromHandlerFunc(pprof.Symbol).Handle)
}

/************************************/
/************处理请求相关**************/
/************************************/

// 启动监听
func (easyWeb *EasyWeb) Run(addr ...string) (err error) {
	defer func() {
		debugPrintError(err)
	}()
	port := resolveAddress(addr...)
	log.Printf("[EasyWeb] Listening and serving HTTP on %s\n", port)
	// 在终止前执行一些清理内存的操作
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGQUIT)
	go func() {
		<-sig
		time.AfterFunc(10*time.Second, func() {
			os.Exit(0)
		})
		if RecoveryLogger != nil {
			RecoveryLogger.Close()
		}
		if Logger != nil {
			Logger.Close()
		}
		if beforeQuitFunc != nil {
			beforeQuitFunc()
		}
		os.Exit(0)
	}()
	err = http.ListenAndServe(port, easyWeb)
	return
}

// 实现ServeHTTP接口
func (easyWeb *EasyWeb) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := easyWeb.pool.Get().(*Context)
	c.response = c.response.reset(w)
	c.request = req
	c.reset()
	easyWeb.handleHTTPRequest(c)
	easyWeb.pool.Put(c)
}

// 处理请求的核心
func (easyWeb *EasyWeb) handleHTTPRequest(c *Context) {
	httpMethod := c.request.Method
	unescape := true
	path := c.request.URL.RawPath
	if path == "" {
		path = c.request.URL.Path
		unescape = false
	}
	// session init
	if enableSession {
		var err error
		c.session, err = globalSessions.SessionStart(c.response, c.request)
		if err != nil {
			c.response.WriteHeader(503)
			c.response.WriteString("Service Session Unavailable")
			c.response.WriteHeaderNow()
			return
		}
		defer func() {
			if c.session != nil {
				c.session.SessionRelease(c.response)
			}
		}()
	}
	// 寻找匹配的handler
	t := easyWeb.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method == httpMethod {
			root := t[i].root
			handlers, params, tsr := root.getValue(path, c.params, unescape)
			if handlers != nil {
				c.handlers = handlers
				c.params = params
				c.Next()
				c.response.WriteHeaderNow()
				return
			}
			if httpMethod != "CONNECT" && path != "/" {
				if tsr {
					redirectTrailingSlash(c)
					return
				}
			}
			break
		}
	}
	// 查找其他和这个route同名但是method不同的handler,405
	for _, tree := range easyWeb.trees {
		if tree.method != httpMethod {
			if handlers, _, _ := tree.root.getValue(path, nil, unescape); handlers != nil {
				c.handlers = easyWeb.allNoMethodHandlerChain
				serveError(c, 405)
				return
			}
		}
	}
	// 处理404
	c.handlers = easyWeb.allNoRouteHandlerChain
	serveError(c, 404)
}

// 默认的405
func defaultNoMethodHandler() HandlerFunc {
	return func(context *Context) {
		context.response.WriteString("405 method not allowed")
	}
}

// 默认的404
func defaultNoRouteHandler() HandlerFunc {
	return func(context *Context) {
		context.response.WriteString("404 not found")
	}
}

// 处理错误code的方法
func serveError(c *Context, code int) {
	c.response.WriteHeader(code)
	c.Next()
	c.response.WriteHeaderNow()
	return
}

// 处理/
func redirectTrailingSlash(c *Context) {
	req := c.request
	path := req.URL.Path
	code := 301
	if req.Method != "GET" {
		code = 307
	}
	if len(path) > 1 && path[len(path)-1] == '/' {
		req.URL.Path = path[:len(path)-1]
	} else {
		req.URL.Path = path + "/"
	}
	debugPrint("redirecting request %d: %s --> %s", code, path, req.URL.String())
	http.Redirect(c.response.writer, req, req.URL.String(), code)
}
