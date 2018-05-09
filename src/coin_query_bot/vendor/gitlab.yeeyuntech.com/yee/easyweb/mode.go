/**
 * Created by angelina on 2017/8/25.
 */

package easyweb

import (
	"io"
	"os"
)

const (
	DebugMode   string = "debug"
	ReleaseMode string = "release"
	TestMode    string = "test"
)

const (
	debugCode   = iota
	releaseCode
	testCode
)

var easyWebMode = debugCode
var modeName = DebugMode

// 中间件logger的默认输出
var DefaultWriter io.Writer = os.Stdout
// 中间件recovery的默认输出
var DefaultErrorWriter io.Writer = os.Stderr

// 设置运行模式 debug / release
func SetMode(value string) {
	switch value {
	case DebugMode:
		easyWebMode = debugCode
	case ReleaseMode:
		easyWebMode = releaseCode
	case TestMode:
		easyWebMode = testCode
	default:
		panic("easyWebMode mode unknown: " + value)
	}
	modeName = value
}

// 返回运行模式
func Mode() string {
	return modeName
}
