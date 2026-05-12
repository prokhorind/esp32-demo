package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type SensorData struct {
	RoomID      string  `json:"room_id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Timestamp   string  `json:"timestamp"`
}

// Simulated rooms — each acts as an independent ESP8266
var rooms = []string{"room-a"}

func main() {

	opts := mqtt.NewClientOptions()

	opts.AddBroker("tcp://sparrow.rmq.cloudamqp.com:1883")
	opts.SetClientID("go-esp8266-emulator")
	opts.SetUsername("username")
	opts.SetPassword("password")

	opts.AutoReconnect = true
	opts.ConnectRetry = true
	opts.ConnectRetryInterval = 5 * time.Second
	opts.SetCleanSession(true)

	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("MQTT connected")
		subscribeCommands(c)
	}

	opts.OnConnectionLost = func(c mqtt.Client, err error) {
		fmt.Println("MQTT connection lost:", err)
	}

	client := mqtt.NewClient(opts)

	fmt.Println("Connecting to MQTT...")

	token := client.Connect()

	ok := token.WaitTimeout(10 * time.Second)

	if !ok {
		panic("MQTT connection timeout")
	}

	if token.Error() != nil {
		panic(token.Error())
	}

	fmt.Println("Connected to RabbitMQ MQTT")

	// Publish telemetry in background
	go publishLoop(client)

	// Block forever — subscription callbacks run in background goroutines
	select {}
}

func publishLoop(c mqtt.Client) {
	for {
		for _, roomID := range rooms {

			data := SensorData{
				RoomID:      roomID,
				Temperature: randomFloat(20, 30),
				Humidity:    randomFloat(40, 70),
				Timestamp:   time.Now().Format(time.RFC3339),
			}

			payload, err := json.Marshal(data)
			if err != nil {
				fmt.Println("JSON marshal failed:", err)
				continue
			}

			topic := fmt.Sprintf("classroom/%s/telemetry", roomID)

			fmt.Printf("[%s] Publishing → %s\n", roomID, string(payload))

			token := c.Publish(topic, 1, false, payload)

			ok := token.WaitTimeout(5 * time.Second)
			if !ok {
				fmt.Printf("[%s] Publish timeout\n", roomID)
				continue
			}
			if token.Error() != nil {
				fmt.Printf("[%s] Publish failed: %v\n", roomID, token.Error())
			}
		}

		time.Sleep(30 * time.Second)
	}
}

// subscribeCommands listens on classroom/<room>/commands for each simulated room
func subscribeCommands(c mqtt.Client) {

	for _, roomID := range rooms {

		room := roomID // capture loop variable

		topic := fmt.Sprintf("classroom/%s/commands", room)

		token := c.Subscribe(topic, 1, func(_ mqtt.Client, msg mqtt.Message) {
			fmt.Printf("[%s] ← Received command on topic [%s]: %s\n", room, msg.Topic(), string(msg.Payload()))
		})

		token.Wait()

		if token.Error() != nil {
			fmt.Printf("Failed to subscribe to commands for [%s]: %v\n", room, token.Error())
		} else {
			fmt.Printf("Subscribed to commands for [%s] on topic: %s\n", room, topic)
		}
	}
}

func randomFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}
