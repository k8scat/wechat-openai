FROM golang:1.20.2 AS builder
WORKDIR /app
COPY . .
RUN go build -o bin/wechat-openai main.go

FROM scratch
WORKDIR /app
COPY --from=builder /app/bin/wechat-openai .
EXPOSE 8080
ENTRYPOINT ["./wechat-openai", "-config", "config/config.yml"]
