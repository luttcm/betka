# Bet MVP

MVP букмекерской платформы с демо-ставками на виртуальную валюту.

## Стек

- Go + Gin
- PostgreSQL
- Redis
- Docker / Docker Compose

## Быстрый старт

1. Скопировать окружение:

```bash
cp .env.example .env
```

2. Запустить инфраструктуру, API, worker и frontend:

```bash
docker compose up --build
```

[`docker-compose.yml`](docker-compose.yml) поднимает сервисы `migrate`, `api`, `worker`, `frontend`, `postgres`, `redis`.
Сервис `migrate` автоматически применяет SQL-миграции через Goose перед стартом `api`/`worker`.

3. Проверить health endpoint:

```bash
curl http://localhost:3000/health
```

4. Открыть frontend:

```bash
xdg-open http://localhost:3001
```

или просто перейти в браузере на `http://localhost:3001`.

## Переменные окружения

Базовые переменные находятся в `.env.example`:

- `PORT` — внешний порт API (по умолчанию `3000`)
- `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_PORT`
- `REDIS_PORT`
- `AUTH_JWT_SECRET`, `AUTH_TOKEN_TTL`
- `EMAIL_FROM`, `EMAIL_VERIFY_BASE_URL`, `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD`
- `BOOTSTRAP_ADMIN_EMAIL`, `BOOTSTRAP_ADMIN_PASSWORD` (опционально, для dev-инициализации админа)

Дополнительно присутствуют `DATABASE_URL` и `REDIS_URL` для локального запуска backend без Docker.
В Docker Compose DSN формируются автоматически из базовых переменных.

Если `SMTP_HOST`/`SMTP_PORT` не заданы, backend использует log-stub отправитель писем (ссылка подтверждения email будет в логах API).

Если почта временно недоступна, можно создать/обновить админа при старте API через
`BOOTSTRAP_ADMIN_EMAIL` + `BOOTSTRAP_ADMIN_PASSWORD`.
Пользователь будет создан (или обновлён), получит роль `admin`, а `email_verified=true`.

## Хранение данных: PostgreSQL + fallback

Backend больше не ограничен только in-memory режимом:

- при доступной `DATABASE_URL` сервисы `auth/events/wallet/bets` пишут и читают данные из PostgreSQL;
- при недоступной БД API логирует проблему подключения и автоматически использует in-memory fallback (для dev/тестовых сценариев).

Проверка, что запись действительно идёт в БД:

```bash
docker compose up --build
docker compose exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT COUNT(*) FROM users;"
docker compose exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT COUNT(*) FROM events;"
docker compose exec postgres psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c "SELECT COUNT(*) FROM bets;"
```

Если значения растут после вызовов API, персистентность работает корректно.

## Auth flow (MVP)

- `POST /v1/auth/register` создаёт пользователя и отправляет письмо для подтверждения email.
- До подтверждения email `POST /v1/auth/login` возвращает `403`.
- После перехода по ссылке `GET /v1/auth/verify-email?token=...` логин начинает выдавать JWT.

## Структура

- `backend/` — API и worker
- `frontend/` — web-клиент (Next.js)
- `docs/` — архитектура, доменные правила, тестирование, безопасность
- `.github/workflows/` — CI pipeline
- `plans/` — утвержденный план MVP

## Локальный запуск backend без Docker

```bash
cd backend
go run ./cmd/api
```

## Миграции (Goose)

Миграции лежат в `backend/migrations`.

Ручной запуск миграций локально:

```bash
cd backend
go run github.com/pressly/goose/v3/cmd/goose@v3.24.1 \
  -dir ./migrations \
  postgres "postgresql://bet_user:bet_password@localhost:5432/bet_mvp?sslmode=disable" up
```
