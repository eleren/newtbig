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
package dispatcher

import (
	"fmt"
	"path/filepath"

	"github.com/eleren/newtbig/common"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
)

type proxy struct {
	msgKey    string
	bcKey     string
	reqKey    string
	pubKey    string
	sendKey   string
	key       string
	serType   string
	serName   string
	appName   string
	ver       string
	app       module.App
	serStatus []byte
}

func (p *proxy) init(serType string, key string, app module.App, status []byte) {
	p.app = app
	p.key = key
	p.appName = app.Options().AppName
	p.ver = app.Options().Version
	p.serType = filepath.Base(serType)
	p.serName = filepath.Base(key)
	p.msgKey = fmt.Sprintf("%s.%s.%s", p.appName, p.serType, p.serName)
	p.bcKey = fmt.Sprintf("%s.%s", p.msgKey, common.BROATCAST)
	p.reqKey = fmt.Sprintf("%s_%s", p.msgKey, common.REQUEST)
	p.pubKey = fmt.Sprintf("%s.%s", p.msgKey, common.PUBLISH)
	p.sendKey = fmt.Sprintf("%s.%s", p.msgKey, common.SEND)
	p.serStatus = status
}

func (p *proxy) OnClose() error {
	return nil
}

func (p *proxy) GetKey() string {
	return p.key
}

func (p *proxy) GetMsgKey() string {
	return p.msgKey
}

func (p *proxy) GetBCKey() string {
	return p.bcKey
}

func (p *proxy) sendMsg(msg *framepb.Msg) error {
	err := p.app.GetRPC().SendMsg(p.sendKey, msg)
	return err
}

func (p *proxy) publish(msg *framepb.Msg) error {
	err := p.app.GetRPC().PublishMsg(p.pubKey, msg)
	return err
}

func (p *proxy) request(msg *framepb.Msg) (*framepb.Msg, error) {
	return p.app.GetRPC().RequestMsg(p.reqKey, msg)
}

func (p *proxy) broadcast(msg *framepb.Msg) error {
	err := p.app.GetRPC().PublishMsg(p.bcKey, msg)
	return err
}

func (p *proxy) update(status []byte) error {
	p.serStatus = status
	return nil
}

func (p *proxy) GetStatus() []byte {
	return p.serStatus
}

func (p *proxy) GetSerName() string {
	return p.serName
}
