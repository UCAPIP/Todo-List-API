# Todo List API

🇬🇧 [English](#english) | 🇷🇺 [Русский](#русский)

---

## English

A RESTful API for a personal to-do list, written in Go. Users can register, log in and manage their own tasks. Authentication is based on JWT tokens, passwords are stored as bcrypt hashes, and every user can see and change only their own tasks. It was built as a learning project to practice authentication, database design, REST conventions, error handling and security basics.

> The original assignment names the resource `/todos`. This implementation exposes it as `/tasks`.

### Features

- User registration and login with JWT tokens
- Password hashing with bcrypt
- Full CRUD for tasks
- Per-user data isolation: you only see your own tasks
- `403 Forbidden` when you try to change someone else's task
- Pagination for the task list
- Request validation and consistent JSON error responses
- Request body size limit (1 MB)
- SQLite storage with parameterized queries (protection against SQL injection)

### Tech stack

- Go 1.22+ (method and wildcard patterns in the standard `net/http` router)
- SQLite via [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) (pure Go, no CGO required)
- [`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt) for tokens
- [`golang.org/x/crypto/bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt) for password hashing
- [`joho/godotenv`](https://github.com/joho/godotenv) for loading a local `.env` file

### Project structure

| File | Purpose |
|---|---|
| `main.go` | Opens the database, loads configuration, registers routes, starts the server |
| `handlers.go` | HTTP handlers, input validation, pagination parameters |
| `middleware.go` | JWT creation and verification, authentication middleware |
| `repository.go` | Database schema and all SQL queries |
| `responses.go` | JSON response helpers and shared errors |

### Getting started

**1. Clone the repository and install dependencies**

```bash
git clone [https://github.com/UCAPIP/Blogging-Platform-API](https://github.com/UCAPIP/Todo-List-API)
cd <your-repo-folder>
go mod tidy
```

**2. Configure the secret**

Tokens are signed with a secret that is read from the `JWT_SECRET` environment variable (at least 32 characters). The easiest way is a `.env` file in the project folder:

```
JWT_SECRET=put-a-long-random-string-of-at-least-32-characters-here
```

Copy `.env.example` to `.env` and replace the value. Never commit `.env` to Git.

Alternatively, set the variable for the current terminal session.

Windows PowerShell:

```powershell
$env:JWT_SECRET = "put-a-long-random-string-of-at-least-32-characters-here"
```

bash:

```bash
export JWT_SECRET="put-a-long-random-string-of-at-least-32-characters-here"
```

**3. Run the server**

```bash
go run .
```

The API is available at `http://localhost:8080`. On the first start the SQLite file `store.db` and the `Users` and `Tasks` tables are created automatically in the folder you run the server from.

### Authentication

1. Register or log in to get a token.
2. Send the token in the `Authorization` header of every request to `/tasks`:

```
Authorization: Bearer <your-token>
```

Tokens are JWTs signed with HS256 and are valid for 24 hours. A missing, malformed, invalid or expired token results in `401 Unauthorized`. The `/register` and `/login` endpoints are public.

### API reference

#### Endpoints

| Method | Path | Auth | Description | Success | Errors |
|---|---|---|---|---|---|
| `POST` | `/register` | no | Create a user and get a token | `201 Created` | `400`, `409` |
| `POST` | `/login` | no | Log in and get a token | `200 OK` | `400`, `401` |
| `GET` | `/tasks` | yes | List your tasks (paginated) | `200 OK` | `400`, `401` |
| `GET` | `/tasks/{id}` | yes | Get one of your tasks | `200 OK` | `400`, `401`, `404` |
| `POST` | `/tasks` | yes | Create a task | `201 Created` | `400`, `401` |
| `PUT` | `/tasks/{id}` | yes | Update a task (all fields required) | `200 OK` | `400`, `401`, `403`, `404` |
| `DELETE` | `/tasks/{id}` | yes | Delete a task | `204 No Content` | `400`, `401`, `403`, `404` |

#### Register

```
POST /register
```

```json
{
  "name": "John Doe",
  "email": "john@doe.com",
  "password": "password123"
}
```

Rules: `name` and `email` are required, the password must be 8 to 72 characters long. If a user with this email already exists, the API responds with `409 Conflict`.

Response `201 Created`:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Login

```
POST /login
```

```json
{
  "email": "john@doe.com",
  "password": "password123"
}
```

Response `200 OK`:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

An unknown email and a wrong password produce the same `401 Unauthorized` response, so the API does not reveal which emails are registered.

#### Task object

```json
{
  "id": 1,
  "title": "Buy groceries",
  "description": "Buy milk, eggs, and bread",
  "createdAt": "2026-10-22T09:33:49Z",
  "updatedAt": "2026-10-22T09:33:49Z",
  "userId": 7
}
```

`id`, `createdAt`, `updatedAt` and `userId` are set by the server. The client sends only `title` and `description`, and both must not be empty.

#### Create a task

```
POST /tasks
```

```json
{
  "title": "Buy groceries",
  "description": "Buy milk, eggs, and bread"
}
```

Response `201 Created` with the created task.

#### Update a task

```
PUT /tasks/1
```

```json
{
  "title": "Buy groceries",
  "description": "Buy milk, eggs, bread, and cheese"
}
```

Response `200 OK` with the updated task. If the task belongs to another user, the API responds with `403 Forbidden`. If it does not exist, `404 Not Found`.

#### Delete a task

```
DELETE /tasks/1
```

Response `204 No Content` with an empty body. `403 Forbidden` for another user's task, `404 Not Found` if it does not exist.

#### List tasks (pagination)

```
GET /tasks?page=1&limit=10
```

| Parameter | Default | Rules |
|---|---|---|
| `page` | `1` | positive integer |
| `limit` | `10` | positive integer, values above `100` are capped to `100` |

An invalid value (for example `page=0` or `limit=abc`) results in `400 Bad Request`.

Response `200 OK`:

```json
{
  "data": [
    {
      "id": 1,
      "title": "Buy groceries",
      "description": "Buy milk, eggs, bread",
      "createdAt": "2026-10-22T09:33:49Z",
      "updatedAt": "2026-10-22T09:33:49Z",
      "userId": 7
    }
  ],
  "page": 1,
  "limit": 10,
  "total": 25
}
```

`total` is the number of all tasks of the current user, not just the tasks on the current page. A page beyond the end returns an empty `data` array.

#### Error format

All errors are returned as JSON:

```json
{
  "error": "Validation failed",
  "details": "Title is required"
}
```

`details` is optional.

#### Status codes

| Code | Meaning |
|---|---|
| `200` | OK |
| `201` | Created |
| `204` | Deleted, no body |
| `400` | Invalid JSON, validation error, invalid id or invalid query parameter |
| `401` | Missing or invalid token, or wrong credentials |
| `403` | The task belongs to another user |
| `404` | Task not found |
| `409` | User already exists |
| `500` | Unexpected server error |

For `GET /tasks/{id}`, a task that belongs to another user is reported as `404`, so its existence is not revealed.

### Usage examples

**Windows PowerShell**

```powershell
# register and keep the token
$body = @{ name = "John Doe"; email = "john@doe.com"; password = "password123" } | ConvertTo-Json
$resp = Invoke-RestMethod -Method Post -Uri http://localhost:8080/register -ContentType "application/json" -Body $body
$token = $resp.token

# create a task
$headers = @{ Authorization = "Bearer $token" }
$task = @{ title = "Buy groceries"; description = "Buy milk, eggs, and bread" } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri http://localhost:8080/tasks -Headers $headers -ContentType "application/json" -Body $task

# list tasks (page 1, 10 per page)
Invoke-RestMethod -Uri "http://localhost:8080/tasks?page=1&limit=10" -Headers $headers
```

**bash / macOS / Linux / Git Bash**

```bash
# register (the response contains the token)
curl -i -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@doe.com","password":"password123"}'

# login
curl -i -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@doe.com","password":"password123"}'

# create a task (replace <token>)
curl -i -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy groceries","description":"Buy milk, eggs, and bread"}'

# list, update, delete
curl -i "http://localhost:8080/tasks?page=1&limit=10" -H "Authorization: Bearer <token>"
curl -i -X PUT http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy groceries","description":"Buy milk, eggs, bread, and cheese"}'
curl -i -X DELETE http://localhost:8080/tasks/1 -H "Authorization: Bearer <token>"
```

### Testing

The script `test-api.ps1` is a PowerShell smoke test. It starts from fresh users on every run and checks registration, login, token errors, user isolation, `403`/`404` rules, delete and pagination. Start the server and run:

```powershell
.\test-api.ps1
```

### Security notes

- Passwords are hashed with bcrypt and never returned or logged.
- The JWT secret comes from the environment and is never stored in the code. Keep `.env` and `store.db` out of Git (add them to `.gitignore`).
- The token parser accepts only the HS256 algorithm.
- All SQL queries are parameterized.
- Every query on tasks is scoped to the authenticated user.
- In production the API must be served over HTTPS.

### Roadmap

- [ ] Filtering and sorting for the task list
- [ ] Unit and integration tests (`net/http/httptest`)
- [ ] Rate limiting and throttling
- [ ] Refresh token mechanism
- [ ] Foreign key and `NOT NULL` constraint for `Tasks.user_id`
- [ ] Email normalization (lowercase) and format validation
- [ ] Server timeouts and graceful shutdown

---

## Русский

REST API для личного списка дел на Go. Пользователи могут зарегистрироваться, войти и управлять своими задачами. Аутентификация построена на JWT-токенах, пароли хранятся в виде bcrypt-хэшей, а каждый пользователь видит и меняет только свои задачи. Проект учебный: на нём отрабатываются аутентификация, проектирование базы, принципы REST, обработка ошибок и основы безопасности.

> В исходном задании ресурс называется `/todos`. В этой реализации он доступен как `/tasks`.

### Возможности

- Регистрация и вход с выдачей JWT-токена
- Хэширование паролей через bcrypt
- Полный CRUD для задач
- Изоляция данных: пользователь видит только свои задачи
- `403 Forbidden` при попытке изменить чужую задачу
- Пагинация списка задач
- Валидация запросов и единый JSON-формат ошибок
- Ограничение размера тела запроса (1 МБ)
- Хранение в SQLite, параметризованные запросы (защита от SQL-инъекций)

### Стек

- Go 1.22+ (шаблоны с методом и параметрами пути в стандартном роутере `net/http`)
- SQLite через [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) (чистый Go, CGO не нужен)
- [`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt) для токенов
- [`golang.org/x/crypto/bcrypt`](https://pkg.go.dev/golang.org/x/crypto/bcrypt) для хэширования паролей
- [`joho/godotenv`](https://github.com/joho/godotenv) для загрузки локального файла `.env`

### Структура проекта

| Файл | Назначение |
|---|---|
| `main.go` | Открывает базу, загружает настройки, регистрирует маршруты, запускает сервер |
| `handlers.go` | HTTP-обработчики, валидация данных, параметры пагинации |
| `middleware.go` | Создание и проверка JWT, middleware аутентификации |
| `repository.go` | Схема базы и все SQL-запросы |
| `responses.go` | Помощники для JSON-ответов и общие ошибки |

### Быстрый старт

**1. Склонируй репозиторий и подтяни зависимости**

```bash
git clone [<ссылка-на-репозиторий>](https://github.com/UCAPIP/Todo-List-API)
cd <папка-репозитория>
go mod tidy
```

**2. Настрой секрет**

Токены подписываются секретом, который читается из переменной окружения `JWT_SECRET` (не короче 32 символов). Проще всего положить его в файл `.env` в папке проекта:

```
JWT_SECRET=сюда-длинная-случайная-строка-не-короче-32-символов
```

Скопируй `.env.example` в `.env` и подставь свой секрет. Файл `.env` нельзя коммитить в Git.

Можно задать переменную и только для текущей сессии терминала.

Windows PowerShell:

```powershell
$env:JWT_SECRET = "сюда-длинная-случайная-строка-не-короче-32-символов"
```

bash:

```bash
export JWT_SECRET="сюда-длинная-случайная-строка-не-короче-32-символов"
```

**3. Запусти сервер**

```bash
go run .
```

API будет доступен по адресу `http://localhost:8080`. При первом запуске файл SQLite `store.db` и таблицы `Users` и `Tasks` создаются автоматически в той папке, откуда запущен сервер.

### Аутентификация

1. Зарегистрируйся или войди, чтобы получить токен.
2. Передавай токен в заголовке `Authorization` в каждом запросе к `/tasks`:

```
Authorization: Bearer <твой-токен>
```

Токены это JWT, подписанные алгоритмом HS256, срок жизни 24 часа. Отсутствующий, некорректный, поддельный или просроченный токен даёт `401 Unauthorized`. Эндпоинты `/register` и `/login` открытые.

### Справочник по API

#### Эндпоинты

| Метод | Путь | Токен | Описание | Успех | Ошибки |
|---|---|---|---|---|---|
| `POST` | `/register` | нет | Создать пользователя и получить токен | `201 Created` | `400`, `409` |
| `POST` | `/login` | нет | Войти и получить токен | `200 OK` | `400`, `401` |
| `GET` | `/tasks` | да | Список своих задач (с пагинацией) | `200 OK` | `400`, `401` |
| `GET` | `/tasks/{id}` | да | Получить одну свою задачу | `200 OK` | `400`, `401`, `404` |
| `POST` | `/tasks` | да | Создать задачу | `201 Created` | `400`, `401` |
| `PUT` | `/tasks/{id}` | да | Обновить задачу (все поля обязательны) | `200 OK` | `400`, `401`, `403`, `404` |
| `DELETE` | `/tasks/{id}` | да | Удалить задачу | `204 No Content` | `400`, `401`, `403`, `404` |

#### Регистрация

```
POST /register
```

```json
{
  "name": "John Doe",
  "email": "john@doe.com",
  "password": "password123"
}
```

Правила: `name` и `email` обязательны, пароль от 8 до 72 символов. Если пользователь с такой почтой уже существует, API отвечает `409 Conflict`.

Ответ `201 Created`:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Вход

```
POST /login
```

```json
{
  "email": "john@doe.com",
  "password": "password123"
}
```

Ответ `200 OK`:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Неизвестная почта и неверный пароль дают одинаковый ответ `401 Unauthorized`, поэтому API не раскрывает, какие почты зарегистрированы.

#### Объект задачи

```json
{
  "id": 1,
  "title": "Buy groceries",
  "description": "Buy milk, eggs, and bread",
  "createdAt": "2026-10-22T09:33:49Z",
  "updatedAt": "2026-10-22T09:33:49Z",
  "userId": 7
}
```

Поля `id`, `createdAt`, `updatedAt` и `userId` выставляет сервер. Клиент отправляет только `title` и `description`, и оба не должны быть пустыми.

#### Создание задачи

```
POST /tasks
```

```json
{
  "title": "Buy groceries",
  "description": "Buy milk, eggs, and bread"
}
```

Ответ `201 Created` с созданной задачей.

#### Обновление задачи

```
PUT /tasks/1
```

```json
{
  "title": "Buy groceries",
  "description": "Buy milk, eggs, bread, and cheese"
}
```

Ответ `200 OK` с обновлённой задачей. Если задача принадлежит другому пользователю, API отвечает `403 Forbidden`. Если её нет, `404 Not Found`.

#### Удаление задачи

```
DELETE /tasks/1
```

Ответ `204 No Content` с пустым телом. `403 Forbidden` для чужой задачи, `404 Not Found`, если задачи нет.

#### Список задач (пагинация)

```
GET /tasks?page=1&limit=10
```

| Параметр | По умолчанию | Правила |
|---|---|---|
| `page` | `1` | положительное целое число |
| `limit` | `10` | положительное целое число, значения больше `100` урезаются до `100` |

Некорректное значение (например, `page=0` или `limit=abc`) даёт `400 Bad Request`.

Ответ `200 OK`:

```json
{
  "data": [
    {
      "id": 1,
      "title": "Buy groceries",
      "description": "Buy milk, eggs, bread",
      "createdAt": "2026-10-22T09:33:49Z",
      "updatedAt": "2026-10-22T09:33:49Z",
      "userId": 7
    }
  ],
  "page": 1,
  "limit": 10,
  "total": 25
}
```

`total` это количество **всех** задач текущего пользователя, а не только задач на текущей странице. Страница за пределами списка возвращает пустой массив `data`.

#### Формат ошибок

Все ошибки возвращаются в JSON:

```json
{
  "error": "Validation failed",
  "details": "Title is required"
}
```

Поле `details` необязательное.

#### Коды статусов

| Код | Значение |
|---|---|
| `200` | OK |
| `201` | Создано |
| `204` | Удалено, тела нет |
| `400` | Некорректный JSON, ошибка валидации, неверный id или параметр запроса |
| `401` | Нет токена, токен недействителен или неверные учётные данные |
| `403` | Задача принадлежит другому пользователю |
| `404` | Задача не найдена |
| `409` | Пользователь уже существует |
| `500` | Непредвиденная ошибка сервера |

Для `GET /tasks/{id}` чужая задача отдаётся как `404`, чтобы не раскрывать, что она существует.

### Примеры запросов

**Windows PowerShell**

```powershell
# регистрация и сохранение токена
$body = @{ name = "John Doe"; email = "john@doe.com"; password = "password123" } | ConvertTo-Json
$resp = Invoke-RestMethod -Method Post -Uri http://localhost:8080/register -ContentType "application/json" -Body $body
$token = $resp.token

# создать задачу
$headers = @{ Authorization = "Bearer $token" }
$task = @{ title = "Buy groceries"; description = "Buy milk, eggs, and bread" } | ConvertTo-Json
Invoke-RestMethod -Method Post -Uri http://localhost:8080/tasks -Headers $headers -ContentType "application/json" -Body $task

# список задач (страница 1, по 10 на странице)
Invoke-RestMethod -Uri "http://localhost:8080/tasks?page=1&limit=10" -Headers $headers
```

**bash / macOS / Linux / Git Bash**

```bash
# регистрация (в ответе придёт токен)
curl -i -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@doe.com","password":"password123"}'

# вход
curl -i -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"john@doe.com","password":"password123"}'

# создать задачу (подставь <token>)
curl -i -X POST http://localhost:8080/tasks \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy groceries","description":"Buy milk, eggs, and bread"}'

# список, обновление, удаление
curl -i "http://localhost:8080/tasks?page=1&limit=10" -H "Authorization: Bearer <token>"
curl -i -X PUT http://localhost:8080/tasks/1 \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy groceries","description":"Buy milk, eggs, bread, and cheese"}'
curl -i -X DELETE http://localhost:8080/tasks/1 -H "Authorization: Bearer <token>"
```

### Тестирование

Скрипт `test-api.ps1` это дымовой тест на PowerShell. При каждом запуске он создаёт новых пользователей и проверяет регистрацию, вход, ошибки токена, изоляцию пользователей, правила `403`/`404`, удаление и пагинацию. Запусти сервер и выполни:

```powershell
.\test-api.ps1
```

### Заметки по безопасности

- Пароли хэшируются через bcrypt и никогда не возвращаются и не пишутся в логи.
- Секрет JWT берётся из окружения и не хранится в коде. Файлы `.env` и `store.db` держи вне Git (добавь их в `.gitignore`).
- Парсер токена принимает только алгоритм HS256.
- Все SQL-запросы параметризованы.
- Каждый запрос к задачам ограничен аутентифицированным пользователем.
- В продакшене API нужно отдавать только по HTTPS.

### Планы

- [ ] Фильтрация и сортировка списка задач
- [ ] Юнит- и интеграционные тесты (`net/http/httptest`)
- [ ] Rate limiting и throttling
- [ ] Механизм refresh-токенов
- [ ] Внешний ключ и `NOT NULL` для `Tasks.user_id`
- [ ] Нормализация почты (нижний регистр) и проверка формата
- [ ] Таймауты сервера и graceful shutdown
