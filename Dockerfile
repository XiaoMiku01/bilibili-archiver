FROM golang:latest AS builder

ARG VERSION
WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -trimpath \
    -ldflags "-s -w -X main.Version=${VERSION}" \
    -o bilibili-archiver main.go

FROM alpine:latest

# 添加构建参数接收宿主机用户ID和组ID
ARG USER_ID=1000
ARG GROUP_ID=1000

RUN apk update && apk add --no-cache ffmpeg shadow bash

# 创建与宿主机UID/GID匹配的用户
RUN groupadd -g ${GROUP_ID} appgroup && \
    useradd -u ${USER_ID} -g appgroup -s /bin/bash -m appuser

WORKDIR /app
COPY --from=builder /app/bilibili-archiver /app/bilibili-archiver
RUN chown appuser:appgroup /app/bilibili-archiver

WORKDIR /data
# 设置目录权限
RUN chown -R appuser:appgroup /data
RUN chsh -s /bin/bash appuser
# 切换到新创建的用户
USER appuser

ENTRYPOINT ["/app/bilibili-archiver"]
CMD ["--help"]