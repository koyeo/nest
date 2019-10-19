package secret

import (
	"encoding/base64"
)

func TextEncrypt(text, salt string) (cipher string, err error) {
	bs, err := AesEncrypt([]byte(text), []byte(salt))
	if err != nil {
		return
	}
	cipher = base64.StdEncoding.EncodeToString(bs)
	return
}

func TextDecrypt(cipher, salt string) (text string, err error) {
	bs, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return
	}
	bs, err = AesDecrypt(bs, []byte(salt))
	if err != nil {
		return
	}
	text = string(bs)
	return
}
