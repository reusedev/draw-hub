# ---------- builder ----------
FROM docker.m.daocloud.io/library/golang:1.25 AS builder

WORKDIR /app

# ✅ 替换 Debian 源（阿里云）
RUN sed -i 's|deb.debian.org|mirrors.aliyun.com|g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y \
        pkg-config \
        libheif-dev \
        build-essential \
    && rm -rf /var/lib/apt/lists/*

# ✅ Go 国内代理
RUN go env -w GOPROXY=https://goproxy.cn,direct \
    && go env -w GOSUMDB=sum.golang.google.cn

# ✅ GitHub 加速（关键！）
RUN git config --global url."https://ghproxy.com/https://github.com/".insteadOf "https://github.com/"

COPY . .

RUN go mod download

# ⚠️ 必须开启 CGO
ENV CGO_ENABLED=1

RUN go build -ldflags="-s -w" -o main .

# ---------- runtime ----------
FROM docker.m.daocloud.io/library/debian:bookworm-slim

WORKDIR /app

# ✅ 替换 Debian 源
RUN sed -i 's|deb.debian.org|mirrors.aliyun.com|g' /etc/apt/sources.list.d/debian.sources \
    && apt-get update \
    && apt-get install -y \
        libheif1 \
        ca-certificates \
        tzdata \
    && rm -rf /var/lib/apt/lists/*

ENV TZ=Asia/Shanghai

COPY --from=builder /app/main .

CMD ["./main"]