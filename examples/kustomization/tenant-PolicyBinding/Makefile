
all: build

build: miniosdkgo miniosdkjava awssdkpython

miniosdkgo:
	@cd miniosdk/go/ && \
	  go install && GOOS=linux GOARCH=amd64 go build -o client
	@cd miniosdk/go/ && \
		docker buildx build --platform=linux/amd64 -t minio/operator-sts-example:miniosdk-go .

miniosdkjava:
	@cd miniosdk/java/ && \
	  mvn package
	@cd miniosdk/java/ && \
		docker buildx build --platform=linux/amd64 -t minio/operator-sts-example:miniosdk-java .

awssdkpython:
	@cd awssdk/python/ && \
		docker buildx build --platform=linux/amd64 -t minio/operator-sts-example:awssdk-python .
