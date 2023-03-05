package ses

import (
	"fmt"
	"log"
	"strings"
	"sync"

	
)

type SES struct {
	VectorClock *VectorClock
	Queue       []QueueItem
	lock        sync.Mutex
}

type QueueItem struct {
	TimeMsg           *LogicClock
	SourceVectorClock *VectorClock
	Packet            []byte
}

var logger_send *log.Logger = log.New(log.Writer(), "__sender_log__", log.Ldate|log.Ltime|log.Lshortfile)
var logger_receive *log.Logger = log.New(log.Writer(), "__receiver_log__", log.Ldate|log.Ltime|log.Lshortfile)

func NewSES(instanceID, numberProcess int) *SES {
	vectorClock := NewVectorClock(instanceID, numberProcess)
	return &SES{
		VectorClock: vectorClock,
		Queue:       []QueueItem{},
		lock:        sync.Mutex{},
	}
}

func (s *SES) String() string {
	return fmt.Sprintf("", s.VectorClock, s.Queue)
}

func (s *SES) SerializeSES(packet []byte) []byte {
	return s.VectorClock.SerializeVectorClock(packet)
}

func (s *SES) DeserializeSES(packet []byte) (*VectorClock, []byte) {
	vectorClock, remainingPacket := s.VectorClock.DeserializeVectorClock(packet)
	return vectorClock, remainingPacket
}

func (s *SES) MergeSES(sourceVectorClock *VectorClock) {
	for i := 0; i < s.VectorClock.NumberProcess; i++ {
		if i != s.VectorClock.InstanceID && i != sourceVectorClock.InstanceID {
			s.VectorClock.Merge(sourceVectorClock, i, i)
		}
	}
	s.VectorClock.Merge(sourceVectorClock, sourceVectorClock.InstanceID, s.VectorClock.InstanceID)
	s.VectorClock.Increase()
}

func (lc *LogicClock) canDeliver(sourceVectorClock *LogicClock) bool {
	for i := 0; i < lc.NumberProcess; i++ {
		// t_m > t_p
		if sourceVectorClock.Clock[i] > lc.Clock[i] {
			return false
		}
	}
	return true
}

func (s *SES) GetSenderLog(destinationID int, packet []byte) string {
	stringStream := &strings.Builder{}
	fmt.Fprintln(stringStream, "Send Packet Info:")
	fmt.Fprintf(stringStream, "\tSender ID: %d\n", s.VectorClock.InstanceID)
	fmt.Fprintf(stringStream, "\tReceiver ID: %d\n", destinationID)
	fmt.Fprintf(stringStream, "\tPacket Content: %v\n", packet)
	fmt.Fprintf(stringStream, "\tSender Clock:\n")
	fmt.Fprintf(stringStream, "\t\tLocal logical clock: %d\n", s.VectorClock.GetClock(s.VectorClock.InstanceID))
	fmt.Fprintln(stringStream, "\t\tLocal process vectors:")
	for i := 0; i < s.VectorClock.NumberProcess; i++ {
		if i != s.VectorClock.InstanceID && !s.VectorClock.GetClock(i).IsNull() {
			fmt.Fprintf(stringStream, "\t\t\t<P_%d: %v>\n", i, s.VectorClock.GetClock(i))
		}
	}
	return stringStream.String()
}

func (s *SES) GetDeliverLog(tm *LogicClock, sourceVC *VectorClock, packet []byte, status string, header string, printCompare bool) string {
	stringStream := &strings.Builder{}
	fmt.Fprintf(stringStream, "Received Packet Info %s:\n", header)
	fmt.Fprintf(stringStream, "\tSender ID: %d\n", sourceVC.InstanceID)
	fmt.Fprintf(stringStream, "\tReceiver ID: %d\n", s.VectorClock.InstanceID)
	fmt.Fprintf(stringStream, "\tPacket Content: %v\n", packet)
	fmt.Fprintf(stringStream, "\tPacket Clock:\n")
	fmt.Fprintf(stringStream, "\t\tt_m: %d\n", tm)
	fmt.Fprintf(stringStream, "\t\ttP_snd: %d\n", sourceVC.GetClock(sourceVC.InstanceID))
	fmt.Fprintf(stringStream, "\tReceiver Logical Clock (tP_rcv):\n")
	fmt.Fprintf(stringStream, "\t\t%v\n", s.VectorClock.GetClock(s.VectorClock.InstanceID))
	fmt.Fprintf(stringStream, "\tStatus: %s\n", status)
	if printCompare {
		fmt.Fprintf(stringStream, "\tDelivery Condition: %d > %d\n", s.VectorClock.GetClock(s.VectorClock.InstanceID), tm)
	}
	return stringStream.String()
}

func (s *SES) Deliver(packet []byte) {
	s.lock.Lock() // synchronize
	sourceVectorClock, packet := s.DeserializeSES(packet)
	timeProcess := s.VectorClock.GetClock(s.VectorClock.InstanceID) // timestamp send message.
	t_m := sourceVectorClock.GetClock(s.VectorClock.InstanceID)     // timestamp of Process i in the packet
	if timeProcess.canDeliver(t_m) {                                //??????
		// Deliver ???????????(t_m.Clock < timeProcess.Clock)
		//logger_receive.Info(s.GetDeliverLog(t_m, sourceVectorClock, packet, "delivering", "BEFORE DELIVERED", true))
		s.MergeSES(sourceVectorClock)
	} else {
		// Queue
		s.Queue = append(s.Queue, QueueItem{t_m, sourceVectorClock, packet})
		//logger_receive.Info(s.getDeliverLog(t_m, sourceVectorClock, packet, "buffered", "BEFORE BUFFERED", true))
		breakFlag := false
		for !breakFlag {
			breakFlag = true
			for index, item := range s.Queue {
				if timeProcess.canDeliver(t_m) { // ??
					//logger_receive.Info(s.getDeliverLog(item.t_m, item.sourceVectorClock, item.packet, "delivering from buffer", "BEFORE DELIVERED FROM BUFFERED", true))
					s.MergeSES(item.SourceVectorClock)
					s.Queue = append(s.Queue[:index], s.Queue[index+1:]...)
					breakFlag = false
					break
				}
			}
		}
	}
	s.lock.Unlock()
}

func (s *SES) Send(destinationID int, packet []byte) []byte {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.VectorClock.Increase()
	//logger_send.Info(sc.getSenderLog(destinationID, packet))
	result := s.SerializeSES(packet)
	s.VectorClock.SelfMerge(s.VectorClock.InstanceID, destinationID)
	return result
}
