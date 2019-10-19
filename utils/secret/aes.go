package secret

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
)

func AesEncrypt(plainText, key []byte) (cipherText []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}
	plainBytes := []byte(plainText)
	blockSize := block.BlockSize()
	plainBytes = PKCS7Padding(plainBytes, blockSize)
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	cipherText = make([]byte, len(plainBytes))
	blockMode.CryptBlocks(cipherText, plainBytes)
	return
}

func AesDecrypt(cipherText, key []byte) (plainText []byte, err error) {
	keyBytes := []byte(key)
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return
	}
	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, keyBytes[:blockSize])
	plainData := make([]byte, len(cipherText))
	blockMode.CryptBlocks(plainData, cipherText)
	plainText = PKCS7UnPadding(plainData)
	return
}

func PKCS7Padding(cipherText []byte, blockSize int) []byte {
	padding := blockSize - len(cipherText)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(cipherText, paddingText...)
}

func PKCS7UnPadding(plainText []byte) []byte {
	length := len(plainText)
	unPadding := int(plainText[length-1])
	return plainText[:(length - unPadding)]
}
