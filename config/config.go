package config

import "os"

type WechatApplet struct {
	AppId     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
	AppToken  string `json:"app_token"`   // 微信公众平台消息推送配置的 Token
	AppAESKey string `json:"app_aes_key"` // 微信公众平台消息推送配置的 EncodingAESKey
}

var (
	WechatAppletEnv   = &WechatApplet{}
	DefaultProjectEnv = map[string]string{
		"APP_ID":      "wxxxxxxxxxx",
		"APP_SECRET":  "....",
		"APP_TOKEN":   "...",
		"APP_AES_KEY": "....",
	}
)

func init() {
	InitEnvConfig()
}

func InitEnvConfig() {
	WechatAppletEnv.AppId = GetEnv("APP_ID")
	WechatAppletEnv.AppSecret = GetEnv("APP_SECRET")
	WechatAppletEnv.AppToken = GetEnv("APP_TOKEN")
	WechatAppletEnv.AppAESKey = GetEnv("APP_AES_KEY")
}

func GetEnv(key string) string {
	env := os.Getenv(key)

	if env != "" || string(env) != "" {
		return env
	}

	localEnv := DefaultProjectEnv[key]
	return localEnv
}
