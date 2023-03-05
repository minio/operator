
all: build

build: miniosdkgo miniosdkjava awssdkpython

miniosdkdotnet:
	@cd minio-sdk/dotnet/ && \
		docker build -q -t miniodev/operator-sts-example:minio-sdk-dotnet .

miniosdkgo:
	@cd minio-sdk/go/ && \
		docker build -q -t miniodev/operator-sts-example:minio-sdk-go .

miniosdkjava:
	@cd minio-sdk/java/ && \
		docker build -q -t miniodev/operator-sts-example:minio-sdk-java .

miniosdkjavascript:
	@cd minio-sdk/javascript/ && \
		docker build -q -t miniodev/operator-sts-example:minio-sdk-javascript .

awssdkpython:
	@cd aws-sdk/python/ && \
		docker build -q -t miniodev/operator-sts-example:aws-sdk-python .

miniosdkpython:
	@cd minio-sdk/python/ && \
		docker build -q --platform=linux/amd64 -t miniodev/operator-sts-example:minio-sdk-python .
