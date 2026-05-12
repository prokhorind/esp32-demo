package main

type SensorData struct {
	RoomID      string  `json:"room_id"`
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
}

type Command struct {
	Command string `json:"command"`
}
