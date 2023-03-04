
all: build

build: miniosdkgo miniosdkjava awssdkpython

miniosdkdotnet:
	@cd minio-sdk/dotnet/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-dotnet .

miniosdkgo:
	@cd minio-sdk/go/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-go .

miniosdkjava:
	@cd minio-sdk/java/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-java .

miniosdkjavascript:
	@cd minio-sdk/javascript/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-javascript .

awssdkpython:
	@cd aws-sdk/python/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:aws-sdk-python .

miniosdkpython:
	@cd minio-sdk/python/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-python .
