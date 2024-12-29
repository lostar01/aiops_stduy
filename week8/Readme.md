1. 尝试修改实战三，并接入 Ollama 实现自托管大模型（Qwen2）推理
```
  llmEndpoint: "http://192.168.1.111:11434/v1"
  llmToken: "ollama"
  llmModel: "qwen2.5"
```

修改实战四，配置 RagFlow 接入 Ollama 实现自托管大模型推理
<img src=./img/ragflow.png/>

#### RagFlow 核心API
创建对话
》GET /api/new_conversation?user_id=xxx  
》data.id = conversation_id

基于运维知识库获取回复
》POST /api/completion
》conversation_id、message、stream=false
》回复： data.answer

#### Ragflow operator 
```
go mod init github.com/lostar01/rag-log-operator
kubebuilder init --domain=aiops.com
kubebuilder create api --group log --version v1 --kind RagLogPilot
```
