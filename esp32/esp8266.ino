#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>
#include <DHT.h>

// ─── Sensor ────────────────────────────────────────────────────────────────────
#define DHTPIN 4
#define DHTTYPE DHT22
DHT dht(DHTPIN, DHTTYPE);

// ─── Identity ──────────────────────────────────────────────────────────────────
// Change this per device — each room has its own ESP8266
const char* room_id = "room-a";

// ─── Network ───────────────────────────────────────────────────────────────────
const char* ssid          = "Denys-mobile";
const char* password      = "pass";
const char* mqtt_server   = "mqtt_server";

// ─── Topics ────────────────────────────────────────────────────────────────────
// RabbitMQ topic exchange uses dots as separators
char telemetry_topic[64];
char commands_topic[64];

WiFiClient espClient;
PubSubClient client(espClient);

// ─── Command handler ───────────────────────────────────────────────────────────
// Called when a message arrives on classroom/<room_id>/commands
void onCommand(char* topic, byte* payload, unsigned int length) {

  String message;

  for (unsigned int i = 0; i < length; i++) {
    message += (char)payload[i];
  }

  Serial.print("Command received: ");
  Serial.println(message);

  // Parse and act on the command
  StaticJsonDocument<128> doc;
  DeserializationError err = deserializeJson(doc, message);

  if (err) {
    Serial.println("Failed to parse command JSON");
    return;
  }

  const char* cmd = doc["command"];

  if (strcmp(cmd, "ping") == 0) {
    Serial.println("Pong!");
  } else {
    Serial.print("Unknown command: ");
    Serial.println(cmd);
  }
}

// ─── WiFi ──────────────────────────────────────────────────────────────────────
void setup_wifi() {
  delay(10);
  Serial.print("Connecting to ");
  Serial.println(ssid);
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }
  Serial.println("\nWiFi connected");
}

// ─── MQTT reconnect ────────────────────────────────────────────────────────────
void reconnect() {
  while (!client.connected()) {
    Serial.print("Attempting MQTT connection...");

    if (client.connect(room_id, "login", "pass")) {
      Serial.println("connected");

      // Subscribe to commands sent from the server to this room
      client.subscribe(commands_topic);
      Serial.print("Subscribed to: ");
      Serial.println(commands_topic);

    } else {
      Serial.print("failed, rc=");
      Serial.print(client.state());
      Serial.println(" — retrying in 5s");
      delay(5000);
    }
  }
}

// ─── Setup ─────────────────────────────────────────────────────────────────────
void setup() {
  Serial.begin(9600);
  dht.begin();

  // Build topics from room_id
  snprintf(telemetry_topic, sizeof(telemetry_topic), "classroom.%s.telemetry", room_id);
  snprintf(commands_topic,  sizeof(commands_topic),  "classroom.%s.commands",  room_id);

  setup_wifi();

  client.setServer(mqtt_server, 1883);
  client.setCallback(onCommand);
}

// ─── Loop ──────────────────────────────────────────────────────────────────────
void loop() {
  if (!client.connected()) {
    reconnect();
  }
  client.loop();

  float h = dht.readHumidity();
  float t = dht.readTemperature();

  if (isnan(h) || isnan(t)) {
    Serial.println("Failed to read from DHT sensor!");
    return;
  }

  Serial.print("Temp: "); Serial.print(t); Serial.print("°C  ");
  Serial.print("Hum: ");  Serial.print(h); Serial.println("%");

  StaticJsonDocument<200> doc;
  doc["room_id"]     = room_id;
  doc["temperature"] = t;
  doc["humidity"]    = h;
  doc["timestamp"]   = "2023-10-27T10:00:00Z"; // Replace with NTP for real timestamps

  char buffer[256];
  serializeJson(doc, buffer);

  Serial.print("Publishing to ");
  Serial.print(telemetry_topic);
  Serial.print(": ");
  Serial.println(buffer);

  client.publish(telemetry_topic, buffer);

  delay(10000); // 10 second interval
}
