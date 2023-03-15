package constant

import "time"

const (
	INT_SIZE    = 4
	PORT_OFFSET = 1200
	MAX_MESSAGE = 20
	IP_ADDR     = "127.0.0.1"
	MIN_DELAY    = 100 * time.Millisecond
	MAX_DELAY    = 1000 * time.Millisecond
)
