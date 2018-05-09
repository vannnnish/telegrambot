/**
 * Created by angelina on 2017/5/20.
 */

package yeeXss

import (
	"regexp"
	"github.com/yeeyuntech/yeego/yeeStrings"
)

// 黑名单标签
var BlackLabel = []string{"<iframe>", "</iframe>", "<script>", "</script>", "javascript", "xssm", "script"}

// XssFilter
// 过滤Xss
func XssBlackLabelFilter(s string) string {
	reg, err := regexp.Compile("(?i)" + yeeStrings.StringArrayToString(BlackLabel, "|"))
	if err != nil {
		return ""
	}
	if reg.MatchString(s) {
		s = reg.ReplaceAllString(s, "")
	}
	return s
}
