/**
 * Created by angelina on 2017/4/17.
 */

package yeeStrconv

import "strconv"

// AtoIDefault0
// 字符串转int,不成功则0
func AtoIDefault0(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// FormatInt
// 格式化int为string
func FormatInt(i int) string {
	return strconv.Itoa(i)
}

// FormatFloat
// 格式化float64为string
func FormatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// FormatFloatPrec0
// 格式化float64位字符串，只取整数位
func FormatFloatPrec0(f float64) string {
	return strconv.FormatFloat(f, 'f', 0, 64)
}

// FormatFloatPrec0
// 格式化float64位字符串，2位小数点
func FormatFloatPrec2(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

// FormatFloatPrec0
// 格式化float64位字符串,4位小数点
func FormatFloatPrec4(f float64) string {
	return strconv.FormatFloat(f, 'f', 4, 64)
}

// ParseFloat64
// 将string解析为float64
func ParseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// ParseFloat64Default0
// 将string解析为float64，失败为0
func ParseFloat64Default0(s string) float64 {
	out, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return out
}

// IntArrayToStringArr
// int数组转换为string数组
func IntArrayToStringArr(i []int) []string {
	strArr := make([]string, len(i))
	for k, v := range i {
		strArr[k] = FormatInt(v)
	}
	return strArr
}
