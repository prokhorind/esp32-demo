package main

import (
	"database/sql"
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

	// Wildcard binding: receives from classroom/<any_room>/telemetry
	err = ch.QueueBind(
		q.Name,
		"classroom.*.telemetry",
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

			if data.RoomID == "" || (data.Temperature == 0 && data.Humidity == 0) {
				log.Printf("Skipping invalid/non-telemetry message on routing key %s: %s", msg.RoutingKey, string(msg.Body))
				continue
			}

			log.Printf("Received from room [%s]: temp=%.1f hum=%.1f",
				data.RoomID, data.Temperature, data.Humidity)

			SaveTelemetry(data)
		}
	}()
}

func SaveTelemetry(data SensorData) {

	_, err := DB.Exec(`
        INSERT INTO telemetry (room_id, temperature, humidity)
        VALUES ($1, $2, $3)
    `,
		data.RoomID,
		data.Temperature,
		data.Humidity,
	)

	if err != nil {
		log.Println(err)
	}
}

func GetRooms() ([]string, error) {

	rows, err := DB.Query(`
        SELECT DISTINCT room_id
        FROM telemetry
        ORDER BY room_id
    `)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var rooms []string

	for rows.Next() {
		var room string
		if err := rows.Scan(&room); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func GetLatestTelemetry(roomID string) (SensorData, error) {

	stmt, err := DB.Prepare(`
        SELECT room_id, temperature, humidity
        FROM telemetry
        WHERE room_id = $1
        ORDER BY created_at DESC
        LIMIT 1
    `)
	if err != nil {
		return SensorData{RoomID: roomID}, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(roomID)

	var s SensorData

	err = row.Scan(
		&s.RoomID,
		&s.Temperature,
		&s.Humidity,
	)

	if err == sql.ErrNoRows {
		return SensorData{RoomID: roomID}, sql.ErrNoRows
	}

	return s, err
}

func GetAverageTelemetry(roomID, from, to string) (SensorData, error) {

	stmt, err := DB.Prepare(`
        SELECT
            COALESCE(AVG(temperature), 0),
            COALESCE(AVG(humidity), 0)
        FROM telemetry
        WHERE room_id = $1
        AND created_at BETWEEN $2 AND $3
    `)
	if err != nil {
		return SensorData{RoomID: roomID}, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(roomID, from, to)

	s := SensorData{RoomID: roomID}

	err = row.Scan(
		&s.Temperature,
		&s.Humidity,
	)

	return s, err
}
