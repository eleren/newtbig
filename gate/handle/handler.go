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
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
)

func HandleHeartBeatReq(session interface{}, req *framepb.Msg) (rst *framepb.Msg) {
	rst = new(framepb.Msg)
	rst.ID = common.Msg_Heartbeat
	session.(module.Session).SendMsg(rst)
	log.Logger.Debug("HeartBeatReq")
	return
}
