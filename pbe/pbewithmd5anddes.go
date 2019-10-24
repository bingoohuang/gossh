package pbe

import (
	"strings"

	"github.com/bingoohuang/gossh/util"

	"crypto/cipher"
	"crypto/des"
	"crypto/md5"
	"crypto/rand"
)

func getDerivedKey(password, salt string, iterations int) ([]byte, []byte) {
	key := md5.Sum([]byte(password + salt))
	for i := 0; i < iterations-1; i++ {
		key = md5.Sum(key[:])
	}
	return key[:8], key[8:]
}

func Encrypt(plainText, password string, iterations int) (string, error) {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	encText, err := doEncrypt(plainText, password, salt, iterations)
	if err != nil {
		return "", err
	}

	return util.Base64SafeEncode(append(salt, encText...)), nil
}

func Decrypt(cipherText, password string, iterations int) (string, error) {
	msgBytes, err := util.Base64SafeDecode(cipherText)
	if err != nil {
		return "", err
	}

	salt := msgBytes[:8]
	encText := msgBytes[8:]
	return doDecrypt(encText, password, salt, iterations)
}

func EncryptWithFixedSalt(plainText, password, fixedSalt string, iterations int) (string, error) {
	salt := make([]byte, 8)
	copy(salt[:], fixedSalt)

	encText, err := doEncrypt(plainText, password, salt, iterations)
	if err != nil {
		return "", err
	}
	return util.Base64SafeEncode(encText), nil
}

func DecryptWithFixedSalt(cipherText, password, fixedSalt string, iterations int) (string, error) {
	msgBytes, err := util.Base64SafeDecode(cipherText)
	if err != nil {
		return "", err
	}

	salt := make([]byte, 8)
	copy(salt[:], fixedSalt)
	encText := msgBytes[:]
	return doDecrypt(encText, password, salt, iterations)
}

func doEncrypt(plainText, password string, salt []byte, iterations int) ([]byte, error) {
	padNum := byte(8 - len(plainText)%8)
	for i := byte(0); i < padNum; i++ {
		plainText += string(padNum)
	}

	dk, iv := getDerivedKey(password, string(salt), iterations)
	block, err := des.NewCipher(dk)
	if err != nil {
		return nil, err
	}

	encrypter := cipher.NewCBCEncrypter(block, iv)
	encrypted := make([]byte, len(plainText))
	encrypter.CryptBlocks(encrypted, []byte(plainText))

	return encrypted, nil
}

func doDecrypt(encText []byte, password string, salt []byte, iterations int) (string, error) {
	dk, iv := getDerivedKey(password, string(salt), iterations)
	block, err := des.NewCipher(dk)

	if err != nil {
		return "", err
	}

	decrypter := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(encText))
	decrypter.CryptBlocks(decrypted, encText)

	decryptedString := strings.TrimRight(string(decrypted), "\x01\x02\x03\x04\x05\x06\x07\x08")

	return decryptedString, nil
}
