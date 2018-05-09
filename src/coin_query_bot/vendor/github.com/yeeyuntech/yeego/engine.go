/**
 * Created by angelina on 2017/4/18.
 */

package yeego

import (
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
)

const (
	// 返回格式
	RET_JSON = 1
	RET_XML  = 2
)

var (
	// 全局的echo实例
	Echo       *echo.Echo
	ReturnType int = RET_JSON
)

// NewEcho
// 初始化全局的echo实例
func NewEcho() *echo.Echo {
	Echo = echo.New()
	return Echo
}

// NewReqAndRes
// 根据echo.Context初始化request&response
func NewReqAndRes(c echo.Context) (req *Request, res *Response) {
	req = NewRequest(c)
	res = NewResponse(c, req)
	req.sessionInit()
	return
}

// StaticFiles
// 设置静态目录
// Static("/static", "assets")
// 请求：/static/js/main.js
// 拿到文件:/assets/js/main.js
func StaticFiles(prefix, root string) {
	Echo.Static(prefix, root)
}

// SetRetType
// 设置返回格式
func SetRetType(i int) {
	if !(i == RET_JSON || i == RET_XML) {
		panic("RetType 类型错误")
	}
	ReturnType = i
}

// Recover
// 打印请求异常信息
func Recover() {
	Echo.Use(middleware.Recover())
}

// Debug
// 是否开启debug
func Debug(b bool) {
	Echo.Debug = b
}

// Logger
// 打印请求信息
func Logger() {
	Echo.Use(middleware.Logger())
}

// Gzip
// 开启gzip压缩
func Gzip() {
	Echo.Use(middleware.Gzip())
}

// BodyLimit
// 设置Body大小 eg:10M
func BodyLimit(str string) {
	Echo.Use(middleware.BodyLimit(str))
}

// AddTrailingSlash
// 自动添加末尾斜杠
func AddTrailingSlash() {
	Echo.Use(middleware.AddTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))
}

// RemoveTrailingSlash
// 自动删除末尾斜杠
func RemoveTrailingSlash() {
	Echo.Use(middleware.RemoveTrailingSlashWithConfig(middleware.TrailingSlashConfig{
		RedirectCode: http.StatusMovedPermanently,
	}))
}

// CORS
// 跨域访问设置
// DefaultCORSConfig = CORSConfig{
// Skipper:      DefaultSkipper,
// AllowOrigins: []string{"*"},
// AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
// }
func CORS(conf middleware.CORSConfig) {
	Echo.Use(middleware.CORSWithConfig(conf))
}
