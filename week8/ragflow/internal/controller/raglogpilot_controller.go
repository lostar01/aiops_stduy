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
	"flag"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/google/uuid"
	logv1 "github.com/lostar01/rag-log-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
)

// RagLogPilotReconciler reconciles a RagLogPilot object
type RagLogPilotReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	KubeClient *kubernetes.Clientset
}

// +kubebuilder:rbac:groups=log.aiops.com,resources=raglogpilots,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=log.aiops.com,resources=raglogpilots/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=log.aiops.com,resources=raglogpilots/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RagLogPilot object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *RagLogPilotReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var raglogpilot logv1.RagLogPilot

	if err := r.Get(ctx, req.NamespacedName, &raglogpilot); err != nil {
		logger.Error(err, "unable to fetch RagLogPilot")

		return ctrl.Result{}, err
	}

	// 检查一下是否有conversationId
	if raglogpilot.Status.ConversationId == "" {
		// 创建新的对话
		conversationId, err := r.createNewConversation(raglogpilot)
		if err != nil {
			logger.Error(err, "unable to create new conversation")
			return ctrl.Result{}, err
		}

		// 把 ConversationId 更新到 Status里
		raglogpilot.Status.ConversationId = conversationId
		if err := r.Status().Update(ctx, &raglogpilot); err != nil {
			logger.Error(err, "unable to update RagLogPilot Status")
		}
	}

	// 获取目标 namespace 的所有pod
	var pods corev1.PodList
	if err := r.List(ctx, &pods, client.InNamespace(raglogpilot.Spec.WorkloadNameSpace)); err != nil {
		logger.Error(err, "unable to list pods")
		return ctrl.Result{}, err
	}

	for _, pod := range pods.Items {
		logString, err := r.getPogLogs(pod)
		if err != nil {
			logger.Error(err, "unable to get pod logs")
			continue
		}
		var errorlog []string
		logLines := strings.Split(logString, "\n")
		for _, line := range logLines {
			if strings.Contains(line, "ERROR") {
				errorlog = append(errorlog, line)
			}
		}

		if len(errorlog) > 0 {
			combinedErrorlog := strings.Join(errorlog, "\n")
			fmt.Println(combinedErrorlog)

			// 调用Ragflow API 寻求解决方案
			answer, err := r.queryRagSystem(combinedErrorlog, raglogpilot)
			if err != nil {
				logger.Error(err, "unable to query rag system")
				continue
			}

			err = r.sendFeishuAlert(raglogpilot.Spec.FeishuWebhook, answer)
			if err != nil {
				fmt.Println(err.Error())
			}
			logger.Info("RAG system response", "answer", answer)
		}
	}

	return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

func (r *RagLogPilotReconciler) sendFeishuAlert(webhook, analysis string) error {
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
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return nil
}

func (r *RagLogPilotReconciler) queryRagSystem(podLog string, ragLogPilot logv1.RagLogPilot) (string, error) {
	payload := map[string]interface{}{
		"session_id": ragLogPilot.Status.ConversationId,
		"question":   fmt.Sprintf("以下是获取到的日志：%s， 请基于运维专家知识库进行解答，如果不知道，就说不知道", podLog),
		"stream":     false,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/chats/%s/completions", ragLogPilot.Spec.RagFlowEndpoint, ragLogPilot.Spec.ChatId), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ragLogPilot.Spec.RaglowToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["data"].(map[string]interface{})["answer"].(string), nil
}

func (r *RagLogPilotReconciler) getPogLogs(pod corev1.Pod) (string, error) {
	tailLines := int64(20)
	logOptions := &corev1.PodLogOptions{TailLines: &tailLines}
	req := r.KubeClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, logOptions)

	// stream
	logSteam, err := req.Stream(context.TODO())
	if err != nil {
		return "", err
	}

	defer logSteam.Close()

	var logBuffer bytes.Buffer
	if _, err := logBuffer.ReadFrom(logSteam); err != nil {
		return "", err
	}
	return logBuffer.String(), nil
}

func (r *RagLogPilotReconciler) createNewConversation(ragLogPilot logv1.RagLogPilot) (string, error) {
	payload := map[string]interface{}{
		"name": fmt.Sprintf("ragflow-%s", uuid.NewString()[:6]),
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/chats/%s/sessions", ragLogPilot.Spec.RagFlowEndpoint, ragLogPilot.Spec.ChatId), bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ragLogPilot.Spec.RaglowToken))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result["data"].(map[string]interface{})["id"].(string), nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RagLogPilotReconciler) SetupWithManager(mgr ctrl.Manager) error {
	var kubeConfig *string
	var config *rest.Config

	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flag.String("kubeConfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path")
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", *kubeConfig); err != nil {
			return err
		}
	}

	r.KubeClient, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&logv1.RagLogPilot{}).
		Named("raglogpilot").
		Complete(r)
}
