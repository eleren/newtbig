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
package main

import (
	"github.com/eleren/newtbig/app"
	"github.com/eleren/newtbig/common"
	"github.com/eleren/newtbig/demo/app/handle"
	"github.com/eleren/newtbig/demo/msg"
	"github.com/eleren/newtbig/demo/process"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
)

func main() {
	app := app.NewApp(
		module.Debug(true),
		module.Version("1.0.0"),
		module.Conf("app"),
		module.Process(process.NewJsonProcessor()),
	)
	app.OnRegiste(registe)
	app.Run()

}

func registe(app module.App) {
	log.Logger.Info("app registe")
	app.RegisteSer(msg.Msg_Login_rst, "gate")
	app.RegisteSer(msg.Msg_Start_rst, "gate")
	app.RegisteSer(common.Msg_Kick_Rst, "gate")

	app.Register(common.Msg_Kick, handle.HandleKickReq)
	app.Register(msg.Msg_Login, handle.HandleLoginReq)
	app.Register(msg.Msg_Start, handle.HandleStartReq)
}
