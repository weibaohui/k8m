package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

var secretKey = "secret-key-16-ok"

// pkcs7Padding 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	// 判断缺少几位长度。最少1，最多 blockSize
	padding := blockSize - len(data)%blockSize
	// 补足位数。把切片[]byte{byte(padding)}复制padding个
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padText...)
}

// pkcs7UnPadding 填充的反向操作
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("加密字符串错误！")
	}
	// 获取填充的个数
	unPadding := int(data[length-1])
	return data[:(length - unPadding)], nil
}

// AesEncrypt 加密
func AesEncrypt(data []byte) ([]byte, error) {
	key := []byte(secretKey)
	// 创建加密实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 判断加密块的大小
	blockSize := block.BlockSize()
	// 填充
	encryptBytes := pkcs7Padding(data, blockSize)
	// 初始化加密数据接收切片
	crypted := make([]byte, len(encryptBytes))
	// 使用cbc加密模式
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	// 执行加密
	blockMode.CryptBlocks(crypted, encryptBytes)
	return crypted, nil
}

// AesDecrypt 解密
func AesDecrypt(ciphertextBase64 string) ([]byte, error) {
	// 解码 Base64 密文
	data, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		return nil, err
	}

	key := []byte(secretKey)
	// 创建实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	// 获取块的大小
	blockSize := block.BlockSize()
	// 使用cbc
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	// 初始化解密数据接收切片
	crypted := make([]byte, len(data))
	// 执行解密
	blockMode.CryptBlocks(crypted, data)
	// 去除填充
	crypted, err = pkcs7UnPadding(crypted)
	if err != nil {
		return nil, err
	}
	return crypted, nil
}
