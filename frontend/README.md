# Frontend (Config Service Admin)

Web-клиент для backend API.

## Запуск в составе всего стека (рекомендуется)

Из корня проекта:

```bash
docker compose up -d --build
```

Frontend будет доступен на `http://localhost:3000`.

В docker-режиме frontend отдает статические файлы через Nginx и проксирует API-запросы в backend-контейнер (`app:8080`) внутри общей сети.

## Локальный dev-режим

```bash
npm install
echo "VITE_API_BASE_URL=/" > .env.local
npm run dev
```

Dev URL: `http://localhost:5173`.

## Тесты и сборка

```bash
npm test
npm run build
```

## Быстрая диагностика

Если в UI `Failed to fetch`:

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:3000/health
```

Если первый запрос не `200`, backend не запущен.
Если первый `200`, а второй нет — проблема с фронтовым контейнером/прокси. 