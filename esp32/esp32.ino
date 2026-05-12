#include <WiFi.h>
#include <PubSubClient.h>

const char* WIFI_SSID = "YOUR_WIFI";
const char* WIFI_PASSWORD = "YOUR_PASSWORD";

const char* MQTT_SERVER = "192.168.0.100"; // IP вашого Docker host
const int MQTT_PORT = 1883;

const char* MQTT_USER = "admin";
const char* MQTT_PASSWORD = "admin";

WiFiClient espClient;
PubSubClient client(espClient);

void connectWiFi() {

  WiFi.begin(WIFI_SSID, WIFI_PASSWORD);

  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }

  Serial.println();
  Serial.println("WiFi connected");
}

void connectMQTT() {

  while (!client.connected()) {

    Serial.println("Connecting MQTT...");

    if (
      client.connect(
        "esp32-client",
        MQTT_USER,
        MQTT_PASSWORD
      )
    ) {

      Serial.println("MQTT connected");

    } else {

      Serial.print("Failed: ");
      Serial.println(client.state());

      delay(2000);
    }
  }
}

void setup() {

  Serial.begin(115200);

  connectWiFi();

  client.setServer(MQTT_SERVER, MQTT_PORT);
}

void loop() {

  if (!client.connected()) {
    connectMQTT();
  }

  client.loop();

  float temperature = 24.5;
  float humidity = 55.0;
  int light = 800;

  String payload = "{";
  payload += "\"temperature\":";
  payload += temperature;
  payload += ",";
  payload += "\"humidity\":";
  payload += humidity;
  payload += ",";
  payload += "\"light\":";
  payload += light;
  payload += "}";

  Serial.println(payload);

  client.publish(
    "classroom/telemetry",
    payload.c_str()
  );

  delay(5000);
}