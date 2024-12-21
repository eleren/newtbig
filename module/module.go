// Copyright 2020 newtbig Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package module

import (
	"context"
	"time"

	"github.com/eleren/newtbig/etcdclient"
	"github.com/eleren/newtbig/msg/framepb"
)

type FuncVerify func(interface{}) bool
type DisposeFunc func(ctl interface{}, data *framepb.Msg) (rst *framepb.Msg)

type Server interface {
	BeforeStart() error
	Start() error
	BeforeStop() error
	AfterStart() error
	Stop() error
	AfterStop() error
	Count() uint32
}

type Service interface {
	Server() Server
	Run() error
	String() string
}

type Gate interface {
	App
	SetVerify(_func FuncVerify) error
	GetVerify() FuncVerify
}

type App interface {
	UpdateOptions(opts ...Option) error
	Run(opts ...Option) error
	OnInit() error
	OnDestroy() error
	Options() Options
	OnRegiste(func(app App)) error
	GetProcessID() string
	WorkDir() string
	SetVerify(_func FuncVerify) error
	GetVerify() FuncVerify
	GetEClient() *etcdclient.ETCDClient
	GetRPC() RPC
	GetDispatcher() Dispatcher
	Register(msgID uint32, msgHandler DisposeFunc)
	RegisteSer(id uint32, key string) error
	IsGate() bool
}

type Processor interface {
	GetApp() App
	Init(app App)
	Marshal(msg *framepb.Msg) ([]byte, error)
	Unmarshal(data []byte) (*framepb.Msg, error)
	HandleMsg(ctl interface{}, msg *framepb.Msg) *framepb.Msg
	Register(msgID uint32, msgHandler DisposeFunc)
	Dump()
}

type Session interface {
	OnStart(Connect) error
	SendByte(data []byte) error
	SendMsg(msg *framepb.Msg) error
	OnStop() error
	GetUID() uint64
	Kick() error
	SetHeartBeatTime(time time.Time)
}

type Connect interface {
	IsClosed() bool
	OnClose()
	OnSend(d []byte) error
	OnReceive() (*framepb.Msg, error)
	OnString() string
	OnSendMsg(d *framepb.Msg) error
	OnReceiveMsg() (*framepb.Msg, error)
}

type Dispatcher interface {
	RegisteSer(id uint32, key string) error
	GetServer(sType string, id uint64) ([]byte, error)
	GetGate(id uint64) ([]byte, error)
	GetStatus() ([]byte, error)
	CheckRoute(key uint64) bool
	Start() error
	Stop() error
	Clear() error
	Dispatch(msg *framepb.Msg) error
	Request(msg *framepb.Msg) (*framepb.Msg, error)
	Broadcast(msg *framepb.Msg) error
	SendMsg(msg *framepb.Msg) error
}

type RPC interface {
	Call(ctx context.Context, msg *framepb.Msg) (*framepb.Msg, error)
	CallNR(msg *framepb.Msg) (err error)
	CallS(msg *framepb.Msg) (err error)
	CallB(msg *framepb.Msg) (err error)
	Start() (err error)
	Stop() (err error)
	CallBack() chan *framepb.Msg

	SendMsg(route string, msg *framepb.Msg) error
	PublishMsg(route string, msg *framepb.Msg) error
	RequestMsg(route string, msg *framepb.Msg) (*framepb.Msg, error)
}

type CallBack interface {
	Respond(data []byte) error
}
