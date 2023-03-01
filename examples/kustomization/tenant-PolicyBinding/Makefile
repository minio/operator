
all: build

build: miniosdkgo miniosdkjava awssdkpython

miniosdkdotnet:
	@cd miniosdk/dotnet/ && \
		docker buildx build -q --platform=linux/amd64 -t minio/operator-sts-example:miniosdk-dotnet .

miniosdkgo:
	@cd miniosdk/go/ && \
	  go install && GOOS=linux GOARCH=amd64 go build -o client
	@cd miniosdk/go/ && \
		docker buildx build -q --platform=linux/amd64 -t minio/operator-sts-example:miniosdk-go .

miniosdkjava:
	@cd miniosdk/java/ && \
	  mvn package
	@cd miniosdk/java/ && \
		docker buildx build -q --platform=linux/amd64 -t minio/operator-sts-example:miniosdk-java .

miniosdkjavascript:
	@cd miniosdk/javascript/ && \
		docker buildx build -q --platform=linux/amd64 -t minio/operator-sts-example:miniosdk-javascript .

awssdkpython:
	@cd awssdk/python/ && \
		docker buildx build -q --platform=linux/amd64 -t minio/operator-sts-example:awssdk-python .

miniosdkpython:
	@cd miniosdk/python/ && \
		docker buildx build -q --platform=linux/amd64 -t minio/operator-sts-example:miniosdk-python .