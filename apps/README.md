# Smart Home Sensor Management API

## Prerequisites

- Docker and Docker Compose

## Services

| Сервис | Язык | Порт | Описание |
|---|---|---|---|
| `app` (smart_home) | Go | 8080 | Монолит-фасад (Strangler Fig), проксирует в микросервисы |
| `device-management` | Python (FastAPI) | 8082 | CRUD устройств, PostgreSQL, публикация событий |
| `temperature-telemetry` | Java (Spring Boot) | 8083 | Телеметрия, кэш, подписка на RabbitMQ |
| `temperature-api` | Go | 8081 | Имитация внешнего датчика |
| `postgres` | — | 5432 | БД `smarthome` (монолит) + `device_management` (микросервис) |
| `rabbitmq` | — | 5672 / 15672 | Брокер доменных событий |

### Архитектура MVP (Strangler Fig)

```
Клиент → Монолит (Go) :8080
           ├─ REST → Device Management (Python) :8082 → PostgreSQL
           ├─ REST → Temperature Telemetry (Java) :8083 → temperature-api
           └─ fallback → temperature-api :8081

Device Management ──RabbitMQ──► Temperature Telemetry
  (device.created / updated / deleted)   (инвалидация кэша)
```

Монолит при `USE_MICROSERVICES=true` делегирует Create/Get Sensors в микросервисы.
Операции Update/Delete пока остаются в монолитной БД (постепенная миграция).

## Getting Started

### Option 1: Using Docker Compose (Recommended)

The easiest way to start the application is to use Docker Compose:

```bash
./init.sh
```

This script will:

1. Build and start the PostgreSQL and application containers
2. Wait for the services to be ready
3. Display information about how to access the API

Alternatively, you can run Docker Compose directly:

```bash
docker-compose up -d
```

The API will be available at http://localhost:8080

### Option 2: Manual setup

If you prefer to run the application without Docker:

1. Start the PostgreSQL database:

```bash
docker-compose up -d postgres
```

2. Build and run the application:

```bash
go build -o smarthome
./smarthome
```

## API Testing

A Postman collection is provided for testing the API. Import the `smarthome-api.postman_collection.json` file into Postman to get started.

## API Endpoints

- `GET /health` - Health check
- `GET /api/v1/sensors` - Get all sensors
- `GET /api/v1/sensors/:id` - Get a specific sensor
- `POST /api/v1/sensors` - Create a new sensor
- `PUT /api/v1/sensors/:id` - Update a sensor
- `DELETE /api/v1/sensors/:id` - Delete a sensor
- `PATCH /api/v1/sensors/:id/value` - Update a sensor's value and status
