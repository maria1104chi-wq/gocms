# Go 后端 Dockerfile
# 文件名: Dockerfile
# 路径: /workspace/backend/Dockerfile

FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装必要的构建依赖
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod ./

# 下载依赖 (由于没有go.sum，先跳过)
# RUN go mod download

# 复制源代码
COPY . .

# 编译应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# 生产镜像
FROM alpine:latest

WORKDIR /app

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 从builder阶段复制编译好的二进制文件
COPY --from=builder /app/main .

# 创建上传目录
RUN mkdir -p /app/static/uploads

# 暴露端口
EXPOSE 8080

# 设置时区
ENV TZ=Asia/Shanghai

# 启动应用
CMD ["./main"]
