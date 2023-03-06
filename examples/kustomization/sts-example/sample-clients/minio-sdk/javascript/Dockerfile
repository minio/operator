
FROM ubuntu:latest

RUN mkdir /app
WORKDIR /app
COPY client /app/
COPY start.sh /app/
RUN chmod +x /app/start.sh

RUN \
    apt-get update && \
    apt-get install -y curl ca-certificates golang-go

CMD ["./app/client"]
