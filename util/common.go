// common
package util

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha3"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"log"
	"os"
	"reflect"
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

	randKey = make([]byte, 32) // 32字节=256位
	if _, err := rand.Read(randKey); err != nil {
		LogError(err)
	}

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

	// md5Hash := md5.New()
	hash := sha3.New512()
	// hash := sha512.New()
	_, err := hash.Write(plaintext)
	if err != nil {
		LogError(err)
		return "", err
	}

	for _, value := range hash.Sum(nil) {
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

func RSAEncrypt(plaintext []byte) ([]byte, error) {
	return rsa.EncryptOAEP(sha3.New512(), rand.Reader, &privateKey.PublicKey, plaintext, nil)
}

func RSADecrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptOAEP(sha3.New512(), rand.Reader, privateKey, ciphertext, nil)
}
func AESEncrypt(plaintext, key []byte) ([]byte, error) {

	iv := make([]byte, aes.BlockSize) // 16字节IV
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

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
	blockMode := cipher.NewCBCEncrypter(block, iv)
	blockMode.CryptBlocks(ciphertext, plaintext)

	return append(iv, ciphertext...), nil
}

func AESDecrypt(ciphertext, key []byte) ([]byte, error) {

	iv := ciphertext[:aes.BlockSize]
	text := ciphertext[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if nil != err {
		LogError(err)
		return nil, err
	}

	plaintext := make([]byte, len(text))
	blockMode := cipher.NewCBCDecrypter(block, iv)
	blockMode.CryptBlocks(plaintext, text)

	// PKCS7UnPadding
	length := len(plaintext)
	paddLen := int(plaintext[length-1])
	return plaintext[:(length - paddLen)], nil
}

var randKey = []byte{127, 203, 93, 145, 251, 180, 86, 246, 151, 233, 207, 61, 84, 250, 88, 97, 51, 175, 41, 99, 143, 225, 107, 94, 39, 24, 227, 113, 141, 230, 0, 133}

var publicRandKey = []byte{241, 251, 197, 239, 193, 229, 227, 241, 199, 211, 223, 233, 241, 229, 223, 199, 193, 233, 229, 229, 199, 223, 193, 233, 197, 193, 197, 211, 241, 197, 233, 229}

var privateKey, _ = rsa.GenerateKey(rand.Reader, 4096) // 使用 OaepSHA3_512 时，RSA密钥长度≥3072位（SHA3-384最低要求）‌

// var publicKey = &privateKey.PublicKey

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

func GetPublicKey() []byte {
	pubASN1, _ := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)

	ciphertext, err := AESEncryptWithPublicRand(pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	}))
	if nil != err {
		LogError(err)
		return nil
	}
	// LogDebug(pem.EncodeToMemory(&pem.Block{
	// 	Type:  "PUBLIC KEY",
	// 	Bytes: pubASN1,
	// }))
	return ciphertext
}

func AESEncryptWithPublicRand(plaintext []byte) ([]byte, error) {
	return AESEncrypt(plaintext, publicRandKey)
}

func AESDecryptWithPublicRand(ciphertext []byte) ([]byte, error) {
	return AESDecrypt(ciphertext, publicRandKey)
}

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

// func CheckSign(obj interface{}) bool {

// 	// 1. 获取结构体的反射值
// 	val := reflect.ValueOf(obj)
// 	typ := val.Type() // 等同于 reflect.TypeOf(obj)

// 	// 2. 遍历结构体字段
// 	for i := 0; i < typ.NumField(); i++ {
// 		field := typ.Field(i)      // 获取字段信息
// 		fieldValue := val.Field(i) // 获取字段值

// 		// 3. 获取字段元信息
// 		fmt.Printf("字段名: %-10s 类型: %-15v 值: %v\n",
// 			field.Name,
// 			field.Type,
// 			fieldValue.Interface())

// 		// 4. 获取标签信息
// 		if tag, ok := field.Tag.Lookup("json"); ok {
// 			fmt.Printf("\tJSON标签: %s\n", tag)
// 		}
// 	}

// 	jsonData, err := json.Marshal(obj)
// 	if err != nil {
// 		fmt.Println("序列化失败:", err)
// 		return false
// 	}

// 	// ordered := make(map[string]interface{})

// 	// err = json.Unmarshal(jsonData, &ordered)
// 	// if err != nil {
// 	// 	fmt.Println("序列化失败:", err)
// 	// 	return false
// 	// }

// 	// jsonData1, err := json.Marshal(ordered)
// 	// if err != nil {
// 	// 	fmt.Println("序列化失败:", err)
// 	// 	return false
// 	// }

// 	// 3. 输出结果
// 	fmt.Println("序列化JSON:", string(jsonData))

// 	return true
// }

func ModifyField(obj interface{}, fieldName string, newValue interface{}) {

	field := reflect.ValueOf(obj).Elem().FieldByName(fieldName)

	if field.IsValid() && field.CanSet() {
		field.Set(reflect.ValueOf(newValue))
	}
}
