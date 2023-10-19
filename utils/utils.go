package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"

	"go-wechat-event/config"
)

// 验证消息体正确性
func CheckSignature(signature string, encparams ...string) bool {
	// 将 encparams 和 token 参数进行字典序排序
	params := append(encparams, config.WechatAppletEnv.AppToken)
	sort.Strings(params)

	// 将参数拼接成一个字符串
	str := strings.Join(params, "")

	// 对拼接后的字符串进行 SHA1 计算
	h := sha1.New()
	io.WriteString(h, str)
	hashed := hex.EncodeToString(h.Sum(nil))

	// 将计算得到的 SHA1 值与微信服务器传来的 signature 进行对比
	return hashed == signature
}

// 解密微信推送的数据
func DecryptEventMessage(text string) (string, error) {
	// Base64解码
	ciphertext, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return "", err
	}

	// 解密密钥
	aesKey, err := base64.StdEncoding.DecodeString(config.WechatAppletEnv.AppAESKey + "=")
	if err != nil {
		return "", err
	}

	// 解密算法
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	blockMode := cipher.NewCBCDecrypter(block, aesKey[:16])
	plaintext := make([]byte, len(ciphertext))
	blockMode.CryptBlocks(plaintext, ciphertext)

	// PKCS#7去填充
	padding := int(plaintext[len(plaintext)-1])
	plaintext = plaintext[:len(plaintext)-padding]

	// 提取16个随机字节和4个字节的消息长度
	// randomBytes := plaintext[:16]
	msgLengthBytes := plaintext[16:20]
	msgLength := int(msgLengthBytes[0])<<24 | int(msgLengthBytes[1])<<16 | int(msgLengthBytes[2])<<8 | int(msgLengthBytes[3])

	// 提取消息内容和AppID
	content := plaintext[20 : 20+msgLength]
	appIDBytes := plaintext[20+msgLength:]

	fmt.Println("string(appIDBytes):    ", string(appIDBytes))
	fmt.Println("string(content):       ", string(content))

	// 验证AppID
	if !strings.Contains(string(appIDBytes), config.WechatAppletEnv.AppId) {
		return "", fmt.Errorf("AppID validation failed")
	}

	return string(content), nil
}

func PKCS7Decode(data []byte) []byte {
	padding := int(data[len(data)-1])
	if padding < 1 || padding > 32 {
		padding = 0
	}
	return data[:len(data)-padding]
}
