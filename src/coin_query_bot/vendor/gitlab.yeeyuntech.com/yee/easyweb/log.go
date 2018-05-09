/**
 * Created by angelina on 2017/9/1.
 * Copyright © 2017年 yeeyuntech. All rights reserved.
 */

package easyweb

import (
	l4g "gitlab.yeeyuntech.com/yee/easyweb/log4go"
	"github.com/yeeyuntech/yeego/yeeFile"
	"os"
)

const (
	recoveryLogFileName = "logs/recovery/recovery.log"
	defaultLogFileName  = "logs/log/web.log"
)

var RecoveryLogger l4g.Logger
var Logger l4g.Logger

func defaultRecoveryWriter() {
	if RecoveryLogger != nil {
		return
	}
	RecoveryLogger = l4g.NewLogger()
	if IsDebugging() {
		RecoveryLogger.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())
	} else {
		err := yeeFile.MkdirForFile(recoveryLogFileName)
		if err != nil {
			panic("can not create file dir " + recoveryLogFileName)
		}
		os.Create(recoveryLogFileName)
		flw := l4g.NewFileLogWriter(recoveryLogFileName, true)
		flw.SetFormat("[%D %T] [%L] (%S) %M")
		flw.SetRotateDaily(true)
		RecoveryLogger.AddFilter("file", l4g.DEBUG, flw)
	}
}

// 设置默认的
func defaultLogger() {
	if Logger != nil {
		return
	}
	Logger = l4g.NewLogger()
	if IsDebugging() {
		Logger.AddFilter("stdout", l4g.DEBUG, l4g.NewConsoleLogWriter())
	} else {
		err := yeeFile.MkdirForFile(defaultLogFileName)
		if err != nil {
			panic("can not create file dir " + defaultLogFileName)
		}
		os.Create(defaultLogFileName)
		flw := l4g.NewFileLogWriter(defaultLogFileName, true)
		flw.SetFormat("[%D %T] [%L] (%S) %M")
		flw.SetRotateDaily(true)
		Logger.AddFilter("file", l4g.DEBUG, flw)
	}
}
