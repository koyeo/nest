package secret

import (
	"crypto/hmac"
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func Md5(text []byte) string {
	m := md5.New()
	m.Write(text)
	return hex.EncodeToString(m.Sum(nil))
}

func FileMd5(path string) (hash string, err error) {

	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer func() {
		err = f.Close()
	}()

	h := md5.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}

	hash = hex.EncodeToString(h.Sum(nil))

	return
}

func Md5HMac(salt, text []byte) []byte {
	mac := hmac.New(md5.New, salt)
	mac.Write(text)
	return mac.Sum(nil)
}
