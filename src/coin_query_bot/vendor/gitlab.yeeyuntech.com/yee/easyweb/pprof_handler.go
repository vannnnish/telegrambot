/**
 * Created by angelina on 2017/9/2.
 * Copyright © 2017年 yeeyuntech. All rights reserved.
 */

package easyweb

import (
	"sync"
	"net/http"
)

type customHTTPHandler struct {
	serveHTTP func(w http.ResponseWriter, r *http.Request)
}

func (c *customHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.serveHTTP(w, r)
}

type customEasyWebHandler struct {
	httpHandler http.Handler

	wrappedHandleFunc HandlerFunc
	once              sync.Once
}

func (ceh *customEasyWebHandler) Handle(c *Context) {
	ceh.once.Do(func() {
		ceh.wrappedHandleFunc = ceh.mustWrapHandleFunc(c)
	})
	ceh.wrappedHandleFunc(c)
}

func (ceh *customEasyWebHandler) mustWrapHandleFunc(c *Context) HandlerFunc {
	return func(c *Context) {
		ceh.httpHandler.ServeHTTP(c.Response(), c.Request())
	}
}

func fromHTTPHandler(httpHandler http.Handler) *customEasyWebHandler {
	return &customEasyWebHandler{httpHandler: httpHandler}
}

func fromHandlerFunc(serveHTTP func(w http.ResponseWriter, r *http.Request)) *customEasyWebHandler {
	return &customEasyWebHandler{httpHandler: &customHTTPHandler{serveHTTP: serveHTTP}}
}
