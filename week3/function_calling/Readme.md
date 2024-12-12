### Function_caling
#### 实践1：
```
python3 main.py 
输入指令:帮我修改gateway 的配置 ， vendor 修改 aliply

 ChatGPT want to call function： [ChatCompletionMessageToolCall(id='call_yoqq4l2n', function=Function(arguments='{"key":"vendor","service_name":"gateway","value":"aliply"}', name='modify_config'), type='function', index=0)]
{'key': 'vendor', 'service_name': 'gateway', 'value': 'aliply'}

 函数调用的参数： gateway vendor aliply
LLM Res： 你需要使用配置管理工具来修改gateway的配置文件。这里假设你使用的是etcd作为配置存储。

首先，你需要连接到etcd，并找到gateway配置文件：


# 使用etcdctl客户端连接到etcd
etchectl get /gateway/config


输出如下：

{
  "service": "gateway",
  "key": "vendor",
  "value": "ly/aliply"
}


然后，你可以使用以下命令修改vendor的值：

# 使用etcdctl客户端更新vendor的值
etchectl set /gateway/config/vendor aliply


或者，如果你想使用Python命令，您需要在代码中调用etcd API：

###python
import etcd3

client = etcd3.Client()
config = client.get('/gateway/config')
config = dict(config)

# 修改vendor的值
config['vendor'] = 'aliply'

# 更新配置
client.put('/gateway/config', config)


在上述示例中，`etchectl set`命令或者Python代码都是用于修改_gateway配置文件的。当你确认配置正确后，就会成功修改vendor的值为aliply。
```

#### 实践2
```
python3 main.py 
输入指令:帮我重启 gateway 服务

 ChatGPT want to call function： [ChatCompletionMessageToolCall(id='call_mqyfxu0x', function=Function(arguments='{"service_name":"gateway"}', name='restart_service'), type='function', index=0)]
{'service_name': 'gateway'}

 准备重启  gateway
LLM Res： .gateway服务已经成功重启。
```

#### 实践3
```
python3 main.py 
输入指令:帮我部署一个deployment，镜像是nginx

 ChatGPT want to call function： [ChatCompletionMessageToolCall(id='call_fl8azpap', function=Function(arguments='{"image":"nginx","resource_type":"deployment"}', name='apply_manifest'), type='function', index=0)]
{'image': 'nginx', 'resource_type': 'deployment'}

 应用manifest deployment nginx
LLM Res： 这个命令会创建一个名为“test-deployment”的Deployment，它使用docker镜像作为容器，镜像中包含了Nginx服务。然后它会自动创建和应用manifest（清单文件），从而实现部署。

在该命令后，你可以通过 `kubectl get pods` 来查看是否已经有 pod 被创建，并且 `get deployment` 来查看deployment的状态。
```