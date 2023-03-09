package ses

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
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

// var generalLog, senderLog, receiverLog *os.File

func InitLog(instanceID, numberProcess int) {
	var err error
	folder := fmt.Sprintf("./static/logs/%02d_process", numberProcess)
	os.Mkdir(folder, os.ModePerm)

	senderLog, err := os.Create(
		fmt.Sprintf("./static/logs/%02d_process/%02d__sender.log", numberProcess, instanceID),
	)
	if err != nil {
		log.Fatalf("Failed to open sender log file: %v", err)
	}

	receiverLog, err := os.Create(
		fmt.Sprintf("./static/logs/%02d_process/%02d__receiver.log",numberProcess, instanceID),
	)
	if err != nil {
		log.Fatalf("Failed to open receiver log file: %v", err)
	}
	log.New(senderLog, "__sender_log__ ", log.Ldate|log.Ltime)
	log.New(receiverLog, "__receiver_log__ ", log.Ldate|log.Ltime)
}

func NewSES(instanceID, numberProcess int) *SES {
	vectorClock := NewVectorClock(instanceID, numberProcess)
	return &SES{
		VectorClock: vectorClock,
		Queue:       []QueueItem{},
		lock:        sync.Mutex{},
	}
}

func (s *SES) String() string {
	return fmt.Sprintf("%s, %s", s.VectorClock, s.Queue)
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
	currentTime := time.Now()
	nano := currentTime.Nanosecond()
	formattedTime := currentTime.Format("2006-01-02T15:04:05.") + fmt.Sprintf("%03d", nano/1000000) + "Z"
	fmt.Fprintf(stringStream, "Current Time: %v\n", formattedTime)
	fmt.Fprintln(stringStream, "Send Packet Info:")
	fmt.Fprintf(stringStream, "\tSender ID: %d\n", s.VectorClock.InstanceID)
	fmt.Fprintf(stringStream, "\tReceiver ID: %d\n", destinationID)
	fmt.Fprintf(stringStream, "\tPacket Content: %v\n", string(packet))
	fmt.Fprintf(stringStream, "\tSender Clock:\n")

	fmt.Fprintf(stringStream, "\t\tLocal logical clock:%v\n", s.VectorClock.GetLogicalClock(s.VectorClock.InstanceID))
	fmt.Fprintln(stringStream, "\t\tLocal process vectors:")
	for i := 0; i < s.VectorClock.NumberProcess; i++ {
		if i != s.VectorClock.InstanceID && !s.VectorClock.GetClock(i).IsNull() {
			fmt.Fprintf(stringStream, "\t\t\t<P_%d: %v>\n", i, s.VectorClock.GetClock(i))
		}
	}
	// fmt.Println("\n")
	// fmt.Print(stringStream.String())
	// fmt.Println("\n\n")
	return stringStream.String()
}

func (s *SES) GetDeliverLog(tm *LogicClock, sourceVC *VectorClock, packet []byte, status string, header string, printCompare bool) string {
	stringStream := &strings.Builder{}
	currentTime := time.Now()
	nano := currentTime.Nanosecond()
	formattedTime := currentTime.Format("2006-01-02T15:04:05.") + fmt.Sprintf("%03d", nano/1000000) + "Z"
	fmt.Fprintf(stringStream, "Current Time: %v\n", formattedTime)
	fmt.Fprintf(stringStream, "Received Packet Info %s:\n", header)
	fmt.Fprintf(stringStream, "\tSender ID: %d\n", sourceVC.InstanceID)
	fmt.Fprintf(stringStream, "\tReceiver ID: %d\n", s.VectorClock.InstanceID)
	fmt.Fprintf(stringStream, "\tPacket Content: %v\n", string(packet))
	fmt.Fprintf(stringStream, "\tPacket Clock:\n")
	fmt.Fprintf(stringStream, "\t\tt_m: %d\n", tm.Clock)
	fmt.Fprintf(stringStream, "\t\ttP_send: %d\n", sourceVC.GetLogicalClock(s.VectorClock.InstanceID))
	fmt.Fprintf(stringStream, "\tReceiver Logical Clock (tP_rcv):\n")
	fmt.Fprintf(stringStream, "\t\t%v\n", s.VectorClock.GetClock(s.VectorClock.InstanceID))
	fmt.Fprintf(stringStream, "\tStatus: %s\n", status)
	if printCompare {
		fmt.Fprintf(stringStream, "\tDelivery Condition: %d > %d\n", s.VectorClock.GetLogicalClock(s.VectorClock.InstanceID), tm.Clock)
	}
	return stringStream.String()
}

func (s *SES) writeSenderFile(data string, id int, numberProcess int) {
	path := fmt.Sprintf("./static/logs/%02d_process/%02d__sender.log", numberProcess, id)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if _, err := fmt.Fprintln(file, data+"\n-------------\n"); err != nil {
		panic(err)
	}
}
func (s *SES) writeReceiverFile(data string, numberProcess int, id int) {
	path := fmt.Sprintf("./static/logs/%02d_process/%02d__receiver.log", numberProcess, id)
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if _, err := fmt.Fprintln(file, data+"\n-------------\n"); err != nil {
		panic(err)
	}
}

func (s *SES) Deliver(packet []byte) {
	s.lock.Lock() // synchronize
	sourceVectorClock, packet := s.DeserializeSES(packet)
	timeProcess := s.VectorClock.GetClock(s.VectorClock.InstanceID) // timestamp send message.
	t_m := sourceVectorClock.GetClock(s.VectorClock.InstanceID)     // timestamp of Process i in the packet

	if timeProcess.canDeliver(t_m) { //??????
		// Deliver ???????????(t_m.Clock < timeProcess.Clock)
		s.writeReceiverFile(s.GetDeliverLog(t_m, sourceVectorClock, packet, "delivering", "BEFORE DELIVERED", true), s.VectorClock.NumberProcess, s.VectorClock.InstanceID)
		s.MergeSES(sourceVectorClock)
	} else {
		// Queue
		s.Queue = append(s.Queue, QueueItem{t_m, sourceVectorClock, packet})
		s.writeReceiverFile(s.GetDeliverLog(t_m, sourceVectorClock, packet, "buffered", "BEFORE DELIVERED", true), s.VectorClock.NumberProcess, s.VectorClock.InstanceID)
		breakFlag := false
		for !breakFlag {
			breakFlag = true
			for index, item := range s.Queue {
				// fmt.Println("hi: ", item.TimeMsg)
				if timeProcess.canDeliver(item.TimeMsg) { // ??
					s.writeReceiverFile(s.GetDeliverLog(item.TimeMsg, item.SourceVectorClock, item.Packet, "delivering from buffer", "BEFORE DELIVERED FROM BUFFERED", true),s.VectorClock.NumberProcess, s.VectorClock.InstanceID)
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
	//fmt.Println("\n\n num: ", s.VectorClock.NumberProcess)
	s.writeSenderFile(s.GetSenderLog(destinationID, packet), s.VectorClock.InstanceID,  s.VectorClock.NumberProcess)
	result := s.SerializeSES(packet)
	s.VectorClock.SelfMerge(s.VectorClock.InstanceID, destinationID)
	return result
}
