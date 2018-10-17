package login

import (
	"fmt"
	"toy/network"
	pb "toy/protocol"
)

func OnModLoginC2SLogin(s *network.Session, request interface{}) error {
	fmt.Println("recv login ", request.(*pb.LoginC2SLogin))

	//encode
	p := &pb.LoginS2CLogin{
		Code: 200,
		LoginInfo: &pb.PLoginInfo{
			Id:   1,
			Name: "SpiderMan.",
		},
	}
	cmd := pb.ModLoginS2CLogin
	body := p
	return s.Send(cmd, body)
}

func OnModLoginS2CLogin(s *network.Session, request interface{}) error {
	fmt.Println("recv login ", request)
	return nil
}

func Notify(s *network.Session, code uint32) error {
	p := &pb.LoginS2CNotify{
		Code: code,
	}
	cmd := pb.ModLoginS2CNotify
	body := p
	return s.Send(cmd, body)
}

func OnModLoginC2SHeartBeat(s *network.Session, request interface{}) error {
	p := &pb.LoginS2CHeartBeat{}
	cmd := pb.ModLoginS2CHeartBeat
	body := p
	return s.Send(cmd, body)
}

func OnModLoginS2CHeartBeat(s *network.Session, request interface{}) error {
	fmt.Println("recv hb ", request.(*pb.LoginS2CHeartBeat))
	return nil
}
