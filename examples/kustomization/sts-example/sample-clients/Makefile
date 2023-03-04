
all: build

build: miniosdkgo miniosdkjava awssdkpython

miniosdkdotnet:
	@cd minio-sdk/dotnet/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-dotnet . --output type=docker

miniosdkgo:
	@cd minio-sdk/go/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-go . --output type=docker

miniosdkjava:
	@cd minio-sdk/java/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-java . --output type=docker

miniosdkjavascript:
	@cd minio-sdk/javascript/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-javascript . --output type=docker

awssdkpython:
	@cd aws-sdk/python/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:aws-sdk-python . --output type=docker

miniosdkpython:
	@cd minio-sdk/python/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-python . --output type=docker
