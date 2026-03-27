# ---------- builder ----------
FROM docker.m.daocloud.io/library/golang:1.24 AS builder

WORKDIR /app

# 使用国内源（可选）
RUN apt-get update \
    && apt-get install -y \
        pkg-config \
        libheif-dev \
        build-essential \
    && rm -rf /var/lib/apt/lists/*

# Go 代理
RUN go env -w GOPROXY=https://goproxy.cn,direct

COPY . .

RUN go mod download

# ⚠️ 必须开启 CGO
ENV CGO_ENABLED=1

RUN go build -o main .

# ---------- runtime ----------
FROM docker.m.daocloud.io/library/debian:bookworm-slim

WORKDIR /app

# 安装运行时依赖
RUN apt-get update \
    && apt-get install -y \
        libheif1 \
        ca-certificates \
        tzdata \
    && rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Shanghai

COPY --from=builder /app/main .

CMD ["./main"]