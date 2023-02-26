
FROM ubuntu:latest

RUN \
    apt-get update && \
    apt-get install -y curl ca-certificates golang-go

RUN mkdir /app
WORKDIR /app
COPY client /app/

CMD ["./app/client"]
