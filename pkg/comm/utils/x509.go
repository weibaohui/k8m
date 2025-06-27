package utils

import (
	"crypto/x509"
	"encoding/pem"
)

// ParseCertificate  判断证书类型 进行解析
func ParseCertificate(data []byte) (*x509.Certificate, error) {
	// 1. 尝试解析 PEM 格式
	block, _ := pem.Decode(data)
	if block != nil {
		data = block.Bytes
	}
	cert, err := x509.ParseCertificate(data)
	if err != nil {
		return nil, err
	}
	return cert, nil

}
