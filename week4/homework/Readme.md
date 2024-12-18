1. 用 Coze 实现个人运维知识库，并分享链接
https://www.coze.cn/store/agent/7449388012720717859?bot_id=true
<img src=./core.png>

2. 部署 RAGFlow，实现内部运维知识库（选做）
Ensure vm.max_map_count >= 262144:
```bash copy
echo 'vm.max_map_count=262144' >> /etc/sysctl.conf
sysctl -p
git clone https://github.com/infiniflow/ragflow.git
cd ragflow
docker compose -f docker/docker-compose.yml up -d
```
<img src=./Ragflow.png>

