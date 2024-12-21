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
	"fmt"
	"net/http"
	"strconv"

	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
	"google.golang.org/protobuf/proto"
)

type Handle struct {
	app module.App
}

func NewHandle(app module.App) *Handle {
	h := new(Handle)
	h.app = app
	return h
}

func (h *Handle) HandleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	uidStr, ok := r.URL.Query()["uid"]
	if !ok && len(uidStr) == 0 {
		log.Logger.Errorf("HandleLogin param code err ")
		return
	}
	uidS := uidStr[0]
	uid, err := strconv.ParseInt(uidS, 10, 64)
	if err != nil {
		log.Logger.Errorf("HandleLogin parseint uidS :%s err :%s", uidS, err.Error())
		fmt.Fprintf(w, "fail")
		return
	}
	data, err1 := h.app.GetDispatcher().GetGate(uint64(uid))
	if err1 != nil || data == nil {
		log.Logger.Errorf("HandleLogin dis getserver uid :%d err :%s", uid, err1.Error())
		fmt.Fprintf(w, "fail")
		return
	}
	status := &framepb.Status{}
	err = proto.Unmarshal(data, status)
	if err != nil {
		log.Logger.Error("HandleLogin Unmarshal err :", err.Error())
		return
	}

	log.Logger.Debugf("HandleLogin uid :%s  addr : %s  Count :%d", uidS, status.Addr, status.Count)
	fmt.Fprintf(w, status.Addr)
}

func (h *Handle) HandleClear(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Logger.Debugf("HandleClear ")
	err1 := h.app.GetDispatcher().Clear()
	if err1 != nil {
		log.Logger.Errorf("HandleClear dis getserver err :%s", err1.Error())
		fmt.Fprintf(w, "fail")
		return
	}

	fmt.Fprintf(w, "OK")
}

func (h *Handle) HandleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Logger.Debugf("HandleStatus ")
	data, err := h.app.GetDispatcher().GetStatus()
	if err != nil {
		log.Logger.Errorf("HandleStatus dis GetStatus err :%s", err.Error())
		fmt.Fprintf(w, "fail")
		return
	}

	statusInfo := &framepb.ServiceInfo{}
	err = proto.Unmarshal(data, statusInfo)
	if err != nil {
		log.Logger.Error("HandleStatus Unmarshal info err :", err.Error())
		return
	}
	for _, service := range statusInfo.Sers {
		log.Logger.Infof("HandleStatus service name :%s", service.Name)
		for _, sData := range service.GetStatus() {
			status := &framepb.Status{}
			err = proto.Unmarshal(sData, status)
			if err != nil {
				log.Logger.Error("HandleStatus Unmarshal status err :", err.Error())
				continue
			}
			log.Logger.Infof("HandleStatus service status : %v", status)
		}
	}

	fmt.Fprintf(w, "OK")
}
