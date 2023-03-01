FROM ubuntu:latest

RUN \
    apt-get update && \
    apt-get install -y curl ca-certificates python3 python3-pip

RUN mkdir app
WORKDIR /app
COPY requirements.txt /app/requirements.txt
RUN pip3 install -r requirements.txt
COPY main.py /app/main.py
CMD ["python3", "/app/main.py"]
