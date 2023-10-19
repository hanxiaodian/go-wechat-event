package wechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

// 客服消息请求结构体
type WeChatMessageRequest struct {
	ToUser  string `json:"touser"`
	MsgType string `json:"msgtype"`
	Text    struct {
		Content string `json:"content"`
	} `json:"text"`
}

// 发送客服消息
func SendWeChatMessage(messageBody map[string]interface{}) error {
	// 序列化消息体
	body, err := json.Marshal(messageBody)
	if err != nil {
		return err
	}

	// 发送请求 TODO: AK
	resp, err := http.Post("https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=YOUR_ACCESS_TOKEN",
		"application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// 解析响应
	var respData map[string]interface{}
	if err := json.Unmarshal(respBody, &respData); err != nil {
		return err
	}

	// 检查是否发送成功
	if errCode, ok := respData["errcode"].(float64); ok && errCode != 0 {
		errMsg := respData["errmsg"].(string)
		return errors.New(fmt.Sprintf("Failed to send message: %s", errMsg))
	}

	// 发送成功
	return nil
}
