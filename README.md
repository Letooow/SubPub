# SubPub realisation

Сервис подписок на канал и публикаций сообщений:

1. Функция Subscribe возвращает подписку на канал публикаций.
2. С помощью функции Unsubscribe можно отписаться от канала.
3. Функция Publish делает публикацию в канал с именем "..." для всех подписчиков

### ***Реализован Graceful shutdown***

Сервис учитывает переменные окружения ***HTTP_HOST*** и ***HTTP_PORT*** для создания сервера.

Команда для генерации gRPC файлов:

___protoc  --go_out=./generated/ --go_opt=paths=source_relative --go-grpc_out=./generated/ --go-grpc_opt=paths=source_relative ./pubsub.proto___

