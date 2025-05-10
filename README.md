# SubPub

gRPC-сервис, реализующий паттерн публикации и подписки (Pub/Sub) с поддержкой потоковой передачи сообщений.

## Структура проекта

```
subpub/
├── cmd/
│   ├── config/
│   │   └── congig.go
│   └── server/
│       ├── config.yaml
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── api.go
│   │   ├── api_test.go
│   │   ├── subpub.go
│   │   └── subscription.go
│   └── gateways/
│       ├── generated/
│       │   ├── pubsub.pb.go
│       │   └── pubsub_grpc.pb.go       
│       ├── pubsub.proto
│       ├── server.go
│       └── pubsub.go
├── proto/
│   └── pubsub.proto
├── go.mod
└── README.md
```

## Конфигурация

Сервер использует следующие переменные окружения:

* `HTTP_HOST` — адрес прослушивания (по умолчанию `localhost`)
* `HTTP_PORT` — порт прослушивания (по умолчанию `8080`)

И ***.yaml*** конфиг:

```yaml
server:
  host: "localhost"
  port: 8080
log:
  level: "info"
```

## Запуск

1. Генерация gRPC-кода:

   ```bash
   protoc --go_out=./internal/gateways/generated/ \
          --go_opt=paths=source_relative \
          --go-grpc_out=./internal/gateways/generated/ \
          --go-grpc_opt=paths=source_relative \
          ./proto/pubsub.proto
   ```

2. Запуск сервера:

   ```bash
   go run ./cmd/main.go
   ```

### Публикация сообщения

```bash
grpcurl -plaintext -import-path proto -proto pubsub.proto \
  -d '{"key": "news", "data": "Новость"}' \
  localhost:8080 PubSub/Publish
```


### Подписка на поток

```bash
grpcurl -plaintext -import-path proto -proto pubsub.proto \
  -d '{"key": "news"}' \
  localhost:8080 PubSub/Subscribe
```
