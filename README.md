# notes-api

Простой REST API на Go для CRUD-операций над пользователями.

Несмотря на название `notes-api`, на текущий момент API работает с сущностью **User** (`id`, `name`, `age`, `created_at`).

Стек:
- `net/http` (стандартный `http.ServeMux`, роутинг в стиле `POST /users`)
- `PostgreSQL` — основное хранилище (драйвер `github.com/jackc/pgx/v5`)
- `Redis` — кэширование `GET /users/{id}` на 24 часа (`github.com/redis/go-redis/v9`)
- `goose` — SQL-миграции (`migrations/`)
- Конфигурация — через переменные окружения (`.env`)

## Структура проекта

```text
.
├── cmd/server/main.go                 # точка входа: подключение к PG + Redis, запуск HTTP-сервера
├── internal/
│   ├── user/
│   │   ├── user.go                    # модель User + Validate()
│   │   ├── handler.go                 # HTTP-хендлеры, регистрация роутов
│   │   └── repository.go              # SQL-запросы к Postgres + работа с Redis-кэшем
│   └── respond/
│       └── respond.go                 # хелперы JSON / Error / ServerError
├── migrations/
│   └── 20260927084546_create_users_table.sql  # таблица users
├── Makefile                           # шорткаты для миграций и запуска
├── .env.example                       # пример настроек
├── go.mod / go.sum
```

Логика:
1. `main.go` читает `DATABASE_URL`, `REDIS_URL`, `PORT` из окружения, проверяет `pool.Ping()`, поднимает `http.ListenAndServe`.
2. `Handler.Register(mux)` вешает 5 роутов.
3. `Repository` ходит в Postgres напрямую через `pgxpool`, а `GetById` сначала смотрит в Redis по ключу `user:{id}`.

## API

| Метод  | Путь          | Описание            | Коды ответа          |
|--------|---------------|---------------------|----------------------|
| POST   | `/users`      | Создать пользователя| `201`, `400`, `500`  |
| GET    | `/users`      | Список всех         | `200`, `500`         |
| GET    | `/users/{id}` | Один по id (из кэша/БД) | `200`, `400`, `404`, `500` |
| PUT    | `/users/{id}` | Обновить            | `200`, `400`, `404`, `500` |
| DELETE | `/users/{id}` | Удалить             | `200`, `400`, `404`, `500` |

Модель:

```json
{
  "id": 1,
  "name": "Ivan",
  "age": 25,
  "created_at": "2026-09-27T08:45:46Z"
}
```

Примеры:

```bash
# создать
curl -X POST localhost:3000/users \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ivan","age":25}'

# все
curl localhost:3000/users

# один
curl localhost:3000/users/1

# обновить
curl -X PUT localhost:3000/users/1 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Petr","age":30}'

# удалить
curl -X DELETE localhost:3000/users/1
```

Валидация: `name` — обязателен, `age` > 0, иначе `400`.

## Что нужно для запуска

1. **Go 1.25+**
   ```bash
   go version
   ```
2. **PostgreSQL** (локально или в Docker), созданная БД, например `notes`:
   ```bash
   createdb notes
   # или через psql:
   # CREATE DATABASE notes;
   ```
3. **Redis** (локально или в Docker):
   ```bash
   redis-cli ping
   # должен ответить PONG
   ```
4. **goose** — CLI для миграций:
   ```bash
   go install github.com/pressly/goose/v3/cmd/goose@latest
   goose -version
   ```

## Запуск по шагам

```bash
# 1. Клонировать и зайти в проект
git clone <repo-url> notes-api
cd notes-api

# 2. Создать конфиг из примера
cp .env.example .env
```

Отредактируйте `.env`:

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/notes?sslmode=disable
REDIS_URL=redis://localhost:6379
PORT=3000
```

```bash
# 3. Поставить зависимости
go mod tidy

# 4. Применить миграции (создаст таблицу users)
make migrate-up

# 5. Запустить сервер
make run
# Server is running on port 3000
```

Без `Makefile` то же самое вручную:

```bash
export $(cat .env | xargs)  # или source .env под вашу shell
goose -dir migrations postgres "$DATABASE_URL" up
go run cmd/server/main.go
```

## Что такое Makefile и зачем он здесь

`Makefile` — это файл с именованными командами (таргетами) для `make`. Вместо того чтобы каждый раз вспоминать длинные команды, вы пишете коротко `make run`.

В этом проекте `Makefile`:

```make
include .env

export

migrate-up:
	goose -dir migrations postgres "${DATABASE_URL}" up

migrate-down:
	goose -dir migrations postgres "${DATABASE_URL}" down

migrate-new:
	goose -dir migrations create $(name) sql

migrate-status:
	goose -dir migrations postgres "${DATABASE_URL}" status

run:
	go run cmd/server/main.go
```

Разбор построчно:

- `include .env` — подгружает переменные (`DATABASE_URL`, `REDIS_URL`, `PORT`) из файла `.env` как переменные `make`.
- `export` (пустой, без списка) — экспортирует все эти переменные в окружение дочерних процессов. Поэтому `go run cmd/server/main.go` внутри `make run` уже видит `os.Getenv("DATABASE_URL")`, а `goose` видит `"${DATABASE_URL}"`. Без этого пришлось бы каждый раз делать `export` вручную.
- `migrate-up` — применить все новые миграции из `migrations/`.
- `migrate-down` — откатить последнюю миграцию (в данном проекте — `DROP TABLE users`).
- `migrate-status` — показать, какие миграции применены, а какие нет.
- `migrate-new` — создать пустой файл миграции: `make migrate-new name=add_notes_table`.
- `run` — запустить сервер (`cmd/server/main.go`). Порт берётся из `PORT`.

То есть для повседневной работы достаточно двух команд: `make migrate-up` и `make run`.
