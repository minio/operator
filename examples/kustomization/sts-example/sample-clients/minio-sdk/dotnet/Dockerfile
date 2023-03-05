FROM ubuntu:22.10

RUN apt-get update && \
  apt-get install -y apt-transport-https \
  dotnet-sdk-7.0 curl ca-certificates

RUN curl https://packages.microsoft.com/config/ubuntu/22.10/packages-microsoft-prod.deb --output packages-microsoft-prod.deb && \
    dpkg -i packages-microsoft-prod.deb && \
    rm packages-microsoft-prod.deb

RUN mkdir app
WORKDIR /app
COPY dotnet.csproj /app
COPY Program.cs /app
RUN dotnet publish

ENTRYPOINT ["/app/bin/Debug/net7.0/dotnet"]
