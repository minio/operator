
all: build

build: miniosdkgo miniosdkjava awssdkpython

miniosdkdotnet:
	@cd miniosdk/dotnet/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-dotnet .

miniosdkgo:
	@cd miniosdk/go/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-go .

miniosdkjava:
	@cd miniosdk/java/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-java .

miniosdkjavascript:
	@cd miniosdk/javascript/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-javascript .

awssdkpython:
	@cd awssdk/python/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:aws-sdk-python .

miniosdkpython:
	@cd miniosdk/python/ && \
		docker buildx build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-python .