/**
 * Created by angelina on 2017/8/24.
 */

package easyweb

import (
	"net/http"
	"strings"
	"path"
)

type IRouter interface {
	IRoutes
	Group(string, ...HandlerFunc) *RouterGroup
}

type IRoutes interface {
	Use(...HandlerFunc) IRoutes

	Any(string, HandlerFunc, ...HandlerFunc) IRoutes
	GET(string, HandlerFunc, ...HandlerFunc) IRoutes
	POST(string, HandlerFunc, ...HandlerFunc) IRoutes
	DELETE(string, HandlerFunc, ...HandlerFunc) IRoutes
	PATCH(string, HandlerFunc, ...HandlerFunc) IRoutes
	PUT(string, HandlerFunc, ...HandlerFunc) IRoutes
	OPTIONS(string, HandlerFunc, ...HandlerFunc) IRoutes
	HEAD(string, HandlerFunc, ...HandlerFunc) IRoutes

	StaticFile(string, string) IRoutes
	Static(string, string) IRoutes
	StaticFS(string, http.FileSystem) IRoutes
}

// RouterGroup is used internally to configure router, a RouterGroup is associated with a prefix
// and an array of handlers (middleware).
// RouterGroup用来在内部配置路由，通常与一个前缀以及一组handlers(中间件)相关
type RouterGroup struct {
	Handlers HandlersChain
	basePath string
	easyWeb  *EasyWeb
	root     bool
}

// 通常用来确保某个struct实现了某个interface{}的全部方法
var _ IRouter = &RouterGroup{}

// Use adds middleware to the group, see example code in github.
// 添加中间件
func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}

// Group creates a new router group. You should add all the routes that have common middlwares or the same path prefix.
// For example, all the routes that use a common middlware for authorization could be grouped.
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		easyWeb:  group.easyWeb,
	}
}

func (group *RouterGroup) BasePath() string {
	return group.basePath
}

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.easyWeb.addRoute(httpMethod, absolutePath, handlers)
	return group.returnObj()
}

// 第二个参数为主要路由，第三个参数为中间件
func (group *RouterGroup) POST(relativePath string, mainHandler HandlerFunc, middleware ...HandlerFunc) IRoutes {
	middleware = append(middleware, mainHandler)
	return group.handle("POST", relativePath, middleware)
}

func (group *RouterGroup) GET(relativePath string, mainHandler HandlerFunc, middleware ...HandlerFunc) IRoutes {
	middleware = append(middleware, mainHandler)
	return group.handle("GET", relativePath, middleware)
}

func (group *RouterGroup) DELETE(relativePath string, mainHandler HandlerFunc, middleware ...HandlerFunc) IRoutes {
	middleware = append(middleware, mainHandler)
	return group.handle("DELETE", relativePath, middleware)
}

func (group *RouterGroup) PATCH(relativePath string, mainHandler HandlerFunc, middleware ...HandlerFunc) IRoutes {
	middleware = append(middleware, mainHandler)
	return group.handle("PATCH", relativePath, middleware)
}

func (group *RouterGroup) PUT(relativePath string, mainHandler HandlerFunc, middleware ...HandlerFunc) IRoutes {
	middleware = append(middleware, mainHandler)
	return group.handle("PUT", relativePath, middleware)
}

func (group *RouterGroup) OPTIONS(relativePath string, mainHandler HandlerFunc, middleware ...HandlerFunc) IRoutes {
	middleware = append(middleware, mainHandler)
	return group.handle("OPTIONS", relativePath, middleware)
}

func (group *RouterGroup) HEAD(relativePath string, mainHandler HandlerFunc, middleware ...HandlerFunc) IRoutes {
	middleware = append(middleware, mainHandler)
	return group.handle("HEAD", relativePath, middleware)
}

func (group *RouterGroup) Any(relativePath string, mainHandler HandlerFunc, middleware ...HandlerFunc) IRoutes {
	middleware = append(middleware, mainHandler)
	group.handle("GET", relativePath, middleware)
	group.handle("POST", relativePath, middleware)
	group.handle("PUT", relativePath, middleware)
	group.handle("PATCH", relativePath, middleware)
	group.handle("HEAD", relativePath, middleware)
	group.handle("OPTIONS", relativePath, middleware)
	group.handle("DELETE", relativePath, middleware)
	group.handle("CONNECT", relativePath, middleware)
	group.handle("TRACE", relativePath, middleware)
	return group.returnObj()
}

// StaticFile registers a single route in order to server a single file of the local filesystem.
// router.StaticFile("favicon.ico", "./resources/favicon.ico")
func (group *RouterGroup) StaticFile(relativePath, filepath string) IRoutes {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static file")
	}
	handler := func(c *Context) {
		c.File(filepath)
	}
	group.GET(relativePath, handler)
	group.HEAD(relativePath, handler)
	return group.returnObj()
}

// Static serves files from the given file system root.
// Internally a http.FileServer is used, therefore http.NotFound is used instead
// of the Router's NotFound handler.
// To use the operating system's file system implementation,
// use :
//     router.Static("/static", "/var/www")
func (group *RouterGroup) Static(relativePath, root string) IRoutes {
	return group.StaticFS(relativePath, Dir(root, false))
}

// StaticFS works just like `Static()` but a custom `http.FileSystem` can be used instead.
// Gin by default user: gin.Dir()
func (group *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) IRoutes {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	handler := group.createStaticHandler(relativePath, fs)
	urlPattern := path.Join(relativePath, "/*filepath")

	// Register GET and HEAD handlers
	group.GET(urlPattern, handler)
	group.HEAD(urlPattern, handler)
	return group.returnObj()
}

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutePath := group.calculateAbsolutePath(relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs))
	_, nolist := fs.(*onlyfilesFS)
	return func(c *Context) {
		if nolist {
			c.response.WriteHeader(404)
		}
		fileServer.ServeHTTP(c.response, c.request)
	}
}

func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(group.Handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

func (group *RouterGroup) returnObj() IRoutes {
	if group.root {
		return group.easyWeb
	}
	return group
}
