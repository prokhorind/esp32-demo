package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func StartConsumer() {

	var conn *amqp.Connection
	var err error

	for {

		conn, err = amqp.Dial(
			os.Getenv("RABBITMQ_URL"),
		)

		if err == nil {
			break
		}

		log.Println("RabbitMQ not ready... retrying")

		time.Sleep(3 * time.Second)
	}

	log.Println("Connected to RabbitMQ")
	ch, err := conn.Channel()

	if err != nil {
		log.Fatal(err)
	}

	q, err := ch.QueueDeclare(
		"classroom_queue",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	err = ch.QueueBind(
		q.Name,
		"#",
		"amq.topic",
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatal(err)
	}

	go func() {

		for msg := range msgs {

			var data SensorData

			err := json.Unmarshal(msg.Body, &data)

			if err != nil {
				log.Println(err)
				continue
			}

			log.Println("Received:", data)

			SaveTelemetry(data)
		}
	}()
}

func SaveTelemetry(data SensorData) {

	_, err := DB.Exec(
		`
        INSERT INTO telemetry
        (temperature, humidity)
        VALUES ($1, $2)
        `,
		data.Temperature,
		data.Humidity,
	)

	if err != nil {
		log.Println(err)
	}
}

func GetLatestTelemetry() (SensorData, error) {

	row := DB.QueryRow(`
        SELECT temperature, humidity
        FROM telemetry
        ORDER BY created_at DESC
        LIMIT 1
    `)

	var s SensorData

	err := row.Scan(
		&s.Temperature,
		&s.Humidity,
	)

	return s, err
}

func GetAverageTelemetry(
	from string,
	to string,
) (SensorData, error) {

	row := DB.QueryRow(`
        SELECT
            COALESCE(AVG(temperature), 0),
            COALESCE(AVG(humidity), 0)
        FROM telemetry
        WHERE created_at
        BETWEEN $1 AND $2
    `,
		from,
		to,
	)

	var s SensorData

	err := row.Scan(
		&s.Temperature,
		&s.Humidity,
	)

	return s, err
}
