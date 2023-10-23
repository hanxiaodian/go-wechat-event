package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"go-wechat-event/service/wechat"
	"go-wechat-event/utils"
)

var logger = logrus.New()

func main() {
	r := gin.Default()

	// 微信服务器验证地址
	r.GET("/wechat/event", handleVerification)
	// 接收微信服务器推送的事件
	r.POST("/wechat/event", handleEvent)

	// 启动服务
	r.Run(":8080")
}

func handleVerification(c *gin.Context) {
	// 从请求参数中获取验证所需的参数
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	echostr := c.Query("echostr")

	check := utils.CheckSignature(signature, timestamp, nonce)
	if !check {
		// 验证失败
		c.String(http.StatusBadRequest, "Invalid signature")
	}

	// 验证成功，返回 echostr 给微信服务器
	c.String(http.StatusOK, echostr)
}

type EventPushBody struct {
	ToUserName string
	Encrypt    string
}

type EventMsg struct {
	ToUserName   string `json:"ToUserName"`   // 小程序的原始ID
	FromUserName string `json:"FromUserName"` // 发送消息的人的 openid
	CreateTime   int64  `json:"CreateTime"`   // 消息创建时间，整型
	MsgId        int64  `json:"MsgId"`        // 消息id，64位整型
	MsgType      string `json:"MsgType"`      // 消息类型
	Event        string `json:"Event"`        // 微信服务器推送事件类型
	Content      string `json:"Content"`      // MsgType 为 text 时有效，用户发的文本内容
	PicUrl       string `json:"PicUrl"`       // MsgType 为 image 时有效，用户发的图片链接（由系统生成）
	MediaId      string `json:"MediaId"`      // MsgType 为 image 时有效，图片消息媒体id，可以调用[获取临时素材]((getTempMedia)接口拉取数据。
	Title        string `json:"Title"`        // MsgType 为 miniprogrampage 时有效，用户发送小程序卡片标题
	AppId        string `json:"AppId"`        // MsgType 为 miniprogrampage 时有效，用户发送小程序appid
	PagePath     string `json:"PagePath"`     // MsgType 为 miniprogrampage 时有效，用户发送小程序页面路径
	ThumbUrl     string `json:"ThumbUrl"`     // MsgType 为 miniprogrampage 时有效，用户发送封面图片的临时cdn链接
	ThumbMediaId string `json:"ThumbMediaId"` // MsgType 为 miniprogrampage 时有效，用户发送封面图片的临时素材id
	SessionFrom  string `json:"SessionFrom"`  // MsgType 是 event 且 Event 是 user_enter_tempsession 时有效，开发者在客服会话按钮设置的 session-from 属性
}

func handleEvent(c *gin.Context) {
	// 从请求 query 参数中获取验证所需的参数
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	openid := c.Query("openid")
	// encrypt_type := c.Query("encrypt_type")
	msgsignature := c.Query("msg_signature")

	signatureCheck := utils.CheckSignature(signature, timestamp, nonce)
	if !signatureCheck {
		// 验证失败
		c.String(http.StatusBadRequest, "Invalid signature")
	}

	// 从请求 request body 中获取事件推送参数
	var eventData EventPushBody
	err := c.ShouldBindJSON(&eventData)
	if err != nil {
		c.String(http.StatusBadRequest, "invalid params")
		return
	}

	msgSignatureCheck := utils.CheckSignature(msgsignature, timestamp, nonce, eventData.Encrypt)
	if !msgSignatureCheck {
		// 验证失败
		c.String(http.StatusBadRequest, "invalid message signature")
	}

	// 解密推送消息
	decryptedData, err := utils.DecryptEventMessage(eventData.Encrypt)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to decrypt event message")
		return
	}

	// 解析推送消息的类型
	var eventMsg EventMsg
	if err := json.Unmarshal([]byte(decryptedData), &eventMsg); err != nil {
		c.String(400, "Failed to parse decrypted message")
		return
	}

	switch eventMsg.MsgType {
	case "text":
		// 处理文本消息
		// 具体逻辑...
		var messageBody = map[string]interface{}{
			"touser":  openid,
			"msgtype": "text",
			"text": map[string]string{
				"content": "购买地址是：xxx",
			},
		}
		// 发送客服消息
		err := wechat.SendWeChatMessage(messageBody)
		if err != nil {
			// 异常处理，提示或日志输出，但是要返回给微信服务器 success，不然会出现 “该小程序客服暂时无法提供服务，请稍后再试”
			logger.Error(err)
		}

	case "image":
		// 处理图片消息
		// 具体逻辑...

	case "voice":
		// 处理语音消息
		// 具体逻辑...

	case "video":
		// 处理视频消息
		// 具体逻辑...

	case "location":
		// 处理地理位置消息
		// 具体逻辑...

	case "link":
		// 处理链接消息
		// 具体逻辑...

	case "event":
		// 解析事件类型
		var event struct {
			Event string `json:"Event"`
			// 其他字段根据具体事件类型添加
		}
		if err := json.Unmarshal([]byte(decryptedData), &event); err != nil {
			logger.Error(err)
		}

		switch eventMsg.Event {
		case "user_enter_tempsession":
			// 用户进入客服会话事件通知
			if eventMsg.SessionFrom != "" {
				fmt.Println(eventMsg.SessionFrom)
			}
		default:
			// 未知事件类型
		}
	default:
		// 未知消息类型
	}

	// 处理完消息或事件逻辑后，返回响应给微信服务器
	c.String(200, "success")
}
