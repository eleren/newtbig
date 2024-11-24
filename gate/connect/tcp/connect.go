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
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"

	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	pb "github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
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
	opts          module.Options
	conn          net.Conn
	hCache        []byte
	dCache        []byte
	cache         []byte
	readDeadline  time.Duration
	writeDeadline time.Duration
	byteOrder     binary.ByteOrder
	headSize      uint16
	isClose       bool
	reader        *bufio.Reader
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
	c.opts = opts
	c.readDeadline = time.Duration(opts.ReadDeadline) * time.Second
	c.writeDeadline = time.Duration(opts.WriteDeadline) * time.Second
	c.cache = make([]byte, c.opts.MaxPacketSize)
	c.dCache = make([]byte, c.opts.MaxPacketSize)
	c.headSize = opts.NetHeadSize
	c.hCache = make([]byte, c.headSize)
	c.isClose = false
	if opts.ByteOrder {
		c.byteOrder = binary.BigEndian
	} else {
		c.byteOrder = binary.LittleEndian
	}
	c.conn = _conn
	c.reader = bufio.NewReaderSize(c.conn, int(c.opts.MaxPacketSize))
	return nil
}

func (c *Connect) OnSend(d []byte) error {
	if c.isClose {
		return fmt.Errorf("connect :%s is close", c.OnString())
	}
	c.conn.SetWriteDeadline(time.Now().Add(c.writeDeadline))
	sz := len(d)
	if sz+int(c.headSize) > int(c.opts.MaxPacketSize) || sz <= 0 {
		return fmt.Errorf("Connect %s send data size :%d err", c.OnString(), sz)
	}
	if c.headSize == 2 {
		c.byteOrder.PutUint16(c.cache, uint16(sz))
	} else {
		c.byteOrder.PutUint32(c.cache, uint32(sz))
	}
	copy(c.cache[c.headSize:], d)
	_, err := c.conn.Write(c.cache[:sz+int(c.headSize)])
	return err
}

func (c *Connect) OnReceive() (*pb.Msg, error) {
	if c.isClose {
		return nil, fmt.Errorf("connect :%s is closed", c.OnString())
	}
	c.conn.SetReadDeadline(time.Now().Add(c.readDeadline))
	rLen, err := io.ReadFull(c.reader, c.hCache)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("Connect :%s head closed s ", c.OnString())
		} else if err == io.ErrClosedPipe {
			return nil, fmt.Errorf("Connect :%s head closed c ", c.OnString())
		}
		return nil, fmt.Errorf("Connect :%s head err :%s", c.OnString(), err.Error())
	}
	if rLen < int(c.headSize) {
		log.Logger.Warnf("Connect :%s head rLen err", c.OnString())
		return nil, nil
	}

	var size int
	if c.headSize == 2 {
		size = int(c.byteOrder.Uint16(c.hCache))
	} else {
		size = int(c.byteOrder.Uint32(c.hCache))
	}

	if size+int(c.headSize) > int(c.opts.MaxPacketSize) || size <= 0 {
		return nil, fmt.Errorf("Connect %s receive data size :%d err", c.OnString(), size)
	}

	cache := c.dCache[:size]
	rLen, err = io.ReadFull(c.reader, cache)
	if err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("Connect :%s data closed s ", c.OnString())
		} else if err == io.ErrClosedPipe {
			return nil, fmt.Errorf("Connect :%s data closed c ", c.OnString())
		}
		return nil, fmt.Errorf("Connect :%s data err :%s", c.OnString(), err.Error())
	}
	if rLen < size {
		log.Logger.Warnf("Connect :%s data rLen :%d size :%d err", c.OnString(), rLen, size)
		return nil, nil
	}
	msg, err1 := c.opts.Process.Unmarshal(cache)
	if err1 != nil {
		utils.Put(msg)
		return nil, fmt.Errorf("Connect :%s OnReceive Unmarshal :%s ", c.OnString(), err1.Error())
	}

	return msg, nil
}

func (c *Connect) OnString() string {
	return fmt.Sprintf("%s_%s", "tcp", c.conn.RemoteAddr().String())
}

func (c *Connect) OnSendMsg(d *pb.Msg) error {
	return nil
}

func (c *Connect) OnReceiveMsg() (*pb.Msg, error) {
	return nil, nil
}
