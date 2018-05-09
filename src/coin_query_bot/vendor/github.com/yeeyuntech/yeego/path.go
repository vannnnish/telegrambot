/**
 * Created by WillkYang on 2017/3/17.
 */

package yeego

import (
	"os"
	"path/filepath"
	"github.com/yeeyuntech/yeego/yeeFile"
	"fmt"
	"strings"
)

var WORK_PATH string

// GetCurrentPath
// 获取项目路径下面的一些目录，不存在直接panic
func GetCurrentPath(dirPath string) string {
	if WORK_PATH != "" {
		return WORK_PATH
	}
	appPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}
	workPath, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	appDirPath := filepath.Join(appPath, dirPath)
	if !yeeFile.FileExists(appDirPath) {
		appDirPath = filepath.Join(workPath, dirPath)
		if !yeeFile.FileExists(appDirPath) {
			panic(fmt.Sprintf("dirPath:[%s] can not find in %s and %s", dirPath, appPath, workPath))
		}
	}
	WORK_PATH = strings.Replace(appDirPath, "\\", "/", -1)
	return WORK_PATH
}
