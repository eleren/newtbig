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
	"fmt"
	"net/http"

	"github.com/eleren/newtbig/app"
	"github.com/eleren/newtbig/demo/login/handle"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/utils"
)

func main() {
	app := app.NewApp(
		module.Debug(true),
		module.Version("1.0.0"),
		module.Conf("login"),
		module.NetType("http"),
	)
	initLogin(app)
	app.Run()

}

func initLogin(app module.App) {
	log.Logger.Info("login init")
	utils.SafeGO(func() {
		handle := handle.NewHandle(app)
		addr := fmt.Sprintf("%s:%s", app.Options().ServerIP, app.Options().ServerPort)
		log.Logger.Infof("Server Start addr :%s", addr)
		mux := http.NewServeMux()
		mux.HandleFunc("/newtbig/login", handle.HandleLogin)
		mux.HandleFunc("/newtbig/clear", handle.HandleClear)
		mux.HandleFunc("/newtbig/status", handle.HandleStatus)
		err := http.ListenAndServe(addr, mux)
		if err != nil {
			log.Logger.Fatalf("Server Start httpServer addr: %s err :%s ", addr, err.Error())
			panic(err)
		}
	})
}
