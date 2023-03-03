package ses

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type VectorClock struct {
	InstanceID int
	NumberProcess  int
	vectors    []*LogicClock
}

func NewVectorClock(instanceID int, numberProcess int) *VectorClock {
	vc := &VectorClock{
		InstanceID: instanceID,
		NumberProcess:  numberProcess,
		vectors:    make([]*LogicClock, numberProcess),
	}

	for i := 0; i < numberProcess; i++ {
		vc.vectors[i] = NewLogicClock(numberProcess, i, i == instanceID)
	}

	return vc
}

func (vc *VectorClock) String() string {
	result := fmt.Sprintf("(%d,%d)", vc.NumberProcess, vc.InstanceID)
	for i := 0; i < vc.NumberProcess; i++ {
		result += fmt.Sprintf("\n%s", vc.vectors[i])
	}
	return result
}

func (vc *VectorClock) Serialize(packet []byte) []byte {
	var data bytes.Buffer

	binary.Write(&data, binary.LittleEndian, vc.InstanceID)

	for i := 0; i < vc.NumberProcess; i++ {
		data.Write(vc.vectors[i].Serialize())
	}

	return append(data.Bytes(), packet...)
}

func DeserializeVectorClock(packet []byte, numberProcess int) (*VectorClock, []byte) {
	dataSize := INT_SIZE * (numberProcess*numberProcess + 1)
	data, packet := packet[:dataSize], packet[dataSize:]

	instanceID := int(binary.BigEndian.Uint32(data[:INT_SIZE]))
	vc := NewVectorClock(numberProcess, instanceID)

	data = data[INT_SIZE:]

	for i := 0; i < vc.NumberProcess; i++ {
		vc.vectors[i].Deserialize(data[INT_SIZE*vc.NumberProcess*i : INT_SIZE*vc.NumberProcess*(i+1)])
	}

	return vc, packet
}

func (vc *VectorClock) Increase() {
	vc.vectors[vc.InstanceID].Increase()
}

func (vc *VectorClock) SelfMerge(sourceID int, destinationID int) {
	vc.vectors[destinationID].Merge(vc.vectors[sourceID])
}

func (vc *VectorClock) Merge(sourceVC *VectorClock, sourceID int, destinationID int) {
	vc.vectors[destinationID].Merge(sourceVC.vectors[sourceID])
}

func (vc *VectorClock) GetClock(index int) *LogicClock {
	return vc.vectors[index]
}
