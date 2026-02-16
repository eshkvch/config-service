# Frontend (Config Service Admin)

Простой web-клиент для backend API.

## Требования

- Node.js 18+ (рекомендуется 20)
- npm
- запущенный backend на `http://localhost:8080`

## Установка

```bash
npm install
```

## Запуск в dev

Рекомендуемая конфигурация (через Vite proxy):

```bash
echo "VITE_API_BASE_URL=/" > .env.local
npm run dev
```

Открыть: `http://localhost:5173`.

## Тесты и сборка

```bash
npm test
npm run build
```

## Быстрая диагностика

Если в UI ошибка `Failed to fetch`:

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:5173/health
```

- если первый запрос не `200`, backend не запущен;
- если первый `200`, но второй нет, проблема в dev proxy или конфиге frontend.
