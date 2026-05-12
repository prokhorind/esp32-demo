# Втеча з Arduino IDE Serial Monitor, або Скільки технологій треба, щоб показати 24°C
A real-time classroom environment monitoring system. ESP8266 devices measure temperature and humidity in each room and publish readings over MQTT. A Go backend consumes the data, stores it in PostgreSQL, and exposes a REST API. A React dashboard displays live readings and lets you send commands back to devices.

---

## Architecture

```
ESP8266 (DHT22)          Go Client (emulator)
      │                         │
      │  MQTT publish           │  MQTT publish
      │  classroom/<room>/telemetry
      ▼                         ▼
┌─────────────────────────────────────┐
│         CloudAMQP (RabbitMQ)        │
│                                     │
│  amq.topic exchange                 │
│  binding: classroom.*.telemetry     │
│        → classroom_queue (AMQP)     │
│                                     │
│  classroom.<room>.commands          │
│        → MQTT subscription queue   │
└──────────┬──────────────┬───────────┘
           │              │
           ▼              ▼
    Go Backend         ESP8266 / Go Client
    (consumer)         (command receiver)
           │
           ▼
      PostgreSQL
      (Neon.tech)
           │
           ▼
      REST API (:8080)
           │
           ▼
    React Frontend (:3000)
```

### Topic routing

RabbitMQ's MQTT plugin translates `/` (MQTT) ↔ `.` (AMQP) for topic separators. This means:

| Direction | Topic format | Example |
|---|---|---|
| ESP8266 / Go client → broker (MQTT) | `classroom/<room>/telemetry` | `classroom/room-a/telemetry` |
| Broker → backend consumer (AMQP) | `classroom.<room>.telemetry` | `classroom.room-a.telemetry` |
| Backend → broker (AMQP) | `classroom.<room>.commands` | `classroom.room-a.commands` |
| Broker → ESP8266 / Go client (MQTT) | `classroom/<room>/commands` | `classroom/room-a/commands` |

---

## Components

### ESP8266 (`esp32/esp8266.ino`)

Reads temperature and humidity from a DHT22 sensor every 10 seconds and publishes JSON to `classroom/<room_id>/telemetry`. Subscribes to `classroom/<room_id>/commands` and handles incoming commands (currently: `ping`).

**Configuration** — edit these constants before flashing:

```cpp
const char* room_id    = "room-a";   // unique per device
const char* ssid       = "...";      // WiFi network
const char* password   = "...";      // WiFi password
const char* mqtt_server = "...";     // MQTT broker hostname
```

**Dependencies** (install via Arduino Library Manager):
- `ESP8266WiFi`
- `PubSubClient`
- `ArduinoJson`
- `DHT sensor library`

**Payload published:**
```json
{
  "room_id": "room-a",
  "temperature": 23.5,
  "humidity": 58.2,
  "timestamp": "2026-05-13T10:00:00Z"
}
```

---

### Go Client (`client/`)

A desktop emulator that acts as a software ESP8266. Subscribes to `classroom/<room>/commands` via MQTT and prints received commands to stdout. Useful for testing the command pipeline without physical hardware.

**Run:**
```bash
go run ./client
```

**Output:**
```
Connecting to MQTT...
Connected to RabbitMQ MQTT
MQTT connected
Subscribed to commands for [room-a] on topic: classroom/room-a/commands
[room-a] ← Received command on topic [classroom/room-a/commands]: {"command":"ping"}
```

To add more simulated rooms, extend the `rooms` slice in `client/main.go`:
```go
var rooms = []string{"room-a", "room-b"}
```

---

### Backend (`backend/`)

A Go service with three responsibilities:

1. **AMQP consumer** — listens on `classroom_queue` (bound to `classroom.*.telemetry`) and writes incoming telemetry to PostgreSQL
2. **REST API** — serves telemetry data and accepts commands from the frontend
3. **Command publisher** — publishes commands to `amq.topic` with AMQP routing key `classroom.<room>.commands`

**Environment variables:**

| Variable | Description |
|---|---|
| `RABBITMQ_URL` | AMQP(S) connection string |
| `DATABASE_URL` | PostgreSQL connection string |

**Database schema:**
```sql
CREATE TABLE telemetry (
    id         SERIAL PRIMARY KEY,
    room_id    TEXT NOT NULL DEFAULT 'default',
    temperature FLOAT,
    humidity   FLOAT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

#### REST API

**`GET /rooms`**

Returns a list of all room IDs that have sent telemetry.

```json
["room-a", "room-b"]
```

---

**`GET /telemetry/latest?room_id=<id>`**

Returns the most recent reading for a room.

```json
{
  "room_id": "room-a",
  "temperature": 23.5,
  "humidity": 58.2
}
```

| Status | Condition |
|---|---|
| `200` | Reading found |
| `400` | `room_id` not provided |
| `404` | No data for this room yet |

---

**`GET /telemetry/average?room_id=<id>&from=<iso>&to=<iso>`**

Returns average temperature and humidity over a time range.

```
GET /telemetry/average?room_id=room-a&from=2026-05-06T00:00:00Z&to=2026-05-13T00:00:00Z
```

```json
{
  "room_id": "room-a",
  "temperature": 22.8,
  "humidity": 55.1
}
```

| Status | Condition |
|---|---|
| `200` | Always (returns 0 values if no data in range) |
| `400` | Any of `room_id`, `from`, or `to` missing |

---

**`POST /commands/:room_id`**

Sends a command to a specific room. The backend publishes to `amq.topic` with routing key `classroom.<room_id>.commands`, which the MQTT plugin delivers to subscribed devices.

```json
{ "command": "ping" }
```

```json
{
  "status": "sent",
  "room_id": "room-a",
  "command": "ping"
}
```

| Status | Condition |
|---|---|
| `200` | Command published |
| `400` | Invalid request body |
| `503` | RabbitMQ channel not available |

**Supported commands:**

| Command | ESP8266 behaviour |
|---|---|
| `ping` | Prints `Pong!` to serial |

---

### Frontend (`frontend/`)

A React + Vite dashboard that polls the backend every 10 seconds and displays:

- Room selector (auto-populated from `/rooms`)
- Current temperature and humidity
- Today's average temperature and humidity
- Weekly average temperature and humidity
- Command panel with a **Ping** button per room

**Environment variable** (set at build time):

| Variable | Description |
|---|---|
| `VITE_API_URL` | Backend base URL, e.g. `http://localhost:8080` |

---

## Running locally

**Prerequisites:** Docker, Docker Compose

```bash
docker compose up --build
```

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |

The Go client runs separately outside Docker:

```bash
go run ./client
```

---

## RabbitMQ notes

The project uses [CloudAMQP](https://www.cloudamqp.com) as the managed broker.

- **Exchange:** `amq.topic` (built-in topic exchange)
- **Telemetry queue:** `classroom_queue` — durable, bound to `classroom.*.telemetry`
- **Command queues:** created automatically by the MQTT plugin per connected client (`mqtt-subscription-<clientId>qos1`)

If you see stale routing behaviour after changing binding patterns, delete `classroom_queue` from the RabbitMQ management UI and restart the backend — it will be recreated with the correct binding.
