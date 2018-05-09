/**
 * Created by angelina on 2017/8/24.
 */

package easyweb

import (
	"net/http"
	"io"
)

const (
	defaultStatus = 200
)

type Response struct {
	writer http.ResponseWriter
	// http返回状态码
	status int
	// 返回的数据大小，默认-1代表未返回
	size int
	// 是否已经返回数据
	written bool
}

var _ http.ResponseWriter = &Response{}

// 重置Response
func (r *Response) reset(writer http.ResponseWriter) *Response {
	r.writer = writer
	r.size = 0
	r.status = defaultStatus
	r.written = false
	return r
}

func (r *Response) Header() http.Header {
	return r.writer.Header()
}

// 改变status code
func (r *Response) WriteHeader(code int) {
	if code > 0 && r.status != code {
		r.status = code
	}
}

// 返回的时候WriteHeader
func (r *Response) WriteHeaderNow() {
	if !r.written {
		r.writer.WriteHeader(r.status)
		r.written = true
	}
}

// 输出[]byte
func (r *Response) Write(data []byte) (n int, err error) {
	r.WriteHeaderNow()
	n, err = r.writer.Write(data)
	r.size += n
	return
}

// 输出string
func (r *Response) WriteString(data string) (n int, err error) {
	r.WriteHeaderNow()
	n, err = io.WriteString(r.writer, data)
	r.size += n
	return
}

// 返回ResponseWriter
func (r *Response) ResponseWriter() http.ResponseWriter {
	return r.writer
}

// 返回状态码
func (r *Response) Status() int {
	return r.status
}

// 返回数据大小
func (r *Response) Size() int {
	return r.size
}

// 是否已经返回
func (r *Response) Written() bool {
	return r.written
}
