# source_html — поиск закупок ЕИС

Сервисы **storage**, **search** и **ui** + PostgreSQL. Материалы по интеграции с ЕИС — в [`info/`](./info/).

## Безопасность v1

- **Trust network:** HTTP API **без аутентификации** (осознанное решение v1). Разворачивайте только во внутреннем или доверенном периметре.
- В `docker-compose` наружу по умолчанию опубликован **только UI** (порт `UI_PORT`, по умолчанию 8080). `storage` и `search` доступны только во внутренней сети compose (`expose`, без `ports` на 0.0.0.0).
- Секреты ЕИС (логин/пароль 223, пути к сертификатам mTLS 44, код отправителя) задаются **только через env** (файл `.env`, не в git). Не коммитьте реальные значения.

## Быстрый старт (Docker Compose)

```bash
cp .env.example .env
# Отредактируйте .env: POSTGRES_PASSWORD, EIS_* при необходимости реального поиска в ЕИС

docker compose up --build -d
```

- UI: http://localhost:8080  
- Конфиг для фронтенда: http://localhost:8080/config.json — базовые пути `/api/storage` и `/api/search` (прокси через UI во внутренние сервисы).
- Миграции БД: `migrations/001_init.up.sql` применяется при **первом** создании тома Postgres (`docker-entrypoint-initdb.d`).

Проверка health:

```bash
curl -s http://localhost:8080/healthz
curl -s http://localhost:8080/config.json
```

### Отладка API напрямую (не для прод)

Раскомментируйте или добавьте в `docker-compose.yml` для сервисов `storage` / `search`:

```yaml
# ports:
#   - "8081:8081"   # storage
#   - "8082:8082"   # search
```

## Локальный запуск без Docker

```bash
export DATABASE_URL='postgres://...'
export STORAGE_URL='http://127.0.0.1:8081'
export SEARCH_URL='http://127.0.0.1:8082'
go run ./cmd/storage
go run ./cmd/search
go run ./cmd/ui
```

## Документация ЕИС

- [Обзор для человека](./info/HUMAN_EIS_OVERVIEW.md)
- [Инструкция для агента/разработчика](./info/AGENT_EIS_INTEGRATION.md)
- [Анализ](./info/ANALYSIS.md)
- [Инструкция работы](./info/WORK_INSTRUCTIONS.md)

## Образы

| Сервис   | Dockerfile            | Порт (внутри сети) |
|----------|------------------------|--------------------|
| storage  | `Dockerfile.storage`   | 8081               |
| search   | `Dockerfile.search`    | 8082               |
| ui       | `Dockerfile.ui`        | 8080 (publish)     |

Переменные окружения — см. [`.env.example`](./.env.example).
