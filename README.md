# wechat-openai

微信公众号对接 OpenAI ChatGPT

## Quick Start

修改 [config.yml](./config.yml) 和 [docker-compose.yml](./docker-compose.yml) 中的配置

```bash
docker compose up -d
```

## Storage

支持内存存储和 Redis 存储

```yaml
storage: "memory" # memory or redis
redis:
  host: "redis"
  port: 6379
  password: ""
```

## License

[MIT](./LICENSE)
