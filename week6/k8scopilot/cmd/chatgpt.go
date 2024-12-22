/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lostar01/k8scopilot/utils"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/restmapper"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// chatgptCmd represents the chatgpt command
var chatgptCmd = &cobra.Command{
	Use:   "chatgpt",
	Short: "use chatgpt interaction",
	Long:  `can use this command interact with chatgpt`,
	Run: func(cmd *cobra.Command, args []string) {
		startChat()
	},
}

func startChat() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("我是 K8S Copilot, 请问有什么可以帮助你？")

	for {
		fmt.Printf(">")
		if scanner.Scan() {
			input := scanner.Text()
			if input == "exit" {
				fmt.Println("再见！")
				break
			}
			if input == "" {
				continue
			}

			fmt.Println("你说：", input)
			response := processInput(input)
			fmt.Println(response)
		}
	}
}

func processInput(input string) string {
	client, err := utils.NewOpenAIClient()
	if err != nil {
		return err.Error()
	}

	// response, err := client.SendMessage("你现在是一个 K8S Copilot, 你要帮助用户生成 YAML 文件内容，仅输出 YAML 里面的内容即可，不要把 YAML 放在 ```代码块里```", input)
	response := functionCalling(input, client)
	return response
}

func functionCalling(input string, client *utils.OpenAI) string {
	// 定义第一个函数, 部署K8S资源
	f1 := openai.FunctionDefinition{
		Name:        "generateK8sResource",
		Description: "生成和部署K8S资源",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"namespace": {
					Type:        jsonschema.String,
					Description: "资源所在的命名空间",
				},
				"resource_type": {
					Type:        jsonschema.String,
					Description: "K8S 标准资源类型，例如： Pod、 Deployment、Service 等",
				},
				"resource_name": {
					Type:        jsonschema.String,
					Description: "资源名称，例如： nginx-app 等",
				},
				"image": {
					Type:        jsonschema.String,
					Description: "镜像名称",
				},
			},
			Required: []string{"namespace", "resource_type", "resource_name", "image"},
		},
	}

	t1 := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f1,
	}

	// 定义查询 K8S 资源
	f2 := openai.FunctionDefinition{
		Name:        "queryResource",
		Description: "查询 K8S 资源",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"namespace": {
					Type:        jsonschema.String,
					Description: "资源所在的命名空间",
				},
				"resource_type": {
					Type:        jsonschema.String,
					Description: "K8S 标准资源类型，例如： Pod、 Deployment、Service 等",
				},
			},
			Required: []string{"namespace", "resource_type"},
		},
	}

	t2 := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f2,
	}

	// 用来删除 K8S 资源
	f3 := openai.FunctionDefinition{
		Name:        "deleteResource",
		Description: "删除 K8S 资源",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"namespace": {
					Type:        jsonschema.String,
					Description: "资源所在的命名空间",
				},
				"resource_type": {
					Type:        jsonschema.String,
					Description: "K8S 标准资源类型，例如： Pod、 Deployment、Service 等",
				},
				"resource_name": {
					Type:        jsonschema.String,
					Description: "K8S 资源名称，例如： nginx等",
				},
			},
			Required: []string{"namespace", "resource_type", "resource_name"},
		},
	}

	t3 := openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &f3,
	}

	dialogue := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleUser, Content: input},
	}

	resp, err := client.Client.CreateChatCompletion(context.TODO(),
		openai.ChatCompletionRequest{
			// Model: openai.GPT4o,
			Model:    "qwen2.5",
			Messages: dialogue,
			Tools:    []openai.Tool{t1, t2, t3},
		},
	)
	if err != nil {
		return err.Error()
	}

	msg := resp.Choices[0].Message
	if len(msg.ToolCalls) != 1 {
		return fmt.Sprintf("未找到合适的工具调用， %v", len(msg.ToolCalls))
	}

	// 组装对话历史
	dialogue = append(dialogue, msg)
	// return fmt.Sprintf("OpenAI 希望能请求函数 %s, 参数 %s", msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)

	callResult, err := callFunction(client, msg.ToolCalls[0].Function.Name, msg.ToolCalls[0].Function.Arguments)
	if err != nil {
		return err.Error()
	}
	return callResult
}

func callFunction(client *utils.OpenAI, name, arguments string) (string, error) {
	if name == "generateK8sResource" {
		params := struct {
			Namespace     string `json:"namespace"`
			Resource_name string `json:"resource_name"`
			Resource_type string `json:"resource_type"`
			Image         string `json:"image"`
		}{}
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return "", err
		}
		return generateK8sResource(client, params.Namespace, params.Resource_type, params.Resource_name, params.Image)
	}
	if name == "queryResource" {
		params := struct {
			Namespace     string `json:"namespace"`
			Resource_type string `json:"resource_type"`
		}{}
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return "", err
		}
		return queryResource(params.Namespace, params.Resource_type)
	}
	if name == "deleteResource" {
		params := struct {
			Namespace     string `json:"namespace"`
			Resource_type string `json:"resource_type"`
			Resource_name string `json:"resource_name"`
		}{}
		if err := json.Unmarshal([]byte(arguments), &params); err != nil {
			return "", err
		}
		return deleteResource(params.Namespace, params.Resource_type, params.Resource_name)
	}
	return "", nil
}

func generateK8sResource(client *utils.OpenAI, namespace string, resource_type string, resource_name string, image string) (string, error) {

	yamlContent, err := client.SendMessage("你现在是一个 K8S 资源生成器，你要帮助用户生成 YAML 文件内容，仅输出 YAML 里面的内容即可，不要把 YAML 放在 ```代码块里```",
		fmt.Sprintf("帮我生成一个%s资源，Namespace 为%s, 资源名称为%s,镜像名为%s", resource_type, namespace, resource_name, image))
	yamlContent = strings.Replace(yamlContent, "```yaml", "#", -1)
	yamlContent = strings.Replace(yamlContent, "```", "#", -1)
	if err != nil {
		return "", nil
	}

	// return yamlContent, nil
	// return fmt.Sprintf("帮我生成一个%s资源，Namespace 为%s, 资源名称为%s,镜像名为%s", resource_type, namespace, resource_name, image), nil
	clientGo, err := utils.NewClientGo(kubeconfig)
	if err != nil {
		return "", err
	}

	resources, err := restmapper.GetAPIGroupResources(clientGo.DiscoveryClient)
	if err != nil {
		return "", err
	}
	// 把 YAML 转成 Unstructured
	unstructuredObj := &unstructured.Unstructured{}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode([]byte(yamlContent), nil, unstructuredObj)
	if err != nil {
		return "", nil
	}

	// 创建 mapper
	mapper := restmapper.NewDiscoveryRESTMapper(resources)
	// 从 unstructuredObj 中获取GVK
	gvk := unstructuredObj.GroupVersionKind()
	mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return "", err
	}

	namespace = unstructuredObj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	// 使用 dynamicClient 创建资源
	_, err = clientGo.DynamicClient.Resource(mapping.Resource).Namespace(namespace).Create(context.TODO(), unstructuredObj, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("资源 %s 创建成功", unstructuredObj.GetName()), nil
}

func queryResource(namespace, resourceType string) (string, error) {
	clientGo, err := utils.NewClientGo(kubeconfig)
	if err != nil {
		return "", err
	}
	resourceType = strings.ToLower(resourceType)

	var gvr schema.GroupVersionResource

	switch resourceType {
	case "deployment":
		gvr = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case "service":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "pod":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	default:
		return "", fmt.Errorf("不支持的资源类型 %s", resourceType)
	}

	// 通过 dynamiClient 获取资源
	resourceList, err := clientGo.DynamicClient.Resource(gvr).Namespace(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	result := ""
	for _, item := range resourceList.Items {
		result += fmt.Sprintf("%s %s\n", resourceType, item.GetName())
	}

	return result, nil
}

func deleteResource(namespace, resourceType, resource_name string) (string, error) {
	clientGo, err := utils.NewClientGo(kubeconfig)
	if err != nil {
		return "", err
	}
	resourceType = strings.ToLower(resourceType)

	var gvr schema.GroupVersionResource

	switch resourceType {
	case "deployment":
		gvr = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}
	case "service":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	case "pod":
		gvr = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	default:
		return "", fmt.Errorf("不支持的资源类型 %s", resourceType)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf(">您即将删除 namespace: %s 的 %s %s,请确认是否删除(Y/N):", namespace, resourceType, resource_name)
		if scanner.Scan() {
			input := scanner.Text()
			if input == "Y" || input == "y" {
				// 通过 dynamiClient 获取资源
				err = clientGo.DynamicClient.Resource(gvr).Namespace(namespace).Delete(context.TODO(), resource_name, metav1.DeleteOptions{})
				if err != nil {
					return "", err
				}
				break
			} else if input == "N" || input == "n" {
				return fmt.Sprintf("取消删除 %s %s", resourceType, resource_name), nil
			} else {
				continue
			}
		}
	}

	return fmt.Sprintf("成功删除 %s %s", resourceType, resource_name), nil
}

func init() {
	askCmd.AddCommand(chatgptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// chatgptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// chatgptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
