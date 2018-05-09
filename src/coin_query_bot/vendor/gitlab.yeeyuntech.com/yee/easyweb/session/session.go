/**
 * Created by angelina on 2017/8/24.
 */

package session

import (
	"net/http"
	"fmt"
	"net/textproto"
	"net/url"
	"encoding/hex"
	"crypto/rand"
	"time"
)

// 存储引擎接口
type Store interface {
	Set(key, value interface{}) error     // 设置session
	Get(key interface{}) interface{}      // 获取session
	Delete(key interface{}) error         // 删除某个session
	SessionID() string                    // 返回该session的ID
	SessionRelease(w http.ResponseWriter) // 释放资源 存储session到对应的存储提供者
	Flush() error                         // 清除session
}

var (
	_ Store = &CookieSessionStore{}
	_ Store = &MemSessionStore{}
)

// 包含了全局session的处理方法
type Provider interface {
	SessionInit(gcLifeTime int64, config string) error // 初始化session
	SessionRead(sid string) (Store, error)             // 读取session
	SessionExist(sid string) bool                      // 是否存在
	SessionDestroy(sid string) error                   // 清除
	SessionAll() int                                   // 获取全部存活session数量
	SessionGC()                                        // gc
}

var (
	_ Provider = &CookieProvider{}
	_ Provider = &MemProvider{}
)

var provides = make(map[string]Provider)

func Register(name string, provide Provider) {
	if provide == nil {
		panic("session: Register provide is nil")
	}
	if _, dup := provides[name]; dup {
		panic("session: Register called twice for provider " + name)
	}
	provides[name] = provide
}

// session配置
type ManagerConfig struct {
	CookieName              string `json:"cookieName"`              // 一般以cookie实现，前台cookie名称
	GcLifeTime              int64 `json:"gcLifeTime"`               // gc时间
	MaxLifeTime             int64 `json:"maxLifeTime"`              // 后台session最大时间
	CookieLifeTime          int `json:"cookieLifeTime"`             // 前台cookie生命周期
	ProviderConfig          string `json:"providerConfig"`          //
	SessionIDLength         int64  `json:"sessionIDLength"`         //
	SessionNameInHTTPHeader string `json:"sessionNameInHttpHeader"` // session在header中的名称
}

// session 管理器
type Manager struct {
	provider Provider
	config   *ManagerConfig
}

// 新建session管理者
func NewManager(provideName string, cf *ManagerConfig) (*Manager, error) {
	provider, ok := provides[provideName]
	if !ok {
		return nil, fmt.Errorf("session: unknown provide %q (forgotten import?)", provideName)
	}
	if cf.MaxLifeTime == 0 {
		cf.MaxLifeTime = cf.GcLifeTime
	}
	strMimeHeader := textproto.CanonicalMIMEHeaderKey(cf.SessionNameInHTTPHeader)
	if cf.SessionNameInHTTPHeader != strMimeHeader {
		strErrMsg := "SessionNameInHttpHeader (" + cf.SessionNameInHTTPHeader + ") has the wrong format, it should be like this : " + strMimeHeader
		return nil, fmt.Errorf(strErrMsg)
	}
	err := provider.SessionInit(cf.MaxLifeTime, cf.ProviderConfig)
	if err != nil {
		return nil, err
	}
	if cf.SessionIDLength == 0 {
		cf.SessionIDLength = 32
	}
	return &Manager{
		provider,
		cf,
	}, nil
}

// 从请求中获取sessionID
// 首先从cookie中读取，然后从url参数中获取，最后从header中获取
func (manager *Manager) getSid(r *http.Request) (string, error) {
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		var sid string
		err = r.ParseForm()
		if err != nil {
			return "", err
		}
		sid = r.FormValue(manager.config.CookieName)
		if sid == "" {
			sids, isFound := r.Header[manager.config.SessionNameInHTTPHeader]
			if isFound && len(sids) != 0 {
				return sids[0], nil
			}
		}
		return sid, nil
	}
	return url.QueryUnescape(cookie.Value)
}

// 获取session store
// 存在则直接获取，不存在则创建
func (manager *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session Store, err error) {
	sid, err := manager.getSid(r)
	if err != nil {
		return nil, err
	}
	if sid != "" && manager.provider.SessionExist(sid) {
		return manager.provider.SessionRead(sid)
	}
	// 生成
	i := 0
	for {
		sid, err = manager.sessionID()
		if err != nil {
			return nil, err
		}
		if manager.provider.SessionExist(sid) {
			i++
			if i > 3 {
				break
			}
			continue
		} else {
			break
		}
	}
	session, err = manager.provider.SessionRead(sid)
	if err != nil {
		return nil, err
	}
	cookie := &http.Cookie{
		Name:     manager.config.CookieName,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		Domain:   "",
	}
	if manager.config.CookieLifeTime > 0 {
		cookie.MaxAge = manager.config.CookieLifeTime
		cookie.Expires = time.Now().Add(time.Duration(manager.config.CookieLifeTime) * time.Second)
	}
	http.SetCookie(w, cookie)
	r.AddCookie(cookie)
	r.Header.Set(manager.config.SessionNameInHTTPHeader, sid)
	w.Header().Set(manager.config.SessionNameInHTTPHeader, sid)
	return
}

// 清除session
func (manager *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {
	r.Header.Del(manager.config.SessionNameInHTTPHeader)
	w.Header().Del(manager.config.SessionNameInHTTPHeader)
	cookie, err := r.Cookie(manager.config.CookieName)
	if err != nil || cookie.Value == "" {
		return
	}
	sid, _ := url.QueryUnescape(cookie.Value)
	manager.provider.SessionDestroy(sid)
	expiration := time.Now()
	cookie = &http.Cookie{Name: manager.config.CookieName,
		Path: "/",
		HttpOnly: true,
		Expires: expiration,
		MaxAge: -1}
	http.SetCookie(w, cookie)
}

// 获取session store
func (manager *Manager) GetSessionStore(sid string) (sessions Store, err error) {
	sessions, err = manager.provider.SessionRead(sid)
	return
}

// session gc
func (manager *Manager) GC() {
	manager.provider.SessionGC()
	time.AfterFunc(time.Duration(manager.config.GcLifeTime)*time.Second, func() { manager.GC() })
}

// 获取活跃session数量
func (manager *Manager) GetActiveSession() int {
	return manager.provider.SessionAll()
}

// 创建sessionID
func (manager *Manager) sessionID() (string, error) {
	b := make([]byte, manager.config.SessionIDLength)
	n, err := rand.Read(b)
	if n != len(b) || err != nil {
		return "", fmt.Errorf("Could not successfully read from the system CSPRNG.")
	}
	return hex.EncodeToString(b), nil
}
