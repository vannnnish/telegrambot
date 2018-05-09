[TOC]
# easyweb
web框架

## 快速开始

```go
import(
	"gitlab.yeeyuntech.com/yee/easyweb"
	"gitlab.yeeyuntech.com/yee/easyweb/middleware"
)

web := easyweb.New()
web.Use(middleware.Recovery())
web.Use(middleware.Logger())
web.StaticFile("/staticfile", "view/1.txt")
web.GET("/test", func(context *easyweb.Context) {
	context.String(200, "test")
})
web.GET("/html", func(context *easyweb.Context) {
	context.RenderData("name", "angelina")
	context.Render200("html")
})
err := web.Run()
if err != nil {
	panic(err.Error())
}
```

## 配置

有一些可以配置的选项

```go
// 获取默认的配置选项，然后可以自行更改后再设置
conf := easyweb.DefaultConfig()
// 运行模式 debug/release
conf.Mode = easyweb.DebugMode
// 模板文件路径
conf.TemplateDir = "view"
// 是否开启pprof
conf.Pprof = false
 // 是否开启session
conf.SessionOn = false
// 运行的端口
conf.Port = ":8080"
// 重新设置配置
easyweb.SetConfig(conf)
// 还可以单独调用函数设置一些配置
// 运行模式
easyweb.SetMode(easyweb.ReleaseMode)
// web应用退出前执行的操作(注意，不会响应kill -9信号)
easyweb.SetQuitFunc(func() {
		fmt.Println("结束前执行的一些操作")
})
// 添加模板函数
easyweb.AddTplFunc("tplFun", func() {})

// 上面是一些全局设置，下面有一些初始化后的设置
web := easyweb.New()
// 设置404返回
web.SetDefault404Handler(func(context *easyweb.Context) {
	context.String(404, "404 not found")
})
// 设置405返回
web.SetDefault405Handler(func(context *easyweb.Context) {
	context.String(405, "405 method not allowed")
})
```
注意：全局的配置应该在 ```easyweb.New()```之前就已经调用

## 中间件

中间件函数与处理路由的函数是一致的,均为```HandlerFunc func(*Context)```类型

```go
// 添加全局中间件
web := easyweb.New()
web.Use(middleware.Recovery())
web.Use(middleware.Logger())
// 添加路由中间件参考路由文档
```

可以自己实现一些中间件并调用(参考框架的中间件实现)

## 路由

  路由采用的是gin框架的路由（[详情](https://github.com/gin-gonic/gin)）,自己做了一些特殊处理

```go
web := easyweb.New()
web.GET("/someGet", getting, middlewares...)
web.POST("/somePost", posting, middlewares...)
web.PUT("/somePut", putting, middlewares...)
web.DELETE("/someDelete", deleting, middlewares...)
web.PATCH("/somePatch", patching, middlewares...)
web.HEAD("/someHead", head, middlewares...)
web.OPTIONS("/someOptions", options, middlewares...)
web.Any("/anyMethods", options, middlewares...)
// 参数路由
// 匹配 /user/angelina /user/anyname 不会匹配 /user /user/
web.GET("/user/:name", func(context *easyweb.Context) {
	name := context.PathParam("name")
	context.String(200, "Hello %s", name)
})
// 匹配 /user/angelina 以及/user/anyname/create
web.GET("/user/:name/*action", func(context *easyweb.Context) {
	name := context.PathParam("name")
	action := c.PathParam("action")
	message := name + " is " + action
	context.String(200, message)
})
```
总结,```web.method("请求的url",主体处理函数,中间件列表)```
method代表请求的方法，中间件从前向后执行

## 处理输入

```go
以下均为*easyweb.Context的方法
// 获取ip
ClientIP() string
// 获取请求方法(GET/POST/...)
Method() string
// 获取header
Header(key string) string
// 获取请求的body体
Body() ([]byte, error)
// 获取cookie
Cookie() (*http.Cookie, error)
// 获取header中Content-Type的值
ContentType() string
// 是否为ws
IsWebSocket() bool
// 获取路由参数(/input/:username)
PathParam(key string) string
// 获取url请求参数,默认为""
Query(key string) string
// 可以设置默认值
DefaultQuery(key, defaultVal string) string
// 第二个字段返回是否存在
GetQuery(key string) (string, bool)
// 返回对应的slice
QueryArray(key string) []string
// 第二个字段返回是否存在
GetQueryArray(key string) ([]string, bool)
// 获取post请求参数
PostForm(key string) string
// 可以设置默认值
DefaultPostForm(key, defaultVal string) string
// 第二个字段返回是否存在
GetPostForm(key string) (string, bool)
// 返回对应的slice
PostFormArray(key string) []string
// 第二个字段返回是否存在
GetPostFormArray(key string) ([]string, bool)
// 获取指定key的第一个上传文件
FormFile(name string) (*multipart.FileHeader, error)
// 存储上传文件
SaveUploadedFile(file *multipart.FileHeader, dst string) error
// 绑定body体数据到结构体,默认自动选择binding
Bind(obj interface{}) error

// 链式调用获取请求参数
Param(key string)
GetParam(key string)
PostParam(key string)
SetDefault(val string)
SetDefaultInt(i int)
GetString() string
MustGetString() string
GetInt() int
MustGetInt() int
GetFloat() float64
MustGetFloat() float64
GetError() error

c.Param("input1").SetDefault("default").MustGetString()
if err := c.GetError();err!=nil{
    fmt.Println("input1参数错误" + err.Error())
}
```

## 处理输出

```go
// 设置返回的状态码
Status(code int)
// 设置cookie
SetCookie(cookie *http.Cookie)
// 设置RenderData，为页面提供渲染数据
RenderData(key string, val interface{})
// 下载文件
Attachment(file, name string)
// 输出JSON
JSON(code int, obj interface{}) error
// 输出格式化的JSON
IndentedJSON(code int, obj interface{}) error
// 输出string
String(code int, format string, values ...interface{}) error
// 渲染html页面(如果是view/index.html,可以省略view以及html,直接Render(200,index))
Render(code int, name string) error
// 输出html
HTML(code int, html string) error
// 输出特定格式data
Data(code int, contentType string, data []byte) error
// 跳转
Redirect(code int, url string)
// 相当于Render(200,name)
Render200(name string) error
// 对JSON输出封装 {data:"",code:"1",msg:""}
Success(msg string, obj interface{}) error
// 对JSON输出封装
Fail(code int, msg string) error
// 对JSON输出封装,相当于Fail(-1,msg)
FailWithDefaultCode(msg string) error
```

## 日志

默认日志在debug环境下输出到stdout中
在release环境下输出到文件中，文件位置为logs/log/web.log
暂时不支持自己设定，后面可以加上~~~


```go
easyweb.Logger.Debug("debug")
easyweb.Logger.Error("error")
```

## 其他

### session

开启session

```go
conf := easyweb.DefaultConfig()
conf.SessionOn = true
easyweb.SetConfig(conf)
```

使用session

```go
// 设置session
c.Session().Set("data", "sessiondata")
// 获取session
c.Session().Get("data")
// 删除某个session
c.Session().Delete("data")
// 清空session
c.Session().Flush()
```

### pprof

开启pprof

```go
conf := easyweb.DefaultConfig()
conf.Pprof = true
easyweb.SetConfig(conf)
```
设置为true后会增加一些pprof的相关web路由

```
/debug/pprof/
/debug/pprof/heap
/debug/pprof/goroutine
/debug/pprof/block
/debug/pprof/threadcreate
/debug/pprof/cmdline
/debug/pprof/profile
/debug/pprof/symbol
```

## 总结

欢迎使用，有问题可以提issue或者直接qq我哈~

