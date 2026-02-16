# Environment Config Service

Сервис для управления конфигурациями различных окружений с использованием DDD и чистой архитектуры.

## Архитектура 

Проект следует принципам Domain-Driven Design (DDD) и чистой архитектуры:

- **Domain Layer** (`internal/domain/`): Доменные сущности, value objects и интерфейсы репозиториев
- **Application Layer** (`internal/application/`): Use cases (бизнес-логика)
- **Infrastructure Layer** (`internal/infrastructure/`): Реализации репозиториев, работа с БД
- **Presentation Layer** (`internal/presentation/`): HTTP handlers

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

### Примеры запросов

#### Создание конфигурации
```bash
curl -X POST http://localhost:8080/configs/production/database_url \
  -H "Content-Type: application/json" \
  -d '{"value": "postgres://localhost:5432/mydb"}'
```

#### Получение конфигурации
```bash
curl http://localhost:8080/configs/production/database_url
```

#### Получение всех конфигураций окружения
```bash
curl http://localhost:8080/configs/production
```

#### Обновление конфигурации
```bash
curl -X PUT http://localhost:8080/configs/production/database_url \
  -H "Content-Type: application/json" \
  -d '{"value": "postgres://localhost:5432/newdb"}'
```

#### Удаление конфигурации
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

3. Запустите PostgreSQL через docker-compose:
```bash
docker-compose up -d postgres
```

4. Примените миграции (они применяются автоматически при первом запуске PostgreSQL)

5. Запустите приложение:
```bash
go run cmd/main.go
```

### Docker

Запуск всего стека через docker-compose:
```bash
docker-compose up -d
```

Сервис будет доступен по адресу `http://localhost:8080`

## Тестирование

Запуск модульных тестов:
```bash
go test ./...
```

Запуск тестов с покрытием:
```bash
go test -cover ./...
```

## Структура проекта

```
config-service/
├── cmd/
│   └── main.go                 # Точка входа
├── config/
│   └── config.go               # Загрузка и валидация конфигурации
├── internal/
│   ├── domain/                 # Domain layer
│   │   ├── entity/             # Доменные сущности
│   │   └── repository/         # Интерфейсы репозиториев
│   ├── application/            # Application layer
│   │   └── usecase/            # Use cases
│   ├── infrastructure/         # Infrastructure layer
│   │   └── database/           # Реализация работы с БД
│   │       └── queries/        # SQL запросы
│   ├── presentation/           # Presentation layer
│   │   └── handler/            # HTTP handlers
│   └── di/                     # Dependency injection
├── migrations/                 # Миграции БД
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Переменные окружения

- `DATABASE_URL` - строка подключения к PostgreSQL (обязательно)
- `PORT` - порт для HTTP сервера (по умолчанию: 8080)

## Особенности реализации

- **DDD**: Доменная логика изолирована в domain layer
- **Чистая архитектура**: Слои разделены, зависимости направлены внутрь
- **Dependency Injection**: Используется Uber FX
- **Валидация**: Конфигурация валидируется при загрузке
- **SQL в отдельных файлах**: Все SQL запросы вынесены в отдельные файлы
- **Без ORM**: Используется чистый database/sql с абстракциями
- **Модульные тесты**: Покрытие тестами domain и application слоев

