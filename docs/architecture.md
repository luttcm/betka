# Architecture

## Repo strategy

- Формат репозитория: monorepo (`backend/`, `frontend/`, `docs/`, `plans/`, `.github/`)
- Причина: быстрый цикл изменений API ↔ frontend и единый CI/CD для MVP

## Контуры

- API: Go (Gin) REST
- Frontend: Web UI (MVP-клиент в `frontend/`)
- Domain: events, bets, odds, wallet, moderation, settlement
- Infra: PostgreSQL, Redis, background workers

## Принципы

- Модульная декомпозиция
- Транзакционность финансовых операций
- Аудит критичных действий
- Идемпотентность создания ставок
