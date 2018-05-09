/**
 * Created by angelina on 2017/8/30.
 */

package render

import (
	"net/http"
	"bytes"
	"html/template"
	"sync"
	"strings"
	"path/filepath"
	"os"
	"io/ioutil"
	"regexp"
	"fmt"
	"github.com/yeeyuntech/yeego/yeeStrings"
	"github.com/yeeyuntech/yeego/yeeFile"
	"errors"
)

var htmlContentType = []string{"text/html; charset=utf-8"}

type HTML struct {
	Name string
	Data interface{}
}

func (h HTML) Render(w http.ResponseWriter) error {
	h.WriteContentType(w)
	if t, ok := cacheTemplates[h.Name]; ok {
		var b []byte
		buf := bytes.NewBuffer(b)
		var err error
		if t.Lookup(h.Name) != nil {
			err = t.ExecuteTemplate(buf, h.Name, h.Data)
		} else {
			err = t.Execute(buf, h.Data)
		}
		w.Write([]byte(template.HTML(Htmlunquote(buf.String()))))
		return err
	}
	panic("can't find templateFile in the path:" + h.Name)
}

func (h HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, htmlContentType)
}

/*********************************/
/************template*************/
/*********************************/

var (
	//  caching map and supported template file extensions.
	cacheTemplates = make(map[string]*template.Template)
	tplFuncMap     = make(template.FuncMap)
	templatesLock  sync.RWMutex
	//  stores the template extension which will build
	templateExt        = []string{"tpl", "html"}
	templateDir string = "view"
)

func AddFuncMap(key string, fn interface{}) {
	tplFuncMap[key] = fn
}

type templateFile struct {
	root  string
	files map[string][]string
}

// visit will make the paths into two part,the first is subDir (without tf.root),the second is full path(without tf.root).
// if tf.root="views" and
// paths is "views/errors/404.html",the subDir will be "errors",the file will be "errors/404.html"
// paths is "views/admin/errors/404.html",the subDir will be "admin/errors",the file will be "admin/errors/404.html"
func (tf *templateFile) visit(paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}
	if !HasTemplateExt(paths) {
		return nil
	}

	replace := strings.NewReplacer("\\", "/")
	file := strings.TrimLeft(replace.Replace(paths[len(tf.root):]), "/")
	subDir := filepath.Dir(file)

	tf.files[subDir] = append(tf.files[subDir], file)
	return nil
}

// HasTemplateExt return this path contains supported template extension of beego or not.
func HasTemplateExt(paths string) bool {
	for _, v := range templateExt {
		if strings.HasSuffix(paths, "."+v) {
			return true
		}
	}
	return false
}

// 设置模板文件路径
func SetTemplateDir(dir string) {
	templateDir = dir
}

// 编译模板文件
func Build() error {
	return BuildTemplate(templateDir)
}

// BuildTemplate will build all template files in a directory.
// it makes beego can render any template file in view directory.
func BuildTemplate(dir string, files ...string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return errors.New("dir open err")
	}
	self := &templateFile{
		root:  dir,
		files: make(map[string][]string),
	}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return self.visit(path, f, err)
	})
	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
		return err
	}
	buildAllFiles := len(files) == 0
	for _, v := range self.files {
		for _, file := range v {
			if buildAllFiles || yeeStrings.IsInSlice(files, file) {
				templatesLock.Lock()
				ext := filepath.Ext(file)
				var t *template.Template
				if len(ext) == 0 {
					t, err = getTemplate(self.root, file, v...)
				} else {
					t, err = getTemplate(self.root, file, v...)
				}
				if err != nil {
					panic(err)
				} else {
					cacheTemplates[file] = t
				}
				templatesLock.Unlock()
			}
		}
	}
	return nil
}

func getTplDeep(root, file, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileAbsPath string
	if strings.HasPrefix(file, "../") {
		fileAbsPath = filepath.Join(root, filepath.Dir(parent), file)
	} else {
		fileAbsPath = filepath.Join(root, file)
	}
	if e := yeeFile.FileExists(fileAbsPath); !e {
		panic("can't find template file:" + file)
	}
	data, err := ioutil.ReadFile(fileAbsPath)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile("{{" + "[ ]*template[ ]+\"([^\"]+)\"")
	allSub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allSub {
		if len(m) == 2 {
			tl := t.Lookup(m[1])
			if tl != nil {
				continue
			}
			if !HasTemplateExt(m[1]) {
				continue
			}
			t, _, err = getTplDeep(root, m[1], file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

func getTemplate(root, file string, others ...string) (t *template.Template, err error) {
	t = template.New(file).Delims("{{", "}}").Funcs(tplFuncMap)
	var subMods [][]string
	t, subMods, err = getTplDeep(root, file, "", t)
	if err != nil {
		return nil, err
	}
	t, err = _getTemplate(t, root, subMods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func _getTemplate(t0 *template.Template, root string, subMods [][]string, others ...string) (t *template.Template, err error) {
	t = t0
	for _, m := range subMods {
		if len(m) == 2 {
			tpl := t.Lookup(m[1])
			if tpl != nil {
				continue
			}
			//first check filename
			for _, otherFile := range others {
				if otherFile == m[1] {
					var subMods1 [][]string
					t, subMods1, err = getTplDeep(root, otherFile, "", t)
					if err != nil {
					} else if subMods1 != nil && len(subMods1) > 0 {
						t, err = _getTemplate(t, root, subMods1, others...)
					}
					break
				}
			}
			//second check define
			for _, otherFile := range others {
				fileAbsPath := filepath.Join(root, otherFile)
				data, err := ioutil.ReadFile(fileAbsPath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile("{{" + "[ ]*define[ ]+\"([^\"]+)\"")
				allSub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allSub {
					if len(sub) == 2 && sub[1] == m[1] {
						var subMods1 [][]string
						t, subMods1, err = getTplDeep(root, otherFile, "", t)
						if err != nil {
						} else if subMods1 != nil && len(subMods1) > 0 {
							t, err = _getTemplate(t, root, subMods1, others...)
						}
						break
					}
				}
			}
		}

	}
	return
}

func init() {
	tplFuncMap["dateformat"] = DateFormat
	tplFuncMap["date"] = Date
	tplFuncMap["compare"] = Compare
	tplFuncMap["compare_not"] = CompareNot
	tplFuncMap["not_nil"] = NotNil
	tplFuncMap["not_null"] = NotNil
	tplFuncMap["substr"] = Substr
	tplFuncMap["html2str"] = HTML2str
	tplFuncMap["str2html"] = Str2html
	tplFuncMap["htmlquote"] = Htmlquote
	tplFuncMap["htmlunquote"] = Htmlunquote
	tplFuncMap["renderform"] = RenderForm
	tplFuncMap["assets_js"] = AssetsJs
	tplFuncMap["assets_css"] = AssetsCSS
	tplFuncMap["map_get"] = MapGet

	// Comparisons
	tplFuncMap["eq"] = eq // ==
	tplFuncMap["ge"] = ge // >=
	tplFuncMap["gt"] = gt // >
	tplFuncMap["le"] = le // <=
	tplFuncMap["lt"] = lt // <
	tplFuncMap["ne"] = ne // !=
}
