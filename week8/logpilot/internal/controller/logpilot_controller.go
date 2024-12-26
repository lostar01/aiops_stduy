/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	aiopscomv1 "github.com/lostar01/llm-log-operator/api/v1"
	openai "github.com/sashabaranov/go-openai"
)

// LogPilotReconciler reconciles a LogPilot object
type LogPilotReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=aiops.com,resources=logpilots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aiops.com,resources=logpilots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=aiops.com,resources=logpilots/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LogPilot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *LogPilotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var logPilot aiopscomv1.LogPilot
	if err := r.Get(ctx, req.NamespacedName, &logPilot); err != nil {
		logger.Error(err, "unable to fetch LogPilot")
		return ctrl.Result{}, err
	}

	// 计算查询时间范围
	currentTime := time.Now().Unix()
	preTimeStamp := logPilot.Status.PreTimeStamp

	fmt.Printf("preTimeStamp: %s\n", preTimeStamp)

	var preTime int64
	if preTimeStamp == "" {
		preTime = currentTime - 5
	} else {
		preTime, _ = strconv.ParseInt(preTimeStamp, 10, 64)
	}

	// Loki 查询
	lokiQuey := logPilot.Spec.LokiPromQL

	// 纳秒级别时间戳
	endTime := currentTime * 100000000
	startTime := (preTime - 5) * 100000000

	if startTime >= endTime {
		logger.Info("startTime >= endTime")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
	}

	startTimeForUpdate := currentTime
	lokiURL := fmt.Sprintf("%s/loki/api/v1/query_range?query=%s&start=%d&end=%d",
		logPilot.Spec.LokiURL, url.QueryEscape(lokiQuey), startTime, endTime)
	// lokiURL := fmt.Sprintf("%s/loki/api/v1/query_range?query=%s&start=%d&end=%d", logPilot.Spec.LokiURL, url.QueryEscape(lokiQuey), startTime, endTime)
	fmt.Printf("lokiURL: %s\n", lokiURL)
	lokiLogs, err := r.queryLoki(lokiURL)
	fmt.Println(lokiLogs)
	if err != nil {
		logger.Error(err, "unable to query Loki")
		return ctrl.Result{}, err
	}

	// 如果有日志的话就调用LLM 去分析
	if lokiLogs != "" {
		fmt.Println("send log to llm")
		analyticsResult, err := r.analyzeLogsWithLLM(logPilot.Spec.LLMEndpoint, logPilot.Spec.LLMToken, logPilot.Spec.LLMModel, lokiLogs)
		if err != nil {
			logger.Error(err, "unable to analyze logs with LLM")
			return ctrl.Result{}, err
		}

		// 如果LLM 返回的结果需要发送飞书通知
		if analyticsResult.HasError {
			err := r.sendFeishuAlert(logPilot.Spec.FeishuWebhook, analyticsResult.Analysis)
			if err != nil {
				logger.Error(err, "unable to send feishu alter")
				return ctrl.Result{}, err
			}
		}
	}

	// 更新状态中preTimeStamp
	logPilot.Status.PreTimeStamp = strconv.FormatInt(startTimeForUpdate, 10)
	if err := r.Status().Update(ctx, &logPilot); err != nil {
		logger.Error(err, "unable to update LogPilot status")
		return ctrl.Result{}, err
	}

	// 10 秒后再次执行

	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}

func (r *LogPilotReconciler) sendFeishuAlert(webhook, analysis string) error {
	message := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": analysis,
		},
	}

	// 将消息内容序列化为 JSON
	messageBody, err := json.Marshal(message)
	if err != nil {
		return err
	}

	// 创建 HTTP POST 请求
	req, err := http.NewRequest("POST", webhook, bytes.NewBuffer(messageBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// 发出请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send Feishu alert, status code: %d", resp.StatusCode)
	}
	return nil
}

type LLMAanalysisResult struct {
	HasError bool   // 是否有错误日志
	Analysis string // LLM 的分析结果
}

func (r *LogPilotReconciler) analyzeLogsWithLLM(endpoint, token, model, logs string) (*LLMAanalysisResult, error) {
	config := openai.DefaultConfig(token)
	config.BaseURL = endpoint
	client := openai.NewClientWithConfig(config)

	resp, err := client.CreateChatCompletion(

		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf("你是一个运维专家，以下日志是从日志系统里获取到的日志，请分析日志错误等级，请分析日志错误等级，如果遇到严重问题，例如外部系统请求失败、系统故障、致命错误、数据库连接异常等严重问题时，给出简短建议,对于你认为严重的且需要通知运维人员的，在内容里返回[feishu]标识: %s\n", logs),
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return r.parseLLMResponse(&resp), nil
}

func (r *LogPilotReconciler) parseLLMResponse(resp *openai.ChatCompletionResponse) *LLMAanalysisResult {
	result := &LLMAanalysisResult{
		Analysis: resp.Choices[0].Message.Content,
	}

	// 判断结果是否包含[feishu]
	if strings.Contains(strings.ToLower(result.Analysis), "[feishu]") {
		result.HasError = true
	} else {
		result.HasError = false
	}
	return result
}

func (r *LogPilotReconciler) queryLoki(lokiURL string) (string, error) {
	resp, err := http.Get(lokiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var lokiResponse map[string]interface{}
	err = json.Unmarshal(body, &lokiResponse)
	if err != nil {
		return "", err
	}

	data, ok := lokiResponse["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("data not found")
	}

	// 检查 result 是否为空
	result, ok := data["result"].([]interface{})
	if !ok || len(result) == 0 {
		fmt.Print("result not found")
		return "", nil
	}

	//

	return string(body), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogPilotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&aiopscomv1.LogPilot{}).
		Named("logpilot").
		Complete(r)
}
