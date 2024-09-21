###Instaall
```
#demo6 dir
helm install vote-test . --namespace test --create-namespace
```

###Result
```
# k get pod -n test
NAME                      READY   STATUS      RESTARTS   AGE
db-6d9f87bb9b-7z89k       1/1     Running     0          5s
helm-hook-demo-cgd8k      0/1     Completed   0          5s
redis-77fccb7f9-f7rvw     1/1     Running     0          5s
result-55585899f6-rx4sr   1/1     Running     0          5s
vote-944ffb7d-km8ld       1/1     Running     0          5s
worker-5878d7c555-jpvb9   1/1     Running     0          5s

### log for helm hook
# k logs -f helm-hook-demo-cgd8k -n test
this is helm hook pre-install job

```
