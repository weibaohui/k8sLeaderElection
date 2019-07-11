all: build push deploy

build:
	docker build -t weibh/k8s-leader-election .
push:
	docker push weibh/k8s-leader-election
deploy:
	kubectl apply -f deployment.yaml
