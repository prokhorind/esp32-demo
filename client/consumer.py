import paho.mqtt.client as mqtt

BROKER = "sparrow.rmq.cloudamqp.com"
PORT = 1883

USERNAME = "username"
PASSWORD = "password"

TOPIC = "classroom/+/telemetry"


def on_connect(client, userdata, flags, rc):

    if rc == 0:
        print("Connected successfully!")

        client.subscribe(TOPIC)

    else:
        print("Connection failed:", rc)


def on_message(client, userdata, msg):

    print("Topic:", msg.topic)
    print("Payload:", msg.payload.decode())


client = mqtt.Client()

# Login/password
client.username_pw_set(USERNAME, PASSWORD)

# Callbacks
client.on_connect = on_connect
client.on_message = on_message

# Connect to MQTT broker
client.connect(BROKER, PORT, 60)

print("Waiting for telemetry...")

client.loop_forever()