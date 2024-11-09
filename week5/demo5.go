package main

import (
  "context"
  "fmt"

  "flag"
  corev1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/apimachinery/pkg/watch"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/tools/clientcmd"
)

func main() {
  kubeconfig := flag.String("kubeconfig","/home/lostar/.kube/config", "location of kubeconfig file")
  config, err := clientcmd.BuildConfigFromFlags("",*kubeconfig)
  if err != nil {
    panic(err)
  }

  clientset,_ := kubernetes.NewForConfig(config)

  timeOut := int64(60)
  watcher,_ := clientset.CoreV1().Pods("default").Watch(context.Background(), metav1.ListOptions{TimeoutSeconds: &timeOut})

  for event := range watcher.ResultChan() {
    item := event.Object.(*corev1.Pod)

    switch event.Type {
    case watch.Added:
	    processPod(item.GetName())
    case watch.Modified:
    case watch.Bookmark:
    case watch.Error:
    case watch.Deleted:
    }
  }

}

func processPod(name string) {
  fmt.Println("new Pod: ", name)
}
