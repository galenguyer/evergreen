package core

import "time"

type Metadata struct {
	Filename string    `json:"filename"`
	Expiry   time.Time `json:"expiry"`
}
