package constant

import "time"

const (
	INT_SIZE    = 4
	PORT_OFFSET = 12000
	MAX_MESSAGE = 150
	IP_ADDR     = "127.0.0.1"
	MIN_DELAY    = 10 * time.Millisecond
	MAX_DELAY    = 100 * time.Millisecond
)
