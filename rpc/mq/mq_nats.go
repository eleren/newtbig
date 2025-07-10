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
package mq

import (
	"fmt"

	"sync"
	"time"

	"github.com/eleren/newtbig/common"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

var nats_js nats.JetStreamContext
var natsJSOnce sync.Once
var nats_conn *nats.Conn
var natsOnce sync.Once

var NATS_REQUEST_TIMEOUT = 10

type StreamMsg struct {
	Subj string
	Data []byte
}

type NatsClient struct {
	nats_js          nats.JetStreamContext
	nats_conn        *nats.Conn
	msg_send_chan    chan *StreamMsg
	msg_receive_chan chan *StreamMsg
	root             string
	key              string
	pubKey           string
	reqKey           string
	bcKey            string
	sendKey          string
	streamSubs       []string
	subs             *nats.Subscription
	opts             module.Options
	done             chan struct{}
	callBack         chan *framepb.Msg
	app              module.App
}

func NewNatsClient(app module.App) (*NatsClient, error) {
	hosts := ""
	for i, host := range app.Options().NatsHosts {
		if i == 0 {
			hosts = fmt.Sprintf("nats://%s", host)
		} else {
			hosts = fmt.Sprintf("%s,nats://%s", hosts, host)
		}
	}

	if hosts == "" {
		return nil, fmt.Errorf("NewNatsClient hosts is nil ")
	}

	log.Logger.Info("rpc init hosts :", hosts)
	natsClient := new(NatsClient)
	natsClient.app = app
	natsClient.opts = app.Options()
	nc, err := nats.Connect(hosts)
	if err != nil {
		log.Logger.Errorf("NewNatsClient Connect err: %s ", err.Error())
		return nil, err
	}
	natsClient.nats_conn = nc
	nc.SetClosedHandler(natsClient.CloseHandler)

	nj, err1 := nc.JetStream()
	if err1 != nil {
		log.Logger.Errorf("NewNatsClient JetStream nil err:%s ", err1.Error())
		return nil, err1
	}
	natsClient.nats_js = nj

	natsClient.Init()

	return natsClient, nil
}

func (nc *NatsClient) Init() {
	nc.callBack = make(chan *framepb.Msg, nc.opts.QMsgLen)
	nc.msg_send_chan = make(chan *StreamMsg, nc.opts.QMsgLen)
	nc.msg_receive_chan = make(chan *StreamMsg, nc.opts.QMsgLen)
	nc.done = make(chan struct{})
	nc.root = fmt.Sprintf("%s.%s", nc.opts.AppName, nc.opts.ServerType)
	nc.key = fmt.Sprintf("%s.%s", nc.root, nc.opts.ServerName)
	nc.pubKey = fmt.Sprintf("%s.%s", nc.key, common.PUBLISH)
	nc.bcKey = fmt.Sprintf("%s.%s", nc.key, common.BROATCAST)
	nc.reqKey = fmt.Sprintf("%s_%s", nc.key, common.REQUEST)
	nc.sendKey = fmt.Sprintf("%s_%s", nc.key, common.SEND)
	nc.streamSubs = []string{fmt.Sprintf("%s.*", nc.root)}

}
func (c *NatsClient) CallBack() chan *framepb.Msg {
	return c.callBack
}

func (nc *NatsClient) CloseHandler(conn *nats.Conn) {
	log.Logger.Info("nc conn closed ")
}

func (nc *NatsClient) PublishMsg(subj string, msg *framepb.Msg) error {
	defer func() {
		utils.Put(msg)
	}()
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	err = nc.nats_conn.Publish(subj, data)
	if err != nil {
		return err
	}
	return nil
}

func (nc *NatsClient) Publish(subj string, data []byte) error {
	err := nc.nats_conn.Publish(subj, data)
	if err != nil {
		return err
	}
	return nil
}

func (nc *NatsClient) RequestMsg(subj string, msg *framepb.Msg) (*framepb.Msg, error) {
	defer func() {
		utils.Put(msg)
	}()
	data, err1 := proto.Marshal(msg)
	if err1 != nil {
		return nil, err1
	}
	msg1, err := nc.nats_conn.Request(subj, data, time.Second*time.Duration(NATS_REQUEST_TIMEOUT))
	if err != nil {
		return nil, fmt.Errorf("nc RequestMsg err :%s req :%d msgId :%d uid :%d  key:%d err:%s ", subj, msg.Seq, msg.ID, msg.UID, msg.Key, err.Error())
	}
	rst := utils.Get()
	err = proto.Unmarshal(msg1.Data, rst)
	if err != nil {
		return nil, err
	}
	return rst, nil
}

func (nc *NatsClient) GetNatsConn() *nats.Conn {
	return nats_conn
}

func (nc *NatsClient) GetNatsJs() nats.JetStreamContext {
	return nats_js
}

func (nc *NatsClient) Done() {
	if nc.done != nil {
		close(nc.done)
	}
	if nc.nats_conn != nil && !nc.nats_conn.IsClosed() {
		nc.nats_conn.Close()
	}
	close(nc.callBack)
	close(nc.msg_receive_chan)
	close(nc.msg_send_chan)

}

func (nc *NatsClient) StartPublishHandle() (err error) {
	key := fmt.Sprintf("%s.*", nc.key)
	log.Logger.Infof("nc start publish sub :%s", key)
	nc.subs, err = nc.nats_conn.SubscribeSync(key)
	if err != nil {
		return err
	}

	if nc.subs == nil {
		return fmt.Errorf("nc pub sub:%s is nil ", key)
	}

	for {
		select {
		case _, ok := <-nc.done:
			if !ok {
				log.Logger.Info("nc pub client  has done !")
				nc.subs.Unsubscribe()
				return
			}
		default:
			m, err := nc.subs.NextMsg(time.Minute)
			if err != nil && err == nats.ErrTimeout {
				if !nc.subs.IsValid() {
					nc.subs, err = nc.nats_conn.SubscribeSync(fmt.Sprintf("%s.*", nc.key))
					if err != nil {
						log.Logger.Warnf("nc pub error: %s'", err.Error())
						continue
					}
				}
				continue
			} else if err != nil {
				log.Logger.Warnf("nc pub  error1 :%s", err.Error())
				if !nc.subs.IsValid() {
					nc.subs, err = nc.nats_conn.SubscribeSync(fmt.Sprintf("%s.*", nc.key))
					if err != nil {
						log.Logger.Warnf("nc pub error2 :%s ", err.Error())
						continue
					}
				}
				continue
			}
			if m == nil {
				log.Logger.Warnf("nc pub sub:%s receive is nil ", nc.key)
				continue
			}
			err = nc.handler(m)
			if err != nil {
				log.Logger.Warnf("nc pub handler sub:%s err :%s ", nc.key, err.Error())
			}
		}
	}
}

func (nc *NatsClient) StartRequestHandle() error {
	log.Logger.Infof("nc start request sub :%s", nc.reqKey)
	subs, errS := nc.nats_conn.Subscribe(nc.reqKey, func(m *nats.Msg) {
		if m != nil {
			log.Logger.Debugf("request data:%s  sub:%s ", m.Data, m.Subject)
			msg := utils.Get()
			err := proto.Unmarshal(m.Data, msg)
			if err != nil {
				log.Logger.Errorf("request sub:%s ummashal err:%s", m.Subject, err.Error())
				err = m.Respond([]byte("msg ummarshal err"))
				if err != nil {
					log.Logger.Errorf("request sub:%s respond1 err:%s", m.Subject, err.Error())
				}
				return
			}
			rst := nc.app.Options().Process.HandleMsg(nc.app.GetRPC(), msg)
			if rst != nil {
				data, err := proto.Marshal(rst)
				if err != nil {
					err = m.Respond([]byte("msg marshal err"))
					if err != nil {
						log.Logger.Errorf("request sub:%s respond2 err:%s", m.Subject, err.Error())
					}
					return
				}
				err = m.Respond(data)
				if err != nil {
					log.Logger.Errorf("request sub:%s respond3 err:%s", m.Subject, err.Error())
				}
				return
			}
		} else {
			log.Logger.Warn("request sub  msg is nil ")
		}
	})
	if errS != nil {
		return fmt.Errorf("request sub :%s  err: %s", nc.reqKey, errS.Error())
	}

	select {
	case _, ok := <-nc.done:
		if !ok {
			log.Logger.Info("request  has done ")
		}
	}
	subs.Unsubscribe()
	return nil
}
