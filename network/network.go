package network

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
	"toy/pkg"
	"toy/pkg/setting"

	"golang.org/x/sync/errgroup"

	"github.com/golang/protobuf/proto"
)

type Network struct {
	addr   string
	router *ProtoRouter

	sessionPool sync.Pool
	sessions    *pkg.SafeMap

	isServer bool

	group  *errgroup.Group
	cancel context.CancelFunc

	// decors
	defaultTimerInterval int
	defaultRecvTimeout   int
	defaultSendTimeout   int

	// callback
	connectFn func(*Session) error
	timeoutFn func(*Session) error
	closedFn  func(*Session) error
}

func NewNetwork() *Network {
	return &Network{
		router:   NewProtoRouter(),
		sessions: pkg.NewSafeMap(),
		sessionPool: sync.Pool{
			New: func() interface{} {
				return newSession()
			},
		},
	}
}

func (s *Network) Setup(addr string, defaultTimerInterval, defaultRecvTimeout, defaultSendTimeout int) {
	s.addr = addr
	s.defaultTimerInterval = defaultTimerInterval
	s.defaultRecvTimeout = defaultRecvTimeout
	s.defaultSendTimeout = defaultSendTimeout
}

func (n *Network) SetupGroup() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	group, newCtx := errgroup.WithContext(ctx)
	n.group = group
	n.cancel = cancel
	return newCtx
}

func (s *Network) GetAddr() string {
	return s.addr
}

func (s *Network) GetRouter() *ProtoRouter {
	return s.router
}

func (n *Network) GetSession() *Session {
	return n.sessionPool.Get().(*Session)
}

func (n *Network) PutSession(s *Session) {
	n.sessionPool.Put(s)
}

func (n *Network) IsServer() bool {
	return n.isServer
}

func (n *Network) GetSessionNum() uint32 {
	return uint32(n.sessions.GetNum())
}

func (n *Network) RegistOnConnect(fn func(*Session) error) {
	n.connectFn = fn
}

func (n *Network) RegistOnClosed(fn func(*Session) error) {
	n.closedFn = fn
}

func (n *Network) RegistOnTimeout(fn func(*Session) error) {
	n.timeoutFn = fn
}

func (n *Network) Routine(ctx context.Context, conn net.Conn) {
	n.group.Go(func() error {
		n.HandleConn(ctx, conn)
		return nil
	})
}

func (n *Network) ConnectWithCtx(ctx context.Context, timeout time.Duration) error {
	var err error
	var conn net.Conn
	if timeout > 0 {
		conn, err = net.DialTimeout("tcp", n.GetAddr(), timeout)
		if err != nil {
			log.Println("dial error:", err)
			return nil
		}
	} else {
		conn, err = net.Dial("tcp", n.GetAddr())
		if err != nil {
			log.Println("dial timeout:", err)
			return nil
		}
	}
	n.Routine(ctx, conn)
	return nil
}

func (n *Network) ServeWithCtx(ctx context.Context) {
	ln, err := net.Listen("tcp", n.GetAddr())
	if err != nil {
		fmt.Println(err)
		return
	}

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
				continue
			}

			n.Routine(ctx, conn)
		}
	}()

	n.Wait()
}

func (n *Network) HandleConn(topCtx context.Context, conn net.Conn) {
	s := n.GetSession()
	s.Init(conn,
		time.Duration(n.defaultTimerInterval)*time.Second,
		time.Duration(n.defaultRecvTimeout)*time.Second,
		time.Duration(n.defaultSendTimeout)*time.Second)

	defer func() {
		onClosed(n, s)
		n.PutSession(s)
	}()

	onConnect(n, s)

	ctx, cancel := context.WithCancel(context.Background())
	group, newCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		defer cancel()
		result := SendLoop(topCtx, newCtx, n, s)
		return result
	})
	group.Go(func() error {
		defer cancel()
		result := RecvLoop(topCtx, newCtx, n, s)
		return result
	})

	if err := group.Wait(); err != nil {
		fmt.Println(err)
	}
}

func SendLoop(topCtx context.Context, ctx context.Context, n *Network, s *Session) error {
	defer func() {
		fmt.Println("sendloop done")
		//close(s.sendCh)
		s.conn.Close()
	}()

	for {
		select {
		case <-topCtx.Done():
			return nil
		case <-ctx.Done():
			return nil
		case b, ok := <-s.sendCh:
			if !ok {
				return fmt.Errorf("sendch closed.")
			}
			if err := writeAll(s.conn, b, 0); err != nil {
				return err
			}
		case <-s.tmTicker.C:
			if err := onTimeout(n, s); err != nil {
				fmt.Println(err)
				continue
			}
			s.tmTicker.Reset(s.tmInterval)
		}
	}
	return nil
}

func RecvLoop(topCtx context.Context, ctx context.Context, n *Network, s *Session) error {
	defer func() {
		fmt.Println("recvloop done")
	}()

	sizebuf := make([]byte, 4)
	msgbuf := make([]byte, setting.NetworkSetting.DefaultBufLen)

	for {
		select {
		case <-topCtx.Done():
			return nil
		case <-ctx.Done():
			return nil
		default:
			if s.tmInterval > 0 {
				err := s.conn.SetReadDeadline(time.Now().Add(s.tmRecvTimeout))
				if err != nil {
					fmt.Printf("SetReadDeadline err:%s\n", err)
					return err
				}
			}

			// 2.读包头
			if _, err := io.ReadFull(s.conn, sizebuf); err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					continue
				}
				fmt.Printf("Read size err:%s\n", err)
				return err
			}
			// size
			size := binary.BigEndian.Uint32(sizebuf)
			if int(size) > setting.NetworkSetting.MaxBufLen {
				return fmt.Errorf("recv limit overflow(%d)", size)
			}

			if int(size) > len(msgbuf) {
				msgbuf = make([]byte, int(size))
			}

			// message
			msgbuf = msgbuf[:size]
			num, err := io.ReadFull(s.conn, msgbuf)
			if err != nil {
				if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
					fmt.Println("Read msg timeout:%s\n", nerr)
					return nerr
				}
				return err
			}

			if num < int(size) {
				return fmt.Errorf("Read less(%d) than (%d)", num, size)
			}

			// unpack
			msg := s.msgPool.Get().(*Message)
			msg.Unpack(msgbuf)

			// doing
			// TODO 判断message类型
			// switch msg.Ctrl {
			// case MESSAGE:
			// case CLOSE:
			//}

			err = onMessage(n, s, msg)
			if err != nil {
				fmt.Println("onMessage error:%s\n", err)
				// do without continue
			}
			s.msgPool.Put(msg)
		}
	}
}

func (s *Session) Send(cmd uint32, pbmsg proto.Message) error {
	/*
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("panic:%s\n", err)
			}
		}()
	*/
	data, err := proto.Marshal(pbmsg)
	if err != nil {
		fmt.Printf("marshaling error:%s\n", err)
		return err
	}

	b := MakePack(cmd, data)
	select {
	case s.sendCh <- b: // pointA
		{
			// do nothing
		}
	}
	return nil
}

func writeAll(conn net.Conn, b []byte, try int) error {
	if try >= 3 {
		return fmt.Errorf("writeAll try more than 3 times!!!")
	}

	n, err := conn.Write(b)
	if err != nil {
		fmt.Printf("writeAll err :%s\n", err)
		return err
	}

	if n != len(b) {
		size := len(b)
		copy(b, b[n:])
		b = b[:(size - n)]
		return writeAll(conn, b, try+1)
	}

	return nil
}

func (n *Network) WaitGroup() {
	if err := n.group.Wait(); err != nil {
		fmt.Println(err)
	}
}

func (n *Network) WaitSig() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	select {
	case <-quit:
		{
			fmt.Println("recv quit signal")
		}
	}
}

func (n *Network) Wait() {
	defer func() {
		if err := n.group.Wait(); err != nil {
			fmt.Println(err)
		}
		fmt.Println("All done.")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	select {
	case <-quit:
		{
			fmt.Println("recv quit signal, and starting group wait.")
			n.cancel()
		}
	}
}

func onConnect(n *Network, s *Session) error {
	if n.connectFn != nil {
		return n.connectFn(s)
	}
	return nil
}

func onClosed(n *Network, s *Session) error {
	if n.closedFn != nil {
		return n.closedFn(s)
	}
	return nil
}

func onTimeout(n *Network, s *Session) error {
	if n.timeoutFn != nil {
		return n.timeoutFn(s)
	}
	return nil
}

func onMessage(n *Network, s *Session, msg *Message) error {
	routeFn := n.router.GetRouter(msg.Cmd)
	constructFn := n.router.GetConstructor(msg.Cmd)
	if routeFn != nil && constructFn != nil {
		pbMsg := constructFn()
		err := proto.Unmarshal(msg.Body, pbMsg)
		if err != nil {
			fmt.Println("unmarshaling error: ", err)
		} else {
			routeFn(s, pbMsg)
		}
	}
	return nil
}
