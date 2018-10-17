package network

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
	"toy/pkg/setting"

	"github.com/gofrs/uuid"
)

const (
	statusClosed int32 = iota
	statusWaiting
	statusLogon
)

type Session struct {
	conn       net.Conn
	status     int32
	uuid       string
	id         int
	sendCh     chan []byte
	remoteAddr string

	msgPool sync.Pool

	tmTicker      *time.Timer
	tmInterval    time.Duration
	tmRecvTimeout time.Duration
	tmSendTimeout time.Duration

	// 统计
	inPacks  uint64
	inBytes  uint64
	outPacks uint64
	outBytes uint64
}

func newSession() *Session {
	s := &Session{}
	return s
}

func (s *Session) Init(conn net.Conn, tmInterval, tmRecvTimeout, tmSendTimeout time.Duration) {
	uuid, _ := uuid.NewV4()
	s.uuid = uuid.String()
	s.conn = conn
	s.sendCh = make(chan []byte, setting.NetworkSetting.SendQueueLen)
	s.remoteAddr = conn.RemoteAddr().String()
	s.tmInterval = tmInterval
	s.tmRecvTimeout = tmRecvTimeout
	s.tmSendTimeout = tmSendTimeout
	atomic.StoreInt32(&s.status, statusWaiting)
	s.tmTicker = time.NewTimer(tmInterval)
	s.msgPool = sync.Pool{
		New: func() interface{} {
			return &Message{}
		},
	}
}

func (s *Session) IsValid() bool {
	if atomic.LoadInt32(&s.status) == statusClosed {
		return false
	}
	return true
}

func (s *Session) SetClosed() {
	atomic.StoreInt32(&s.status, statusClosed)
}
