/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"context"
	"fmt"

	"github.com/lostar01/k8scopilot/utils"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// eventCmd represents the event command
var eventCmd = &cobra.Command{
	Use:   "event",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		eventLog, err := getPodEventsAndLogs()
		if err != nil {
			fmt.Println(err)
		}
		result, err := sendToChatGPT(eventLog)
		fmt.Println(result)
	},
}

// 把日志发送给 ChatGPT 给出建议
func sendToChatGPT(podInfo map[string][]string) (string, error) {
	client, err := utils.NewOpenAIClient()
	if err != nil {
		return "", err
	}
	combinedInfo := "找到以下 Pod Warning 事件和日志：\n"
	for podName, info := range podInfo {
		combinedInfo += fmt.Sprintf("Pod: %s\n", podName)
		for _, i := range info {
			combinedInfo += fmt.Sprintf("%s\n", i)
		}
		combinedInfo += "\n"
	}

	fmt.Println(combinedInfo)

	// 构建 chatgpt 的请求信息
	message := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "你现在是一个 Kubernetes 专家，你要帮助用户诊断 K8S 的问题",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: fmt.Sprintf("以下是多个 Pod Event 事件对应的日志：%s\n, 请主要针对 Pod Log 给出实质性，可操作的建议", combinedInfo),
		},
	}

	// 请求 chatgpt
	resp, err := client.Client.CreateChatCompletion(
		context.TODO(),
		openai.ChatCompletionRequest{
			Model:    "qwen2.5",
			Messages: message,
		},
	)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func getPodEventsAndLogs() (map[string][]string, error) {
	clientGo, err := utils.NewClientGo(kubeconfig)
	if err != nil {
		return nil, err
	}
	result := make(map[string][]string)

	// 获取 Warning 级别的事件
	events, err := clientGo.Clientset.CoreV1().Events("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "type=Warning",
	})
	if err != nil {
		return nil, err
	}

	for _, event := range events.Items {
		podName := event.InvolvedObject.Name
		namespace := event.InvolvedObject.Namespace
		message := event.Message

		if event.InvolvedObject.Kind == "Pod" {
			logOptions := &corev1.PodLogOptions{}
			req := clientGo.Clientset.CoreV1().Pods(namespace).GetLogs(podName, logOptions)
			podLogs, err := req.Stream(context.TODO())
			if err != nil {
				result[podName] = append(result[podName], fmt.Sprintf("Event: %s", err.Error()))
				result[podName] = append(result[podName], fmt.Sprintf("Namespace: %s", namespace))
				return result, nil
			}
			defer podLogs.Close()

			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(podLogs)
			if err != nil {
				continue
			}

			// 如果有日志
			result[podName] = append(result[podName], fmt.Sprintf("Event: %s", message))
			result[podName] = append(result[podName], fmt.Sprintf("Namespace: %s", namespace))
			// 日志信息
			result[podName] = append(result[podName], fmt.Sprintf("Log: %s", buf.String()))
		}
	}
	return result, nil
}

func init() {
	analyzeCmd.AddCommand(eventCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// eventCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// eventCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
