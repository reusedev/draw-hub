# 构建阶段
FROM docker.m.daocloud.io/library/golang:1.24-alpine AS builder

WORKDIR /app

# 配置 Go 模块代理和 Alpine 镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && go env -w GOPROXY=https://goproxy.cn,direct

# 复制源代码
COPY . .

# 安装依赖
RUN go mod download

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 运行阶段
FROM docker.m.daocloud.io/library/alpine:latest

WORKDIR /app

# 配置 Alpine 镜像源并安装必要的运行时依赖
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk update \
    && apk add --no-cache ca-certificates tzdata curl
# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .