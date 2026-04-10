# Frontend (MVP)

Frontend-часть MVP на Next.js + TypeScript.

## Технологии

- Next.js (App Router)
- React + TypeScript
- TanStack Query
- React Hook Form + Zod
- Tailwind CSS

## Реализовано в этой итерации

- Каталог событий: `GET /v1/events`
- Карточка события: `GET /v1/events/:id`
- Ставка из карточки события: `POST /v1/bets` + `Idempotency-Key`
- Создание события: `POST /v1/events`
- Кошелёк: `GET /v1/wallet`
- Транзакции кошелька: `GET /v1/wallet/transactions`
- Мои ставки: `GET /v1/bets/my`
- Очередь модерации: `GET /v1/moderation/events`
- Одобрение события модератором/админом: `POST /v1/moderation/events/:id/approve`
- Отклонение события модератором/админом: `POST /v1/moderation/events/:id/reject`
- Регистрация: `POST /v1/auth/register`
- Вход: `POST /v1/auth/login`
- Подтверждение email: `GET /v1/auth/verify-email?token=...`

Маршруты:

- `/` — каталог
- `/events/[id]` — карточка события
- `/events/new` — создание события
- `/wallet` — кошелёк и транзакции
- `/bets/my` — мои ставки + фильтр статуса
- `/moderation` — вкладка модерации (только `moderator/admin`)
- `/auth/register` — регистрация
- `/auth/login` — вход
- `/auth/verify` — подтверждение email по токену

## Дизайн-система (текущая реализация)

- Визуальная база приведена к стилистике из [`DESIGN.md`](../DESIGN.md): акцентный синий `#0052ff`, тёмные hero-блоки и pill-кнопки.
- Общие UI-классы вынесены в [`globals.css`](src/app/globals.css): `btn-primary`, `btn-secondary`, `btn-danger`, `panel`, `panel-dark`, `text-input`.
- Унифицированы состояния загрузки/ошибки/пустых данных через компонент `ui-states`.

## Локальный запуск

1. Установить Node.js 20+ и npm 10+.
2. В каталоге `frontend/` установить зависимости:

   ```bash
   npm install
   ```

3. Создать `.env.local` по примеру `.env.example`.
4. Убедиться, что backend доступен по `NEXT_PUBLIC_API_BASE_URL`.
5. Запустить dev-сервер:

   ```bash
   npm run dev
   ```

По умолчанию фронтенд стартует на порту `3001`.

## Запуск через Docker Compose (рекомендуется)

Из корня репозитория:

```bash
docker compose up --build
```

После старта:

- frontend: `http://localhost:3001`
- backend api: `http://localhost:3000`

Внутри compose frontend использует reverse-proxy через Next.js rewrite:

- `NEXT_PUBLIC_API_BASE_URL=/api`
- `BACKEND_URL=http://api:3000`

Это нужно, чтобы запросы из браузера не ходили на внутренний хост `api` напрямую.

## Важно для MVP

- JWT хранится в `localStorage` после входа.
- Для создания события нужен авторизованный пользователь.
- Для ставок/кошелька/моих ставок требуется авторизация.
- Раздел модерации виден и доступен только ролям `moderator/admin`.
- Если SMTP не настроен, токен подтверждения email можно взять из логов API и вставить на страницу `/auth/verify`.
