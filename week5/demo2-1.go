package main

import (
  "context"
  "flag"
  "fmt"

  appsv1 "k8s.io/api/apps/v1"
  corev1 "k8s.io/api/core/v1"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/tools/clientcmd"
)

func main() {
  //加载 kubeconfig 配置
  kubeconfig := flag.String("kubeconfig", "/home/lostar/.kube/config", "location of kubeconfig file")
  config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
  if err != nil {
    fmt.Printf("error %s", err.Error())
  }

  // 创建 clientset
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    panic(err.Error())
  }

  // 定义 Nginx Deployment
  replicas := int32(1)
  deployment := &appsv1.Deployment{
    ObjectMeta: metav1.ObjectMeta{
        Name: "nginx-deployment",
    },
    Spec: appsv1.DeploymentSpec{
      Replicas: &replicas,
      Selector: &metav1.LabelSelector{
        MatchLabels: map[string]string {
          "app": "nginx",
	},
      },
      Template: corev1.PodTemplateSpec{
        ObjectMeta: metav1.ObjectMeta{
          Labels: map[string]string{
            "app": "nginx",
	  },
	},
	Spec: corev1.PodSpec{
          Containers: []corev1.Container{
            {
                Name: "nginx",
		Image: "nginx:1.14.2",
		Ports: []corev1.ContainerPort{
                    {
                      ContainerPort: 80,
		    },
		},
	    },
	  },
	},
      },
    },
  }

  // 创建 Deployment
  deploymentsClient := clientset.AppsV1().Deployments(corev1.NamespaceDefault)
  result, err := deploymentsClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
  if err != nil {
    panic(err)
  }

  fmt.Printf("Created deployment %q. \n", result.GetObjectMeta().GetName())
}
