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
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/eleren/newtbig/gate/connect/tcp"
	"github.com/eleren/newtbig/gate/session"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/utils"
)

func (s *Server) run_tcp() error {
	var lis net.Listener
	if s.opts.RunMode == "release" {
		confPath := fmt.Sprintf("%s/config/tls/", s.opts.WorkDir)
		crt, err := tls.LoadX509KeyPair(confPath+"server.crt", confPath+"server.key")
		if err != nil {
			log.Logger.Fatalf("tcp tls listen fail err : %s", err.Error())
			return err
		}
		tlsCfg := &tls.Config{Certificates: []tls.Certificate{crt}}
		lis, err = tls.Listen("tcp", ":"+s.Port, tlsCfg)
		if err != nil {
			log.Logger.Fatalf("tcp tls listen fail err : %s", err.Error())
			return err
		}
		log.Logger.Info("tcp tls listening on: ", lis.Addr())
	} else {
		var err error
		lis, err = net.Listen("tcp", ":"+s.Port)
		if err != nil {
			log.Logger.Fatalf("tcp listen fail err : %s", err.Error())
			return err
		}
		log.Logger.Info("tcp listening on: ", lis.Addr())
	}

	defer lis.Close()
	s.setListener(lis)
	var delay time.Duration
	for {
		conn, err := lis.Accept()
		if err != nil {
			if utils.IsTimeoutError(err) {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay *= 2
				}
				if max := time.Second; delay > max {
					delay = max
				}
				log.Logger.Warnf("tcp accept timeout retrying :%d  err :%s", delay, err.Error())
				time.Sleep(delay)
				continue
			}
			//Go 1.16+
			if errors.Is(err, net.ErrClosed) {
				log.Logger.Warnf("tcp accept closed err :", err.Error())
				return nil
			}

			return err
		}
		delay = 0
		if session.GetSessionMgr().Count() > s.opts.MaxConn {
			log.Logger.Warn("tcp connect too much err  ip:", conn.RemoteAddr())
			conn.Close()
			continue
		}

		connect, errC := tcp.NewConnect(conn, s.opts)
		if errC != nil || connect == nil {
			log.Logger.Error("tcp connect init err ! ", errC.Error())
			conn.Close()
			continue
		}

		ss := session.NewSession(s.getApp())
		if ss != nil {
			err := ss.OnStart(connect)
			if err != nil {
				ss.OnStop()
			}
		}
	}
}
