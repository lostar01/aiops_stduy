package main

import (
  "context"
  "flag"
  "fmt"
  "os"
  "path/filepath"

  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/runtime/schema"
  "k8s.io/client-go/dynamic"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/restmapper"
  "k8s.io/client-go/tools/clientcmd"
  "k8s.io/client-go/util/homedir"
)

func main() {
  // 解析命令行参数
  if len(os.Args) != 3 {
    fmt.Printf("Usage: %s get <resource>\n", os.Args[0])
    os.Exit(1)
  }

  command := os.Args[1]
  kind := os.Args[2]

  if command != "get" {
    fmt.Println("Unsupported command:", command)
    os.Exit(1)
  }

  // 加载 kubeconfig 配置
  var kubeconfig *string

  if home := homedir.HomeDir(); home != "" {
    kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "[optional] absolute path to the kubeconfig file")
  } else {
    kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
  }

  flag.Parse()

  config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
  if err != nil {
    panic(err.Error())
  }

  // 创建 dynamic client
  dynamicClient, err := dynamic.NewForConfig(config)
  if err != nil {
    panic(err)
  }

  // 获取客户端和映射器
  clientset, err := kubernetes.NewForConfig(config)

  discoveryClient := clientset.Discovery()
  apiGroupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
  if err != nil {
    panic(err)
  }

  mapper := restmapper.NewDiscoveryRESTMapper(apiGroupResources)

  // 动态映射 kind 到 GVR
  // gvk := schema.FromAPIVersionAndKind("mygroup.example.com/v1alpha1", kind)
  // 还可以用这个方法
  gvk := schema.GroupVersionKind{
    Group: "mygroup.example.com",
    Version: "v1alpha1",
    Kind: kind,
  }

  mapping, err := mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
  if err != nil {
    panic(err)
  }
  // mapping.Resource 就是 GVR， 这样就实现 GVK -> GVR 的转化

  // 获取资源
  resourceInterface := dynamicClient.Resource(mapping.Resource).Namespace("default")
  resources, err := resourceInterface.List(context.TODO(), metav1.ListOptions{})
  if err != nil {
    panic(err)
  }

  // 打印资源
  for _, resource := range resources.Items {
    fmt.Printf("Name: %s, Namespace: %s, UID: %s\n", resource.GetName(), resource.GetNamespace(), resource.GetUID())
  }
}
