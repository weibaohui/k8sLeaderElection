FROM golang:alpine as builder
WORKDIR /go/src/github.com/weibaohui/k8sLeaderElection
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags '-d -w -s ' -a -installsuffix cgo -o /app .


FROM busybox
WORKDIR /
COPY --from=builder /app .
EXPOSE 80
CMD ["./app","--kubeconfig="]