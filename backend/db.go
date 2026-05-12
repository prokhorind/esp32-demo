package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {

	var err error

	DB, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))

	if err != nil {
		log.Fatal(err)
	}

	query := `
    CREATE TABLE IF NOT EXISTS telemetry (
        id SERIAL PRIMARY KEY,
        temperature FLOAT,
        humidity FLOAT,
        light INTEGER,
        created_at TIMESTAMP DEFAULT NOW()
    )
    `

	_, err = DB.Exec(query)

	if err != nil {
		log.Fatal(err)
	}
}
