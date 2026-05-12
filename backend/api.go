package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

var mqChannel *amqp.Channel

func StartAPI() {

	// Reuse AMQP connection for publishing commands
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))

	if err == nil {
		mqChannel, err = conn.Channel()
		if err != nil {
			fmt.Println("Failed to open AMQP channel for commands:", err)
			mqChannel = nil
		}
	} else {
		fmt.Println("Failed to connect AMQP for commands:", err)
	}

	r := gin.Default()
	r.Use(cors.Default())

	// -------------------------------------------------------
	// GET /rooms — list all rooms that have sent data
	// -------------------------------------------------------
	r.GET("/rooms", func(c *gin.Context) {

		rooms, err := GetRooms()

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, rooms)
	})

	// -------------------------------------------------------
	// GET /telemetry/latest?room_id=room-a
	// -------------------------------------------------------
	r.GET("/telemetry/latest", func(c *gin.Context) {

		roomID := c.Query("room_id")

		if roomID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "room_id is required"})
			return
		}

		data, err := GetLatestTelemetry(roomID)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "no data for this room yet"})
			return
		}

		if err != nil {
			log.Printf("GetLatestTelemetry error (room=%s): %v", roomID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, data)
	})

	// -------------------------------------------------------
	// GET /telemetry/average?room_id=room-a&from=...&to=...
	// -------------------------------------------------------
	r.GET("/telemetry/average", func(c *gin.Context) {

		roomID := c.Query("room_id")
		from := c.Query("from")
		to := c.Query("to")

		if roomID == "" || from == "" || to == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "room_id, from and to are required"})
			return
		}

		data, err := GetAverageTelemetry(roomID, from, to)

		if err != nil && err != sql.ErrNoRows {
			log.Printf("GetAverageTelemetry error (room=%s): %v", roomID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, data)
	})

	// -------------------------------------------------------
	// POST /commands/:room_id
	// Body: { "command": "ping" }
	//
	// Publishes to MQTT topic: classroom/<room_id>/commands
	// The ESP8266 in that room receives and acts on it.
	// -------------------------------------------------------
	r.POST("/commands/:room_id", func(c *gin.Context) {

		roomID := c.Param("room_id")

		var cmd Command

		if err := c.ShouldBindJSON(&cmd); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if mqChannel == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "MQTT publisher not available"})
			return
		}

		// RabbitMQ topic exchange uses dots as word separators.
		// The MQTT plugin binds MQTT topic classroom/room-a/commands
		// as classroom.room-a.commands in amq.topic — so we must publish with dots.
		topic := fmt.Sprintf("classroom.%s.commands", roomID)

		log.Printf("Publishing command [%s] to topic: %s", cmd.Command, topic)

		err := mqChannel.Publish(
			"amq.topic",
			topic,
			false,
			false,
			amqp.Publishing{
				ContentType: "application/json",
				Body:        []byte(fmt.Sprintf(`{"command":"%s"}`, cmd.Command)),
			},
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "sent",
			"room_id": roomID,
			"command": cmd.Command,
		})
	})

	r.Run(":8080")
}
