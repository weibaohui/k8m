FROM node:18-alpine  AS node-builder

WORKDIR /app

ADD ui .

RUN npm i -g pnpm && pnpm install && pnpm build

FROM golang:1.24-alpine  AS golang-builder

ENV GOPROXY="https://goproxy.io"

WORKDIR /app

ADD . .
COPY --from=node-builder /app/dist ./ui/dist

RUN sed -i "s/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g" /etc/apk/repositories \
    && apk upgrade && apk add --no-cache --virtual .build-deps \
    ca-certificates gcc g++ curl upx

RUN go build -o k8m . && upx -9 k8m

### build final image
FROM alpine:3.21

WORKDIR /app

ENV TZ=Asia/Shanghai

COPY --from=golang-builder /app/k8m .

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk upgrade && apk add --no-cache curl bash inotify-tools kubectl alpine-conf busybox-extras sqlite tzdata \
    && apk del alpine-conf && rm -rf /var/cache/* && chmod +x k8m

#k8m Server
EXPOSE 3618
#MCP Server
EXPOSE 3619 

CMD /app/k8m