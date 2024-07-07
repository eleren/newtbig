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
package ws

import (
	"io"
	"time"

	"fmt"
	"net"

	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	pb "github.com/eleren/newtbig/msg/framepb"

	"github.com/gorilla/websocket"
)

func NewConnect(conn *websocket.Conn, opts module.Options) module.Connect {
	c := new(Connect)
	err := c.OnInit(conn, opts)
	if err != nil {
		log.Error("connect :%s init err:", c.OnString(), err)
		return nil
	}
	return c
}

type Connect struct {
	conn *websocket.Conn
	IP   net.IP
	Port string

	header        []byte
	readDeadline  time.Duration
	writeDeadline time.Duration
	maxPacketSize int64
	isClose       bool
}

func (c *Connect) OnClose() {
	if !c.isClose {
		c.isClose = true
		if c.conn != nil {
			c.conn.Close()
		}
	}
}

func (c *Connect) IsClosed() bool {
	return c.isClose
}

func (c *Connect) OnInit(_conn *websocket.Conn, opts module.Options) error {
	c.header = make([]byte, 2)
	c.readDeadline = time.Duration(opts.ReadDeadline) * time.Second
	c.writeDeadline = time.Duration(opts.WriteDeadline) * time.Second
	c.maxPacketSize = opts.MaxPacketSize
	c.isClose = false
	c.conn = _conn
	host, port, err := net.SplitHostPort(_conn.RemoteAddr().String())
	if err != nil {
		return err
	}
	c.IP = net.ParseIP(host)
	c.Port = port

	return nil
}

func (c *Connect) OnSend(d []byte) error {
	if c.isClose {
		return fmt.Errorf("connect :%s is close", c.OnString())
	}
	c.conn.SetWriteDeadline(time.Now().Add(c.writeDeadline))
	return c.conn.WriteMessage(websocket.TextMessage, d)
}

func (c *Connect) OnReceive() ([]byte, error) {
	if c.isClose {
		return nil, fmt.Errorf("connect :%s is close", c.OnString())
	}
	c.conn.SetWriteDeadline(time.Now().Add(c.readDeadline))
	c.conn.SetReadLimit(c.maxPacketSize)
	_, payload, err := c.conn.ReadMessage()
	if err != nil && err != io.EOF {
		return nil, err
	}
	return payload, nil
}

func (c *Connect) OnString() string {
	return fmt.Sprintf("%s_%s:%s", "ws", c.IP, c.Port)
}

func (c *Connect) OnSendMsg(d *pb.Msg) error {
	return nil
}

func (c *Connect) OnReceiveMsg() (*pb.Msg, error) {
	return nil, nil
}
