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
package process

import (
	"encoding/json"
	"errors"

	msgm "github.com/eleren/newtbig/demo/msg"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/process"
	"github.com/eleren/newtbig/utils"
)

type JsonProcessor struct {
	process.Processor
}

func NewJsonProcessor() module.Processor {
	p := new(JsonProcessor)
	p.EnablePerf = true
	return p
}

func (c *JsonProcessor) Marshal(msg *framepb.Msg) ([]byte, error) {
	defer func() {
		utils.Put(msg)
	}()
	if msg == nil {
		return nil, errors.New("message is nil err !")
	}
	var msgJson msgm.Head
	msgJson.Seq = msg.Seq
	msgJson.Cmd = msg.ID
	msgJson.Data = msg.Body
	msgJson.Key = msg.Key
	requestData, err := json.Marshal(msgJson)
	if err != nil || requestData == nil {
		log.Logger.Error("jsonprocessor GetGameMsgData json Marshal2:", err.Error())
		return nil, err
	}

	if uint32(len(requestData)) >= c.MaxPacketSize {
		log.Logger.Error("msgdata is too long,msgid=", msgJson.Cmd, ",len=", len(requestData))
	}
	log.Logger.Info("jsonprocessor Marshal seq:", msgJson.Seq, "  cmd:", msgJson.Cmd)

	return requestData, nil
}

func (c *JsonProcessor) Unmarshal(data []byte) (*framepb.Msg, error) {
	if uint32(len(data)) > c.MaxPacketSize {
		log.Logger.Error("msgdata is too long,len=", len(data))
	}

	request := &msgm.Head{}
	err := json.Unmarshal(data, request)
	if err != nil {
		log.Logger.Error("json Unmarshal err:", err)
		return nil, err
	}

	msg := utils.Get()
	msg.Seq = request.Seq
	msg.ID = request.Cmd
	msg.Body = request.Data
	msg.Key = request.Key
	msg.UID = uint64(request.UID)
	log.Logger.Info("jsonprocessor ummarshal seq:", msg.Seq, "  cmd:", msg.ID)
	return msg, nil

}
