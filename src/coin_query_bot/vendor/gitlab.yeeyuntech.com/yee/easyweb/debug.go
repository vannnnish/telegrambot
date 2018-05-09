/**
 * Created by angelina on 2017/8/25.
 */

package easyweb

import (
	"log"
	"strings"
)

// 是否是debug模式
func IsDebugging() bool {
	return easyWebMode == debugCode
}

const version = "1.0.0"

func printVersion() {
	if easyWebMode == 2 {
		// test mode
		return
	}
	log.Printf(`
EasyWeb Version : %s
	`, version)
}

func debugPrintWARNINGNew() {
	printVersion()
	debugPrint(`[WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using code:	easyWeb.SetMode(easyWeb.ReleaseMode)

`)
}

func debugPrintRoute(httpMethod, absolutePath string, handlers HandlersChain) {
	if IsDebugging() {
		if strings.Contains(absolutePath, "/debug/pprof") {
			return
		}
		nuHandlers := len(handlers)
		handlerName := nameOfFunction(handlers.Last())
		debugPrint("%-6s %-25s --> %s (%d handlers)\n", httpMethod, absolutePath, handlerName, nuHandlers)
	}
}

func debugPrintError(err error) {
	if err != nil {
		debugPrint("[ERROR] %v\n", err)
	}
}

func debugPrint(format string, values ...interface{}) {
	if IsDebugging() {
		log.Printf("[EasyWeb-debug] "+format, values...)
	}
}
