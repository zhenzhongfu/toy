package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
	"toy/mod"
	"toy/network"
	"toy/pkg/setting"
	pb "toy/protocol"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe(":8885", nil))
	}()

	setting.Setup("../../conf/app.ini")

	n := network.NewNetwork()
	n.Setup(":8888", 1, 30, 30)
	ctx := n.SetupGroup()
	mod.Setup(n.GetRouter())
	n.RegistOnConnect(onConnect)
	n.RegistOnClosed(onClosed)
	n.RegistOnTimeout(onTimeout)

	for i := 0; i < 100; i++ {
		n.ConnectWithCtx(ctx, time.Second)
	}

	n.WaitGroup()
}

func onConnect(s *network.Session) error {
	fmt.Println("on connect")
	p := &pb.LoginC2SLogin{
		Accname: "dio",
	}
	cmd := uint32(pb.ModLoginC2SLogin)
	body := p
	s.Send(cmd, body)

	return nil
}

func onClosed(s *network.Session) error {
	fmt.Println("on closed")
	return nil
}

func onTimeout(s *network.Session) error {
	fmt.Println("send hb")
	p := &pb.LoginC2SHeartBeat{}
	cmd := uint32(pb.ModLoginC2SHeartBeat)
	body := p
	return s.Send(cmd, body)
}
