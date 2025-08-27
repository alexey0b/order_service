<h1 align="center">Order Service</h1>

## 📋 Описание проекта

**Order Service** — это микросервис для управления заказами. Сервис обрабатывает заказы из Kafka, сохраняет их в PostgreSQL и предоставляет HTTP API для получения информации о заказах с кешированием.

---

## 🚀 Запуск сервиса 

### Предварительные требования

- Docker и Docker Compose
- Go 1.24+ (для локального запуска)

### Запуск сервиса

1. **Запустите `Docker Engine`**

2. **Клонируйте репозиторий:**

```bash
git clone https://github.com/alexey0b/order_service.git
```
-  Перейдите в корень проекта

3. **Запустите БД:**

```bash
make postgres-start
```

4. **Заведите пользователя для БД:**

```bash
make postgres-create-user NAME=username PASSWORD=password
```

5. **Примените миграции:**

- Установите пакет Goose

```bash
make install-goose
```

- Мигрируйте таблицы в БД

```bash
make migrate-up
```

6. **Выдайте права на созданную БД:**

```bash
make postgres-grant-permissions NAME=username
```

- Чтобы сервис взаимодействовал с БД от лица созданного пользователя, то введите его **NAME** и **PASSWORD** в файле конфигурации у значений полей `user` и `password`:

```yaml
# path: /order_service/config/config.yaml
---
# Database configuration
postgres:
  host: "postgres"
  port: 5432
  database: "my_db"
  user: "user"
  password: "password"
  ssl_mode: "disable"
```

7. **Запустите брокер Kafka:**

```bash
make broker-start
```

8. **Создайте Kafka-топик:**

```bash
make broker-create-topic NAME=topic_name
```

- Введите созданное имя топика **NAME** в файле конфигурации у значения поля `topic`

```yaml
# path: /order_service/config/config.yaml
---
# Kafka configuration
kafka:
  network: "tcp"
  brokers: ["broker:9092"]
  topic: "my-topic"
  group_id: "1"
  poll_timeout: 1000 # in milliseconds
```

9. **Запустите go-сервис:**

```bash
make app-start
```

10. **Отправьте тестовые сообщения на созданный топик:**

```bash
make broker-send-msgs
```

### API c UI интерфейсом доступен по адресу: [http://localhost:8080](http://localhost:8080)

**Пример запроса:**

```bash
curl http://localhost:8080/api/v1/order/b563feb7b2b84b6test
```

---

## 📊 Мониторинг и метрики

- **Запустите Prometheus через docker compose**:

```bash
make promo-start
```

### Prometheus метрики

Сервис предоставляет метрики для мониторинга через Prometheus:

- **Эндпоинт метрик**: [http://localhost:8080/metrics](http://localhost:8080/metrics)
- **Prometheus UI**: [http://localhost:9090](http://localhost:9090)

**Доступные метрики:**
- `app_requests_total` - общее количество запросов
- `app_request_duration_seconds` - время обработки запросов

---

## 🛠️ Технические ресурсы

- **Язык программирования**: Go (Golang)
- **База данных**: PostgreSQL 

### Основные библиотеки 
- **[jmoiron/sqlx](https://github.com/jmoiron/sqlx)** - работа с PostgreSQL
- **[jackc/pgx](https://github.com/jackc/pgx)** - PostgreSQL драйвер
- **[segmentio/kafka-go](https://github.com/segmentio/kafka-go)** - Kafka клиент
- **[hashicorp/golang-lru](https://github.com/hashicorp/golang-lru)** - LRU кеш
- **[pressly/goose](https://github.com/pressly/goose)** - миграции БД
- **[spf13/viper](https://github.com/spf13/viper)** - конфигурация
- **[prometheus/client_golang](https://github.com/prometheus/client_golang)** - метрики Prometheus

### Библиотеки для тестирования

- **[stretchr/testify](https://github.com/stretchr/testify)** - assertions
- **[uber-go/mock](https://github.com/uber-go/mock)** - моки
- **[testcontainers-go](https://github.com/testcontainers/testcontainers-go)** - интеграционные тесты

---

## 📚 Полезные команды

```bash
# Посмотреть все доступные команды
make help

# Запустить unit тесты
make unit-test-start

# Запустить интеграционные тесты
make integration-test-start

# Проверить статус миграций
make migrate-status

```

---
