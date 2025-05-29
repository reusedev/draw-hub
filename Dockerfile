# 构建阶段
FROM golang:1.23-alpine AS builder

WORKDIR /app

# 复制源代码
COPY . .

# 安装依赖
RUN go mod download

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o draw-hub .

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata
# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=builder /app/draw-hub .