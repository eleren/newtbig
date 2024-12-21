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
package proxy

import (
	"fmt"

	"github.com/eleren/newtbig/common"
	"github.com/eleren/newtbig/gate/session"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	pb "github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
)

type Proxy struct {
	proxyCtl chan struct{}
	app      module.App
	out      chan *pb.Msg
}

func NewProxy(app module.App) *Proxy {
	clientProxy := &Proxy{
		app:      app,
		out:      make(chan *pb.Msg, 8192),
		proxyCtl: make(chan struct{}),
	}
	utils.SafeGO(func() {
		clientProxy.run()
	})
	return clientProxy
}

func (p *Proxy) Close() {
	close(p.proxyCtl)

}

func (p *Proxy) GetApp() module.App {
	return p.app

}
func (p *Proxy) SendByte(data []byte) error {
	return nil
}

func (p *Proxy) SendMsg(seq uint32, msgID uint32, uid uint64, key uint64, data []byte) error {
	rst := utils.Get()
	rst.Seq = seq
	rst.ID = msgID
	rst.UID = uid
	rst.Key = key
	rst.Body = data
	p.Send(rst)
	return nil
}

func (p *Proxy) SendSyc1(msg *pb.Msg) error {
	if p.out == nil {
		return fmt.Errorf("Proxy out is nil")
	}
	p.out <- msg
	return nil
}

func (p *Proxy) Send(msg *pb.Msg) error {
	if p.out == nil {
		return fmt.Errorf("Proxy Send out is nil")
	}
	err := p.app.GetRPC().CallNR(msg)
	if err != nil {
		return fmt.Errorf("proxy Send uid:%d call msg:%d  seq:%d,err:%s", msg.UID, msg.ID, msg.Seq, err.Error())
	}
	return nil
}

func (p *Proxy) run() {
	for {
		select {
		case msg := <-p.out:
			switch msg.ID {
			case common.Msg_Kick:
				log.Logger.Debug("proxy msg from server kick id:", msg.UID)
				session.GetSessionMgr().DelSession(msg.UID)
				utils.Put(msg)
				break
			default:
				err := p.app.GetRPC().CallNR(msg)
				if err != nil {
					log.Logger.Errorf("proxy call uid:%d call msg:%d  seq:%d,err:%s", msg.UID, msg.ID, msg.Seq, err.Error())
					continue
				}
			}
		case _, ok := <-p.proxyCtl:
			if !ok {
				log.Logger.Info("server proxy call stop ")
				return
			}

		}
	}
}
