/**
 * Created by angelina-zf on 17/2/25.
 */

// file 文件处理相关函数
package yeeFile

import (
	"time"
	"strings"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"io"
	"path/filepath"
)

// GetBytes
// 通过给定的文件名称或者url地址以及超时时间获取文件的[]byte数据.
func GetBytes(filenameOrURL string, timeout ...time.Duration) ([]byte, error) {
	if strings.Contains(filenameOrURL, "://") {
		if strings.Index(filenameOrURL, "file://") == 0 {
			filenameOrURL = filenameOrURL[len("file://"):]
		} else {
			client := http.DefaultClient
			if len(timeout) > 0 {
				client = &http.Client{Timeout: timeout[0]}
			}
			r, err := client.Get(filenameOrURL)
			if err != nil {
				return nil, err
			}
			defer r.Body.Close()
			if r.StatusCode < 200 || r.StatusCode > 299 {
				return nil, fmt.Errorf("%d: %s", r.StatusCode, http.StatusText(r.StatusCode))
			}
			return ioutil.ReadAll(r.Body)
		}
	}
	return ioutil.ReadFile(filenameOrURL)
}

// SetBytes
// 向指定的文件设置[]byte内容.
func SetBytes(filename string, data []byte) error {
	return ioutil.WriteFile(filename, data, 0660)
}

// AppendBytes
// 向指定的文件追加[]byte内容.
func AppendBytes(filename string, data []byte) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

// GetString
// 通过给定的文件名称或者url地址以及超时时间获取文件的string数据.
func GetString(filenameOrURL string, timeout ...time.Duration) (string, error) {
	bytes, err := GetBytes(filenameOrURL, timeout...)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// SetString
// 向指定的文件设置string内容.
func SetString(filename string, data string) error {
	return SetBytes(filename, []byte(data))
}

// AppendString
// 向指定的文件追加string内容.
func AppendString(filename string, data string) error {
	return AppendBytes(filename, []byte(data))
}

// FileExists
// 文件或者文件夹是否存在.
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

// Mkdir
// 创建文件夹
func Mkdir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// MkdirForFile
// 为某个文件创建目录
func MkdirForFile(path string) (err error) {
	path = filepath.Dir(path)
	return os.MkdirAll(path, os.FileMode(0777))
}

// FileTimeModified
// 返回文件的最后修改时间
// 如果有错误则返回空time.Time.
func FileTimeModified(filename string) time.Time {
	info, err := os.Stat(filename)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// IsDir
// 判断是否是文件夹.
func IsDir(dirname string) bool {
	info, err := os.Stat(dirname)
	return err == nil && info.IsDir()
}

// Find
// 在给定的文件夹中查找某文件.
func Find(searchDirs []string, filenames ...string) (filePath string, found bool) {
	for _, dir := range searchDirs {
		for _, filename := range filenames {
			filePath = path.Join(dir, filename)
			if FileExists(filePath) {
				return filePath, true
			}
		}
	}
	return "", false
}

// GetPrefix
// 获取文件的前缀.
func GetPrefix(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[0:i]
		}
	}
	return ""
}

// GetExt
// 获取文件的后缀.
func GetExt(filename string) string {
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			return filename[i+1:]
		}
	}
	return ""
}

// Copy
// 将文件从原地址拷贝到目的地.
func Copy(source string, dest string) (err error) {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()
	_, err = io.Copy(destFile, sourceFile)
	if err == nil {
		si, err := os.Stat(source)
		if err == nil {
			err = os.Chmod(dest, si.Mode())
		}
	}
	return err
}

// DirSize
// 返回文件夹的大小
func DirSize(path string) int64 {
	var dirSize int64 = 0
	readSize := func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() {
			dirSize += file.Size()
		}
		return nil
	}
	filepath.Walk(path, readSize)
	return dirSize
}
