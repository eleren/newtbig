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
package tcp

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	pb "github.com/eleren/newtbig/msg/framepb"
)

func NewConnect(conn net.Conn, opts module.Options) (module.Connect, error) {
	c := new(Connect)
	err := c.OnInit(conn, opts)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type Connect struct {
	conn          net.Conn
	IP            net.IP
	Port          string
	header        []byte
	cache         []byte
	readDeadline  time.Duration
	writeDeadline time.Duration
	maxPacketSize int64
	byteOrder     binary.ByteOrder
	headSize      uint16
	isClose       bool
}

func (c *Connect) IsClosed() bool {
	return c.isClose
}

func (c *Connect) OnClose() {
	if !c.isClose {
		c.isClose = true
		if c.conn != nil {
			c.conn.Close()
		}
	}
}

func (c *Connect) OnInit(_conn net.Conn, opts module.Options) error {
	log.Logger.Debugf("Connect OnInit readDeadLine :%d   writeDeadLine :%d", opts.ReadDeadline, opts.WriteDeadline)
	c.readDeadline = time.Duration(opts.ReadDeadline) * time.Second
	c.writeDeadline = time.Duration(opts.WriteDeadline) * time.Second
	c.maxPacketSize = opts.MaxPacketSize
	c.cache = make([]byte, c.maxPacketSize)
	c.headSize = opts.NetHeadSize
	c.header = make([]byte, c.headSize)
	c.isClose = false
	if opts.ByteOrder {
		c.byteOrder = binary.BigEndian
	} else {
		c.byteOrder = binary.LittleEndian
	}
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
	sz := len(d)
	if c.headSize == 2 {
		c.byteOrder.PutUint16(c.cache, uint16(sz))
	} else {
		c.byteOrder.PutUint32(c.cache, uint32(sz))
	}
	copy(c.cache[c.headSize:], d)
	_, err := c.conn.Write(c.cache[:sz+int(c.headSize)])
	return err
}

func (c *Connect) OnReceive() ([]byte, error) {
	if c.isClose {
		return nil, fmt.Errorf("connect :%s is close", c.OnString())
	}
	c.conn.SetReadDeadline(time.Now().Add(c.readDeadline))
	_, err := io.ReadFull(c.conn, c.header)
	if err != nil {
		log.Logger.Debugf("Connect :%s OnReceive Read head err : %s", c.OnString(), err.Error())
		return nil, err
	}
	var size int
	if c.headSize == 2 {
		size = int(c.byteOrder.Uint16(c.header))
	} else {
		size = int(c.byteOrder.Uint32(c.header))
	}
	payload := make([]byte, size)
	_, err = io.ReadFull(c.conn, payload)
	if err != nil && err != io.EOF { //errors.Is(err, io.ErrUnexpectedEOF) {
		log.Logger.Debugf("Connect %s OnReceive Read data err :%s", c.OnString(), err.Error())
		return nil, err
	}
	return payload, nil
}

func (c *Connect) OnString() string {
	return fmt.Sprintf("%s_%s:%s", "tcp", c.IP, c.Port)
}

func (c *Connect) OnSendMsg(d *pb.Msg) error {
	return nil
}

func (c *Connect) OnReceiveMsg() (*pb.Msg, error) {
	return nil, nil
}
