package common

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
)

func SaltPassword(salt, password string) string {
	s1 := sha256.New()
	s1.Write([]byte(password))             //sha256 passwd
	str1 := fmt.Sprintf("%x", s1.Sum(nil)) //Sum拿出加密的hash,十六进制输出

	s2 := sha256.New()
	s2.Write([]byte(str1 + salt)) //sha256 salt+sha256passwd
	return fmt.Sprintf("%x", s2.Sum(nil))
}

//MD5 md5加密
func MD5(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))
}

//转换成json
func Obj2Json(s interface{}) string {
	bts, _ := json.Marshal(s)
	return string(bts)
}
func InStringSlice(slice []string, str string) bool {
	for _, item := range slice {
		if str == item {
			return true
		}
	}
	return false
}
