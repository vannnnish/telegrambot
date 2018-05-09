/**
 * Created by angelina on 2017/5/4.
 */

package aes

import (
	"crypto/sha512"
	"crypto/aes"
	"crypto/rand"
	"bytes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// AesEncrypt aes加密
// key为任意长度
// data为任意长度
// 使用随机化iv,多次加密结果不同
// PKCS5Padding
// CBC模式
func AesEncrypt(key []byte, data []byte) ([]byte, error) {
	keyHash := sha512.Sum512(key)
	aesKey := keyHash[:32]
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	cbcIv := make([]byte, 16)
	rand.Read(cbcIv)
	data = PKCS5Padding(data, block.BlockSize())
	crypted := make([]byte, len(data)+16)
	copy(crypted[:16], cbcIv)
	blockMode := cipher.NewCBCEncrypter(block, cbcIv)
	blockMode.CryptBlocks(crypted[16:], data)
	return []byte(base64.StdEncoding.EncodeToString(crypted)), nil
}

// AesDecrypt aes cbc PKCS5Padding 解密
func AesDecrypt(key []byte, data []byte) (origData []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New("AES解密失败")
		}
	}()
	data, _ = base64.StdEncoding.DecodeString(string(data))
	cbcIv := data[:16]
	keyHash := sha512.Sum512(key)
	aesKey := keyHash[:32]
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, cbcIv)
	origData = make([]byte, len(data[16:]))
	blockMode.CryptBlocks(origData, data[16:])
	origData = PKCS5UnPadding(origData)
	return origData, nil
}

// PKCS5Padding 数据填充方式
// 填充为blockSize的整数倍
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// PKCS5UnPadding 解填充
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
