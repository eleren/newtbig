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
package session

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/eleren/newtbig/common"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
)

var ErrClose = errors.New("session close")

func NewSession(app module.App) module.Session {
	session := new(Session)
	err := session.OnInit(app)
	if err != nil {
		log.Error("NewSession init err:", err.Error())
		return nil
	}
	return session
}

type Session struct {
	app               module.Gate
	opts              module.Options
	userId            uint64
	conn              module.Connect
	out               chan []byte
	call              chan *framepb.Msg
	in                chan *framepb.Msg
	Addr              string
	rpmLimit          uint32
	sendCtl           chan struct{}
	receiveCtl        chan struct{}
	LastHeartBeatTime int64
	PacketCount       uint32
	PacketCount1Min   int
	vrifyed           bool
	stopOnce          sync.Once
	broMsgs           map[uint32]bool
}

func (s *Session) OnInit(app module.App) error {
	s.app = app.(module.Gate)
	s.opts = app.Options()
	s.out = make(chan []byte, s.opts.MaxMsgChanLen)
	s.in = make(chan *framepb.Msg, s.opts.MaxMsgChanLen)
	s.call = make(chan *framepb.Msg, s.opts.MaxMsgChanLen)
	s.rpmLimit = s.opts.RpmLimit
	s.sendCtl = make(chan struct{})
	s.receiveCtl = make(chan struct{})
	s.SetHeartBeatTime(time.Now())
	s.broMsgs = make(map[uint32]bool)
	s.broMsgs[common.Msg_BroatCast_Rst] = true
	return nil
}

func (s *Session) OnStart(_conn module.Connect) error {
	s.conn = _conn
	s.Addr = s.conn.OnString()
	log.Logger.Infof("session  new connection :%s", s.Addr)
	utils.SafeGO(func() {
		err := s.run()
		if err != nil && err != ErrClose {
			log.Logger.Error(err.Error())
		}
	})
	utils.SafeGO(func() {
		err := s.receive()
		if err != nil && err != ErrClose {
			log.Logger.Error(err.Error())
		}
	})
	return nil
}

func (s *Session) OnStop() error {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("session :%d  onstop err stack :%v ", s.GetUID(), r)
		}
	}()
	s.stopOnce.Do(func() {
		log.Logger.Infof("session :%d addr :%s onstop once ", s.GetUID(), s.Addr)
		if s.vrifyed {
			GetSessionMgr().DelSession(s.GetUID())
			if err := s.CallKick(); err != nil {
				log.Logger.Warnf("session :%d  callkick err :%s ", s.GetUID(), err.Error())
			}
		}
		s.vrifyed = false
		close(s.sendCtl)
		close(s.receiveCtl)
		s.Save()
		if s.conn != nil && !s.conn.IsClosed() {
			s.conn.OnClose()
		}
	})
	return nil
}

func (s *Session) GetUID() uint64 {
	return s.userId
}

func (s *Session) SetUID(id uint64) {
	s.userId = id
}

func (s *Session) GetProcessor() module.Processor {
	return s.opts.Process
}

func (s *Session) SendMsg(msg *framepb.Msg) error {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("session :%d Send stop err stack : %v ", s.GetUID(), r)
		}
	}()
	if !s.vrifyed {
		log.Logger.Warnf("seesion :%d sendMsg :%d   has stoped yet", s.GetUID(), msg.ID)
		return nil
	}
	if s.conn == nil || s.conn.IsClosed() {
		log.Logger.Warnf("seesion:%d sendMsg :%d   has stoped yet", s.GetUID(), msg.ID)
		return nil
	}
	// log.Logger.Debugf("session :%d  addr :%s SendMsg msg id :%d", s.GetUID(), s.Addr, msg.ID)
	msg.UID = s.userId
	data, err := s.opts.Process.Marshal(msg)
	if err != nil {
		log.Logger.Warnf("session :%d marshal err :%s ", s.GetUID(), err.Error())
		return err
	}
	utils.Put(msg)
	return s.SendByte(data)
}

func (s *Session) SendByte(data []byte) error {
	s.out <- data
	return nil
}

func (s *Session) Call(msg *framepb.Msg) error {
	s.call <- msg
	return nil
}

func (s *Session) receive() error {
	for {
		select {
		case _, ok := <-s.receiveCtl:
			if !ok {
				log.Logger.Warnf("session :%d conn :%s receive close by stop", s.GetUID(), s.Addr)
				return nil
			}
		default:
			payload, err := s.conn.OnReceive()
			if err != nil {
				log.Logger.Warnf("session :%d receive err :%s", s.GetUID(), err.Error())
				s.OnStop()
				return ErrClose
			}
			if payload != nil {
				s.in <- payload
			}
		}
	}
}

func (s *Session) run() error {
	t := time.NewTicker(time.Second * 3)
	tm := time.NewTicker(time.Second * 5)
	defer func() {
		t.Stop()
		tm.Stop()
		close(s.call)
		close(s.out)
		close(s.in)
	}()
	if s.opts.HeartBeatInterval <= 0 {
		return fmt.Errorf("session %d HeartBeatInterval:%d is err ", s.GetUID(), s.opts.HeartBeatInterval)
	}
	for {
		select {
		case <-t.C:
			if time.Now().Unix()-s.LastHeartBeatTime > s.opts.HeartBeatInterval {
				log.Logger.Warnf("session :%d conn :%s heartbeat timeout ", s.GetUID(), s.Addr)
				s.OnStop()
				return ErrClose
			}
		case <-tm.C:
			if !s.timer_work() {
				log.Logger.Warnf("session :%d mini timer", s.GetUID())
				s.OnStop()
				return ErrClose
			}
		case msg := <-s.in:
			if msg == nil {
				log.Logger.Warnf("session :%d receive payload is nil ", s.GetUID())
				continue
			}
			s.PacketCount++
			s.PacketCount1Min++
			err := s.msgIn(msg)
			if err != nil {
				log.Logger.Warnf("session :%d msgIn err:%s", s.GetUID(), err.Error())
			}
		case message := <-s.out:
			if s.conn == nil || s.conn.IsClosed() {
				log.Logger.Warnf("session :%d send conn is closed ", s.GetUID())
				s.OnStop()
				return ErrClose
			}
			err := s.conn.OnSend(message)
			if err != nil {
				log.Logger.Warnf("session :%d send reply data, err:%s ", s.GetUID(), err.Error())
			}
		case msg := <-s.call:
			err := s.app.GetRPC().CallNR(msg)
			if err != nil {
				log.Logger.Warnf("seesion:%d call msg:%d  seq:%d,err:%s", s.GetUID(), msg.ID, msg.Seq, err.Error())
			}
		case _, ok := <-s.sendCtl:
			log.Logger.Debugf("session :%d  addr :%s send close by stop ", s.GetUID(), s.Addr)
			if !ok {
				return ErrClose
			}
		}
	}
}

func (s *Session) msgIn(msg *framepb.Msg) error {
	switch msg.ID {
	case common.Msg_Verify:
		if msg.UID == 0 {
			utils.Put(msg)
			return fmt.Errorf("session :%d verify msg uid is nil:%d", s.GetUID(), msg.Seq)
		}

		if s.app.GetVerify()(msg) {
			s.SetUID(msg.UID)
			s.vrifyed = true
			GetSessionMgr().AddSession(s)
			log.Logger.Debugf("session :%d addr :%s verrify suc ", s.GetUID(), s.Addr)
		} else {
			log.Logger.Debugf("session :%d addr :%s verrify fail ", msg.UID, s.Addr)
			s.vrifyed = false
			msg.ID = common.Msg_Verify_Rst_Fail
			s.SendMsg(msg)
		}
	case common.Msg_Heartbeat:
		if s.vrifyed {
			s.SetHeartBeatTime(time.Now())
			utils.Put(msg)
			if s.opts.HBBack {
				s.SendSign(common.Msg_Heartbeat_Rst)
			}
		}
	case common.Msg_RegBroatCast:
		if s.vrifyed {
			s.setBroCastStatus(msg)
			utils.Put(msg)
		}
	case common.Msg_DisregBroatCast:
		if s.vrifyed {
			s.setBroCastStatus(msg)
			utils.Put(msg)
		}
	default:
		if s.vrifyed {
			s.SetHeartBeatTime(time.Now())
			err := s.app.GetRPC().CallNR(msg)
			if err != nil {
				log.Logger.Errorf("seesion:%d call msg:%d  seq:%d, err:%s", s.GetUID(), msg.ID, msg.Seq, err.Error())
			}
		}
	}
	return nil
}

func (s *Session) timer_work() bool {
	defer func() {
		s.PacketCount1Min = 0
	}()

	if s.PacketCount1Min > int(s.rpmLimit) {
		log.Logger.Warnf("session :%d   count1m: %d, total:%d  RPM", s.GetUID(), s.PacketCount1Min, s.PacketCount)
		return false
	}
	return true
}

func (s *Session) Kick() error {
	if s.userId != 0 {
		msg := utils.Get()
		msg.Seq = 0
		msg.UID = uint64(s.userId)
		msg.Body = nil
		msg.ID = common.Msg_Kick_Rst
		msg.Key = uint64(s.userId)
		data, err := s.opts.Process.Marshal(msg)
		if err != nil {
			utils.Put(msg)
			return fmt.Errorf("session :%d marshal err :%s ", s.GetUID(), err.Error())
		}
		utils.Put(msg)
		if s.conn != nil && !s.conn.IsClosed() {
			return s.conn.OnSend(data)
		}
	}
	return nil
}

func (s *Session) CallKick() error {
	if s.userId != 0 {
		msg := utils.Get()
		msg.Seq = 0
		msg.UID = uint64(s.userId)
		msg.Body = nil
		msg.ID = common.Msg_Kick
		msg.Key = uint64(s.userId)
		return s.app.GetRPC().CallNR(msg)
	}
	return nil
}

func (s *Session) Save() error {
	return nil
}

func (s *Session) SetHeartBeatTime(time time.Time) {
	s.LastHeartBeatTime = time.Unix()
}

func (s *Session) SendSign(msgID uint32) error {
	msg := utils.Get()
	msg.Seq = 0
	msg.UID = uint64(s.userId)
	msg.Body = nil
	msg.ID = msgID
	msg.Key = uint64(s.userId)
	return s.SendMsg(msg)

}

func (s *Session) Broadcast(msg *framepb.Msg) error {
	if msg.ID == 0 {
		return fmt.Errorf("session :%d  broadcast id is nil err ", s.GetUID())
	}
	if msg.ID == common.Msg_ReRoute {
		rst := s.app.GetDispatcher().CheckRoute(s.GetUID())
		if !rst {
			err := s.SendSign(common.Msg_Kick_Rst)
			if err != nil {
				return err
			}
			err = s.OnStop()
			if err != nil {
				return err
			}
		}
	} else if s.broMsgs[msg.ID] {
		return s.SendMsg(msg)
	}
	return nil
}

func (s *Session) setBroCastStatus(msg *framepb.Msg) error {
	log.Logger.Debugf("session :%d  setBroCastStatus msg :%d  body :%s ", s.GetUID(), msg.ID, msg.Body)
	if msg.Body != nil {
		num, err := strconv.ParseUint(string(msg.Body), 10, 32)
		if err != nil {
			return fmt.Errorf("session :%d setBroCastStatus id:%d body :%s err :%s", s.GetUID(), msg.ID, msg.Body, err.Error())
		}
		if msg.ID == common.Msg_RegBroatCast {
			s.broMsgs[uint32(num)] = true
		} else {
			s.broMsgs[uint32(num)] = false
		}
		return nil
	} else {
		return fmt.Errorf("session :%d setBroCastStatus is nil err ", s.GetUID())
	}
}

func (s *Session) RpcMsg(msg *framepb.Msg) error {
	if msg.ID == common.Msg_RegBroatCast || msg.ID == common.Msg_DisregBroatCast {
		return s.setBroCastStatus(msg)
	} else {
		return s.SendMsg(msg)
	}
}
