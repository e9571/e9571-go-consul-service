# 第一阶段：构建 Go 程序 123
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -o app

# 第二阶段：运行环境
FROM alpine:3.18
WORKDIR /app
# 安装 CA 证书以支持 HTTPS 请求（Consul 健康检查可能需要）
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/app .
EXPOSE 3000
CMD ["./app"]