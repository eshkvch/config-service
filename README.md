# Environment Config Service

Учебный проект для лабораторной по DevOps:
- backend на Go + PostgreSQL;
- frontend (Vite) как web-клиент к REST API;
- CI в GitHub Actions (build/test для backend и frontend).

## Технологии
- Go 1.24
- PostgreSQL
- Uber FX для dependency injection
- Uber Zap для логирования
- Go Playground Validator для валидации конфигов
- Joho Godotenv для загрузки переменных окружения
- Node JS для логики веб клиента
- Vite для проксирования

## Структура

- `backend` — REST API, бизнес-логика, доступ к БД, unit-тесты.
- `frontend` — web UI для CRUD-операций.
- `.github/workflows/ci.yml` — CI pipeline.
- `docker-compose.yml` — запуск PostgreSQL и backend.

## Требования

Минимум:
- Docker + Docker Compose;
- Node.js 18+ (рекомендуется 20);
- npm.

Для локального запуска backend без Docker дополнительно:
- Go 1.24.

## Быстрый запуск (рекомендуемый)

Запуск backend + БД в Docker:

```bash
cd ~/Projects/itmo-devops/config-service
sudo docker compose up -d --build
sudo docker compose ps
curl -i http://localhost:8080/health
```

Ожидается `HTTP/1.1 200` и JSON со статусом `ok`.

Запуск frontend:

```bash
cd frontend
npm install
echo "VITE_API_BASE_URL=/" > .env.local
npm run dev
```

Открыть в браузере: `http://localhost:5173`.

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

## Проверка через UI

1. В поле `Environment` ввести, например, `production`.
2. Нажать `Load`.
3. В блоке `Entry Editor` создать запись (`Create`).
4. Выбрать запись из таблицы (`Select`) и обновить (`Update`).
5. Проверить `Lookup` по ключу.
6. Удалить запись (`Delete`).

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

В pipeline настроены jobs:
- backend: `lint`, `build`, `test`;
- frontend: `frontend_build`, `frontend_test`.

Проверка:

```bash
git add .
git commit -m "Run full checks"
git push
```

Далее проверить статус jobs в GitHub Actions.

## Частые проблемы

### `Failed to fetch` в frontend

Проверить, что backend поднят:

```bash
curl -i http://localhost:8080/health
```

Проверить proxy через Vite:

```bash
curl -i http://localhost:5173/health
```

Убедиться, что в `frontend/.env.local` есть:

```env
VITE_API_BASE_URL=/
```

После изменения переменных перезапустить `npm run dev`.

## Полезные endpoints

- `GET /health`
- `GET /doc.json`
- `GET /doc.yaml`
- `POST /configs/{env}/{key}`
- `GET /configs/{env}/{key}`
- `GET /configs/{env}`
- `PUT /configs/{env}/{key}`
- `DELETE /configs/{env}/{key}`
