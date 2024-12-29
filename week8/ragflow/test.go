package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	//	"github.com/google/uuid"
	"net/http"
)

func main() {
	analysis := "根据提供的日志信息和知识库内容，以下是总结"

	webhook := "https://open.feishu.cn/open-apis/bot/v2/hook/302ae652-86bd-4151-a867-ba04788844f3"
	message := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": analysis,
		},
	}

	// fmt.Println(message)
	messageBody, _ := json.Marshal(message)
	req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(messageBody))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	fmt.Println(http.StatusOK)
	if resp.StatusCode != http.StatusOK {
		return
	}
	return
}
