package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type SensorData struct {
	ID          string  `json:"id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Light       int     `json:"light"`
	Timestamp   string  `json:"timestamp"`
}

func main() {

	opts := mqtt.NewClientOptions()

	opts.AddBroker("tcp://localhost:1883")

	opts.SetClientID("go-esp32-emulator")

	opts.SetUsername("admin")
	opts.SetPassword("admin")

	// Auto reconnect
	opts.AutoReconnect = true
	opts.ConnectRetry = true
	opts.ConnectRetryInterval = 5 * time.Second

	// Debug callbacks
	opts.OnConnect = func(c mqtt.Client) {
		fmt.Println("MQTT connected")
	}

	opts.OnConnectionLost = func(
		c mqtt.Client,
		err error,
	) {
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

	rand.Seed(time.Now().UnixNano())

	for {

		data := SensorData{
			ID:          generateID(),
			Temperature: randomFloat(20, 30),
			Humidity:    randomFloat(40, 70),
			Light:       rand.Intn(1000),
			Timestamp:   time.Now().Format(time.RFC3339),
		}

		payload, err := json.Marshal(data)

		if err != nil {

			fmt.Println("JSON marshal failed:", err)

			time.Sleep(2 * time.Second)

			continue
		}

		fmt.Println("-----------------------------------")
		fmt.Println("Publishing message")
		fmt.Println("Topic: classroom/telemetry")
		fmt.Println("Payload:", string(payload))

		token := client.Publish(
			"classroom/telemetry",
			1,     // QoS 1
			false, // retain
			payload,
		)

		ok := token.WaitTimeout(5 * time.Second)

		if !ok {

			fmt.Println("Publish timeout")

			time.Sleep(2 * time.Second)

			continue
		}

		if token.Error() != nil {

			fmt.Println("Publish failed:", token.Error())

			time.Sleep(2 * time.Second)

			continue
		}

		fmt.Println("Message delivered successfully")

		time.Sleep(5 * time.Second)
	}
}

func randomFloat(
	min float64,
	max float64,
) float64 {

	return min + rand.Float64()*(max-min)
}

func generateID() string {

	return fmt.Sprintf(
		"msg-%d",
		time.Now().UnixNano(),
	)
}
