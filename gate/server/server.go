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
package server

import (
	"net"
	"sync"

	"github.com/eleren/newtbig/common"
	"github.com/eleren/newtbig/gate/handle"
	"github.com/eleren/newtbig/gate/session"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/utils"
)

const NETYPE_TCP = "tcp"
const NETYPE_WS = "ws"
const NETYPE_UDP = "udp"
const NETYPE_GRPC = "grpc"

type Server struct {
	sync.RWMutex
	wg          sync.WaitGroup
	lis         net.Listener
	Port        string
	ClientIndex uint64
	netype      string // “tcp”  “udp” “ws”
	opts        module.Options
	app         module.Gate
	exitCB      chan struct{}
}

func NewServer(app module.Gate) *Server {
	s := new(Server)
	s.app = app
	s.opts = app.Options()
	s.Port = s.opts.ServerPort
	s.exitCB = make(chan struct{})
	return s
}

func (s *Server) BeforeStart() error {
	log.Logger.Info("gate Server BeforeStart ")
	return nil
}

func (s *Server) Start() error {
	log.Logger.Info("gate Server Start ")
	handle.Register(s.opts.Process)
	s.netype = s.opts.NetType

	if s.netype == NETYPE_TCP {
		utils.SafeGO(func() {
			err := s.run_tcp()
			if err != nil {
				log.Logger.Errorf("gate Server Start err :", err.Error())
				return
			}
		})
	} else if s.netype == NETYPE_WS {
		utils.SafeGO(func() {
			err := s.run_ws()
			if err != nil {
				log.Logger.Errorf("gate Server Start err :", err.Error())
				return
			}
		})
	}

	utils.SafeGO(func() {
		s.callBack()
	})

	return nil
}

func (s *Server) BeforeStop() error {
	log.Logger.Info("gate Server BeforeStop ")
	return nil
}
func (s *Server) AfterStart() error {
	log.Logger.Info("gate Server AfterStart ")
	return nil
}

func (s *Server) Stop() error {
	log.Logger.Info("gate Server Stop ")
	session.GetSessionMgr().StopAllSession()
	if s.lis != nil {
		log.Logger.Info("server listening off id:", s.string())
		s.lis.Close()
		s.lis = nil
	}
	close(s.exitCB)
	return nil
}

func (s *Server) AfterStop() error {
	log.Logger.Info("gate Server AfterStop ")
	return nil
}

func (s *Server) setListener(listener net.Listener) {
	s.lis = listener
}

func (s *Server) string() string {
	return "gate server"
}

func (s *Server) getApp() module.Gate {
	return s.app
}

func (s *Server) callBack() {
	log.Logger.Info("gate Server callBack running")
	for {
		select {
		case msg, ok := <-s.app.GetRPC().CallBack():
			if !ok {
				log.Logger.Warn("gate Server callBack err")
				return
			}
			if msg != nil {
				log.Logger.Debugf("gate Server callBack seq:%d cmd:%d uid:%d  key:%d  data:%s", msg.Seq, msg.ID, msg.UID, msg.Key, msg.Body)
				if msg.ID == common.Msg_BroatCast_Rst {
					session.GetSessionMgr().BroatCast(msg)
				} else {
					ss := session.GetSessionMgr().GetSession(msg.UID)
					if ss == nil {
						log.Logger.Errorf("Server callback sesseion not found uid:%d  id:%d", msg.UID, msg.Seq)
					} else {
						err := ss.SendMsg(msg)
						if err != nil {
							log.Logger.Errorf("Server callback sesseion sendMsg err uid:%d  id:%d", msg.UID, msg.Seq)
						}
					}
				}
			}
		case _, ok := <-s.exitCB:
			if !ok {
				log.Logger.Info("Server callBack stop")
				return
			}
		}
	}

}

func (s *Server) Count() uint32 {
	return uint32(session.GetSessionMgr().Count())
}
