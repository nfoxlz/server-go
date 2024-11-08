// common
package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/google/uuid"
)

var isDebug bool

var DbBusinessExceptionPrefix string

func init() {
	goos := runtime.GOOS
	switch goos {
	case "windows":
		lineSplit = "\r\n"
	case "linux":
		lineSplit = "\n"
	default:
		lineSplit = "\r"
	}

	// isDebug = false
	// // isDebug = runtime.Callers(0, make([]uintptr, 10)) > 0
	// callers := make([]uintptr, 10)
	// frames := runtime.CallersFrames(callers[:runtime.Callers(0, callers[:])])
	// for frame, more := frames.Next(); more; frame, more = frames.Next() {
	// 	log.Println(frame)
	// 	if strings.Contains(frame.Function, "debug") || strings.Contains(frame.File, "debug") {
	// 		isDebug = true
	// 		return
	// 	}
	// }

	// isDebug = true
	isDebug = strings.Contains(os.Args[0], "debug")
}

func EncryptWithSalt(password string, salt []byte) (string, error) {
	plaintext := []byte(password)
	plaintextLen := len(plaintext)
	saltLen := len(salt)
	for i := 0; i < plaintextLen && i < saltLen; i++ {
		plaintext[i] += salt[i]
	}

	md5Hash := md5.New()
	_, err := md5Hash.Write(plaintext)
	if err != nil {
		LogError(err)
		return "", err
	}

	for _, value := range md5Hash.Sum(nil) {
		salt = append(salt, value)
	}

	return base64.StdEncoding.EncodeToString(salt), nil
}

func Encrypt(password string) (string, error) {
	data, err := uuid.New().MarshalBinary()
	if nil != err {
		LogError(err)
		return "", err
	}

	return EncryptWithSalt(password, data)
}

func Verify(password string, ciphertext string) bool {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if nil != err {
		LogError(err)
		return false
	}

	passwordCiphertext, err := EncryptWithSalt(password, data[:16])
	if nil != err {
		LogError(err)
		return false
	}

	return passwordCiphertext == ciphertext
}

func AESEncrypt(plaintext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if nil != err {
		LogError(err)
		return nil, err
	}

	// PKCS7Padding
	blockSize := block.BlockSize()
	plaintextLen := len(plaintext)
	pad := blockSize - plaintextLen%blockSize
	plaintext = append(plaintext, bytes.Repeat([]byte{byte(pad)}, pad)...)

	ciphertext := make([]byte, len(plaintext))
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	blockMode.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

func AESDecrypt(ciphertext, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if nil != err {
		LogError(err)
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	blockMode := cipher.NewCBCDecrypter(block, key[:block.BlockSize()])
	blockMode.CryptBlocks(plaintext, ciphertext)

	// PKCS7UnPadding
	length := len(plaintext)
	paddLen := int(plaintext[length-1])
	return plaintext[:(length - paddLen)], nil
}

var randKey = []byte{241, 251, 197, 239, 193, 229, 227, 241, 199, 211, 223, 233, 241, 229, 223, 199, 193, 233, 229, 229, 199, 223, 193, 233, 197, 193, 197, 211, 241, 197, 233, 229}

// func init() {
// 	randKey = make([]byte, 32)
// 	var key *big.Int
// 	var err error
// 	for i := 0; i < 32; i++ {
// 		key, err = rand.Prime(rand.Reader, 8)
// 		if nil != err {
// 			LogError(err)
// 			return
// 		}
// 		randKey[i] = key.Bytes()[0]
// 	}
// }

func AESEncryptWithRand(plaintext []byte) ([]byte, error) {
	return AESEncrypt(plaintext, randKey)
}

func AESDecryptWithRand(ciphertext []byte) ([]byte, error) {
	return AESDecrypt(ciphertext, randKey)
}

func AESEncryptWithRandToString(plaintext []byte) (string, error) {
	ciphertext, err := AESEncryptWithRand(plaintext)
	if nil != err {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func AESDecryptWithRandString(ciphertext string) ([]byte, error) {
	ciphertextData, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if nil != err {
		LogError(err)
		return nil, err
	}

	plaintext, err := AESDecryptWithRand(ciphertextData)
	if nil != err {
		return nil, err
	}
	return plaintext, nil
}

func ReadJsonFile(name string, jsonobject any) error {
	buf, err := os.ReadFile(name)
	if nil != err {
		LogError(err)
		log.Println(name)
		return err
	}

	return json.Unmarshal(buf, jsonobject)
}

func UnpackArray[T ~[]E, E any](source T) []any {
	r := make([]any, len(source))
	for i, e := range source {
		r[i] = e
	}
	return r
}

func PackArray[T ~[]E, E any](dest []any) T {
	r := make(T, len(dest))
	for i, e := range dest {
		r[i] = e.(E)
	}
	return r
}

func IsFileExist(name string) bool {
	_, err := os.Stat(name)
	if nil == err {
		return true
	} else {
		if !os.IsNotExist(err) {
			LogError(err)
		}
		return false
	}
}

var lineSplit string

func Split(s string) []string {
	return strings.Split(s, lineSplit)
}

func Min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func MergeMaps[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)

	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}

	return result
}

func ExtractMessage(err error) (int64, string) {
	message := err.Error()

	var errNo int64
	if DbBusinessExceptionPrefix == message[:4] {
		message = message[4:]
		errNo = -1
	} else {
		errNo = 0
	}

	return errNo, message
}
