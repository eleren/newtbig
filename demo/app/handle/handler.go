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
package handle

import (
	"github.com/eleren/newtbig/common"
	"github.com/eleren/newtbig/demo/msg"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
)

func HandleKickReq(cb interface{}, req *framepb.Msg) (rst *framepb.Msg) {
	//kick user
	rst = new(framepb.Msg)
	rst.ID = common.Msg_Kick_Rst
	log.Logger.Info("HandleKickReq ........")

	return
}

func HandleLoginReq(cb interface{}, req *framepb.Msg) (rst *framepb.Msg) {
	//create user
	log.Logger.Info("HandleLoginReq ........")
	rst = utils.Get()
	rst.Seq = req.Seq
	rst.ID = msg.Msg_Login_rst
	rst.UID = req.UID
	rst.Key = req.UID
	rst.Body = nil
	cb.(module.RPC).CallNR(rst)

	return
}

func HandleStartReq(cb interface{}, req *framepb.Msg) (rst *framepb.Msg) {
	//start app
	log.Logger.Info("HandleLoginReq ........")
	rst = utils.Get()
	rst.Seq = req.Seq
	rst.ID = msg.Msg_Start_rst
	rst.UID = req.UID
	rst.Key = req.UID
	rst.Body = nil
	cb.(module.RPC).CallNR(rst)

	return

}
