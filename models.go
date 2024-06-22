package main

import "time"

// Config ...
type Config struct {
	HTTP struct {
		Port int `yaml:"port"`
	} `yaml:"http"`
	System struct {
		CookieName string `yaml:"cookie_name"`
	} `yaml:"system"`
}

// ResponseTicket ...
type ResponseTicket struct {
	AckReference string    `json:"ackReference"`
	AckTimestamp time.Time `json:"ackTimestamp"`
}
