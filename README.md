## 编译
make 
## 安装
	kubectl apply -f deployment.yaml
##
开启三个副本，
只有一个副本执行leader的逻辑。
### 1
开启一个terminal 执行
go run main.go  --kubeconfig=$HOME/.kube/config
### 2
开启一个terminal 执行
go run main.go  --kubeconfig=$HOME/.kube/config
### 3
开启一个terminal 执行
go run main.go  --kubeconfig=$HOME/.kube/config