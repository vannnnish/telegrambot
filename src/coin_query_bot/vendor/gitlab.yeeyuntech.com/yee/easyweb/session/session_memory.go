/**
 * Created by angelina on 2017/9/2.
 * Copyright © 2017年 yeeyuntech. All rights reserved.
 */

package session

import (
	"sync"
	"time"
	"net/http"
	"container/list"
)

var memPder = &MemProvider{list: list.New(), sessions: make(map[string]*list.Element)}

type MemSessionStore struct {
	sid          string                      // session id
	timeAccessed time.Time                   // last access time
	values       map[interface{}]interface{} // session store
	lock         sync.RWMutex
}

func (st *MemSessionStore) Set(key, value interface{}) error {
	st.lock.Lock()
	st.values[key] = value
	st.lock.Unlock()
	return nil
}

func (st *MemSessionStore) Get(key interface{}) interface{} {
	st.lock.RLock()
	if v, ok := st.values[key]; ok {
		st.lock.RUnlock()
		return v
	}
	st.lock.RUnlock()
	return nil
}

func (st *MemSessionStore) Delete(key interface{}) error {
	st.lock.Lock()
	delete(st.values, key)
	st.lock.Unlock()
	return nil
}

func (st *MemSessionStore) SessionID() string {
	return st.sid
}

func (st *MemSessionStore) SessionRelease(w http.ResponseWriter) () {

}

func (st *MemSessionStore) Flush() error {
	st.lock.Lock()
	st.values = make(map[interface{}]interface{})
	st.lock.Unlock()
	return nil
}

type MemProvider struct {
	lock        sync.RWMutex             // locker
	sessions    map[string]*list.Element // map in memory
	list        *list.List               // for gc
	maxLifeTime int64
}

func (pder *MemProvider) SessionInit(gcLifeTime int64, config string) error {
	pder.maxLifeTime = gcLifeTime
	return nil
}

func (pder *MemProvider) SessionRead(sid string) (Store, error) {
	pder.lock.RLock()
	if element, ok := pder.sessions[sid]; ok {
		go pder.SessionUpdate(sid)
		pder.lock.RUnlock()
		return element.Value.(*MemSessionStore), nil
	}
	pder.lock.RUnlock()
	pder.lock.Lock()
	newSession := &MemSessionStore{sid: sid, timeAccessed: time.Now(), values: make(map[interface{}]interface{})}
	element := pder.list.PushFront(newSession)
	pder.sessions[sid] = element
	pder.lock.Unlock()
	return newSession, nil
}

func (pder *MemProvider) SessionExist(sid string) bool {
	pder.lock.RLock()
	if _, ok := pder.sessions[sid]; ok {
		pder.lock.RUnlock()
		return true
	}
	pder.lock.RUnlock()
	return false
}

func (pder *MemProvider) SessionDestroy(sid string) error {
	pder.lock.Lock()
	if element, ok := pder.sessions[sid]; ok {
		delete(pder.sessions, sid)
		pder.list.Remove(element)
		pder.lock.Unlock()
		return nil
	}
	pder.lock.Unlock()
	return nil
}

func (pder *MemProvider) SessionGC() {
	pder.lock.RLock()
	for {
		element := pder.list.Back()
		if element == nil {
			break
		}
		if (element.Value.(*MemSessionStore).timeAccessed.Unix() + pder.maxLifeTime) < time.Now().Unix() {
			pder.lock.RUnlock()
			pder.lock.Lock()
			pder.list.Remove(element)
			delete(pder.sessions, element.Value.(*MemSessionStore).sid)
			pder.lock.Unlock()
			pder.lock.RLock()
		} else {
			break
		}
	}
	pder.lock.RUnlock()
}

func (pder *MemProvider) SessionAll() int {
	return pder.list.Len()
}

// 更新session的最新访问时间
func (pder *MemProvider) SessionUpdate(sid string) error {
	pder.lock.Lock()
	if element, ok := pder.sessions[sid]; ok {
		element.Value.(*MemSessionStore).timeAccessed = time.Now()
		pder.list.MoveToFront(element)
		pder.lock.Unlock()
		return nil
	}
	pder.lock.Unlock()
	return nil
}

func init() {
	Register("memory", memPder)
}
