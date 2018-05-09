package yeego

import (
	"net/http"
	"github.com/labstack/echo"
	"bytes"
	"github.com/yeeyuntech/yeego/yeeTemplate"
	"strings"
	"encoding/json"
	"time"
	"github.com/yeeyuntech/yeego/yeeCrypto/aes"
)

const (
	// DefaultCode
	// 默认错误码0：无错误
	DefaultCode = 0
	// 默认有错误的错误码
	DefaultErrorCode = -1
	// DefaultHttpStatus
	// 默认的http状态码200
	DefaultHttpStatus = http.StatusOK
)

// ResParams
// 返回的参数封装
type ResParams struct {
	Code int         `json:"code";xml:"code"` // 错误码
	Data interface{} `json:"data";xml:"data"` // 数据
	Msg  string      `json:"msg";xml:"msg"`   // 消息
}

// Response
// 对返回的组合封装
type Response struct {
	context    echo.Context
	params     *ResParams
	httpStatus int
	req        *Request
}

// NewResponse
// 新建一个返回
func NewResponse(c echo.Context, re *Request) *Response {
	return &Response{context: c, params: new(ResParams), httpStatus: DefaultHttpStatus, req: re}
}

// Context
// 返回resp的Context
func (resp *Response) Context() echo.Context {
	return resp.context
}

// SetStatus
// 设置返回状态码
func (resp *Response) SetStatus(status int) {
	resp.httpStatus = status
}

// SetMsg
// 设置返回消息
func (resp *Response) SetMsg(msg string) {
	resp.params.Msg = msg
}

// SetData
// 设置返回数据
func (resp *Response) SetData(data interface{}) {
	resp.params.Data = data
}

// Customise(
// 自定义返回内容
func (resp *Response) Customise(code int, data interface{}, msg string) error {
	resp.params.Code = code
	resp.params.Data = data
	resp.params.Msg = msg
	return resp.Ret(resp.params)
}

// Fail
// 返回失败
func (resp *Response) Fail(err error, code int) error {
	resp.params.Code = code
	resp.params.Msg = err.Error()
	return resp.Ret(resp.params)
}

// FailWithDefaultErrorCode
// 默认错误码返回错误
func (resp *Response) FailWithDefaultErrorCode(err error) error {
	resp.params.Code = DefaultErrorCode
	resp.params.Msg = err.Error()
	return resp.Ret(resp.params)
}

// Success
// 返回成功的结果 默认code为0
func (resp *Response) Success(d interface{}) error {
	resp.params.Code = DefaultCode
	resp.params.Data = d
	return resp.Ret(resp.params)
}

// SuccessWithMsg
// 成功并返回msg
func (resp *Response) SuccessWithMsg(msg string) error {
	resp.params.Code = DefaultCode
	resp.params.Msg = msg
	return resp.Ret(resp.params)
}

// Ret
// 返回结果
func (resp *Response) Ret(par interface{}) error {
	resp.SetSession()
	switch ReturnType {
	case 2:
		return resp.context.XML(resp.httpStatus, par)
	default:
		return resp.context.JSON(resp.httpStatus, par)
	}
}

// Write
// 返回row数据
func (resp *Response) Write(data []byte) error {
	resp.SetSession()
	_, err := resp.context.Response().Write(data)
	return err
}

// HTML
// 返回html数据
func (resp *Response) HTML(html string) error {
	resp.SetSession()
	return resp.context.HTML(200, html)
}

// Redirect
// 跳转
func (resp *Response) Redirect(url string) error {
	resp.SetSession()
	return resp.context.Redirect(302, url)
}

// Data
// 设置渲染数据
func (resp *Response) Data(k string, v interface{}) {
	data := resp.context.Get("ResponseData")
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		dataMap = make(map[string]interface{})
	}
	dataMap[k] = v
	resp.context.Set("ResponseData", dataMap)
}

// GetData
// 获取设置的数据
func (resp *Response) GetData() map[string]interface{} {
	data := resp.context.Get("ResponseData")
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		dataMap = make(map[string]interface{})
	}
	return dataMap
}

// Render
// 渲染页面
func (resp *Response) Render(name string) error {
	resp.SetSession()
	var buf bytes.Buffer
	if !strings.Contains(name, ".") {
		name = name + ".html"
	}
	if Config.GetString("app.RunMode") == "dev" {
		err := yeeTemplate.BuildTemplate("view")
		if err != nil {
			panic(err)
		}
	}
	err := yeeTemplate.ExecuteTemplate(&buf, name, resp.context.Get("ResponseData"))
	if err != nil {
		return resp.context.HTML(500, err.Error())
	}
	return resp.context.HTML(200, string(buf.Bytes()))
}

func (resp *Response) SetSession() {
	if resp.req.sessionHasSet && resp.req.sessionMap != nil {
		sessionData, _ := json.Marshal(resp.req.sessionMap)
		sessionValue, _ := aes.AesDecrypt(SessionPsk, sessionData)
		resp.context.SetCookie(&http.Cookie{
			Name:    SessionCookieName,
			Value:   string(sessionValue),
			Path:    "/",
			Expires: time.Now().Add(12 * time.Hour),
		})
	}
}
