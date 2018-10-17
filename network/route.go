package network

import "github.com/golang/protobuf/proto"

type routeFn func(*Session, interface{}) error
type constructFn func() proto.Message

type ProtoRouter struct {
	router      map[uint32]routeFn
	constructor map[uint32]constructFn
}

func NewProtoRouter() *ProtoRouter {
	return &ProtoRouter{
		router:      make(map[uint32]routeFn),
		constructor: make(map[uint32]constructFn),
	}
}

func (r *ProtoRouter) GetRouter(id uint32) routeFn {
	return r.router[id]
}

func (r *ProtoRouter) GetConstructor(id uint32) constructFn {
	return r.constructor[id]
}

func (r *ProtoRouter) Insert(id uint32, fn routeFn, fn2 constructFn) {
	if fn != nil {
		r.router[id] = fn
	}
	if fn2 != nil {
		r.constructor[id] = fn2
	}
}
