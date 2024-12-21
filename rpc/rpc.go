// Copyright 2014 mqant Author. All Rights Reserved.
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
package rpc

import (
	"context"

	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
	mq "github.com/eleren/newtbig/rpc/mq"
	"github.com/eleren/newtbig/utils"
)

type RPCClient struct {
	app         module.App
	nats_client *mq.NatsClient
	opts        module.Options
}

func NewRPCClient(app module.App) (module.RPC, error) {
	rpc_client := new(RPCClient)
	rpc_client.app = app
	rpc_client.opts = app.Options()
	nats_client, err := mq.NewNatsClient(app)
	if err != nil {
		log.Logger.Errorf("NewRPCClient err: %s", err.Error())
		return nil, err
	}
	rpc_client.nats_client = nats_client
	return rpc_client, nil
}

func (c *RPCClient) CallBack() chan *framepb.Msg {
	return c.nats_client.CallBack()
}

func (c *RPCClient) Start() (err error) {
	utils.SafeGO(func() {
		err := c.nats_client.StartPublishHandle()
		if err != nil {
			log.Logger.Warnf("nc start StartPublishHandle err: %s", err.Error())
		}
	})
	utils.SafeGO(func() {
		err := c.nats_client.StartRequestHandle()
		if err != nil {
			log.Logger.Warnf("nc start StartRequestHandle err: %s", err.Error())
		}
	})
	return nil
}

func (c *RPCClient) Stop() (err error) {
	if c.nats_client != nil {
		c.nats_client.Done()
	}
	return nil
}

func (c *RPCClient) Call(ctx context.Context, msg *framepb.Msg) (*framepb.Msg, error) {
	log.Logger.Debugf("rpc call seq:%d cmd:%d uid:%d  key:%d  data:%s", msg.Seq, msg.ID, msg.UID, msg.Key, msg.Body)
	return c.app.GetDispatcher().Request(msg)
}

func (c *RPCClient) CallNR(msg *framepb.Msg) (err error) {
	log.Logger.Debugf("rpc cn seq:%d cmd:%d uid:%d  key:%d  data:%s", msg.Seq, msg.ID, msg.UID, msg.Key, msg.Body)
	return c.app.GetDispatcher().Dispatch(msg)
}

func (c *RPCClient) CallB(msg *framepb.Msg) (err error) {
	log.Logger.Debugf("rpc cb seq:%d cmd:%d uid:%d  key:%d  data:%s", msg.Seq, msg.ID, msg.UID, msg.Key, msg.Body)
	return c.app.GetDispatcher().Broadcast(msg)
}

func (c *RPCClient) CallS(msg *framepb.Msg) (err error) {
	log.Logger.Debugf("rpc s seq:%d cmd:%d uid:%d  key:%d  data:%s", msg.Seq, msg.ID, msg.UID, msg.Key, msg.Body)
	return c.app.GetDispatcher().SendMsg(msg)
}

func (c *RPCClient) SendMsg(rout string, msg *framepb.Msg) error {
	return c.nats_client.SendMsg(rout, msg)
}

func (c *RPCClient) PublishMsg(rout string, msg *framepb.Msg) error {
	return c.nats_client.PublishMsg(rout, msg)
}

func (c *RPCClient) RequestMsg(rout string, msg *framepb.Msg) (*framepb.Msg, error) {
	return c.nats_client.RequestMsg(rout, msg)
}
