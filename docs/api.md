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

Ограничения потока аутентификации:

- После регистрации пользователь получает письмо с ссылкой подтверждения email.
- До подтверждения email endpoint `POST /v1/auth/login` возвращает `403 Forbidden` с ошибкой `email is not verified`.
- После успешного подтверждения (`GET /v1/auth/verify-email?token=...`) логин возвращает JWT access token.

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

## Events

- `GET /v1/events` — публичный список только `approved` событий
- `GET /v1/events/:id` — публичная карточка только `approved` события
- `GET /v1/events/:id/odds` — текущие динамические коэффициенты для события
- `POST /v1/events` — создание события авторизованным пользователем (`Bearer`)
- `POST /v1/events/:id/request-settlement` — запрос на завершение события его создателем (с доказательством)

Пример `POST /v1/events` request:

```json
{
  "title": "Will company X close in 2026?",
  "description": "Community forecast event",
  "category": "business",
  "resolve_at": "2026-12-01T12:00:00Z"
}
```

Пример `POST /v1/events` response:

```json
{
  "id": "evt_1",
  "creator_user_id": "usr_...",
  "title": "Will company X close in 2026?",
  "description": "Community forecast event",
  "category": "business",
  "resolve_at": "2026-12-01T12:00:00Z",
  "status": "pending",
  "created_at": "2026-04-09T09:00:00Z"
}
```

Правила:

- Новое событие создаётся в статусе `pending`.
- Пока событие не одобрено модерацией, оно не попадает в `GET /v1/events` и `GET /v1/events/:id`.
- Коэффициенты рассчитываются динамически из пула открытых ставок по исходам `yes/no` с маржой платформы.
- Для завершения события создатель переводит событие в `settlement_requested`, передавая `evidence_url` или файл (`evidence_file.file_name` + `evidence_file.file_data`).

Пример `GET /v1/events/:id/odds` response:

```json
{
  "event_id": "1",
  "odds": {
    "yes": 1.9000,
    "no": 1.9000
  },
  "margin_bps": 500
}
```

Пример `POST /v1/events/:id/request-settlement` request:

```json
{
  "evidence_url": "https://example.com/proof"
}
```

## RBAC (минимальный каркас)

- `GET /v1/moderation/health` доступен только ролям `moderator` и `admin`.
- `GET /v1/moderation/events` доступен только ролям `moderator` и `admin`.
- `POST /v1/moderation/events/:id/approve` доступен только ролям `moderator` и `admin`.
- `POST /v1/moderation/events/:id/reject` доступен только ролям `moderator` и `admin`, требует JSON body с полем `reason`.
- Для роли `user` endpoint возвращает `403 Forbidden`.
- Для невалидного/отсутствующего Bearer token endpoint'ы под auth middleware возвращают `401 Unauthorized`.

## Admin / Settlement

- `GET /v1/admin/events/settlement-requests` — список событий в статусе `settlement_requested` (только `admin`)
- `POST /v1/admin/events/:id/settle` — финализация события по исходу `winner_outcome` (`yes|no`, только `admin`)
- Повторный settlement для уже завершенного события возвращает `409 Conflict`.

Пример `POST /v1/admin/events/:id/settle` request:

```json
{
  "winner_outcome": "yes"
}
```

Пример `POST /v1/admin/events/:id/settle` response:

```json
{
  "event": {
    "id": "1",
    "status": "settled",
    "winner_outcome": "yes"
  },
  "settled_bets": [
    {
      "id": "10",
      "status": "won"
    },
    {
      "id": "11",
      "status": "lost"
    }
  ]
}
```

Гарантии settlement:

- Settlement выполняется атомарно: в одной транзакции фиксируются статус события, статусы ставок, зачисления победителям и запись аудита.
- При любой ошибке в процессе изменения откатываются целиком.

Пример `POST /v1/moderation/events/:id/reject` request:

```json
{
  "reason": "Недостаточно данных для проверки события"
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
