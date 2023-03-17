package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"unsafe"
)

func CalculateMD5(bytesOrString any) string {
	var data []byte
	switch d := bytesOrString.(type) {
	case []byte:
		data = d
	case string:
		data = *(*[]byte)(unsafe.Pointer(&d))
	}

	hash := md5.New()
	hash.Write(data)
	return hex.EncodeToString(hash.Sum(nil))
}

func CalculateResponse(username, realm, password string, method, uri, nonce string) string {
	ha1 := CalculateMD5(fmt.Sprintf("%s:%s:%s", username, realm, password))
	ha2 := CalculateMD5(fmt.Sprintf("%s:%s", method, uri))
	return CalculateMD5(fmt.Sprintf("%s:%s:%s", ha1, nonce, ha2))
}
