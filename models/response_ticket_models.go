package models

import "time"

// ResponseTicket ...
type ResponseTicket struct {
	AckReference string    `json:"ackReference"`
	AckTimestamp time.Time `json:"ackTimestamp"`
}
