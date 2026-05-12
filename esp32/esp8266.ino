#include <ESP8266WiFi.h>
#include <PubSubClient.h>
#include <ArduinoJson.h>
#include <DHT.h>

// Pins and Sensor
#define DHTPIN 4
#define DHTTYPE DHT22
DHT dht(DHTPIN, DHTTYPE);

// Network Credentials
const char* ssid = "Denys-mobile";
const char* password = "pass";
const char* mqtt_server = "172.20.10.10";

WiFiClient espClient;
PubSubClient client(espClient);

void setup() {
  Serial.begin(9600);
  dht.begin();

  setup_wifi();
  client.setServer(mqtt_server, 1883);
}

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

void reconnect() {
  while (!client.connected()) {
    Serial.print("Attempting MQTT connection...");
    // Matches your Go credentials: ClientID, Username, Password
    if (client.connect("arduino-client", "admin", "admin")) {
      Serial.println("connected");
    } else {
      Serial.print("failed, rc=");
      Serial.print(client.state());
      Serial.println(" try again in 5 seconds");
      delay(5000);
    }
  }
}

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

  StaticJsonDocument<200> doc;
  doc["id"] = "msg-" + String(millis());
  doc["temperature"] = t;
  doc["humidity"] = h;
  doc["timestamp"] = "2023-10-27T10:00:00Z"; // ESP32 needs NTP for real timestamps

  char buffer[256];
  serializeJson(doc, buffer);

  Serial.print("Publishing: ");
  Serial.println(buffer);

  // Publish to the same topic as your Go code
  client.publish("classroom/telemetry", buffer);

  delay(10000); // 10 second interval
}