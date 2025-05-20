# Auth Service

Сервис аутентификации на Go + PostgreSQL + Docker

## Быстрый старт

1. Убедитесь, что у вас установлен Docker и docker-compose.
2. В корне проекта выполните:

```powershell
docker-compose -f docker-compose.yml up -d
```

3. Swagger-документация будет доступна по адресу: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## Переменные окружения

Создайте файл `.env` в корне проекта со следующим содержимым (или используйте переменные окружения, как в docker-compose.yml):

```
SERVER_PORT=8080
DB_HOST=db
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=auth_service_db
DB_SSL_MODE=disable
JWT_ACCESS_SECRET=my_super_secret_access_key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_SECRET=my_super_secret_refresh_key
JWT_REFRESH_EXPIRY=720h
WEBHOOK_URL=https://webhook.site/your-test-id
```

## Примеры запросов для PowerShell (Windows)

### 1. Сгенерировать GUID пользователя

В PowerShell выполните:
```powershell
[guid]::NewGuid()
```
Скопируйте полученное значение (например, `b3e1c2a7-2f3b-4c8e-9e2a-1a2b3c4d5e6f`).

### 2. Получить пару токенов (access и refresh)
```powershell
$headers = @{ "User-Agent" = "test-agent" }
Invoke-WebRequest -Uri "http://localhost:8080/auth/login?user_id=<GUID>" -Method POST -Headers $headers
```
Замените `<GUID>` на ваш сгенерированный идентификатор.

**Важно!** Чтобы увидеть полный ответ, используйте:
```powershell
$response = Invoke-WebRequest -Uri "http://localhost:8080/auth/login?user_id=<GUID>" -Method POST -Headers $headers
$response.Content
```
В ответе будут поля `access_token` и `refresh_token`.

### 3. Обновить токены
```powershell
$headers = @{ 
  "User-Agent" = "test-agent"
  "Content-Type" = "application/json"
}
$body = '{"refresh_token": "<refresh_token>"}'
Invoke-WebRequest -Uri "http://localhost:8080/auth/refresh" -Method POST -Headers $headers -Body $body
```
Замените `<refresh_token>` на значение из предыдущего ответа (без скобок и кавычек внутри строки).

### 4. Получить текущего пользователя
```powershell
$headers = @{ "Authorization" = "Bearer <access_token>" }
Invoke-WebRequest -Uri "http://localhost:8080/user/me" -Method GET -Headers $headers
```
Замените `<access_token>` на актуальный токен.

### 5. Деавторизация (выход)
```powershell
$headers = @{ "Authorization" = "Bearer <access_token>" }
Invoke-WebRequest -Uri "http://localhost:8080/auth/logout" -Method POST -Headers $headers
```

## Swagger UI

Визуальная документация и тестирование API доступны по адресу:
[http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## Частые ошибки и их решения

- **refresh_token невалидный** — убедитесь, что вы используете именно тот refresh_token, который получили в ответе на логин, и не добавляете лишних символов (например, скобок).
- **Swagger UI не открывается** — проверьте, что контейнеры запущены (`docker ps`), порт 8080 не занят, и нет ошибок в логах (`docker-compose logs`).
- **В ответе нет refresh_token** — проверьте полный вывод `$response.Content`.
- **Проблемы с PowerShell** — убедитесь, что используете двойные кавычки для строк и правильно экранируете переменные.

---

**Требования тестового задания выполнены:**
- Запуск одной командой
- Swagger UI с ошибками и примерами
- Примеры запросов для Windows
- Документация 