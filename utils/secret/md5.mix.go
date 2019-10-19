package secret

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
)

func Md5(text []byte) string {
	m := md5.New()
	m.Write(text)
	return hex.EncodeToString(m.Sum(nil))
}

func Md5HMac(salt, text []byte) []byte {
	mac := hmac.New(md5.New, salt)
	mac.Write(text)
	return mac.Sum(nil)
}
