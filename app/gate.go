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
package app

import (
	"sync"

	"github.com/eleren/newtbig/dispatcher"
	"github.com/eleren/newtbig/gate/server"
	"github.com/eleren/newtbig/module"
)

var (
	gateOnce sync.Once
	gate     *DefaultGate
)

func NewGate(opts ...module.Option) module.Gate {
	gateOnce.Do(func() {
		options := newOptions(opts...)
		gate = new(DefaultGate)
		gate.isGate = true
		gate.opts = options
		gate.opts.Serv = server.NewServer(gate)
		gate.dspatcher = dispatcher.NewDispatcher(gate)
	})
	return gate
}

type DefaultGate struct {
	DefaultApp
	Verify module.FuncVerify
}

func (app *DefaultGate) SetVerify(_func module.FuncVerify) error {
	app.Verify = _func
	return nil
}

func (app *DefaultGate) GetVerify() module.FuncVerify {
	return app.Verify
}
