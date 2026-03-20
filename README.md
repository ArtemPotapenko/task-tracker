# Task Tracker

Небольшой набор Go-микросервисов для управления задачами.

Сервисы:
- `account-service` — регистрация и логин
- `task-service` — создание и получение задач
- `email-service` — отправка welcome и daily summary писем
- `scheduler-service` — периодическая обработка просроченных задач
- `http-gateway` — HTTP API поверх gRPC

## Запуск

Поднять весь стек:

```bash
docker compose up --build
```

HTTP gateway:

```text
http://localhost:8080
```

Swagger UI:

```text
http://localhost:8080/swagger/account
http://localhost:8080/swagger/task
```

Для task API используется заголовок:

```text
Authorization: Bearer <JWT>
```

## Основные сценарии

Регистрация:

```bash
curl -X POST http://localhost:8080/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"password123","repeatPassword":"password123"}'
```

Логин:

```bash
curl -X POST http://localhost:8080/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","password":"password123"}'
```

Получить задачи на сегодня:

```bash
curl http://localhost:8080/v1/tasks/today \
  -H 'Authorization: Bearer <JWT>'
```

Создать задачу:

```bash
curl -X POST http://localhost:8080/v1/tasks \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <JWT>' \
  -d '{"description":"ship feature","dueDate":1760000000}'
```

## Тесты

Unit-тесты:

```bash
make test-unit
```

Интеграционные тесты:

```bash
make test-integration
```

E2E-тесты:

```bash
make test-e2e
```

Проверка сборки всех пакетов:

```bash
make build
```

Линтер:

```bash
make lint
```

## Proto и OpenAPI

Перегенерировать весь protobuf-код:

```bash
make proto
```

Перегенерировать только внешний API, gateway и OpenAPI:

```bash
make proto-external
```
