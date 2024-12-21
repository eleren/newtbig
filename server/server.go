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
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/server/proxy"
	"github.com/eleren/newtbig/utils"
)

type Server struct {
	app    module.App
	exitCB chan struct{}
	proxy  *proxy.Proxy
	keyMap map[uint64]uint32
}

func NewServer(app module.App) *Server {
	s := new(Server)
	s.app = app
	s.exitCB = make(chan struct{})
	s.proxy = proxy.NewProxy(app)
	s.keyMap = make(map[uint64]uint32)
	return s
}

func (s *Server) BeforeStart() error {
	log.Logger.Info("Server BeforeStart ")
	return nil
}

func (s *Server) Start() error {
	log.Logger.Info("Server Start ")
	utils.SafeGO(func() {
		s.callBack()
	})
	return nil
}

func (s *Server) BeforeStop() error {
	log.Logger.Info("Server BeforeStop ")
	return nil
}
func (s *Server) AfterStart() error {
	log.Logger.Info("Server AfterStart ")
	return nil
}

func (s *Server) Stop() error {
	log.Logger.Info("Server Stop ")
	close(s.exitCB)

	s.proxy.Close()
	return nil
}

func (s *Server) AfterStop() error {
	log.Logger.Info("Server AfterStop ")
	return nil
}

func (s *Server) callBack() {
	log.Logger.Info("Server callBack running")
	for {
		select {
		case msg, ok := <-s.app.GetRPC().CallBack():
			if !ok {
				log.Logger.Warn("Server callBack err")
				return
			}

			if msg != nil {
				// log.Logger.Debugf("Server callBack msg seq:%d cmd:%d uid:%d  key:%d  data:%s", msg.Seq, msg.ID, msg.UID, msg.Key, msg.Body)
				s.keyMap[msg.Key]++
				rst := s.app.Options().Process.HandleMsg(s.app.GetRPC(), msg)
				if rst != nil {
					utils.Put(rst)
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
	return uint32(len(s.keyMap))
}
