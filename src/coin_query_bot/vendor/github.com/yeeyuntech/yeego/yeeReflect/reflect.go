/**
 * Created by angelina-zf on 17/2/25.
 */

// reflect 对反射做一些封装,获取结构体的相关信息
package yeeReflect

import "reflect"

// IsNil
// 判断是否为空,首先对类型判断,如果未知类型,直接为空
func IsNil(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Interface, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

// GetTypeName
// 获取类型的名称
func GetTypeFullName(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		return GetTypeFullName(t.Elem())
	}
	if t.Name() == "" {
		return ""
	}
	if t.PkgPath() == "" {
		return t.Name()
	}
	return t.PkgPath() + "." + t.Name()
}

// IndirectType
// 获取真实类型
func IndirectType(v reflect.Type) reflect.Type {
	switch v.Kind() {
	case reflect.Ptr:
		return IndirectType(v.Elem())
	default:
		return v
	}
	return v
}

// StructGetAllField
// 获取每个结构体的全部字段信息
func StructGetAllField(t reflect.Type) []*reflect.StructField {
	fieldMap := make(map[string]bool)
	return structGetAllField(t, fieldMap, []int{})
}

func structGetAllField(t reflect.Type, fieldMap map[string]bool, indexs []int) (output []*reflect.StructField) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	//匿名字段列表
	anonymousFieldList := []*reflect.StructField{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		f.Index = append(indexs, f.Index...)
		if f.Anonymous {
			anonymousFieldList = append(anonymousFieldList, &f)
		}
		if fieldMap[f.Name] {
			continue
		}
		fieldMap[f.Name] = true
		output = append(output, &f)
	}
	for _, f := range anonymousFieldList {
		output = append(output, structGetAllField(f.Type, fieldMap, f.Index)...)
	}
	return
}
