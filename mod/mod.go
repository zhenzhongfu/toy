package mod

import (
	"toy/mod/login"
	"toy/network"
	pb "toy/protocol"

	"github.com/golang/protobuf/proto"
)

func Setup(router *network.ProtoRouter) {
	//login
	router.Insert(pb.ModLoginC2SLogin, login.OnModLoginC2SLogin, func() proto.Message { return &pb.LoginC2SLogin{} })
	router.Insert(pb.ModLoginS2CLogin, login.OnModLoginS2CLogin, func() proto.Message { return &pb.LoginS2CLogin{} })

	//hb
	router.Insert(pb.ModLoginC2SHeartBeat, login.OnModLoginC2SHeartBeat, func() proto.Message { return &pb.LoginC2SHeartBeat{} })
	router.Insert(pb.ModLoginS2CHeartBeat, login.OnModLoginS2CHeartBeat, func() proto.Message { return &pb.LoginS2CHeartBeat{} })
}
