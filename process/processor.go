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
package process

import (
	"fmt"

	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
	"google.golang.org/protobuf/proto"

	"errors"
	"reflect"
	"runtime/debug"
	"sync"
	"time"
)

type msgInfo struct {
	msgType    reflect.Type
	msgHandler reflect.Value
	handler    interface{}
}

func (m *msgInfo) MsgType() reflect.Type {
	return m.msgType
}

func NewProcessor() module.Processor {
	p := new(Processor)
	p.MaxPacketSize = 4096
	return p
}

type Processor struct {
	app           module.App
	msgInfos      map[uint32]*msgInfo
	mutex         sync.Mutex
	perfRecorder  map[uint32]*perfRecorder
	EnablePerf    bool
	route         map[uint32]string
	MaxPacketSize uint32
}

func (p *Processor) Init(app module.App) {
	p.app = app
	p.msgInfos = make(map[uint32]*msgInfo)
	p.perfRecorder = make(map[uint32]*perfRecorder)
	p.route = make(map[uint32]string)
	p.EnablePerf = false
	p.MaxPacketSize = uint32(p.app.Options().MaxPacketSize)

}

func (c *Processor) Marshal(msg *framepb.Msg) ([]byte, error) {
	defer func() {
		utils.Put(msg)
	}()
	if msg == nil {
		return nil, errors.New("message is nil err !")
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Logger.Error("message Marshal err :", err.Error())
		return nil, err
	}
	if uint32(len(data)) > c.MaxPacketSize {
		log.Logger.Error("requestData is too long msgid=", msg.ID, "  len=", len(data))
	}
	return data, nil
}

func (c *Processor) Unmarshal(data []byte) (*framepb.Msg, error) {
	if len(data) < 2 {
		log.Logger.Error("data too show len = %v", len(data))
		return nil, fmt.Errorf("data too show %v", len(data))
	}

	if uint32(len(data)) > c.MaxPacketSize {
		log.Logger.Error(" proto data is too long ,len=", len(data))
	}

	msg := utils.Get()
	err := proto.Unmarshal(data, msg)
	if err != nil {
		log.Logger.Error("data Unmarshal err:", err.Error())
		return nil, err
	}
	return msg, nil
}

func (p *Processor) HandleMsg(ctl interface{}, msg *framepb.Msg) *framepb.Msg {
	start := time.Now()
	defer func() {
		if err := recover(); err != nil {
			log.Logger.Error("recover HandleMsg error: ", err, ", msgid=", msg.ID, ", stack: ", string(debug.Stack()))
			return
		}
		p.Record(start, msg.ID, "")
	}()

	if msgInfo, ok := p.msgInfos[msg.ID]; ok && msgInfo.msgHandler.Kind() == reflect.Func {
		rets := msgInfo.msgHandler.Call([]reflect.Value{reflect.ValueOf(ctl), reflect.ValueOf(msg)})
		if len(rets) == 1 {
			ret := rets[0].Interface()
			if ret == nil {
				log.Logger.Errorf("Handler call back ret type err")
				return nil
			}
			rst := ret.(*framepb.Msg)
			return rst
		} else {
			log.Logger.Errorf("Handler call back ret len err")
			return nil
		}
	}

	return nil
}

func (c *Processor) Register(msgID uint32, msgHandler module.DisposeFunc) {
	_, ok := c.msgInfos[msgID]
	if ok {
		log.Logger.Info("Msg have registered: ", msgID)
		return
	}

	item := new(msgInfo)
	if msgHandler != nil {
		item.msgType = reflect.TypeOf(msgHandler)
	}
	item.msgHandler = reflect.ValueOf(msgHandler)
	item.handler = msgHandler
	c.msgInfos[msgID] = item
}

func (c *Processor) Record(start time.Time, msgId uint32, msgName string) {
	if c.EnablePerf == false {
		return
	}
	dur := time.Now().Sub(start)
	c.mutex.Lock()
	record, ok := c.perfRecorder[msgId]
	if !ok {
		record = new(perfRecorder)
		record.name = msgName
		record.msgID = msgId
		c.perfRecorder[msgId] = record
	}
	if dur > time.Duration(record.maxTime) {
		record.maxTime = int64(dur)
	}
	record.count++
	record.totalTime += int64(dur)
	c.mutex.Unlock()
}

func (c *Processor) GetApp() module.App {
	return c.app
}

func (c *Processor) Dump() {
	if c.EnablePerf == false {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	var records []*perfRecorder = nil
	for _, v := range c.perfRecorder {
		records = append(records, v)
	}

	log.Logger.Warnf("[analysis],statistics,%v", len(c.perfRecorder))
	{
		averageTime := func(p1, p2 *perfRecorder) bool {
			return p1.totalTime/int64(p1.count) > p2.totalTime/int64(p2.count)
		}
		perfRecorderSortBy(averageTime).Sort(records)
		for i, v := range records {
			if v.count == 0 {
				continue
			}
			log.Logger.Warn("[analysis],avg time cost,%v,name=%v,id=%v,avarageTime=%vms", i+1, v.name, v.msgID, float32(v.totalTime)/float32(v.count)/1e6)
		}
	}
	{
		maxTime := func(p1, p2 *perfRecorder) bool {
			return p1.maxTime > p2.maxTime
		}
		perfRecorderSortBy(maxTime).Sort(records)
		for i, v := range records {
			log.Logger.Warn("[analysis],longest time cost,%v,name=%v,id=%v, longest=%vms", i+1, v.name, v.msgID, float32(v.maxTime)/1e6)
		}
	}
	{
		count := func(p1, p2 *perfRecorder) bool {
			return p1.count > p2.count
		}
		perfRecorderSortBy(count).Sort(records)
		for i, v := range records {
			log.Logger.Warn("[analysis],most frequently,%v,name=%v,id=%v,Call times=%v", i+1, v.name, v.msgID, v.count)
		}
	}
}

func (p *Processor) GetMsgInfo(id uint32) *msgInfo {
	msgInfo, ok := p.msgInfos[id]
	if !ok || msgInfo == nil {
		return nil
	}
	return msgInfo
}
