package ses

import (
	"encoding/binary"
	"fmt"
)

type VectorClock struct {
	InstanceID    int
	NumberProcess int
	Vectors       []*LogicClock
}

func NewVectorClock(instanceID, numberProcess int) *VectorClock {
	vc := &VectorClock{
		InstanceID:    instanceID,
		NumberProcess: numberProcess,
		Vectors:       make([]*LogicClock, numberProcess),
	}

	for i := 0; i < numberProcess; i++ {
		vc.Vectors[i] = NewLogicClock(i, numberProcess, i == instanceID)
	}

	return vc
}

func (vc *VectorClock) String() string {
	result := fmt.Sprintf("(%d,%d)", vc.NumberProcess, vc.InstanceID)
	for i := 0; i < vc.NumberProcess; i++ {
		result += fmt.Sprintf("\n%s", vc.Vectors[i])
	}
	return result
}

func (vc *VectorClock) SerializeVectorClock(packet []byte) []byte {
	data := make([]byte, 0)
	b := make([]byte, binary.MaxVarintLen64)
	x := binary.PutUvarint(b, uint64(vc.InstanceID))
	data = append(data, b[:x]...)

	for i := 0; i < vc.NumberProcess; i++ {
		data = append(data, vc.Vectors[i].Serialize()...)
	}
	return append(data, packet...)
}

func (vc *VectorClock) DeserializeVectorClock(packet []byte) (*VectorClock, []byte) {
	dataSize := INT_SIZE * (vc.NumberProcess*vc.NumberProcess + 1)
	data, packet := packet[:dataSize], packet[dataSize:]

	newInstanceID := int(binary.BigEndian.Uint32(data[0:INT_SIZE]))
	newVectorClock := NewVectorClock(newInstanceID, vc.NumberProcess)
	data = data[INT_SIZE:]


	for i := 0; i < vc.NumberProcess; i++ {
		start := INT_SIZE * vc.NumberProcess * i
		end := INT_SIZE * vc.NumberProcess * (i + 1)
		newVectorClock.Vectors[i] = newVectorClock.Vectors[i].Deserialize(data[start:end])
	}

	return newVectorClock, packet
}
func (vc *VectorClock) Increase() {
	vc.Vectors[vc.InstanceID].Increase()
}

func (vc *VectorClock) SelfMerge(sourceID int, destinationID int) {
	vc.Vectors[destinationID].UpdateClock(vc.Vectors[sourceID])
}

func (vc *VectorClock) Merge(sourceVC *VectorClock, sourceID int, destinationID int) {
	vc.Vectors[destinationID].UpdateClock(sourceVC.Vectors[sourceID])
}

func (vc *VectorClock) GetClock(index int) *LogicClock {
	fmt.Print("getclock: ", vc.Vectors)
	return vc.Vectors[index]
}

func (vc *VectorClock) GetLogicalClock(lc *LogicClock) []int {
	return lc.Clock
}
