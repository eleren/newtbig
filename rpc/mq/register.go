// Copyright 2020 newtbig Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package mq

import (
	"fmt"

	"github.com/eleren/newtbig/gate/session"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/utils"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
)

func (nc *NatsClient) handler(msg *nats.Msg) error {
	// log.Logger.Debugf("handler msg sub:%s subject:%s , reply:%s , data:%s ", msg.Sub.Subject, msg.Subject, msg.Reply, msg.Data)
	if msg.Data == nil {
		return fmt.Errorf("broadCastHandler sub:%s data is nil", msg.Subject)
	}

	if msg.Subject == nc.pubKey {
		err := nc.commonHandler(msg.Data)
		if err != nil {
			return err
		}
	} else if msg.Subject == nc.bcKey {
		err := nc.broadCastHandler(msg.Data)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("handler err sub:%s", msg.Subject)
	}
	return nil
}

func (nc *NatsClient) commonHandler(data []byte) error {
	msg := utils.Get()
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return err
	}
	// log.Logger.Debugf("common msgdata seq:%d cmd:%d uid:%d  key:%d  data:%s", msg.Seq, msg.ID, msg.UID, msg.Key, msg.Body)
	if nc.app.IsGate() {
		return session.GetSessionMgr().Receive(msg)
	} else {
		// log.Logger.Debugf("common handler msg seq:%d cmd:%d uid:%d  key:%d  data:%s", msg.Seq, msg.ID, msg.UID, msg.Key, msg.Body)
		nc.callBack <- msg
	}

	return nil
}

func (nc *NatsClient) broadCastHandler(data []byte) error {

	msg := utils.Get()
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return err
	}
	// log.Logger.Debugf("broatcast msgdata seq:%d cmd:%d uid:%d  key:%d  data:%s", msg.Seq, msg.ID, msg.UID, msg.Key, msg.Body)
	if nc.app.IsGate() {
		return session.GetSessionMgr().BroatCast(msg)
	} else {
		nc.callBack <- msg
	}

	return nil
}

func (nc *NatsClient) streamHandler(data []byte) error {
	msg := utils.Get()
	err := proto.Unmarshal(data, msg)
	if err != nil {
		return err
	}
	log.Logger.Debugf("stream handler msg seq:%d cmd:%d uid:%d  key:%d  data:%s", msg.Seq, msg.ID, msg.UID, msg.Key, msg.Body)
	nc.callBack <- msg
	return nil
}
