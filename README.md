# Dio's Toy
## 环境
- 下载安装protoc
wget https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip
unzip -d /usr/local protoc-3.6.1-linux-x86_64.zip
- 安装proto编译器插件
go get github.com/golang/protobuf/protoc-gen-go
- 安装Go protobuf包
go get github.com/golang/protobuf/proto

## 生成pb
make proto

## 测试
cd example/network
go run server.go
