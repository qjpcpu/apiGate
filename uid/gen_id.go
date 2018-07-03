package uid

import (
	"crypto/md5"
	"errors"
	"fmt"
	"strconv"
	"time"
)

const Epoch int64 = 1288834974657

const ServerBits uint8 = 10
const SequenceBits uint8 = 12

const ServerShift uint8 = SequenceBits
const TimeShift uint8 = SequenceBits + ServerBits

const ServerMax int = -1 ^ (-1 << ServerBits)

const SequenceMask int32 = -1 ^ (-1 << SequenceBits)

var g_uid *GUID

func InitGenerator(serverId, workerNum int) {
	g_uid = NewGUID(serverId, workerNum)
}

func GenUniqueId() string {
	id, err := GenNumId()
	if err != nil {
		return ""
	}
	data := []byte(id)
	return fmt.Sprintf("%x", md5.Sum(data))
}

func GenNumId() (string, error) {
	if g_uid == nil {
		return "", errors.New("id generator not initilized")
	}
	if id, err := g_uid.Gen(); err != nil {
		return "", err
	} else {
		return strconv.FormatInt(id, 10), nil
	}
}

type Worker struct {
	serverId      int
	lastTimestamp int64
	sequence      int32
}

type GUID struct {
	workers chan *Worker
}

func NewWorker(serverId int) *Worker {
	if serverId < 0 || ServerMax < serverId {
		panic(fmt.Errorf("invalid server Id"))
	}
	return &Worker{
		serverId:      serverId,
		lastTimestamp: 0,
		sequence:      0,
	}
}

func NewGUID(serverId, serverNum int) *GUID {
	workers := make(chan *Worker, serverNum)
	for n := 0; n < serverNum; n++ {
		workers <- NewWorker(serverId + n)
	}
	return &GUID{
		workers: workers,
	}
}

func (s *GUID) Gen() (int64, error) {
	worker := <-s.workers
	id, err := worker.Next()
	s.workers <- worker
	return id, err
}

func (s *Worker) Next() (int64, error) {
	t := now()
	if t < s.lastTimestamp {
		return -1, fmt.Errorf("invalid system clock")
	}
	if t == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & SequenceMask
		if s.sequence == 0 {
			t = s.nextMillis()
		}
	} else {
		s.sequence = 0
	}
	s.lastTimestamp = t
	tp := (t - Epoch) << TimeShift
	sp := int64(s.serverId << ServerShift)
	n := tp | sp | int64(s.sequence)

	return n, nil
}

func (s *Worker) nextMillis() int64 {
	t := now()
	for t <= s.lastTimestamp {
		time.Sleep(100 * time.Microsecond)
		t = now()
	}
	return t
}

func now() int64 {
	return time.Now().UnixNano() / 1000000
}
