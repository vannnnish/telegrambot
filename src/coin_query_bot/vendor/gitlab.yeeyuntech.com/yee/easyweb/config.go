/**
 * Created by angelina on 2017/9/2.
 * Copyright © 2017年 yeeyuntech. All rights reserved.
 */

package easyweb

var Config config = config{
	Mode:        "debug",
	TemplateDir: "view",
	Pprof:       false,
	SessionOn:   false,
	Port:        ":8080",
}

type config struct {
	Mode        string `json:"mode"`        // 运行模式  debug / release / test 默认:debug
	TemplateDir string `json:"templateDir"` // 模板文件路径 默认:view
	Pprof       bool `json:"pprof"`         // 是否开启pprof 默认:false
	SessionOn   bool `json:"sessionOn"`     // 是否开启session 默认：false
	Port        string `json:"port"`        // 运行端口 默认:8080
}

// 返回默认配置
func DefaultConfig() config {
	return Config
}

// 设置config
func SetConfig(conf config) {
	Config = conf
	applyConf(conf)
}

// 通过配置文件设置config
func ParseConfig(file string) config {
	conf := Config
	// todo 解析配置文件
	SetConfig(conf)
	return conf
}

// 应用config
func applyConf(conf config) {
	enableSessionOn(conf.SessionOn)
	SetMode(conf.Mode)
}
