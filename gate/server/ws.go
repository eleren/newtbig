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
	"fmt"
	"net/http"

	"github.com/eleren/newtbig/gate/connect/ws"
	"github.com/eleren/newtbig/gate/session"
	log "github.com/eleren/newtbig/logging"
	"github.com/gorilla/websocket"
)

func (s *Server) run_ws() error {
	http.HandleFunc("/newtbig", s.wsPage)

	if s.opts.RunMode == "release" {
		confPath := fmt.Sprintf("%s/config/tls/", s.opts.WorkDir)
		log.Logger.Info(" WebSocket wss  listen:", s.Port)
		err := http.ListenAndServeTLS(":"+s.Port, confPath+"server.crt", confPath+"server.key", nil)
		if err != nil {
			return err
		}
	} else {
		log.Logger.Info(" WebSocket ws  listen:", s.Port)
		err := http.ListenAndServe(":"+s.Port, nil)
		if err != nil {
			return err
		}
	}

	log.Logger.Info("run_ws end")
	return nil
}

func (s *Server) wsPage(resp http.ResponseWriter, req *http.Request) {
	conn, err := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
		log.Logger.Info("ws upgrade", "ua:", r.Header["User-Agent"], "referer:", r.Header["Referer"])
		return true
	}}).Upgrade(resp, req, nil)
	if err != nil {
		http.NotFound(resp, req)
		return
	}
	conn.SetReadLimit(int64(s.opts.MaxPacketSize))
	if session.GetSessionMgr().Count() > s.opts.MaxConn {
		log.Logger.Error("ws connect too much ip:", conn.RemoteAddr())
		conn.Close()
		return
	}
	connect := ws.NewConnect(conn, s.opts)
	if connect == nil {
		log.Logger.Error("ws init err")
		return
	}
	ss := session.NewSession(s.getApp())
	if ss != nil {
		err := ss.OnStart(connect)
		if err != nil {
			ss.OnStop()
		}
	}

}
