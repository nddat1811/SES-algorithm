package ses

import (
	"encoding/binary"
	"fmt"
)

const (
	INT_SIZE    = 4
	INT_REPR    = "big"
	PORT_OFFSET = 60000
	MAX_MESSAGE = 150
)

type LogicClock struct {
	NumberProcess int
	InstanceID    int
	Clock         []int
}

func NewLogicClock(numberProcess, instanceID int, zeroFill bool) *LogicClock {
	clock := make([]int, numberProcess)
	if zeroFill {
		for i := 0; i < numberProcess; i++ {
			clock[i] = 0
		}
	} else {
		for i := 0; i < numberProcess; i++ {
			clock[i] = -1
		}
	}
	return &LogicClock{numberProcess, instanceID, clock}
}

func (lc *LogicClock) String() string {
	return fmt.Sprintf("%v", lc.GetTime())
}

func (lc *LogicClock) GetTime() []int {
	return lc.Clock
}

func (lc *LogicClock) IsNull() bool {
	for _, c := range lc.Clock {
		if c == -1 {
			return true
		}
	}
	return false
}

func (lc *LogicClock) Increase() {
	lc.Clock[lc.InstanceID]++
}

func (lc *LogicClock) UpdateClock(other *LogicClock) {
	if lc.IsNull() {
		for i, c := range other.GetTime() {
			lc.Clock[i] = c
		}
	} else {
		for i, c := range other.GetTime() {
			if lc.Clock[i] < c {
				lc.Clock[i] = c
			}
		}
	}
}

func (c *LogicClock) Serialize() []byte {
	data := make([]byte, 0)
	for i := 0; i < c.NumberProcess; i++ {
		b := make([]byte, INT_SIZE)
		binary.LittleEndian.PutUint32(b, uint32(c.Clock[i]))
		data = append(data, b...)
	}
	return data
}

func (lc *LogicClock) Deserialize(data []byte) *LogicClock {
	newClock := &LogicClock{
		NumberProcess: lc.NumberProcess,
		InstanceID:    lc.InstanceID,
		Clock:         make([]int, lc.NumberProcess),
	}
	for i := 0; i < lc.NumberProcess; i++ {
		newClock.Clock[i] = int(int32(binary.LittleEndian.Uint32(data[INT_SIZE*i : INT_SIZE*(i+1)])))
	}
	return newClock
}

func (lc *LogicClock) Equal(other *LogicClock) bool {
	for i, c := range lc.Clock {
		if c != other.GetTime()[i] {
			return false
		}
	}
	return true
}

func (lc *LogicClock) LessThan(other *LogicClock) bool {
	return lc.LessThanOrEqual(other) && !lc.Equal(other)
}

func (lc *LogicClock) LessThanOrEqual(other *LogicClock) bool {
	for i, c := range lc.Clock {
		if c > other.GetTime()[i] {
			return false
		}
	}
	return true
}
