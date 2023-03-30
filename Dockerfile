FROM golang:1.20.2-alpine3.16 AS builder
WORKDIR /app
COPY . .
RUN go build -o bin/wechat-openai main.go

FROM alpine:3.16
WORKDIR /app
COPY --from=builder /app/bin/wechat-openai .
EXPOSE 8080
ENV TZ=Asia/Shanghai
CMD ["./wechat-openai", "-config", "config/config.yml"]
