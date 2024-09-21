### Install local k3s & crossplane controller
```
terraform plan
terraform apply
```

### create redis related source
```
kubectl apply -f yaml/tencent-vpc.yaml
kubectl apply -f yaml/tencent-subnet.yaml
kubectl apply -f yaml/tencent-securitygrop.yaml

##创建完 vpC 与子网和安全组后需要重新配置 Redis 定义
kubectl apply -f yaml/tencent-redis-instance.yaml

```
#### 个人使用认为crossplane 对相关的资源支持并不是很完善，使用起来可能稍微麻烦！


