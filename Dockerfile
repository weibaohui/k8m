FROM golang:alpine AS builder
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

FROM alpine:latest
COPY --from=builder /app/k8m /usr/local/bin/
CMD ["k8m"]