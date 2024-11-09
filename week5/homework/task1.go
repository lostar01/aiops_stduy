package main

import (
  "flag"
  "fmt"
  "path/filepath"
  "time"

  v1 "k8s.io/api/core/v1"
  "k8s.io/client-go/informers"
  "k8s.io/client-go/kubernetes"
  "k8s.io/client-go/rest"
  "k8s.io/client-go/tools/cache"
  "k8s.io/client-go/tools/clientcmd"
  "k8s.io/client-go/util/homedir"
  "k8s.io/client-go/util/workqueue"
)

// 这两个先定义
type Controller struct {
  indexer cache.Indexer
  queue   workqueue.TypedRateLimitingInterface[string]
  informer cache.Controller
}

func NewController(queue workqueue.TypedRateLimitingInterface[string], indexer cache.Indexer, informer cache.Controller) *Controller {
  return &Controller {
    informer: informer,
    indexer: indexer,
    queue: queue,
  }
}

// 然后定义 main

// 处理下一个
func(c *Controller) processNextItem() bool {
  key, quit := c.queue.Get()
  if quit {
    return false
  }

  defer c.queue.Done(key)

  err := c.syncToStdout(key)
  c.handleErr(err,key)
  return true
}

// 输出日志
func (c *Controller) syncToStdout(key string) error {
  // 通过 key 从 indexer 中获取完整的对象
  obj, exists, err := c.indexer.GetByKey(key)
  if err != nil {
    fmt.Printf("Fetching object with key %s from store failed with %v\n",key,err)
    return err
  }

  if !exists {
    fmt.Printf("Pod %s does not exist anymore \n", key)
  } else {
    pod := obj.(*v1.Pod)
    fmt.Printf("Sync/Add/Update for Pod %s \n", pod.Name)
    if pod.Name == "test-pod" {
      time.Sleep(2 * time.Second)
      return fmt.Errorf("simulated error for pod %s", pod.Name)
    }
  }
  return nil
}

// 错误处理
func (c *Controller) handleErr(err error, key string) {
  if err == nil {
    c.queue.Forget(key)
    return
  }

  if c.queue.NumRequeues(key) < 5 {
    fmt.Printf("Retry %d for key %s\n", c.queue.NumRequeues(key), key)
    // 重新加入队列，并且进行速率限制, 这里会让他过一段时间才会被处理，避免过度重启
    c.queue.AddRateLimited(key)
    return
  }

  c.queue.Forget(key)
  fmt.Printf("Dropping pod %q out of the queue: %v\n", key, err)
}

func main() {
  var err error
  var config *rest.Config

  var kubeconfig *string

  if home := homedir.HomeDir(); home != "" {
    kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "[可选] kubeconfig 绝对路径")
  } else {
    kubeconfig = flag.String("kubeconfig", "", "kubeconfig 绝对路径")
  }

  // 初始化 rest.Config 对象
  if config, err = rest.InClusterConfig(); err != nil {
    if config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig); err != nil {
      panic(err.Error())
    }
  }

  // 创建Clientset 对象
  clientset, err := kubernetes.NewForConfig(config)
  if err != nil {
    panic(err.Error())
  }

  // 初始化 informer factory
  informerFactory := informers.NewSharedInformerFactory(clientset, time.Hour*12)

  // 创建速率限制队列
  queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[string]())

  // 对 Pod 监听
  podInformer := informerFactory.Core().V1().Pods()
  informer := podInformer.Informer()
  informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
	  AddFunc: func(obj interface{}) { onAddPod(obj, queue)},
	  UpdateFunc: func(old, new interface{}) { onUpdatePod(new,queue) },
	  DeleteFunc: func(obj interface{}) { onDeletePod(obj, queue)},
  })

  controller := NewController(queue, podInformer.Informer().GetIndexer(), informer)

  stopper := make(chan struct{})
  defer close(stopper)

  // 启动 informer , List & Watch
  informerFactory.Start(stopper)
  informerFactory.WaitForCacheSync(stopper)

  // 处理队列中的事件
  go func() {
    for {
      if !controller.processNextItem() {
        break
      }
    }
  }()

  <-stopper
}

func onAddPod(obj interface{}, queue workqueue.TypedRateLimitingInterface[string]) {
  // 生成 key
  key, err := cache.MetaNamespaceKeyFunc(obj)
  if err != nil {
    queue.Add(key)
  }
}

func onUpdatePod(new interface{}, queue workqueue.TypedRateLimitingInterface[string]) {
  key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(new)
  if err == nil {
    queue.Add(key)
  }
}

func onDeletePod(obj interface{}, queue workqueue.TypedRateLimitingInterface[string]) {
  key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
  if err == nil {
    queue.Add(key)
  }
}
