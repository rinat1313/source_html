# HTTP API (v1)

Описание для интеграции **внутри доверенной сети**. Аутентификации нет.

## Как обращаться через UI

Браузер и внешние клиенты за reverse-proxy UI используют префиксы:

| Префикс | Сервис | Пример |
|---------|--------|--------|
| `/api/storage` | storage | `GET /api/storage/profiles` |
| `/api/search` | search | `POST /api/search/runs` |

Конфигурация для фронтенда: `GET /config.json` — поля `storagePrefix`, `searchPrefix`.

Прямой доступ к storage (порт 8081) и search (8082) в compose по умолчанию **не опубликован**; пути там с `/api/v1/...`.

## Общие правила

- Формат: JSON, кодировка UTF-8.
- Ошибки: `{ "error": "текст" }` и соответствующий HTTP-код.
- Пагинация списков: query-параметры `offset`, `limit` (если `limit` не задан, по умолчанию **10**).
- Ответы списков закупок: `{ "items": [...], "total": N, "offset": ..., "limit": ... }`.

## storage — профили

| Метод | Путь (через UI) | Описание |
|-------|-----------------|----------|
| GET | `/api/storage/profiles` | Список профилей |
| POST | `/api/storage/profiles` | Создать. Тело: `{ "name": "...", "config": { ... } }` |
| GET | `/api/storage/profiles/{id}` | Один профиль |
| PUT | `/api/storage/profiles/{id}` | Обновить имя и config |
| DELETE | `/api/storage/profiles/{id}` | Удалить (204 без тела) |

### Поля `config`

| Поле | Тип | Описание |
|------|-----|----------|
| `law` | string | `44`, `223` или `both` |
| `keywords` | string[] | Ключевые слова (фильтр AND по title+description) |
| `limit` | number | Лимит сохранений за прогон (> 0) |
| `eis44` | object | `org_region`, `document_types[]`, `period.from`, `period.to` |
| `eis223` | object | `entity_types[]`, `region`, `period.from`, `period.to` |

Секреты ЕИС в config **не сохраняются** (поля с паролями, токенами, сертификатами отбрасываются).

## storage — закупки

| Метод | Путь | Query | Описание |
|-------|------|-------|----------|
| GET | `/api/storage/purchases` | `law`, `q`, `offset`, `limit` | Список в базе |
| GET | `/api/storage/purchases/{id}` | — | Карточка закупки |

Поля закупки в JSON: `id`, `law`, `entity_kind` (`notice` / `contract`), `reg_number`, `guid`, `title`, `description`, `customer_name`, `customer_inn`, `file_links`, `effective_date`, даты создания/обновления.

## storage — результаты прогона

| Метод | Путь | Query | Описание |
|-------|------|-------|----------|
| GET | `/api/storage/search/runs/{id}/results` | `offset`, `limit` | Закупки, привязанные к прогону |

Элементы списка содержат закупку и метаданные привязки к прогону (как в ответе search ниже).

## search — прогоны

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/search/runs` | Запустить прогон. Тело: `{ "profile_id": "uuid" }` **или** `{ "config": { ... } }` без сохранения профиля |
| GET | `/api/search/runs/{id}` | Статус прогона |
| GET | `/api/search/runs/{id}/results` | Результаты с пагинацией |

Ответ прогона (основные поля): `id`, `status` (`pending`, `running`, `completed`, `failed`), `error_message`, `law`, `limit_n`, `matched_saved`, `created_at`, `finished_at`.

При недоступности ЕИС: `status` = `failed`, `error_message` содержит **ЕИС недоступна**; уже сохранённые `matched_saved` не откатываются.

## Служебные

| Метод | Путь | Описание |
|-------|------|----------|
| GET | `/healthz` | UI, storage или search — `{ "status": "ok" }` при живых зависимостях |

## Внутренние endpoints

Пути `/internal/api/v1/...` на storage предназначены **только для сервиса search** (создание прогона, upsert закупки). Внешним клиентам их использовать не нужно.
