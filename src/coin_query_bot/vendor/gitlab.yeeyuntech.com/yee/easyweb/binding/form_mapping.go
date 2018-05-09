/**
 * Created by angelina on 2017/8/28.
 */

package binding

import (
	"reflect"
	"errors"
	"strconv"
)

func mapForm(ptr interface{}, form map[string][]string) error {
	typ := reflect.TypeOf(ptr).Elem()  // 动态类型信息
	val := reflect.ValueOf(ptr).Elem() // 运行时的数据
	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		valueField := val.Field(i)
		if !valueField.CanSet() { // 是否可以设置该字段
			continue
		}
		valueFieldKind := valueField.Kind()         // 获取字段kind
		inputFieldName := typeField.Tag.Get("form") // 根据tag获取输入的参数
		if inputFieldName == "" {
			inputFieldName = typeField.Name // 如果为空，则直接设置为struct的属性键值
			if valueFieldKind == reflect.Struct { // 判断该字段是否为Struct，如果是则递归解析
				err := mapForm(valueField.Addr().Interface(), form)
				if err != nil {
					return err
				}
				continue
			}
		}
		inputValue, ok := form[inputFieldName] // 查询输入参数,不存在则直接continue
		if !ok {
			continue
		}
		inputValueNum := len(inputValue)
		if valueFieldKind == reflect.Slice && inputValueNum > 0 { // 如果该字段是slice类型并且输入有值
			sliceElemKind := valueField.Type().Elem().Kind()                            // 获取slice的单个类型
			slice := reflect.MakeSlice(valueField.Type(), inputValueNum, inputValueNum) // 构造slice
			for i := 0; i < inputValueNum; i++ {
				err := setWithProperType(sliceElemKind, inputValue[i], slice.Index(i)) // 为slice设值
				if err != nil {
					return err
				}
			}
			val.Field(i).Set(slice) // 将slice赋值到struct中
		} else {
			err := setWithProperType(valueFieldKind, inputValue[0], valueField) // 为slice设值
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// 根据kind设置合适类型的值
func setWithProperType(valKind reflect.Kind, val string, valueField reflect.Value) error {
	switch valKind {
	case reflect.Int:
		return setIntField(val, 0, valueField)
	case reflect.Int8:
		return setIntField(val, 8, valueField)
	case reflect.Int16:
		return setIntField(val, 16, valueField)
	case reflect.Int32:
		return setIntField(val, 32, valueField)
	case reflect.Int64:
		return setIntField(val, 64, valueField)
	case reflect.Uint:
		return setUintField(val, 0, valueField)
	case reflect.Uint8:
		return setUintField(val, 8, valueField)
	case reflect.Uint16:
		return setUintField(val, 16, valueField)
	case reflect.Uint32:
		return setUintField(val, 32, valueField)
	case reflect.Uint64:
		return setUintField(val, 64, valueField)
	case reflect.Bool:
		return setBoolField(val, valueField)
	case reflect.Float32:
		return setFloatField(val, 32, valueField)
	case reflect.Float64:
		return setFloatField(val, 64, valueField)
	case reflect.String:
		valueField.SetString(val)
	default:
		return errors.New("Unknown value kind")
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	i, err := strconv.ParseInt(val, 10, bitSize)
	if err != nil {
		return err
	}
	field.SetInt(i)
	return nil
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	i, err := strconv.ParseUint(val, 10, bitSize)
	if err != nil {
		return err
	}
	field.SetUint(i)
	return nil
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return err
	}
	field.SetBool(b)
	return nil
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	f, err := strconv.ParseFloat(val, bitSize)
	if err != nil {
		return err
	}
	field.SetFloat(f)
	return nil
}
