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
)

type Option func(*Options)

type Options struct {
	Context           context.Context
	HostName          string
	PID               string
	Metadata          map[string]string
	Version           string
	RunMode           string
	Debug             bool
	Parse             bool
	WorkDir           string
	AppName           string
	ConfName          string
	ETCDHosts         []string
	ETCDUser          string
	ETCDPassword      string
	ETCDCertFile      string
	ETCDKeyFile       string
	NetType           string
	ServerType        string
	ServerName        string
	ServerIP          string
	ServerIPOut       string
	ServerPort        string
	ServerConTypes    []string
	HeartBeatInterval int64
	ExitWaitTTL       time.Duration
	RegisterInterval  time.Duration
	RegisterTTL       int
	RPCExpired        time.Duration
	RPCMaxCoroutine   uint32
	MaxConn           int32
	MaxPacketSize     int64
	ReadDeadline      int64
	WriteDeadline     int64
	RpmLimit          uint32
	MaxMsgChanLen     uint32
	NetHeadSize       uint16
	ByteOrder         bool
	NatsHosts         []string
	Serv              Server
	Process           Processor
	Rpc               RPC
	DisLen            uint32
	DisRep            uint32
	HBBack            bool
	QMsgLen           uint32
}

func Version(v string) Option {
	return func(o *Options) {
		o.Version = v
	}
}

func Debug(t bool) Option {
	return func(o *Options) {
		o.Debug = t
	}
}

func WorkDir(v string) Option {
	return func(o *Options) {
		o.WorkDir = v
	}
}

func Conf(cName string) Option {
	return func(o *Options) {
		o.ConfName = cName
	}
}

func RegisterTTL(t int) Option {
	return func(o *Options) {
		o.RegisterTTL = t
	}
}

func RegisterInterval(t time.Duration) Option {
	return func(o *Options) {
		o.RegisterInterval = t
	}
}

func ExitWaitTTL(t time.Duration) Option {
	return func(o *Options) {
		o.ExitWaitTTL = t
	}
}

func Parse(t bool) Option {
	return func(o *Options) {
		o.Parse = t
	}
}

func RPCExpired(t time.Duration) Option {
	return func(o *Options) {
		o.RPCExpired = t
	}
}

func RPCMaxCoroutine(t uint32) Option {
	return func(o *Options) {
		o.RPCMaxCoroutine = t
	}
}

func Metadata(md map[string]string) Option {
	return func(o *Options) {
		o.Metadata = md
	}
}

func Wait(b bool) Option {
	return func(o *Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, "wait", b)
	}
}

func NetType(nt string) Option {
	return func(o *Options) {
		o.NetType = nt
	}
}

func Process(p Processor) Option {
	return func(o *Options) {
		o.Process = p
	}
}
func ReadDeadline(t int64) Option {
	return func(o *Options) {
		o.ReadDeadline = t
	}
}

func WriteDeadline(t int64) Option {
	return func(o *Options) {
		o.WriteDeadline = t
	}
}

func NetHeadSize(t uint16) Option {
	return func(o *Options) {
		o.NetHeadSize = t
	}
}

func ByteOrder(t bool) Option {
	return func(o *Options) {
		o.ByteOrder = t
	}
}

func Serv(s Server) Option {
	return func(o *Options) {
		o.Serv = s
	}
}

func DisLen(t uint32) Option {
	return func(o *Options) {
		o.DisLen = t
	}
}

func DisRep(t uint32) Option {
	return func(o *Options) {
		o.DisRep = t
	}
}

func HBBack(t bool) Option {
	return func(o *Options) {
		o.HBBack = t
	}
}

func MaxConn(t int32) Option {
	return func(o *Options) {
		o.MaxConn = t
	}
}

func MaxPacketSize(t int64) Option {
	return func(o *Options) {
		o.MaxPacketSize = t
	}
}

func HeartBeatInterval(t int64) Option {
	return func(o *Options) {
		o.HeartBeatInterval = t
	}
}

func RpmLimit(t uint32) Option {
	return func(o *Options) {
		o.RpmLimit = t
	}
}

func MaxMsgChanLen(t uint32) Option {
	return func(o *Options) {
		o.MaxMsgChanLen = t
	}
}

func QMsgLen(t uint32) Option {
	return func(o *Options) {
		o.QMsgLen = t
	}
}
