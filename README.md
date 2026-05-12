# Smart Classroom

Architecture:


ESP32 -> RabbitMQ -> Go Backend -> PostgreSQL
Reacr -HTTP> Go Service -> HTML 
## Start project

```bash
docker compose up --build
```

ipconfig getifaddr en0

RabbitMQ dashboard:
http://localhost:15672
https://sparrow.rmq.cloudamqp.com/#/

App
https://render.com

Dashboard
http://localhost:5173/

Login:
admin
admin

API:
GET http://localhost:8080/telemetry/latest
GET http://localhost:8080/telemetry/average?from=2026-05-05T13:43:10.380Z&to=2026-05-12T13:43:10.38
