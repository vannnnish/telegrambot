/**
 * Created by angelina-zf on 17/2/25.
 */

// yeego
// test辅助函数，用来快速判断是否相等
package yeego

import (
	"reflect"
	"github.com/yeeyuntech/yeego/yeeReflect"
)

// Equal
// get和expect是否相同,不同则panic
func Equal(get interface{}, expect interface{}, params ...interface{}) {
	if isEqual(get, expect) {
		return
	}
	if len(params) == 0 {
		Print("get:", get, "expect:", expect)
	} else {
		Print("get:", get, "expect:", expect, params)
	}
	panic("Not Equal")
}

// OK
// input是否等于true
func OK(input interface{}, params ...interface{}) {
	Equal(true, input, params...)
}

// NotEqual
// get和expect是否不同
func NotEqual(get interface{}, expect interface{}) {
	if isEqual(get, expect) {
		panic("Equal")
	}
}

func isEqual(a interface{}, b interface{}) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)
	if yeeReflect.IsNil(va) && yeeReflect.IsNil(vb) {
		return true
	}
	return false
}
