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

## Auth

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `GET /v1/auth/verify-email?token=...`
- `GET /v1/me` (требуется `Authorization: Bearer <token>`)

Пример `POST /v1/auth/register` request:

```json
{
  "email": "user@example.com",
  "password": "strong-password"
}
```

Пример `POST /v1/auth/register` response:

```json
{
  "id": "usr_...",
  "email": "user@example.com",
  "role": "user",
  "email_verified": false
}
```

Пример `POST /v1/auth/login` response:

```json
{
  "access_token": "<jwt>",
  "token_type": "Bearer"
}
```

Пример `GET /v1/me` response:

```json
{
  "status": "authorized"
}
```

## RBAC (минимальный каркас)

- `GET /v1/moderation/health` доступен только ролям `moderator` и `admin`.
- Для роли `user` endpoint возвращает `403 Forbidden`.
- Для невалидного/отсутствующего Bearer token endpoint'ы под auth middleware возвращают `401 Unauthorized`.

## Базовые группы endpoint

- Auth
- Events
- Bets
- Wallet
- Moderation
- Admin/Settlement

OpenAPI будет публиковаться из backend-приложения после реализации доменных модулей.
