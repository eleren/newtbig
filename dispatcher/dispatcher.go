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
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/eleren/newtbig/common"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
	"go.etcd.io/etcd/api/v3/mvccpb"
	etcdclient "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"

	"golang.org/x/net/context"
)

func NewDispatcher(app module.App) module.Dispatcher {
	dis := new(Dispatcher)
	err := dis.init(app)
	if err != nil {
		log.Logger.Errorf("NewDispatcher err:%s", err.Error())
		panic(err)
	}
	return dis
}

type Dispatcher struct {
	root           string
	names          map[string]common.RouteType
	services       map[string]*service
	names_provided bool
	mu             sync.RWMutex
	route          map[uint32]string
	opts           module.Options
	app            module.App
	exit           chan struct{}
	myName         string
	myType         string
	gKey           string
}

func (p *Dispatcher) init(app module.App) error {
	p.exit = make(chan struct{})
	p.app = app
	p.opts = app.Options()
	etcdhosts := p.opts.ETCDHosts
	log.Logger.Infof("init etcdhosts :%v", etcdhosts)

	p.root = p.opts.AppName
	p.myType = fmt.Sprintf("%s/%s", p.root, p.opts.ServerType)
	p.services = make(map[string]*service)
	p.names = make(map[string]common.RouteType)
	p.route = make(map[uint32]string)
	names := p.opts.ServerConTypes
	if len(names) > 0 {
		p.names_provided = true
	}

	for _, v := range names {
		values := strings.Split(strings.TrimSpace(v), "|")
		if len(values) == 2 {
			rType, err := strconv.Atoi(values[1])
			if err != nil {
				log.Logger.Warnf("init service names :%s  err :%s", v, err.Error())
			} else {
				p.names[p.root+"/"+values[0]] = common.RouteType(rType)
			}
		} else {
			p.names[p.root+"/"+strings.TrimSpace(v)] = common.Hash
		}
	}
	if p.names_provided {
		p.gKey = fmt.Sprintf("%s/%s", p.root, names[0])
	}
	log.Logger.Infof("all service names:%v", p.names)
	return nil

}

func (p *Dispatcher) Start() error {
	err := p.connectAll(p.root)
	if err != nil {
		log.Logger.Warnf("init err:%s", err.Error())
	}

	utils.SafeGO(func() {
		p.watcher()
	})

	return nil

}

func (p *Dispatcher) Stop() error {
	close(p.exit)

	return nil
}

func (p *Dispatcher) connectAll(directory string) error {
	kAPI := p.app.GetEClient().GetClient()
	log.Logger.Info("connect_all services dic:", directory)
	resp, err := kAPI.Get(context.Background(), directory, etcdclient.WithPrefix())
	if err != nil {
		return fmt.Errorf("connect_all err:%s", err.Error())
	}

	if resp == nil {
		return fmt.Errorf("connect_all resp is nil")
	}

	if len(resp.Kvs) == 0 {
		log.Logger.Info("connect_all no services :", directory)
		return nil
	}

	for _, node := range resp.Kvs {
		log.Logger.Debug("add service key:", string(node.Key), " value:", string(node.Value))
		err := p.addService(string(node.Key), node.Value)
		if err != nil {
			log.Logger.Error("connect_all add service err key:", string(node.Key), " value:", string(node.Value), "err:", err.Error())
		}
	}

	return nil
}

func (p *Dispatcher) Clear() error {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("dis clear recover err : %v ", r)
		}
	}()

	log.Logger.Info("dis clear : ", p.root)
	eClient := p.app.GetEClient().GetClient()
	if eClient == nil {
		return fmt.Errorf("dis clear err : etcd api is nil ")
	}
	delResp, err := eClient.Delete(context.Background(), p.root, etcdclient.WithPrefix())
	if err != nil {
		return fmt.Errorf("dis clear err :%s", err.Error())
	}

	if len(delResp.PrevKvs) != 0 {
		for _, kvpx := range delResp.PrevKvs {
			log.Logger.Infof("dis clear del key :%s  value :%s ", string(kvpx.Key), string(kvpx.Value))
		}
	}

	return nil
}

func (p *Dispatcher) watcher() {
	log.Logger.Info("dis watcher : ", p.root)
	watchRespChan, err := p.app.GetEClient().Watch(p.root)
	if err != nil || watchRespChan == nil {
		log.Logger.Errorf("watchRespChan is nil ")
		return
	}

	for {
		select {
		case watchResp, ok := <-watchRespChan:
			if !ok {
				log.Logger.Errorf("watcher watchRespchan err")
				return
			}
			for _, event := range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT:
					err := p.addService(string(event.Kv.Key), event.Kv.Value)
					if err != nil {
						log.Logger.Errorf("watcher put err:%s", err.Error())
					}

				case mvccpb.DELETE:
					log.Logger.Debug("watcher del key:", string(event.Kv.Key), "Revision:", event.Kv.CreateRevision, event.Kv.ModRevision)
					err := p.removeService(string(event.Kv.Key))
					if err != nil {
						log.Logger.Errorf("watcher del err:%s", err.Error())
					}

				}
			}

		case _, ok := <-p.exit:
			if !ok {
				log.Logger.Infof("watcher :%s stop", p.root)
				return
			}

		}
	}

}

func (p *Dispatcher) addService(key string, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	service_name := filepath.Dir(key)
	service_name = strings.ReplaceAll(service_name, "\\", "/")
	if p.names_provided && p.names[service_name] <= 0 {
		log.Logger.Warnf("servcie %s not provided for key:%s ", service_name, key)
		return nil
	}

	ser := p.services[service_name]
	if ser == nil {
		ser = new(service)
		ser.init(service_name, p.app, p.names[service_name])
		p.services[service_name] = ser
	}

	proxy := ser.GetProxy(key)
	if proxy == nil {
		ids, err := ser.CreateProxy(key, value)
		log.Logger.Debugf("add_service proxy create :%s  value :%s ids :%v", key, value, ids)
		if err != nil {
			return err
		}
	} else {
		err := proxy.update(value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Dispatcher) removeService(key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	service_name := filepath.Dir(key)
	service_name = strings.ReplaceAll(service_name, "\\", "/")
	if p.names_provided && p.names[service_name] <= 0 {
		return fmt.Errorf("servcie %s not provided err", service_name)
	}

	service := p.services[service_name]
	if service == nil {
		return fmt.Errorf("servcie %s not ext err", service_name)
	}
	err := service.DelProxy(key)
	if err != nil {
		return err
	}
	return nil
}

func (p *Dispatcher) registerCallback(path string, callback chan string) error {
	service := p.services[path]
	if service == nil {
		return errors.New("service not exit err ")
	}
	service.register_callback(callback)

	log.Logger.Debug("register callback on:", path)
	return nil
}

func (p *Dispatcher) RegisteSer(id uint32, key string) error {
	if p.route == nil {
		return errors.New("route is nil err !")
	}
	p.route[id] = key
	return nil
}

func (p *Dispatcher) GetMsgKey(id uint32) (string, error) {
	key, ok := p.route[id]
	if !ok {
		return "", errors.New("msg id not exit err")
	}
	return key, nil
}

func (p *Dispatcher) getService(msg *framepb.Msg) (*service, error) {
	if msg == nil {
		return nil, errors.New("Route message is nil err ")
	}
	key, err := p.GetMsgKey(msg.ID)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("%s/%s", p.root, key)
	p.mu.RLock()
	service := p.services[path]
	if service == nil {
		p.mu.RUnlock()
		return nil, fmt.Errorf("Route service:%s not exit err", path)
	}
	p.mu.RUnlock()
	return service, nil
}

func (p *Dispatcher) Dispatch(msg *framepb.Msg) error {
	service, err := p.getService(msg)
	if err != nil {
		return err
	}
	return service.Dispatch(msg)
}

func (p *Dispatcher) Request(msg *framepb.Msg) (*framepb.Msg, error) {
	service, err := p.getService(msg)
	if err != nil {
		return nil, err
	}
	return service.Request(msg)
}

func (p *Dispatcher) Broadcast(msg *framepb.Msg) error {
	service, err := p.getService(msg)
	if err != nil {
		return err
	}
	return service.Broadcast(msg)
}

func (p *Dispatcher) SendMsg(msg *framepb.Msg) error {
	service, err := p.getService(msg)
	if err != nil {
		return err
	}
	return service.Send(msg)
}

func (p *Dispatcher) GetServer(sType string, id uint64) ([]byte, error) {
	proxy, err := p.getProxy(sType, id)
	if err != nil {
		return nil, err
	}
	return proxy.GetStatus(), nil
}

func (p *Dispatcher) getProxy(sType string, id uint64) (*proxy, error) {
	if sType == "" {
		return nil, errors.New("getserver _type is nil err ")
	}
	if id == 0 {
		return nil, errors.New("getserver id is nil err ")
	}

	p.mu.RLock()
	service := p.services[sType]
	if service == nil {
		p.mu.RUnlock()
		return nil, fmt.Errorf("getserver servcie: %s not ext err", sType)
	}
	p.mu.RUnlock()

	proxy, err1 := service.RouteById(id)
	if err1 != nil {
		return nil, err1
	}
	return proxy, nil
}

func (p *Dispatcher) GetGate(id uint64) ([]byte, error) {
	if p.names_provided {
		return p.GetServer(p.gKey, id)
	} else {
		return nil, fmt.Errorf("no server provided err")
	}
}

func (p *Dispatcher) GetStatus() ([]byte, error) {
	status := &framepb.ServiceInfo{
		Sers: make([]*framepb.ServiceStatus, 0),
	}
	for _, service := range p.services {
		status.Sers = append(status.Sers, service.Status())
	}

	data, err := proto.Marshal(status)
	if err != nil {
		log.Logger.Error("dis Marshal err :", err.Error())
		return nil, err
	}
	return data, nil
}

func (p *Dispatcher) CheckRoute(key uint64) bool {
	if p.names_provided {
		proxy, err := p.getProxy(p.myType, key)
		if err != nil {
			return false
		} else {
			if strings.Compare(p.myName, proxy.GetSerName()) == 0 {
				return true
			} else {
				return false
			}
		}
	} else {
		return false
	}
}
