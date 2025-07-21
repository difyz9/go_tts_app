# 多阶段构建 - 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o tts_app .

# 最终镜像 - 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1001 -S tts && \
    adduser -u 1001 -S tts -G tts

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/tts_app .

# 复制配置文件模板
COPY config.yaml ./config.yaml.example

# 创建必要的目录
RUN mkdir -p output temp && \
    chown -R tts:tts /app

# 切换到非root用户
USER tts

# 暴露端口（如果将来添加Web接口）
EXPOSE 8080

# 设置环境变量
ENV TTS_CONFIG_FILE=/app/config.yaml
ENV TTS_INPUT_FILE=/app/input.txt
ENV TTS_OUTPUT_DIR=/app/output

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ./tts_app --help || exit 1

# 默认命令
ENTRYPOINT ["./tts_app"]
CMD ["edge", "--help"]

# 元数据标签
LABEL maintainer="your-email@example.com"
LABEL description="TTS语音合成应用 - 支持腾讯云TTS和Edge TTS"
LABEL version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/difyz9/go-tts-app"
LABEL org.opencontainers.image.description="高性能TTS语音合成应用，支持双引擎、并发处理"
LABEL org.opencontainers.image.licenses="MIT"
