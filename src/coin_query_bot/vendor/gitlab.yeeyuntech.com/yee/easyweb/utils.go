/**
 * Created by angelina on 2017/8/24.
 */

package easyweb

import (
	"path"
	"reflect"
	"runtime"
	"strings"
)

// 获取字符串最后一个字符
func lastChar(str string) uint8 {
	size := len(str)
	if size == 0 {
		panic("The length of the string can't be 0")
	}
	return str[size-1]
}

// 拼接地址
func joinPaths(absolutePath, relativePath string) string {
	if len(relativePath) == 0 {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	appendSlash := lastChar(relativePath) == '/' && lastChar(finalPath) != '/'
	if appendSlash {
		return finalPath + "/"
	}
	return finalPath
}

// 解析监听的端口地址，默认8080
func resolveAddress(addr ...string) string {
	port := ""
	if len(addr) == 0 {
		addr = append(addr, Config.Port)
	}
	if addr[0] == "" {
		port = Config.Port
	} else {
		port = addr[0]
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	return port
}

// 获取函数名称
func nameOfFunction(f interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

// 截取content-type前半部分
func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
