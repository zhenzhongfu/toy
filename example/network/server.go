package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"

	"toy/mod"
	"toy/network"
	"toy/pkg/setting"
)

func main() {
	setting.Setup("../../conf/app.ini")

	//pprof
	go func() {
		fmt.Println(http.ListenAndServe(":8887", nil))
	}()

	srv := network.NewNetwork()
	srv.Setup(":8888", 1, 30, 30)
	ctx := srv.SetupGroup()

	srv.RegistOnConnect(onConnect)
	srv.RegistOnClosed(onClosed)
	srv.RegistOnTimeout(onTimeout)
	mod.Setup(srv.GetRouter())

	// server
	go srv.ServeWithCtx(ctx)
	time.Sleep(time.Second)
	// client
	srv.ConnectWithCtx(ctx, time.Second)

	srv.WaitGroup()
}

func onConnect(s *network.Session) error {
	fmt.Println("on connect")
	return nil
}

func onClosed(s *network.Session) error {
	fmt.Println("on closed")
	return nil
}

func onTimeout(s *network.Session) error {
	fmt.Println("on timeout")
	return nil
}
