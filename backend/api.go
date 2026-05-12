package main

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartAPI() {

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/telemetry/latest", func(c *gin.Context) {

		data, err := GetLatestTelemetry()

		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, data)
	})

	r.GET("/telemetry/average", func(c *gin.Context) {

		from := c.Query("from")
		to := c.Query("to")

		if from == "" || to == "" {

			c.JSON(http.StatusBadRequest, gin.H{
				"error": "from and to are required",
			})

			return
		}

		data, err := GetAverageTelemetry(
			from,
			to,
		)

		if err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		c.JSON(http.StatusOK, data)
	})

	r.Run(":8080")
}
