#### 安装cobra-cli
```bash copy
go install github.com/spf13/cobra-cli@latest
export PATH=~/go/bin:$PATH
```
#### Go init
```bash copy
go mod init "github.com/lostar01/k8scopilot"
```

#### Cobra-cli init
```bash copy
cobra-cli init --author "aiops" --license mit
```

#### 修改cmd/root.go
```
 var rootCmd = &cobra.Command{
	Use:   "k8scopilot",
	Short: "这是一个K8S 的 Pilot 工具",
	Long: `这是一个K8S 的 Pilot 工具，以下是一些用法示例
	1. xxx
	2. xxx
	`,
	Version: "v0.0.1",    #这里设置工具的版本
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}
```

#### 添加命令
```bash copy
cobra-cli add hello
```

#### 添加子命令
```bash copy
cobra-cli add world -p 'helloCmd'
```

#### 定义全局Flag
```
rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", defaultKubeconfig, "Path to kubeconfig")
rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "default", "The namespace to use")
```

#### 获取全局Flag 的值
全局直接调用变量

#### 在子命令定义Flag 
```
#该Flags 只在子命令生效
worldCmd.Flags().StringVarP(&source, "source", "s", "world", "The source of the message")
```

#### 标记过期子命令
```
var worldCmd = &cobra.Command{
	Use:        "world",
	Short:      "an demo for sub-command",
	Long:       `This is example for cobra sub command`,
	Deprecated: "This command is deprecated, please use hello instead",  #这行做过期标记
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kubeconfig:", kubeconfig)
		fmt.Println("namespapce:", namespace)
		fmt.Println("source:", source)
	},
```

#### Homework 
```
 ./k8scopilot ask chatgpt
我是 K8S Copilot, 请问有什么可以帮助你？
>查看 default NS 的deployment
你说： 查看 default NS 的deployment
deployment demo-1
deployment nginx-app

>删除 default NS 的 deployment，资源名称为 demo-1
你说： 删除 default NS 的 deployment，资源名称为 demo-1
>您即将删除 namespace: default 的 deployment demo-1,请确认是否删除(Y/N):n
取消删除 deployment demo-1
>删除 default NS 的 deployment，资源名称为 demo-1
你说： 删除 default NS 的 deployment，资源名称为 demo-1
>您即将删除 namespace: default 的 deployment demo-1,请确认是否删除(Y/N):y
成功删除 deployment demo-1
>查看 default NS 的deployment
你说： 查看 default NS 的deployment
deployment nginx-app
>exit
再见！
```