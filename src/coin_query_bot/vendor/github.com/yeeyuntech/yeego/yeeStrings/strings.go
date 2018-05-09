/**
 * Created by angelina on 2017/4/17.
 */

package yeeStrings

import (
	"regexp"
	"strings"
	"bytes"
	"strconv"
	"github.com/yeeyuntech/yeego/yeeStrconv"
)

// IsInSlice
// 字符串是否在slice中
func IsInSlice(slice []string, s string) bool {
	for _, thisS := range slice {
		if thisS == s {
			return true
		}
	}
	return false
}

// MapFunc
// 对传入的slice的每一个元素进行函数操作
func MapFunc(data []string, f func(string) string) []string {
	size := len(data)
	result := make([]string, size, size)
	for i := 0; i < size; i++ {
		result[i] = f(data[i])
	}
	return result
}

// StringStripHTMLTags
// 过滤掉HTML/XML标签
func StripHTMLTags(text string) string {
	var buf *bytes.Buffer
	tagClose := -1
	tagStart := -1
	for i, char := range text {
		//html标签的开始标志
		if char == '<' {
			if buf == nil {
				buf = bytes.NewBufferString(text)
				buf.Reset()
			}
			buf.WriteString(text[tagClose+1: i])
			tagStart = i
			//html标签的结束标志并且start不为-1,说明已经存在开始标签
		} else if char == '>' && tagStart != -1 {
			tagClose = i
			tagStart = -1
		}
	}
	if buf == nil {
		return text
	}
	buf.WriteString(text[tagClose+1:])
	str := buf.String()
	re, _ := regexp.Compile("\\<[\\S\\s]+?\\>")
	str = re.ReplaceAllStringFunc(str, strings.ToLower)
	//替换掉注释和一些标签
	reg := regexp.MustCompile(`<!--[^>]+>|<iframe[\S\s]+?</iframe>
	|<a[^>]+>|</a>|<script[\S\s]+?</script>|<style[\S\s]+?</style>|<div class="hzh_botleft">[\S\s]+?</div>`)
	str = reg.ReplaceAllString(str, "")
	return str
}

// AddURLParam
// 添加URL参数
func AddURLParam(url, name, value string) string {
	var separator string
	//如果不存在?,则表示是第一个参数
	if strings.IndexRune(url, '?') == -1 {
		separator = "?"
		//如果存在?,则表示不是第一个参数
	} else {
		separator = "&"
	}
	return url + separator + name + "=" + value
}

// StringToIntArray
// 将字符串分割为int切片 有错误忽略
func StringToIntArray(s, sep string) []int {
	arr := strings.Split(s, sep)
	intArr := make([]int, 0)
	for _, v := range arr {
		if v == "" {
			continue
		}
		i, err := strconv.Atoi(v)
		if err != nil {
			return []int{}
		}
		intArr = append(intArr, i)
	}
	return intArr
}

// IntArrayToString
// 将int切片组合成string
func IntArrayToString(arr []int, sep string) string {
	strArr := yeeStrconv.IntArrayToStringArr(arr)
	return strings.Join(strArr, sep)
}

// StringToStringArray
// 将字符串分割为string切片
func StringToStringArray(s, sep string) []string {
	arr := strings.Split(s, sep)
	strArr := make([]string, 0)
	for _, v := range arr {
		if v == "" {
			continue
		}
		strArr = append(strArr, v)
	}
	return strArr
}

// StringArrayToString
// 将string切片组合成string
func StringArrayToString(arr []string, sep string) string {
	return strings.Join(arr, sep)
}
