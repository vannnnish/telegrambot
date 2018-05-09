/**
 * Created by angelina on 2017/8/24.
 */

package easyweb

import (
	"net/http"
	"math"
	"net/url"
	"gitlab.yeeyuntech.com/yee/easyweb/session"
	"errors"
	"strings"
	"net"
	"io/ioutil"
	"gitlab.yeeyuntech.com/yee/easyweb/validation"
	"mime/multipart"
	"io"
	"os"
	"gitlab.yeeyuntech.com/yee/easyweb/binding"
	"gitlab.yeeyuntech.com/yee/easyweb/render"
	"fmt"
)

const (
	defaultMemory      = 32 << 20 // 32 MB
	abortIndex    int8 = math.MaxInt8 / 2

	defaultSuccessCode = 0
	defaultFailCode    = -1

	HeaderXForwardedFor      = "X-Forwarded-For"
	HeaderXRealIP            = "X-Real-IP"
	HeaderConnection         = "Connection"
	HeaderUpgrade            = "Upgrade"
	HeaderContentType        = "Content-Type"
	HeaderContentDisposition = "Content-Disposition"
)

// Context对象的方法不要并发调用~
type Context struct {
	request  *http.Request
	response *Response

	handlers HandlersChain
	index    int8

	// 路由参数
	params Params

	// 自己存储的值，用来在上下文中传递参数
	store Map

	// get请求参数
	query url.Values

	// 默认当前需要获取的参数
	nowParam Param

	// 参数验证
	validation validation.Validation

	// session
	session session.Store

	// 输出到render中的数据data
	renderData map[string]interface{}

	easyWeb *EasyWeb
}

/************************************/
/****Context初始化以及属性相关信息*******/
/************************************/

// 初始化context
func (c *Context) reset() {
	c.handlers = nil
	c.index = -1
	c.params = make([]Param, 0)
	c.store = make(Map)
	c.query = nil
	c.nowParam = Param{}
	c.validation = validation.Validation{}
	c.renderData = make(map[string]interface{})
}

func (c *Context) Request() *http.Request {
	return c.request
}

func (c *Context) Response() *Response {
	return c.response
}

func (c *Context) Session() session.Store {
	return c.session
}

// 返回主处理函数的名称
func (c *Context) HandlerName() string {
	return nameOfFunction(c.handlers.Last())
}

// 返回主处理函数
func (c *Context) Handler() HandlerFunc {
	return c.handlers.Last()
}

/************************************/
/************处理请求相关*************/
/************************************/

// 中间件中使用
func (c *Context) Next() {
	c.index++
	// 当前handler数长度
	s := int8(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// 终止handler，注意这个并不会终止当前的handler，只是不再进行下面的handlers
// eg:在中间件中调用，如判断权限中间件中调用，可以不再进行下面的处理函数
func (c *Context) Abort() {
	c.index = abortIndex
}

// 判断当前context是否已经终止
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// 终止chain，携带status code
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.response.WriteHeaderNow()
	c.Abort()
}

/************************************/
/********处理context自带数据store******/
/************************************/

func (c *Context) StoreSet(key string, value interface{}) {
	if c.store == nil {
		c.store = make(Map)
	}
	c.store[key] = value
}

func (c *Context) StoreGet(key string) (val interface{}, err error) {
	if val, ok := c.store[key]; ok {
		return val, nil
	}
	return nil, errors.New("Key \"" + key + "\" does not exist")
}

func (c *Context) StoreGetString(key string) (s string) {
	if val, ok := c.store[key]; ok && val != nil {
		s, _ = val.(string)
	}
	return
}

func (c *Context) StoreGetInt(key string) (i int) {
	if val, ok := c.store[key]; ok && val != nil {
		i, _ = val.(int)
	}
	return
}

func (c *Context) StoreGetInt64(key string) (i64 int64) {
	if val, ok := c.store[key]; ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

func (c *Context) StoreGetFloat64(key string) (f64 float64) {
	if val, ok := c.store[key]; ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

func (c *Context) StoreGetMapString(key string) (ms map[string]string) {
	if val, ok := c.store[key]; ok && val != nil {
		ms, _ = val.(map[string]string)
	}
	return
}

func (c *Context) StoreGetMapInterface(key string) (mi map[string]interface{}) {
	if val, ok := c.store[key]; ok && val != nil {
		mi, _ = val.(map[string]interface{})
	}
	return
}

/************************************/
/************处理输入相关*************/
/************************************/

// 获取ip
func (c *Context) ClientIP() string {
	clientIP := c.requestHeader(HeaderXForwardedFor)
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}
	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP = strings.TrimSpace(c.requestHeader(HeaderXRealIP))
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP, _, _ = net.SplitHostPort(strings.TrimSpace(c.Request().RemoteAddr))
	return clientIP
}

// 获取请求的方法
func (c *Context) Method() string {
	return c.request.Method
}

// 获取header
func (c *Context) Header(key string) string {
	return c.requestHeader(key)
}

// 获取body
func (c *Context) Body() ([]byte, error) {
	return ioutil.ReadAll(c.request.Body)
}

// 获取cookie
func (c *Context) Cookie(key string) (*http.Cookie, error) {
	return c.request.Cookie(key)
}

// 获取Content-Type
func (c *Context) ContentType() string {
	return filterFlags(c.requestHeader(HeaderContentType))
}

// 判断是否是WebSocket
func (c *Context) IsWebSocket() bool {
	if strings.Contains(strings.ToLower(c.requestHeader(HeaderConnection)), "upgrade") &&
		strings.ToLower(c.requestHeader(HeaderUpgrade)) == "websocket" {
		return true
	}
	return false
}

// 获取路由参数
func (c *Context) PathParam(key string) string {
	return c.params.ByName(key)
}

// 获取get请求参数
func (c *Context) Query(key string) string {
	val, _ := c.GetQuery(key)
	return val
}

// 带参数默认值的get请求
func (c *Context) DefaultQuery(key, defaultVal string) string {
	if val, ok := c.GetQuery(key); ok {
		return val
	}
	return defaultVal
}

// 获取get参数值
// eg : /?a=1&b=
// ("1",true) = c.GetQuery("a")
// ("",true) = c.GetQuery("b")
// ("",false) = c.GetQuery("c")
func (c *Context) GetQuery(key string) (string, bool) {
	if val, ok := c.GetQueryArray(key); ok {
		return val[0], true
	}
	return "", false
}

// 返回指定key对应的slice值
func (c *Context) QueryArray(key string) []string {
	arr, _ := c.GetQueryArray(key)
	return arr
}

// 返回指定key对应的slice值，第二个返回值代表是否存在一个有效值
func (c *Context) GetQueryArray(key string) ([]string, bool) {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	if val, ok := c.query[key]; ok && len(val) > 0 {
		return val, true
	}
	return []string{}, false
}

//
func (c *Context) PostForm(key string) string {
	if val, ok := c.GetPostForm(key); ok {
		return val
	}
	return ""
}

//
func (c *Context) DefaultPostForm(key, defaultVal string) string {
	if val, ok := c.GetPostForm(key); ok {
		return val
	}
	return defaultVal
}

//
func (c *Context) GetPostForm(key string) (string, bool) {
	if val, ok := c.GetPostFormArray(key); ok {
		return val[0], true
	}
	return "", false
}

// 获取post请求指定key对应的slice值
func (c *Context) PostFormArray(key string) []string {
	arr, _ := c.GetPostFormArray(key)
	return arr
}

// 获取post请求指定key对应的slice值，第二个返回值代表是否存在一个有效值
func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	req := c.request
	req.ParseForm()
	req.ParseMultipartForm(defaultMemory)
	if values := req.PostForm[key]; len(values) > 0 {
		return values, true
	}
	if req.MultipartForm != nil && req.MultipartForm.Value != nil {
		if values := req.MultipartForm.Value[key]; len(values) > 0 {
			return values, true
		}
	}
	return []string{}, false
}

// 获取指定key的第一个上传文件
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	_, fh, err := c.request.FormFile(name)
	return fh, err
}

// 解析MultipartForm
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.request.ParseMultipartForm(defaultMemory)
	return c.request.MultipartForm, err
}

// 存储上传文件
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(out, src)
	return nil
}

// 绑定数据到结构体
func (c *Context) Bind(obj interface{}) error {
	b := binding.Default(c.request.Method, c.ContentType())
	return c.BindWith(obj, b)
}

// bind
func (c *Context) BindWith(obj interface{}, b binding.Binding) error {
	return b.Bind(c.request, obj)
}

// 获取请求的header
func (c *Context) requestHeader(key string) string {
	return c.Request().Header.Get(key)
	// todo 使用下面这种方式获取的header有问题 eg:header:xxx 无法获取
	//if values, _ := c.Request().Header[key]; len(values) > 0 {
	//	return values[0]
	//}
	//return ""
}

/************************************/
/************处理输出相关*************/
/************************************/

// bodyAllowedForStatus is a copy of http.bodyAllowedForStatus non-exported function.
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}

// 设置返回的状态码
func (c *Context) Status(code int) {
	c.response.WriteHeader(code)
}

// 设置cookie
func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.response, cookie)
}

// 设置renderData
func (c *Context) RenderData(key string, val interface{}) {
	c.renderData[key] = val
}

// 获取renderData
func (c *Context) GetRenderData() map[string]interface{} {
	return c.renderData
}

// 输出文件
func (c *Context) File(path string) {
	http.ServeFile(c.response, c.request, path)
}

// 浏览器下载文件
func (c *Context) Attachment(file, name string) {
	c.contentDisposition(file, name, "attachment")
}

func (c *Context) contentDisposition(file, name, dispositionType string) {
	c.response.Header().Set(HeaderContentDisposition, fmt.Sprintf("%s; filename=%s", dispositionType, name))
	c.File(file)
}

// 根据render以及code返回数据
func (c *Context) render(code int, r render.Render) error {
	// dev模式则重新buildTemplate
	if Mode() == DebugMode {
		render.Build()
	}
	c.Status(code)
	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.response)
		c.response.WriteHeaderNow()
		return nil
	}
	if err := r.Render(c.response); err != nil {
		return err
	}
	return nil
}

func (c *Context) JSON(code int, obj interface{}) error {
	return c.render(code, render.JSON{Data: obj})
}

func (c *Context) IndentedJSON(code int, obj interface{}) error {
	return c.render(code, render.IndentedJSON{Data: obj})
}

func (c *Context) String(code int, format string, values ...interface{}) error {
	return c.render(code, render.String{Format: format, Data: values})
}

func (c *Context) Render(code int, name string) error {
	if !strings.Contains(name, ".") {
		name = name + ".html"
	}
	return c.render(code, render.HTML{Name: name, Data: c.renderData})
}

func (c *Context) HTML(code int, html string) error {
	return c.Data(200, "text/html; charset=utf-8", []byte(html))
}

func (c *Context) Data(code int, contentType string, data []byte) error {
	return c.render(code, render.Data{ContentType: contentType, Data: data})
}

func (c *Context) Redirect(code int, url string) {
	c.render(code, render.Redirect{
		Request: c.request,
		Code:    code,
		Url:     url,
	})
}

func (c *Context) Render200(name string) error {
	return c.Render(200, name)
}

func (c *Context) Success(obj interface{}) error {
	data := make(map[string]interface{})
	data["data"] = obj
	data["code"] = defaultSuccessCode
	data["msg"] = ""
	return c.JSON(200, data)
}

func (c *Context) SuccessWithMsg(msg string) error {
	data := make(map[string]interface{})
	data["data"] = nil
	data["code"] = defaultSuccessCode
	data["msg"] = msg
	return c.JSON(200, data)
}

func (c *Context) Fail(code int, msg string) error {
	data := make(map[string]interface{})
	data["data"] = nil
	data["code"] = code
	data["msg"] = msg
	return c.JSON(200, data)
}

func (c *Context) FailWithDefaultCode(msg string) error {
	data := make(map[string]interface{})
	data["data"] = nil
	data["code"] = defaultFailCode
	data["msg"] = msg
	return c.JSON(200, data)
}
