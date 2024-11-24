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
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/eleren/newtbig/common"
	"github.com/eleren/newtbig/dispatcher/hash"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
)

type service struct {
	proxyMap  sync.Map
	hashring  *hash.HashRing
	Count     int32
	callbacks []chan string
	serType   string
	app       module.App
	routeType common.RouteType
}

func (s *service) init(serType string, app module.App, routeType common.RouteType) {
	s.app = app
	s.routeType = routeType
	if s.routeType == common.Hash {
		s.hashring = hash.NewHashRing(app.Options().DisRep, app.Options().DisLen)
	}
	s.callbacks = make([]chan string, 256)
	s.serType = serType
}

func (s *service) GetProxy(key string) *proxy {
	cl, err := s.proxyMap.Load(key)
	if err {
		ret, _ := cl.(*proxy)
		return ret
	}
	return nil
}

func (s *service) AddProxy(cl *proxy) bool {
	s.Count = int32(atomic.AddInt32(&s.Count, 1))
	_, loaded := s.proxyMap.LoadOrStore(cl.key, cl)
	return !loaded
}

func (s *service) DelProxy(key string) error {
	if s.routeType == common.Hash {
		s.hashring.Remove(key)
	}
	proxy := s.GetProxy(key)
	if proxy != nil {
		err := proxy.OnClose()
		if err != nil {
			return err
		}
	}
	s.proxyMap.Delete(key)
	s.Count = int32(atomic.AddInt32(&s.Count, -1))
	return nil
}

func (s *service) CreateProxy(key string, value []byte) ([]uint32, error) {
	proxy := new(proxy)
	proxy.init(s.serType, key, s.app, value)
	s.AddProxy(proxy)
	for k := range s.callbacks {
		select {
		case s.callbacks[k] <- key:
		default:
		}
	}

	if s.routeType == common.Hash {
		ids := s.hashring.Add(proxy.GetKey())
		return ids, nil
	} else {
		return nil, nil
	}

}

func (s *service) UpdateStatus(key string, value []byte) error {
	proxy := s.GetProxy(key)
	if proxy == nil {
		return fmt.Errorf("proxy:%s is not exit", key)
	}

	return proxy.update(value)
}

func (s *service) Stop() {

}

func (s *service) Destroy() {
	s.proxyMap.Range(func(key, value interface{}) bool {
		value.(*proxy).OnClose()
		s.Count = int32(atomic.AddInt32(&s.Count, -1))
		s.proxyMap.Delete(key)
		return true
	})
}

func (s *service) Dispatch(msg *framepb.Msg) error {
	proxy, err := s.Route(msg)
	if err != nil {
		return err
	}
	err = proxy.publish(msg)
	return err
}

func (s *service) Request(msg *framepb.Msg) (*framepb.Msg, error) {
	proxy, err := s.Route(msg)
	if err != nil {
		return nil, err
	}
	return proxy.request(msg)
}

func (s *service) Send(msg *framepb.Msg) error {
	proxy, err := s.Route(msg)
	if err != nil {
		return err
	}
	return proxy.sendMsg(msg)
}

func (s *service) Broadcast(msg *framepb.Msg) error {
	s.proxyMap.Range(func(key, value interface{}) bool {
		proxy := value.(*proxy)
		proxy.broadcast(msg)
		return true
	})
	return nil
}

func (s *service) RouteById(id uint64) (pr *proxy, err error) {
	if s.Count == 0 {
		return nil, errors.New("service has no proxy err")
	}
	var index string
	if s.routeType == common.Hash {
		err, index = s.hashring.Get(fmt.Sprintf("%d", id))
		if err != nil {
			return nil, err
		}
	} else {
		index = fmt.Sprintf("route_%d", id)
	}

	proxy := s.GetProxy(index)
	if proxy == nil {
		return nil, errors.New("proxy is nil err ")
	}
	return proxy, nil
}

func (s *service) Route(msg *framepb.Msg) (*proxy, error) {
	return s.RouteById(msg.GetKey())

}

func (s *service) register_callback(callback chan string) error {
	if s.callbacks == nil {
		return errors.New("callbacks not init err ")
	}

	s.proxyMap.Range(func(key, value interface{}) bool {
		callback <- key.(string)
		return true
	})
	return nil
}

func (s *service) Status() *framepb.ServiceStatus {
	status := &framepb.ServiceStatus{
		Name:   s.serType,
		Status: make([][]byte, 0),
	}
	s.proxyMap.Range(func(key, value interface{}) bool {
		status.Status = append(status.Status, value.(*proxy).GetStatus())
		return true
	})
	return status
}
