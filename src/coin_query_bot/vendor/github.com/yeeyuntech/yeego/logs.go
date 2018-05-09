/**
 * Created by angelina-zf on 17/2/25.
 */

// yeego 日志相关的功能
// 依赖： "github.com/Sirupsen/logrus"
// 基于logrus的log类,自己搞成了分级的形式
// 可以设置将error层次的log发送到es
package yeego

import (
	"github.com/Sirupsen/logrus"
	"os"
	"github.com/yeeyuntech/yeego/yeeFile"
	"time"
	"runtime"
	"github.com/yeeyuntech/yeego/yeeTime"
	"sync"
)

var infoLogS *logrus.Logger
var debugLogS *logrus.Logger
var errorLogS *logrus.Logger
var day map[string]int
var logfile map[string]*os.File
var logFileLock *sync.Mutex

type LogFields logrus.Fields

// logPath
// 日志的存储地址，按照时间存储
var logPath string
// timePath
// 日志的存储地址，按照时间存储
var timePath string = yeeTime.DateFormat(time.Now(), "YYYY-MM-DD") + "/"
// ESPath
// es服务器的地址以及端口
var eSPath string = "http://localhost:9200"
// ESFromHost
// 从哪个host发送过来的
var eSFromHost string = "localhost"
// ESIndex
// 要存储在哪个index索引下,默认的type为log
var eSIndex string = "testlog"
// runMode
// 运行环境 默认为dev
var runMode string = "dev"

// MustInitLogs
// 注册log
// @param logpath 日志位置
// @param runmode 运行环境 dev|pro
func MustInitLog(path, mode string) {
	logFileLock = new(sync.Mutex)
	if mode != "" && (mode == "dev" || mode == "pro") {
		runMode = mode
	}
	logPath = path
	if runMode != "dev" {
		if !yeeFile.FileExists(logPath) {
			if createErr := os.MkdirAll(logPath, os.ModePerm); createErr != nil {
				panic("error to create logs path : " + createErr.Error())
			}
		}
	}
	day = make(map[string]int)
	logfile = make(map[string]*os.File)
	infoLogS = logrus.New()
	setLogSConfig(infoLogS, logrus.InfoLevel)
	debugLogS = logrus.New()
	setLogSConfig(debugLogS, logrus.DebugLevel)
	errorLogS = logrus.New()
	setLogSConfig(errorLogS, logrus.ErrorLevel)
}

// DefaultLog
// 默认的log配置 dev模式
func DefaultLog() {
	MustInitLog("", "dev")
}

// MustInitESErrorLog
// 为error级别的log注册es
// @param path es服务器的地址以及端口  eg:http://localhost:9200
// @param host 从哪个host发送过来的 eg:localhost
// @param index  要存储在哪个index索引下,默认的type为log eg:dev
// 有错误则直接panic
/*func MustInitESErrorLog(path, host, index string) {
	eSPath = path
	eSFromHost = host
	eSIndex = index
	//TODO 用的是utc时间需要修改源码为Local
	client, err := elastic.NewClient(elastic.SetURL(eSPath))
	if err != nil {
		panic(err.Error())
	}
	hook, err := elogrus.NewElasticHook(client, eSFromHost, logrus.ErrorLevel, eSIndex)
	if err != nil {
		panic(err.Error())
	}
	errorLogS.Hooks.Add(hook)
}*/

func setLogSConfig(logger *logrus.Logger, level logrus.Level) {
	var err error
	logger.Level = level
	logger.Formatter = new(logrus.JSONFormatter)
	if runMode != "dev" {
		logFileLock.Lock()
		defer logFileLock.Unlock()
		logfile[level.String()], err = os.OpenFile(getLogFullPath(level), os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			logfile[level.String()], err = os.Create(getLogFullPath(level))
			if err != nil {
				panic("error to create logs path : " + err.Error())
			}
		}
	}
	if runMode == "dev" {
		// 加上这个貌似没颜色了~好奇怪啊！！！
		//logger.Out = os.Stdout
	} else {
		logger.Out = logfile[level.String()]
	}
	day[level.String()] = yeeTime.Day()
}

// updateLogFile
// 检测是否跨天了,把记录记录到新的文件目录中
func updateLogFile(level logrus.Level) {
	logFileLock.Lock()
	defer logFileLock.Unlock()
	var err error
	day2 := yeeTime.Day()
	if day2 != day[level.String()] {
		logfile[level.String()].Close()
		logfile[level.String()], err = os.Create(getLogFullPath(level))
		if err != nil {
			Print(err)
		}
		switch level {
		case logrus.InfoLevel:
			infoLogS.Out = logfile[level.String()]
		case logrus.DebugLevel:
			debugLogS.Out = logfile[level.String()]
		case logrus.ErrorLevel:
			errorLogS.Out = logfile[level.String()]
		}
	}
}

// locate
// 找到是哪个文件的哪个地方打出的log
func locate(fields LogFields) LogFields {
	_, path, line, ok := runtime.Caller(3)
	if ok {
		fields["file"] = path
		fields["line"] = line
	}
	return fields
}

// LogDebug
// 记录Debug信息
func LogDebug(str interface{}, data LogFields) {
	if debugLogS == nil {
		DefaultLog()
		//panic("please MustInitLogs first")
	}
	if runMode != "dev" {
		updateLogFile(logrus.DebugLevel)
	}
	debugLogS.WithFields(logrus.Fields(locate(data))).Debug(str)
}

// DefaultLogDebug
// 默认debug
func DefaultLogDebug(str interface{}) {
	LogDebug(str, LogFields{})
}

// LogInfo
// 记录Info信息
func LogInfo(str interface{}, data LogFields) {
	if infoLogS == nil {
		DefaultLog()
		//panic("please MustInitLogs first")
	}
	if runMode != "dev" {
		updateLogFile(logrus.InfoLevel)
	}
	infoLogS.WithFields(logrus.Fields(locate(data))).Info(str)
}

// DefaultLogInfo
// 默认info
func DefaultLogInfo(str interface{}) {
	LogInfo(str, LogFields{})
}

// LogError
// 记录Error信息
func LogError(str interface{}, data LogFields) {
	if errorLogS == nil {
		DefaultLog()
		//panic("please MustInitLogs first")
	}
	if runMode != "dev" {
		updateLogFile(logrus.ErrorLevel)
	}
	errorLogS.WithFields(logrus.Fields(locate(data))).Error(str)
}

// DefaultLogError
// 默认error
func DefaultLogError(str interface{}) {
	LogError(str, LogFields{})
}

func getLogFullPath(l logrus.Level) string {
	os.MkdirAll(logPath+timePath, os.ModePerm)
	return logPath + timePath + l.String() + ".log"
}
