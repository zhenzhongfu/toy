package setting

import (
	"log"

	"github.com/go-ini/ini"
)

type Network struct {
	DefaultSendTimeout   int // 发送超时时间
	DefaultRecvTimeout   int // 接受超时时间
	DefaultTimerInterval int // 计时器时间间隔

	MaxConnection int // 最大允许连接数量

	DefaultBufLen int
	MaxBufLen     int

	SendQueueLen int
	MsgQueueLen  int
}

var NetworkSetting = &Network{}

func Setup(path string) {
	if path == "" {
		path = "conf/app.ini"
	}
	Cfg, err := ini.Load(path)
	if err != nil {
		log.Fatalf("fail to parse 'conf/app.ini': %v", err)
	}

	err = Cfg.Section("network").MapTo(NetworkSetting)
	if err != nil {
		log.Fatalf("Cfg.MapTo Network err: %v", err)
	}
}
