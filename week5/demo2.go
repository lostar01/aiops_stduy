package main
import (
  "context"
  "flag"
  "fmt"

  corev1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/kubernetes/scheme"
  "k8s.io/client-go/rest"
  "k8s.io/client-go/tools/clientcmd"
)

func main() {
  // 生成 kubeconfig 配置
  kubeconfig := flag.String("kubeconfig","/home/lostar/.kube/config", "location of kubeconfig file")
  config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
  if err != nil {
    panic(err)
  }

  // 指定Kubernetes API 路径
  config.APIPath = "api"

  // 指定要使用的API组和版本
  config.GroupVersion = &corev1.SchemeGroupVersion

  // 指定编译器
  config.NegotiatedSerializer = scheme.Codecs

  // 生成RESTClient实例
  restClient, err := rest.RESTClientFor(config)
  if err != nil {
    panic(err)
  }

  // 创建空的结构体，存储pod 列表
  podList := &corev1.PodList{}

  // 构建HTTP 请求参数, 直接定义 GVR
  // Get请求

  restClient.Get().
    // 指定要获取的资源类型
    Namespace("kube-system").
    // 指定要获取的资源类型
    Resource("pods").
    // 设置请求参数，使用metav1.ListOptions结构体设置了Limit 参数为500，并使用schema.ParameterCodec 进行参数编码
    VersionedParams(&metav1.ListOptions{Limit: 500}, scheme.ParameterCodec).
    // 发送请求并获取响应， 使用context.TODO() 作为上下文
    Do(context.TODO()).
    Into(podList)

  for _, v := range podList.Items {
    fmt.Printf("Namespace: %v  Name: %v  Status:  %v \n", v.Namespace, v.Name, v.Status.Phase)
  }
}
