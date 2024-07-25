# 使用官方的 Go 镜像作为构建环境
FROM golang:1.22.5 as builder

# 设置工作目录
WORKDIR /app

# 复制 go mod 和 sum 文件
COPY go.mod go.sum ./

# 下载所有依赖
RUN go mod download

# 复制源代码到容器中
COPY . .

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o myapp .

# 使用 Debian 基础镜像作为运行环境
FROM ubuntu:latest

# 设置工作目录
WORKDIR /root/

RUN apt-get update && \
  apt-get install -y ca-certificates && \
  update-ca-certificates

# 安装 ffmpeg
RUN apt-get update && \
  apt-get install -y ffmpeg && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/*

# 从构建环境中复制构建的可执行文件到当前容器
COPY --from=builder /app/myapp .

ENV PORT=9032
ENV APPLICATION_KAFKA_HOST=cognitube-kafka.servicebus.windows.net
ENV APPLICATION_KAFKA_PORT=9093
ENV APPLICATION_KAFKA_TOPIC=video-reencode
ENV KAFKA_EVENTHUB_NAMESPACE=cognitube-kafka
ENV KAFKA_EVENTHUB_NAME=cognitube
ENV KAFKA_USERNAME=\$ConnectionString
ENV VIDEO_CONTAINER_NAME=video-container
ENV AZURE_BLOB_CONNECTION_STRING="DefaultEndpointsProtocol=https;AccountName=cognitube;AccountKey=a1XDmr4IlO9I/tcsuh1akTaGFgmp+nQEoQdA8SlFpmmn7Zi0HKeDMk3ntxWjGI/HMFpQjzBys2ZX+AStqYVfsg==;EndpointSuffix=core.windows.net"
ENV KAFKA_EVENTHUB_CONNECTION_STRING="Endpoint=sb://cognitube-kafka.servicebus.windows.net/;SharedAccessKeyName=RootManageSharedAccessKey;SharedAccessKey=l9+PMVbv8R4LuCtQlPo5x8PIE8jZqn8O4+AEhEMQoqA="

# 暴露端口
EXPOSE 9032

# 运行应用程序
CMD ["./myapp"]
