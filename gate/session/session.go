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
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/eleren/newtbig/common"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
)

var ClOSE_ERR = errors.New("seesion close")

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
	userId            uint64
	conn              module.Connect
	out               chan []byte
	in                chan []byte
	call              chan *framepb.Msg
	IP                net.IP
	Port              string
	rpmLimit          uint32
	receiveCtl        chan struct{}
	sendCtl           chan struct{}
	LastHeartBeatTime time.Time
	PacketCount       uint32
	PacketCount1Min   int
	Processor         module.Processor
	opts              module.Options
	started           bool
	vrifyed           bool
	stopOnce          sync.Once
	lastTime          int64
	broMsgs           map[uint32]bool
}

func (s *Session) OnClose() error {
	log.Logger.Infof("seesion :%d onclose ", s.GetUID())
	if s.conn != nil && !s.conn.IsClosed() {
		s.conn.OnClose()
	}

	GetSessionMgr().DelSession(s.GetUID())
	s.Save()
	err := s.CallKick()
	if err != nil {
		log.Logger.Warnf("session onstop callkick err:%s", err.Error())
	}

	close(s.receiveCtl)
	return nil
}

func (s *Session) OnInit(app module.App) error {
	s.app = app.(module.Gate)
	s.opts = app.Options()
	s.out = make(chan []byte, s.opts.MaxMsgChanLen)
	s.in = make(chan []byte, s.opts.MaxMsgChanLen)
	s.call = make(chan *framepb.Msg, s.opts.MaxMsgChanLen)
	s.rpmLimit = s.opts.RpmLimit
	s.receiveCtl = make(chan struct{})
	s.sendCtl = make(chan struct{})
	s.Processor = s.opts.Process
	s.SetHeartBeatTime(time.Now())
	s.broMsgs = make(map[uint32]bool)
	s.broMsgs[common.Msg_BroatCast_Rst] = true
	return nil
}

func (s *Session) OnStart(_conn module.Connect) error {
	if s.started {
		return fmt.Errorf("seesion has started yet ")
	}
	s.conn = _conn
	log.Logger.Infof("session  new connection :%s", s.conn.OnString())

	utils.SafeGO(func() {
		err := s.run()
		if err != nil && err != ClOSE_ERR {
			log.Logger.Error(err.Error())
		}
	})
	utils.SafeGO(func() {
		err := s.receive()
		if err != nil && err != ClOSE_ERR {
			log.Logger.Error(err.Error())
		}
	})
	s.started = true
	return nil
}

func (s *Session) OnStop() error {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("session :%d  onstop err stack :%v ", s.GetUID(), r)
		}
	}()
	s.stopOnce.Do(func() {
		log.Logger.Infof("session :%d  onstop:", s.GetUID())
		if !s.started {
			log.Logger.Warnf("seesion :%d onstop has stoped yet", s.GetUID())
			return
		}
		s.started = false
		close(s.sendCtl)
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
	return s.Processor
}

func (s *Session) SendMsg(msg *framepb.Msg) error {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("session :%d Send stop err stack : %v ", s.GetUID(), r)
		}
	}()
	if !s.started {
		log.Logger.Warnf("seesion :%d sendMsg :%d  touid :%d has stoped yet", s.GetUID(), msg.ID, msg.UID)
		return nil
	}
	log.Logger.Debugf("session :%d SendMsg msg id :%d", s.GetUID(), msg.ID)
	msg.UID = s.userId
	data, err := s.Processor.Marshal(msg)
	if err != nil {
		log.Logger.Errorf("session :%d marshal err :%s ", s.GetUID(), err.Error())
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
	if !s.started {
		return fmt.Errorf("seesion :%d Call has stoped yet", s.GetUID())
	}
	s.call <- msg
	return nil
}

func (s *Session) receive() error {
	log.Logger.Infof("session :%d  start receive ", s.GetUID())
	defer func() {
		close(s.in)
	}()
	for {
		select {
		case _, ok := <-s.receiveCtl:
			if !ok {
				log.Logger.Warnf("session :%d  receive close by stop", s.GetUID())
				return nil
			}
		default:
			if s.conn.IsClosed() {
				log.Logger.Warnf("session :%d  receive conn is closed", s.GetUID())
				return nil
			}
			payload, err := s.conn.OnReceive()
			if err != nil {
				log.Logger.Warnf("session :%d receive err :%s", s.GetUID(), err.Error())
				return nil
			}

			if payload != nil && s.started {
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
	}()
	if s.opts.HeartBeatInterval <= 0 {
		return fmt.Errorf("session %d HeartBeatInterval:%d is err ", s.GetUID(), s.opts.HeartBeatInterval)
	}
	for {
		select {
		case <-t.C:
			if time.Now().Sub(s.LastHeartBeatTime).Milliseconds() > time.Second.Milliseconds()*s.opts.HeartBeatInterval {
				log.Logger.Warnf("session :%d heartbeat timeout ", s.GetUID())
				s.OnClose()
				return ClOSE_ERR
			}
		case <-tm.C:
			if !s.timer_work() {
				log.Logger.Warnf("session :%d mini timer", s.GetUID())
				s.OnClose()
				return ClOSE_ERR
			}
		case payload, ok := <-s.in:
			if !ok {
				log.Logger.Warnf("session :%d receive in is  closed  ", s.GetUID())
				s.OnClose()
				return ClOSE_ERR
			}

			if payload == nil {
				log.Logger.Warnf("session :%d receive payload is nil ", s.GetUID())
				continue
			}
			s.PacketCount++
			s.PacketCount1Min++

			msg, err := s.Processor.Unmarshal(payload)
			if err != nil {
				log.Logger.Warnf("session :%d payload err :%s ", s.GetUID(), err.Error())
				utils.Put(msg)
				continue
			}
			switch msg.ID {
			case common.Msg_Verify:
				if msg.UID == 0 {
					log.Logger.Warnf("session :%d verify msg uid is nil:%d", s.GetUID(), msg.Seq)
					utils.Put(msg)
					break
				}

				if s.app.GetVerify()(msg) {
					s.SetUID(msg.UID)
					s.vrifyed = true
					GetSessionMgr().AddSession(s)
					log.Logger.Debugf("session :%d verrify suc ", s.GetUID())
				} else {
					log.Logger.Debugf("session :%d verrify fail ", msg.UID)
					s.vrifyed = false
					msg.ID = common.Msg_Verify_Rst_Fail
					s.SendMsg(msg)
				}

			case common.Msg_Heartbeat:
				if s.vrifyed && s.started {
					s.SetHeartBeatTime(time.Now())
					utils.Put(msg)
				}

			case common.Msg_RegBroatCast:
				if s.vrifyed && s.started {
					s.setBroCastStatus(msg)
					utils.Put(msg)
				}
			case common.Msg_DisregBroatCast:
				if s.vrifyed && s.started {
					s.setBroCastStatus(msg)
					utils.Put(msg)
				}
			default:
				if s.vrifyed && s.started {
					s.SetHeartBeatTime(time.Now())
					err := s.app.GetRPC().CallNR(msg)
					if err != nil {
						log.Logger.Errorf("seesion:%d call msg:%d  seq:%d, err:%s", s.GetUID(), msg.ID, msg.Seq, err.Error())

					}
				}
			}
		case message, ok := <-s.out:
			if !ok {
				log.Logger.Warnf("session :%d receive out is  closed  ", s.GetUID())
				return ClOSE_ERR
			}
			if s.conn == nil || s.conn.IsClosed() {
				log.Logger.Warnf("session :%d send conn is closed ", s.GetUID())
				return ClOSE_ERR
			}
			err := s.conn.OnSend(message)
			if err != nil {
				log.Logger.Warnf("session :%d send reply data, err:%s ", s.GetUID(), err.Error())
			}
		case msg, ok := <-s.call:
			if !ok {
				log.Logger.Warnf("session :%d receive call is  closed  ", s.GetUID())
				return ClOSE_ERR
			}
			err := s.app.GetRPC().CallNR(msg)
			if err != nil {
				log.Logger.Errorf("seesion:%d call msg:%d  seq:%d,err:%s", s.GetUID(), msg.ID, msg.Seq, err.Error())
			}

		case _, ok := <-s.sendCtl:
			if !ok {
				log.Logger.Warnf("session :%d  send close by stop ", s.GetUID())
				s.OnClose()
				return ClOSE_ERR
			}
		}
	}
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
		data, err := s.Processor.Marshal(msg)
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
	s.LastHeartBeatTime = time
	return
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
