/**
 * Created by angelina on 2017/8/28.
 */

package easyweb

import (
	"github.com/yeeyuntech/yeego/yeeStrconv"
	"errors"
	"strconv"
	"io/ioutil"
	"fmt"
	"encoding/pem"
	"crypto/x509"
	"crypto/rsa"
	"crypto/rand"
	"encoding/base64"
	"github.com/buger/jsonparser"
)

// 链式调用获取参数
func (c *Context) ClearParam() {
	c.nowParam = *new(Param)
}

func (c *Context) GetParam(key string) *Context {
	c.ClearParam()
	c.nowParam.Key = key
	c.nowParam.Value = c.Query(key)
	return c
}

func (c *Context) PostParam(key string) *Context {
	c.ClearParam()
	c.nowParam.Key = key
	c.nowParam.Value = c.PostForm(key)
	return c
}

func (c *Context) Param(key string) *Context {
	c.ClearParam()
	c.nowParam.Key = key
	if c.Query(key) != "" {
		c.nowParam.Value = c.Query(key)
	} else {
		c.nowParam.Value = c.PostForm(key)
	}
	return c
}

func (c *Context) SetDefault(val string) *Context {
	if len(c.nowParam.Value) == 0 {
		c.nowParam.Value = val
	}
	return c
}

func (c *Context) SetDefaultInt(i int) *Context {
	if len(c.nowParam.Value) == 0 {
		c.nowParam.Value = yeeStrconv.FormatInt(i)
	}
	return c
}

func (c *Context) GetString() string {
	return c.nowParam.Value
}

func (c *Context) MustGetString() string {
	if len(c.nowParam.Value) == 0 {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, "参数不能为空,参数名称为:"+c.nowParam.Key)
		}
	}
	return c.nowParam.Value
}

func (c *Context) MustGetStringWithError(str string) string {
	if len(c.nowParam.Value) == 0 {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, str)
		}
	}
	return c.nowParam.Value
}

func (c *Context) GetInt() int {
	if len(c.nowParam.Value) == 0 {
		return 0
	}
	i, err := strconv.Atoi(c.nowParam.Value)
	if err != nil {
		return 0
	}
	return i
}

func (c *Context) MustGetInt() int {
	if len(c.nowParam.Value) == 0 {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, "参数不能为空,参数名称为:"+c.nowParam.Key)
		}
		return 0
	}
	i, err := strconv.Atoi(c.nowParam.Value)
	if err != nil {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, "参数类型错误,参数名称为:"+c.nowParam.Key)
		}
		return 0
	}
	return i
}

func (c *Context) MustGetIntWithError(str string) int {
	if len(c.nowParam.Value) == 0 {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, str)
		}
		return 0
	}
	i, err := strconv.Atoi(c.nowParam.Value)
	if err != nil {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, str)
		}
		return 0
	}
	return i
}

func (c *Context) GetBool() bool {
	if len(c.nowParam.Value) == 0 {
		return false
	}
	value, err := strconv.ParseBool(c.nowParam.Value)
	if err != nil {
		return false
	}
	return value
}

func (c *Context) MustGetBool() bool {
	if len(c.nowParam.Value) == 0 {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, "参数不能为空,参数名称为:"+c.nowParam.Key)
		}
		return false
	}
	value, err := strconv.ParseBool(c.nowParam.Value)
	if err != nil {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, "参数类型错误,参数名称为:"+c.nowParam.Key)
		}
		return false
	}
	return value
}

func (c *Context) MustGetBoolWithError(str string) bool {
	if len(c.nowParam.Value) == 0 {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, str)
		}
		return false
	}
	value, err := strconv.ParseBool(c.nowParam.Value)
	if err != nil {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, str)
		}
		return false
	}
	return value
}

func (c *Context) GetFloat() float64 {
	if len(c.nowParam.Value) == 0 {
		return 0
	}
	f, err := strconv.ParseFloat(c.nowParam.Key, 64)
	if err != nil {
		return 0
	}
	return f
}

func (c *Context) MustGetFloat() float64 {
	if len(c.nowParam.Value) == 0 {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, "参数不能为空,参数名称为:"+c.nowParam.Key)
		}
		return 0
	}
	f, err := strconv.ParseFloat(c.nowParam.Key, 64)
	if err != nil {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, "参数类型错误,参数名称为:"+c.nowParam.Key)
		}
		return 0
	}
	return f
}

func (c *Context) MustGetFloatWithError(str string) float64 {
	if len(c.nowParam.Value) == 0 {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, str)
		}
		return 0
	}
	f, err := strconv.ParseFloat(c.nowParam.Key, 64)
	if err != nil {
		if !c.validation.HasErrors() {
			c.validation.SetError(c.nowParam.Key, str)
		}
		return 0
	}
	return f
}

func (c *Context) GetError() error {
	if c.validation.HasErrors() {
		for _, err := range c.validation.Errors {
			return errors.New(err.Message)
		}
	}
	return nil
}

const (
	ServerPrivateKeyPath = "cert/server_private.pem"
)

func (c *Context) RsaGetString(key string) string {
	data := c.Param("data").GetString()
	jsonBytes, err := RsaDecrypt(ServerPrivateKeyPath, []byte(data))
	if err != nil {
		return ""
	}
	str, err := jsonparser.GetString(jsonBytes, key)
	if err != nil {
		return ""
	}
	return str
}

func (c *Context) RsaGetInt(key string, defaultInt int) int {
	data := c.Param("data").GetString()
	jsonBytes, err := RsaDecrypt(ServerPrivateKeyPath, []byte(data))
	if err != nil {
		return defaultInt
	}
	i, err := jsonparser.GetInt(jsonBytes, key)
	if err != nil {
		return defaultInt
	}
	return int(i)
}

func (c *Context) RsaMustGetString(key string, errStr string) string {
	data := c.Param("data").GetString()
	jsonBytes, err := RsaDecrypt(ServerPrivateKeyPath, []byte(data))
	if err != nil {
		if !c.validation.HasErrors() {
			c.validation.SetError(key, errStr)
		}
		return ""
	}
	str, err := jsonparser.GetString(jsonBytes, key)
	if err != nil {
		if !c.validation.HasErrors() {
			c.validation.SetError(key, errStr)
		}
		return ""
	}
	return str
}

func (c *Context) RsaMustGetInt(key string, errStr string) int {
	data := c.Param("data").GetString()
	jsonBytes, err := RsaDecrypt(ServerPrivateKeyPath, []byte(data))
	if err != nil {
		if !c.validation.HasErrors() {
			c.validation.SetError(key, errStr)
		}
		return 0
	}
	i, err := jsonparser.GetInt(jsonBytes, key)
	if err != nil {
		if !c.validation.HasErrors() {
			c.validation.SetError(key, errStr)
		}
		return 0
	}
	return int(i)
}

/*
	RSA
 */

// RsaEncrypt
// rsa加密
// @params publicKeyPath 公钥文件位置
// @params origData 需要加密的数据
// @return 返回的数据是经过base64编码的
func RsaEncrypt(publicKeyPath string, origData []byte) (string, error) {
	publicKey, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		return "", fmt.Errorf("read public key file: %s", err)
	}
	block, _ := pem.Decode(publicKey)
	if block == nil {
		return "", errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", err
	}
	pub := pubInterface.(*rsa.PublicKey)
	returnData, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
	return base64.StdEncoding.EncodeToString(returnData), err
}

// RsaDecrypt
// rsa解密
// 需要对传入的数据进行base64解码
// @params privateKeyPath 私钥文件位置
// @params ciphertext 需要解密的数据
func RsaDecrypt(privateKeyPath string, ciphertext []byte) ([]byte, error) {
	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("read private key file: %s", err)
	}
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return nil, errors.New("private key error!")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	base64Text, _ := base64.StdEncoding.DecodeString(string(ciphertext))
	return rsa.DecryptPKCS1v15(rand.Reader, priv, []byte(base64Text))
}
