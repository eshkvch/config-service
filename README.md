# Environment Config Service

Сервис для управления конфигурациями различных окружений с использованием чистой архитектуры.

## Технологии

- Go 1.24
- PostgreSQL
- Uber FX для dependency injection
- Uber Zap для логирования
- Go Playground Validator для валидации конфигов
- Joho Godotenv для загрузки переменных окружения

## API Endpoints

### Health Check
- `GET /health` - Проверка работоспособности сервиса

### Config Management
- `POST /configs/{env}/{key}` - Создание новой конфигурации
- `GET /configs/{env}/{key}` - Получение конфигурации
- `GET /configs/{env}` - Получение всех конфигураций для окружения
- `PUT /configs/{env}/{key}` - Обновление конфигурации
- `DELETE /configs/{env}/{key}` - Удаление конфигурации

### Swagger Documentation
- `GET /swagger.json` - Swagger документация в формате JSON
- `GET /swagger.yaml` - Swagger документация в формате YAML

**Для просмотра Swagger UI:**
1. Откройте https://editor.swagger.io/
2. Вставьте URL: `http://localhost:8080/swagger.yaml` (если сервис запущен локально)
3. Или используйте файл `docs/swagger.yaml` напрямую

## Примеры запросов

### Создание конфигурации
```bash
curl -X POST http://localhost:8080/configs/production/database_url \
  -H "Content-Type: application/json" \
  -d '{"value": "postgres://localhost:5432/mydb"}'
```

### Получение конфигурации
```bash
curl http://localhost:8080/configs/production/database_url
```

### Получение всех конфигураций окружения
```bash
curl http://localhost:8080/configs/production
```

### Обновление конфигурации
```bash
curl -X PUT http://localhost:8080/configs/production/database_url \
  -H "Content-Type: application/json" \
  -d '{"value": "postgres://localhost:5432/newdb"}'
```

### Удаление конфигурации
```bash
curl -X DELETE http://localhost:8080/configs/production/database_url
```

## Установка и запуск

### Локальная разработка

1. Установите зависимости:
```bash
go mod download
```

2. Создайте файл `.env` на основе `.env.example`:
```bash
cp .env.example .env
```

3. Отредактируйте `.env` файл, указав правильные значения:
```env
DATABASE_URL=postgres://config_user:config_pass@localhost:5432/configdb?sslmode=disable
PORT=8080
```

4. Запустите PostgreSQL через docker-compose:
```bash
docker-compose up -d postgres
```

5. Дождитесь применения миграций (они применяются автоматически при первом запуске PostgreSQL)

6. Запустите приложение:
```bash
go run cmd/main.go
```

### Docker

Запуск всего стека через docker-compose:
```bash
docker-compose up -d
```

Сервис будет доступен по адресу `http://localhost:8080`

### Сброс базы данных


```bash
docker-compose down -v  # Удаляет контейнеры и volumes
docker-compose up -d     # Создает все заново, миграции выполнятся автоматически
```

## Тестирование

Запуск модульных тестов:
```bash
go test ./...
```

Запуск тестов с покрытием:
```bash
go test -cover ./...
```