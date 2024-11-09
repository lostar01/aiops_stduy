package main

import (
  "context"
  "fmt"
  // "flag"

  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/rest"
  // "k8s.io/client-go/tools/clientcmd"
)

func main() {
  // kubeconfig := flag.String("kubeconfig","/home/lostar/.kube/config","location of kubeconfig file")
  // config,err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
  // if err != nil {
  //   fmt.Printf("error %s", err.Error())
  // }

  // demo2 in cluster config
  config, err := rest.InClusterConfig()
  if err != nil {
    fmt.Printf("error %s", err.Error())
  }

  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    fmt.Printf("error %s", err.Error())
  }

  deployment,err := clientset.AppsV1().Deployments("default").List(context.Background(),metav1.ListOptions{})

  for _, d := range deployment.Items {
    fmt.Printf("deployment name %s\n", d.Name)
  }
}
