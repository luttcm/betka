# План фронтенд-итерации MVP

Источник: [plans/mvp-plan.md](plans/mvp-plan.md), [AGENTS.md](AGENTS.md)
После выполнения пункта вычеркивай его

## 1. Цель итерации

Закрыть пользовательский frontend-флоу MVP:

`регистрация → подтверждение email → вход → просмотр событий → ставка → просмотр кошелька и моих ставок`,
а также модерацию для ролей `moderator/admin`.

## 2. Scope

### Что уже есть

- Каталог: [frontend/src/app/page.tsx](frontend/src/app/page.tsx)
- Карточка события: [frontend/src/app/events/[id]/page.tsx](frontend/src/app/events/[id]/page.tsx)
- Создание события: [frontend/src/app/events/new/page.tsx](frontend/src/app/events/new/page.tsx)
- Модерация: [frontend/src/app/moderation/page.tsx](frontend/src/app/moderation/page.tsx)
- Auth страницы: [frontend/src/app/auth/login/page.tsx](frontend/src/app/auth/login/page.tsx), [frontend/src/app/auth/register/page.tsx](frontend/src/app/auth/register/page.tsx), [frontend/src/app/auth/verify/page.tsx](frontend/src/app/auth/verify/page.tsx)

### Что добавляем в этой итерации

1. ~~Ставка в карточке события (`yes/no`, сумма, `Idempotency-Key`)~~
2. ~~Страница кошелька (`/wallet`): баланс + транзакции~~
3. ~~Страница моих ставок (`/bets/my`)~~
4. ~~Ролевую навигацию и защиту UI по ролям~~
5. ~~Унифицированные состояния `loading/error/empty`~~

## 3. Экранная карта

- `/` — каталог событий
- `/events/[id]` — карточка события + форма ставки
- `/events/new` — создание события
- `/wallet` — кошелёк
- `/bets/my` — мои ставки
- `/moderation` — модерация (только moderator/admin)
- `/auth/register`, `/auth/verify`, `/auth/login` — аутентификация

## 4. Технический план по этапам

## Этап A — API и типы

Файлы:

- [frontend/src/lib/types.ts](frontend/src/lib/types.ts)
- [frontend/src/lib/api.ts](frontend/src/lib/api.ts)

Задачи:

- ~~Добавить типы `Wallet`, `WalletTransaction`, `BetItem`, `MyBetsResponse`, `PlaceBetPayload`~~
- ~~Добавить API-методы:~~
  - `getWallet()`
  - `getWalletTransactions()`
  - `getMyBets()`
  - `placeBet()`
- ~~Привести ошибки API к единому формату `ApiError`~~

## Этап B — Ставка на карточке события

Файл: [frontend/src/components/event-details.tsx](frontend/src/components/event-details.tsx)

Задачи:

- ~~Добавить форму ставки (исход + сумма)~~
- ~~Генерировать и передавать `Idempotency-Key`~~
- ~~Показывать ошибки API (`insufficient funds`, `event unavailable`, `unauthorized`)~~
- ~~После успешной ставки инвалидировать кэш кошелька и моих ставок~~

## Этап C — Кошелёк

Новые файлы:

- [frontend/src/app/wallet/page.tsx](frontend/src/app/wallet/page.tsx)
- [frontend/src/components/wallet-balance-card.tsx](frontend/src/components/wallet-balance-card.tsx)
- [frontend/src/components/wallet-transactions-list.tsx](frontend/src/components/wallet-transactions-list.tsx)

Задачи:

- ~~Показать текущий баланс~~
- ~~Показать историю транзакций~~
- ~~Обработать `loading/error/empty`~~

## Этап D — Мои ставки

Новые файлы:

- [frontend/src/app/bets/my/page.tsx](frontend/src/app/bets/my/page.tsx)
- [frontend/src/components/my-bets-list.tsx](frontend/src/components/my-bets-list.tsx)

Задачи:

- ~~Вывести список ставок пользователя~~
- ~~Сортировка по времени размещения~~
- ~~Базовый фильтр по статусу (`open/won/lost/refunded`)~~

## Этап E — Навигация и роли

Файлы:

- [frontend/src/components/auth-nav.tsx](frontend/src/components/auth-nav.tsx)
- [frontend/src/lib/auth-context.tsx](frontend/src/lib/auth-context.tsx)

Задачи:

- ~~Для авторизованного пользователя: ссылки на `/wallet` и `/bets/my`~~
- ~~Для `moderator/admin`: ссылка на `/moderation`~~
- ~~Проверки доступа в UI на основе роли~~

## Этап F — Полировка и документация

Файлы:

- [frontend/README.md](frontend/README.md)
- [frontend/src/app/globals.css](frontend/src/app/globals.css)

Задачи:

- ~~Унифицировать текст ошибок и статусы~~
- ~~Упростить визуальные состояния форм и списков~~
- ~~Обновить инструкции запуска и переменные окружения~~
- ~~Добавить Коэффициенты к ставками в карточке события~~

## 5. Приоритеты backlog

## P0 (обязательно)

- API для кошелька и ставок
- Ставка в карточке события
- `/wallet` и `/bets/my`
- Ролевая навигация
- Корректные `loading/error/empty`

## P1 (желательно)

- Фильтрация ставок
- Улучшение UX сообщений и статусов
- Единое форматирование дат/сумм

## P2 (после MVP)

- Пагинация списков
- Расширенная accessibility-полировка
- UI для subscribe/withdraw после готовности backend

## 6. План спринта (5 рабочих дней)

1. День 1: Этап A (типы + API)
2. День 2: Этап B (ставка в карточке)
3. День 3: Этап C (кошелёк)
4. День 4: Этап D (мои ставки)
5. День 5: Этап E/F (навигация, роли, полировка, README)

## 7. Definition of Done итерации

Итерация считается завершённой, если:

1. Пользователь может поставить ставку из карточки события.
2. На `/wallet` видны актуальный баланс и hold-транзакция.
3. На `/bets/my` видна созданная ставка.
4. Раздел модерации доступен только `moderator/admin`.
5. Ключевые экраны устойчивы к `4xx/5xx` и имеют `loading/error/empty` состояния.

## 8. Post-iteration backend dependency

- [ ] Перевести хранение событий/ставок/кошелька с in-memory на PostgreSQL (персистентная запись в БД), чтобы данные не терялись после рестарта Docker.
