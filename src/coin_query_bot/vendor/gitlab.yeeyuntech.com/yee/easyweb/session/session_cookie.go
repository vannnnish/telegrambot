/**
 * Created by angelina on 2017/9/2.
 * Copyright © 2017年 yeeyuntech. All rights reserved.
 */

package session

import (
	"sync"
	"net/http"
	"crypto/cipher"
	"net/url"
	"encoding/json"
	"crypto/aes"
)

var cookiePder = &CookieProvider{}

// CookieSessionStore Cookie SessionStore
type CookieSessionStore struct {
	sid    string
	values map[interface{}]interface{} // session data
	lock   sync.RWMutex
}

func (st *CookieSessionStore) Set(key, value interface{}) error {
	st.lock.Lock()
	st.values[key] = value
	st.lock.Unlock()
	return nil
}

func (st *CookieSessionStore) Get(key interface{}) interface{} {
	st.lock.RLock()
	if v, ok := st.values[key]; ok {
		st.lock.RUnlock()
		return v
	}
	st.lock.RUnlock()
	return nil
}

func (st *CookieSessionStore) Delete(key interface{}) error {
	st.lock.Lock()
	delete(st.values, key)
	st.lock.Unlock()
	return nil
}

func (st *CookieSessionStore) SessionID() string {
	return st.sid
}

func (st *CookieSessionStore) SessionRelease(w http.ResponseWriter) {
	encodedCookie, err := encodeCookie(cookiePder.block, cookiePder.config.SecurityKey,
		cookiePder.config.SecurityName, st.values)
	if err == nil {
		cookie := &http.Cookie{Name: cookiePder.config.CookieName,
			Value: url.QueryEscape(encodedCookie),
			Path: "/",
			HttpOnly: true,
			Secure: cookiePder.config.Secure,
			MaxAge: cookiePder.config.MaxAge}
		http.SetCookie(w, cookie)
	}
}

func (st *CookieSessionStore) Flush() error {
	st.lock.Lock()
	st.values = make(map[interface{}]interface{})
	st.lock.Unlock()
	return nil
}

type cookieConfig struct {
	SecurityKey  string `json:"securityKey"`  // hash string
	BlockKey     string `json:"blockKey"`     // gob encode hash string. it's saved as aes crypto
	SecurityName string `json:"securityName"` // recognized name in encoded cookie string
	CookieName   string `json:"cookieName"`   // cookie name
	Secure       bool   `json:"secure"`       //
	MaxAge       int    `json:"maxAge"`       // cookie max life time
}

// CookieProvider Cookie session provider
type CookieProvider struct {
	maxLifeTime int64
	config      *cookieConfig
	block       cipher.Block
}

func (pder *CookieProvider) SessionInit(gcLifeTime int64, config string) error {
	pder.config = &cookieConfig{}
	err := json.Unmarshal([]byte(config), pder.config)
	if err != nil {
		return err
	}
	if pder.config.BlockKey == "" {
		pder.config.BlockKey = string(generateRandomKey(16))
	}
	if pder.config.SecurityName == "" {
		pder.config.SecurityName = string(generateRandomKey(20))
	}
	pder.block, err = aes.NewCipher([]byte(pder.config.BlockKey))
	if err != nil {
		return err
	}
	pder.maxLifeTime = gcLifeTime
	return nil
}

func (pder *CookieProvider) SessionRead(sid string) (Store, error) {
	maps, _ := decodeCookie(pder.block,
		pder.config.SecurityKey,
		pder.config.SecurityName,
		sid, pder.maxLifeTime)
	if maps == nil {
		maps = make(map[interface{}]interface{})
	}
	rs := &CookieSessionStore{sid: sid, values: maps}
	return rs, nil
}

func (pder *CookieProvider) SessionExist(sid string) bool {
	return true
}

func (pder *CookieProvider) SessionDestroy(sid string) error {
	return nil
}

func (pder *CookieProvider) SessionGC() {
}

func (pder *CookieProvider) SessionAll() int {
	return 0
}

func init() {
	Register("cookie", cookiePder)
}
