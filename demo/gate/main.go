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
package main

import (
	"github.com/eleren/newtbig/app"
	"github.com/eleren/newtbig/common"
	"github.com/eleren/newtbig/demo/msg"
	"github.com/eleren/newtbig/demo/process"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
)

func main() {
	gate := app.NewGate(
		module.Debug(true),
		module.Version("1.0.0"),
		module.Conf("gate"),
		module.NetType("ws"), //“tcp”  “ws”
		module.Process(process.NewJsonProcessor()),
	)
	gate.SetVerify(Verify)
	gate.OnRegiste(Registe)
	gate.Run()

}

func Registe(app module.App) {
	log.Logger.Info("gate registe")
	app.RegisteSer(msg.Msg_Login, "app")
	app.RegisteSer(msg.Msg_Start, "app")
	app.RegisteSer(common.Msg_Kick, "app")
}

func Verify(msg interface{}) bool {

	return true
}
