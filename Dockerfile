FROM registry.cn-hangzhou.aliyuncs.com/minik8m/golang:alpine AS builder
ARG VERSION
ARG GIT_COMMIT
ARG MODEL
ARG API_KEY
ARG API_URL
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app
RUN CGO_ENABLED=0 go build -ldflags "-s -w  -X main.Version=$VERSION -X main.GitCommit=$GIT_COMMIT -X main.Model=$MODEL -X main.ApiKey=$API_KEY -X main.ApiUrl=$API_URL" \
    -o /app/k8m

FROM registry.cn-hangzhou.aliyuncs.com/minik8m/alpine:latest

WORKDIR /app
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache curl bash inotify-tools kubectl
ADD reload.sh /app/reload.sh
RUN chmod +x /app/reload.sh

COPY --from=builder /app/k8m /app/k8m
ENTRYPOINT ["/app/reload.sh","k8m","/app"]
