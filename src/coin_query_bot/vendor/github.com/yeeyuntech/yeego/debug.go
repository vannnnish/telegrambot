/**
 * Created by angelina-zf on 17/2/25.
 */

// yeego 调试接口，格式化打印数据
package yeego

import (
	"runtime"
	"bytes"
	"fmt"
	"reflect"
	"os"
)

// SimpleColorPrint
// 简单的带颜色输出到stdout
func SimpleColorPrint(objList ...interface{}) {
	var buf = new(bytes.Buffer)
	fmt.Fprintf(buf, "%c[0;0;33m", 0x1B)
	for i := 0; i < len(objList); i++ {
		formatData(buf, objList[i])
	}
	fmt.Fprintf(buf, "%c[0m\n", 0x1B)
	os.Stdout.WriteString(buf.String())
	return
}

// Print
// 格式化打印数据.
func Print(objList ...interface{}) {
	var pc, file, line, ok = runtime.Caller(1)
	if !ok {
		return
	}
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}
	var buf = new(bytes.Buffer)
	fmt.Fprintf(buf, "%c[0;0;32m [yeegoDebug] at %s() [%s:%d]\n", 0x1B, function(pc), file, line)
	for i := 0; i < len(objList); i++ {
		formatData(buf, objList[i])
	}
	fmt.Fprintf(buf, "%c[0m", 0x1B)
	os.Stdout.WriteString(buf.String())
	return
}

// Sprint
// 返回格式化后的数据.
func Sprint(objList ...interface{}) string {
	/*
	Caller报告当前go程调用栈所执行的函数的文件和行号信息。实参skip为上溯的栈帧数，0表示Caller的调用者（Caller所在的调用栈）。
	函数的返回值为调用栈标识符、文件名、该调用在文件中的行号。如果无法获得信息，ok会被设为false。
	 */
	var pc, file, line, ok = runtime.Caller(1)
	if !ok {
		return ""
	}
	for i := len(file) - 1; i > 0; i-- {
		if file[i] == '/' {
			file = file[i+1:]
			break
		}
	}
	var buf = new(bytes.Buffer)
	fmt.Fprintf(buf, "%c[0;0;32m [yeegoDebug] at %s() [%s:%d]\n", 0x1B, function(pc), file, line)
	for i := 0; i < len(objList); i++ {
		formatData(buf, objList[i])
	}
	fmt.Fprintf(buf, "%c[0m", 0x1B)
	return buf.String()
}

// function
// 如果pc中存在函数,则返回函数名.
func function(pc uintptr) []byte {
	/*
	FuncForPC返回一个表示调用栈标识符pc对应的调用栈的*Func；
	如果该调用栈标识符没有对应的调用栈，函数会返回nil。每一个调用栈必然是对某个函数的调用。
	 */
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return []byte("???")
	}
	name := []byte(fn.Name())
	if period := bytes.LastIndex(name, []byte(".")); period >= 0 {
		name = name[period+1:]
	}
	return name
}

// formatData
// 将interface{}格式化后添加到buf中.
func formatData(buf *bytes.Buffer, data interface{}) {
	val := reflect.ValueOf(data)
	kind := val.Kind()
	switch kind {
	case reflect.Bool:
		fmt.Fprint(buf, val.Bool(), "\n")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fmt.Fprint(buf, val.Int(), "\n")
	case reflect.Uint8, reflect.Uint16, reflect.Uint, reflect.Uint32, reflect.Uint64:
		fmt.Fprint(buf, val.Uint(), "\n")
	case reflect.Float32, reflect.Float64:
		fmt.Fprint(buf, val.Float(), "\n")
	case reflect.Complex64, reflect.Complex128:
		fmt.Fprint(buf, val.Complex(), "\n")
	case reflect.UnsafePointer:
		fmt.Fprintf(buf, "unsafe.Pointer(0x%X)\n", val.Pointer())
	case reflect.String:
		fmt.Fprint(buf, val.String(), "\n")
	case reflect.Ptr:
		//TODO
		fmt.Fprintf(buf, "%#v\n", data)
	case reflect.Interface:
		//TODO
		fmt.Fprintf(buf, "%#v\n", data)
	case reflect.Struct:
		//TODO
		fmt.Fprintf(buf, "%#v\n", data)
	case reflect.Map:
		//TODO
		fmt.Fprintf(buf, "%#v\n", data)
	case reflect.Array, reflect.Slice:
		fmt.Fprint(buf, "{\n")
		for i := 0; i < val.Len(); i++ {
			formatData(buf, val.Index(i))
		}
		fmt.Fprint(buf, "}\n")
	case reflect.Chan:
		fmt.Fprint(buf, val.Type(), "\n")
	case reflect.Invalid:
		fmt.Fprint(buf, "invalid", "\n")
	default:
		fmt.Fprint(buf, "unknow", "\n")
	}
}
