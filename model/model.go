package model

import (
	"time"
)

type BasicPayload struct {
	Name string    `json:"name"`
	Uuid string    `json:"uuid"`
	Date time.Time `json:"date"`
}
