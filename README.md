# Environment Config Service

Учебный проект для лабораторной по DevOps.

В репозитории есть:
- backend на Go + PostgreSQL;
- frontend (admin UI) для работы с REST API;
- CI pipeline для build/test backend и frontend.

## Стек

- Go 1.24
- PostgreSQL 15
- Uber FX для dependency injection
- Uber Zap для логирования
- Go Playground Validator для валидации конфигов
- Joho Godotenv для загрузки переменных окружения
- Node JS для логики веб клиента
- Vite для проксирования

## Структура

- `backend` - REST API, бизнес-логика, доступ к БД, unit-тесты.
- `frontend` - web UI для CRUD-операций.
- `.github/workflows/ci.yml` - CI pipeline.
- `docker-compose.yml` - запуск PostgreSQL и backend.

## Требования

Минимум:
- Docker + Docker Compose;
- Node.js 18+ (рекомендуется 20);
- npm.

Для локального запуска backend без Docker дополнительно:
- Go 1.24.

## Быстрый запуск

```bash
docker compose up -d --build
```

Проверка:

```bash
docker compose ps
curl -i http://localhost:8080/health
curl -i http://localhost:3000/health
```

Ожидается HTTP `200`.

## Работа через UI

Открыть: `http://localhost:3000`

Сценарий проверки:
1. Ввести `Environment` (например, `production`) и нажать `Load`.
2. В блоке `Entry Editor` создать запись (`Create`).
3. Выбрать запись из таблицы (`Select`) и обновить (`Update`).
4. Проверить `Lookup` по ключу.
5. Удалить запись (`Delete`).

## Проверка API (CRUD)

```bash
# CREATE
curl -i -X POST http://localhost:8080/configs/production/database_url \
  -H "Content-Type: application/json" \
  -d '{"value":"postgres://localhost:5432/mydb"}'

# READ ONE
curl -i http://localhost:8080/configs/production/database_url

# UPDATE
curl -i -X PUT http://localhost:8080/configs/production/database_url \
  -H "Content-Type: application/json" \
  -d '{"value":"postgres://localhost:5432/newdb"}'

# READ ALL
curl -i http://localhost:8080/configs/production

# DELETE
curl -i -X DELETE http://localhost:8080/configs/production/database_url
```

## Локальная разработка (без Docker для frontend)

Backend можно держать в Docker, frontend запускать локально:

```bash
# из корня проекта
docker compose up -d postgres app

# frontend
cd frontend
npm install
echo "VITE_API_BASE_URL=/" > .env.local
npm run dev
```

Frontend dev server: `http://localhost:5173`.

## Тесты и сборка

Backend:

```bash
cd backend
go test -v ./...
go build -v -o bin/config-service ./cmd/...
```

Frontend:

```bash
cd frontend
npm test
npm run build
```

## CI

Workflow: `.github/workflows/ci.yml`

Jobs:
- `lint` (backend)
- `build` (backend)
- `test` (backend)
- `frontend_build`
- `frontend_test`

## Полезные команды

Остановить стек:

```bash
docker compose down
```

Сбросить БД (удалить volume и поднять заново):

```bash
docker compose down -v
docker compose up -d --build
```

## Частые проблемы

### `Failed to fetch` в UI

Проверить backend и прокси frontend:

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:3000/health
```

Если первый запрос не `200` — backend не поднят.
Если первый `200`, а второй нет — проблема в frontend контейнере/прокси.

### Запускается старая версия frontend

```bash
docker compose up -d --build frontend
```

После этого сделать hard refresh в браузере (`Ctrl+Shift+R`).