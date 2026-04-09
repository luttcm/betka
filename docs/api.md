# API

## Versioning

- Префикс: `/v1`

## Health

- `GET /health`
- `GET /v1/health`

Пример ответа:

```json
{
  "status": "ok",
  "timestamp": "2026-04-09T07:00:00Z"
}
```

## Базовые группы endpoint

- Auth
- Events
- Bets
- Wallet
- Moderation
- Admin/Settlement

OpenAPI будет публиковаться из backend-приложения после реализации доменных модулей.
