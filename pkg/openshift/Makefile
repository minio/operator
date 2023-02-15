GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)

operator-job-linux-amd64:
	@echo "operator-job-linux-amd64"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build create-secret.go

operator-job-linux-arm64:
	@echo "operator-job-linux-amd64"
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build create-secret.go

docker-mac:
	@docker build -t minio/operator-job:v4.5.9 .
	@docker tag docker.io/minio/operator-job:v4.5.9 quay.io/cniackz4/operator-job:v4.5.9
	@docker push quay.io/cniackz4/operator-job:v4.5.9

docker-linux-amd64:
	@docker buildx build --platform linux/amd64 --tag minio/operator-job:v4.5.9 .
	@docker tag docker.io/minio/operator-job:v4.5.9 quay.io/cniackz4/operator-job:v4.5.9
	@docker push quay.io/cniackz4/operator-job:v4.5.9
