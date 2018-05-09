/**
 * Created by WillkYang on 17/2/25.
 */

package yeego

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"github.com/labstack/echo"
	"github.com/yeeyuntech/yeego/validation"
	"net/http"
	"github.com/yeeyuntech/yeego/yeeStrconv"
	"encoding/json"
	"github.com/yeeyuntech/yeego/yeeXss"
	"github.com/yeeyuntech/yeego/yeeCrypto/aes"
)

var SessionCookieName = "yeecmsSession"
var SessionPsk = []byte("UuYqgL0MshCYSsMndpxRaXVLSqdkaA40cfsHCxaxE2TRoAAbQ0k2BCaecqMUXcOi")

// Request
// 封装好的Request请求
type Request struct {
	// 请求上下文
	context echo.Context
	// 请求参数
	params *Param
	// 请求校验
	valid validation.Validation
	// jsonTag 是否是Json数据
	jsonTag bool
	// Json
	json *Json
	// jsonParam
	jsonParam *JsonParam
	// session数据
	sessionMap map[string]string
	// 是否开启了session
	sessionHasSet bool
}

// Param
// 普通请求参数
type Param struct {
	key string
	val string
}

// JsonParam
// Json请求参数
type JsonParam struct {
	key string
	val Json
}

// NewRequest
// 从echo.Context组合请求
func NewRequest(c echo.Context) *Request {
	r := &Request{context: c}
	r.sessionInit()
	return r
}

// Context
// 返回req的Context
func (req *Request) Context() echo.Context {
	return req.context
}

// ContextSet
// 在context中存储信息
func (req *Request) ContextSet(key string, val interface{}) {
	req.context.Set(key, val)
}

// ContextGet
// 从context中获取存储的信息
func (req *Request) ContextGet(key string) interface{} {
	return req.context.Get(key)
}

// ContextGetMap
// 从context中获取存储的信息 map
func (req *Request) ContextGetMap(key string) map[string]string {
	data := req.context.Get(key)
	if data == nil {
		return map[string]string{}
	}
	info, ok := data.(map[string]string)
	if !ok {
		return map[string]string{}
	}
	return info
}

// ContextGetInt
// 从context中获取存储的信息 int
func (req *Request) ContextGetInt(key string) int {
	data := req.context.Get(key)
	if data == nil {
		return 0
	}
	return yeeStrconv.AtoIDefault0(data.(string))
}

// ContextGetString
// 从context中获取存储的信息 string
func (req *Request) ContextGetString(key string) string {
	data := req.context.Get(key)
	if data == nil {
		return ""
	}
	info, ok := data.(string)
	if !ok {
		return ""
	}
	return info
}

// SetCookie
// 设置cookie
func (req *Request) SetCookie(cookie *http.Cookie) {
	req.Context().SetCookie(cookie)
}

// GetCookie
// 根据名称获取cookie,不存在则nil
func (req *Request) GetCookie(cookieName string) *http.Cookie {
	cookie, err := req.Context().Cookie(cookieName)
	if err != nil {
		return nil
	}
	return cookie
}

// sessionInit
// 初始化session
func (req *Request) sessionInit() {
	if req.sessionMap != nil {
		return
	}
	cookie, err := req.context.Cookie(SessionCookieName)
	if err != nil {
		// 这个地方没有cookie是正常情况
		req.sessionMap = map[string]string{}
		return
	}
	output, err := aes.AesDecrypt(SessionPsk, []byte(cookie.Value))
	if err != nil {
		req.sessionMap = map[string]string{}
		return
	}
	err = json.Unmarshal(output, &req.sessionMap)
	if err != nil {
		req.sessionMap = map[string]string{}
		return
	}
}

// SessionSetStr
// 向Session设置字符串
func (req *Request) SessionSetStr(key string, value string) {
	req.sessionInit()
	req.sessionHasSet = true
	req.sessionMap[key] = value
}

// SessionGetStr
// 从Session中取出字符串
func (req *Request) SessionGetStr(key string) string {
	req.sessionInit()
	return req.sessionMap[key]
}

// SessionSetJson
// 向Session设置json
func (req *Request) SessionSetJson(key string, value interface{}) {
	data, _ := json.Marshal(value)
	req.SessionSetStr(key, string(data))
}

// SessionGetJson
// 从Session中取出json
func (req *Request) SessionGetJson(key string, obj interface{}) error {
	out := req.SessionGetStr(key)
	if out == "" {
		return errors.New("Session Empty")
	}
	err := json.Unmarshal([]byte(out), &obj)
	return err
}

// SessionClear
// 清除Session
func (req *Request) SessionClear() {
	req.sessionInit()
	req.sessionHasSet = len(req.sessionMap) > 0
	req.sessionMap = map[string]string{}
}

// GetParam
// get方式获取参数
func (req *Request) GetParam(key string) *Request {
	req.CleanParams()
	req.params.key = key
	req.params.val = req.context.QueryParam(key)
	req.jsonTag = false
	return req
}

// PostParam
// post方式获取参数
func (req *Request) PostParam(key string) *Request {
	req.CleanParams()
	req.params.key = key
	req.params.val = req.context.FormValue(key)
	req.jsonTag = false
	return req
}

// Param
// 通用获取参数方式，尝试get方法获取参数失败后转为post方式获取
func (req *Request) Param(key string) *Request {
	if req.GetParam(key); req.params.val == "" {
		req.PostParam(key)
	}
	return req
}

// PathParam
// 路由参数获取
func (req *Request) PathParam(key string) *Request {
	req.CleanParams()
	req.params.key = key
	req.params.val = req.context.Param(key)
	req.jsonTag = false
	return req
}

// SetJson
// 设置Json相关
func (req *Request) SetJson(json string) {
	req.json = InitJson(json)
}

// JsonParam
// 获取json参数
func (req *Request) JsonParam(keys ...string) *Request {
	req.CleanParams()
	var tmpJson Json
	if req.json != nil {
		tmpJson = *(req.json)
	}
	for _, v := range keys {
		tmpJson.Get(v)
		req.jsonParam.key += v
	}
	req.jsonParam.val = tmpJson
	req.jsonTag = true
	return req
}

// SetDefault
// 设置默认参数值
func (req *Request) SetDefault(val string) *Request {
	if req.jsonTag == true {
		if req.jsonParam.val == *new(Json) {
			defJson := fmt.Sprintf(`{"index":"%s"}`, val)
			req.jsonParam.val = *InitJson(defJson).Get("index")
		}
	} else {
		if len(req.params.val) == 0 {
			req.params.val = val
		}
	}
	return req
}

// SetDefaultInt
// 设置默认参数值int
func (req *Request) SetDefaultInt(val int) *Request {
	if req.jsonTag == true {
		if req.jsonParam.val == *new(Json) {
			defJson := fmt.Sprintf(`{"index":"%d"}`, val)
			req.jsonParam.val = *InitJson(defJson).Get("index")
		}
	} else {
		if len(req.params.val) == 0 {
			req.params.val = yeeStrconv.FormatInt(val)
		}
	}
	return req
}

// GetJson
// 获取设置好的Json
func (req *Request) GetJson() *Json {
	return &(req.jsonParam.val)
}

// CleanParams
// 清除全部的参数缓存
func (req *Request) CleanParams() {
	req.params = new(Param)
	req.jsonParam = new(JsonParam)
}

// CleanError
// 清除全部的错误
func (req *Request) CleanError() {
	req.valid.Clear()
}

// getParamValue
// 获取当前参数的值
func (req *Request) getParamValue() string {
	return req.params.val
}

// getParamKey
// 获取当前参数的键
func (req *Request) getParamKey() string {
	return req.params.key
}

// GetString
// 获取string类型参数
func (req *Request) GetString() string {
	return req.getParamValue()
}

// MustGetString
// 必须获取string类型参数，否则报错
func (req *Request) MustGetString() string {
	if req.getParamValue() == "" {
		req.valid.SetError(req.getParamKey(), "参数不能为空，参数名称：")
		return ""
	}
	return req.getParamValue()
}

// GetInt
// 获取int类型参数
// 若参数不存在则默认返回0
func (req *Request) GetInt() int {
	if req.getParamValue() == "" {
		return 0
	}
	if value, err := strconv.Atoi(req.getParamValue()); err != nil {
		req.valid.SetError(req.getParamKey(), "参数不是int类型，参数名称：")
		return -1
	} else {
		return value
	}
}

// MustGetInt
// 必须获取int类型参数
// 为空则报错
func (req *Request) MustGetInt() int {
	if req.getParamValue() == "" {
		req.valid.SetError(req.getParamKey(), "参数不能为空，参数名称：")
		return -1
	}
	if value, err := strconv.Atoi(req.getParamValue()); err != nil {
		req.valid.SetError(req.getParamKey(), "参数不是int类型，参数名称：")
		return -1
	} else {
		return value
	}
}

// GetFloat
// 获取参数并转化为float64类型
func (req *Request) GetFloat() float64 {
	if req.getParamValue() == "" {
		return 0
	}
	if value, err := strconv.ParseFloat(req.getParamValue(), 64); err != nil {
		req.valid.SetError(req.getParamKey(), "参数不是float类型，参数名称：")
		return -1
	} else {
		return value
	}
}

// MustGetFloat
// 必须获取float64类型参数
func (req *Request) MustGetFloat() float64 {
	if req.getParamValue() == "" {
		req.valid.SetError(req.getParamKey(), "参数不能为空，参数名称：")
		return 0
	}
	if value, err := strconv.ParseFloat(req.getParamValue(), 64); err != nil {
		req.valid.SetError(req.getParamKey(), "参数不是float类型，参数名称：")
		return -1
	} else {
		return value
	}
}

// GetBool
// 获取bool类型参数
func (req *Request) GetBool() bool {
	if value, err := strconv.ParseBool(req.getParamValue()); err != nil {
		req.valid.SetError(req.getParamKey(), "参数不是bool类型，参数名称：")
		return false
	} else {
		return value
	}
}

// MustGetBool
// 必须获取bool类型值
func (req *Request) MustGetBool() bool {
	if req.getParamValue() == "" {
		req.valid.SetError(req.getParamKey(), "参数不能为空，参数名称：")
		return false
	}
	if value, err := strconv.ParseBool(req.getParamValue()); err != nil {
		req.valid.SetError(req.getParamKey(), "参数不是bool类型，参数名称：")
		return false
	} else {
		return value
	}
}

// GetError
// 获取当前请求的错误信息
// 会返回第一个error
func (req *Request) GetError() error {
	if req.valid.HasErrors() {
		for _, err := range req.valid.Errors {
			return errors.New(err.Message + err.Key)
		}
	}
	return nil
}

// XssBlackLabelFilter
// xss黑名单过滤
func (req *Request) XssBlackLabelFilter() *Request {
	req.params.val = yeeXss.XssBlackLabelFilter(req.params.val)
	return req
}

// Min
// 检查参数最小值
func (req *Request) Min(value int) *Request {
	if checkVal, err := strconv.Atoi(req.getParamValue()); err != nil {
		req.valid.SetError(req.getParamKey(), "参数不是int类型，参数名称：")
	} else {
		req.valid.Min(checkVal, value, req.getParamKey()).Message("参数不能小于%d，参数名称:", value)
	}
	return req
}

// Max
// 检查参数最大值
func (req *Request) Max(value int) *Request {
	if checkVal, err := strconv.Atoi(req.getParamValue()); err != nil {
		req.valid.SetError(req.getParamKey(), "参数不是int类型，参数名称：")
	} else {
		req.valid.Max(checkVal, value, req.getParamKey()).Message("参数不能大于%d，参数名称:", value)
	}
	return req
}

// MinLength
// 检查参数最小长度
func (req *Request) MinLength(length int) *Request {
	req.valid.MinSize(req.getParamValue(), length, req.getParamKey()).Message("参数长度不能小于%d，参数名称:", length)
	return req
}

// MaxLength
// 检查参数最大长度
func (req *Request) MaxLength(length int) *Request {
	req.valid.MaxSize(req.getParamValue(), length, req.getParamKey()).Message("参数长度不能大于%d，参数名称:", length)
	return req
}

// Phone
// 检查参数是否为手机号或固话
// ^((\\+86)|(86))?(1(([35][0-9])|[8][0-9]|[7][06789]|[4][579]))\\d{8}$
func (req *Request) Phone() *Request {
	req.valid.Phone(req.getParamValue(), req.getParamKey()).Message("号码格式不正确，参数名称：")
	return req
}

// Tel
// 检查参数是否为固话
// ^(0\\d{2,3}(\\-)?)?\\d{7,8}$
func (req *Request) Tel() *Request {
	req.valid.Tel(req.getParamValue(), req.getParamKey()).Message("固话号码格式不正确，参数名称：")
	return req
}

// Mobile
// 检查参数是否为手机号
func (req *Request) Mobile() *Request {
	req.valid.Mobile(req.getParamValue(), req.getParamKey()).Message("手机号码格式不正确，参数名称：")
	return req
}

// Email
// 检查参数是否为Email
func (req *Request) Email() *Request {
	req.valid.Email(req.getParamValue(), req.getParamKey()).Message("邮箱地址格式不正确，参数名称：")
	return req
}

// ZipCode
// 检查参数是否为邮政编码
func (req *Request) ZipCode() *Request {
	req.valid.ZipCode(req.getParamValue(), req.getParamKey()).Message("邮政编码格式不正确，参数名称：")
	return req
}

// Numeric
// 检查参数是否为数字
func (req *Request) Numeric() *Request {
	req.valid.Numeric(req.getParamValue(), req.getParamKey()).Message("数字格式不正确，参数名称：")
	return req
}

// Alpha
// 检查参数是否为Alpha字符
func (req *Request) Alpha() *Request {
	req.valid.Alpha(req.getParamValue(), req.getParamKey()).Message("Alpha格式不正确，参数名称：")
	return req
}

// AlphaNumeric
// 检查参数是否为数字或Alpha字符
func (req *Request) AlphaNumeric() *Request {
	req.valid.AlphaNumeric(req.getParamValue(), req.getParamKey()).Message("AlphaNumeric格式不正确，参数名称：")
	return req
}

// AlphaDash
// 检查参数是否为Alpha字符或数字或横杠-_
func (req *Request) AlphaDash() *Request {
	req.valid.AlphaDash(req.getParamValue(), req.getParamKey()).Message("AlphaDash格式不正确，参数名称：")
	return req
}

// IP
// 检查参数是否为IP地址
func (req *Request) IP() *Request {
	req.valid.IP(req.getParamValue(), req.getParamKey()).Message("IP地址格式不正确，参数名称：")
	return req
}

// Match
// 检查参数是否匹配正则
func (req *Request) Match(match string) *Request {
	req.valid.Match(req.getParamValue(), regexp.MustCompile(match), req.getParamKey()).Message("正则校验失败，参数名称：")
	return req
}

// NoMatch
// 检查参数是否不匹配正则
func (req *Request) NoMatch(match string) *Request {
	req.valid.NoMatch(req.getParamValue(), regexp.MustCompile(match), req.getParamKey()).Message("正则校验失败，参数名称：")
	return req
}
