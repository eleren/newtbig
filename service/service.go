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
package service

import (
	"context"
	"fmt"

	"time"

	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
	"google.golang.org/protobuf/proto"
)

func NewService(ctx context.Context, app module.App) module.Service {
	appName := app.Options().AppName
	serverType := app.Options().ServerType
	serverName := app.Options().ServerName
	key := appName + "/" + serverType + "/" + serverName
	value := app.Options().ServerIPOut + ":" + app.Options().ServerPort
	status := &framepb.Status{
		Key:   key,
		Addr:  value,
		Count: 0,
	}
	return &service{
		ctx:    ctx,
		app:    app,
		server: app.Options().Serv,
		Key:    key,
		Addr:   value,
		Status: status,
	}
}

type service struct {
	ctx    context.Context
	app    module.App
	server module.Server
	Key    string
	Addr   string
	Status *framepb.Status
}

func (s *service) register() error {
	s.Status.Count = s.server.Count()
	data, err := proto.Marshal(s.Status)
	if err != nil {
		log.Logger.Error("register Marshal err :", err.Error())
		return err
	}

	eClient := s.app.GetEClient().GetClient()
	_, err = eClient.Put(context.Background(), s.Key, string(data))
	if err != nil {
		return fmt.Errorf("server registe err:%s", err.Error())
	}
	return nil
}

func (s *service) deregister() error {
	log.Logger.Info("deregister key:", s.Key)
	eClient := s.app.GetEClient().GetClient()
	if eClient != nil {
		_, err := eClient.Delete(context.Background(), s.Key)
		if err != nil {
			return fmt.Errorf("server deregister err:%s", err.Error())
		}
	}

	return nil
}

func (s *service) run(exit chan struct{}) {
	if s.app.Options().RegisterInterval <= 0 {
		log.Logger.Warn("service run  RegisterInterval err")
		return
	}
	t := time.NewTicker(time.Second * s.app.Options().RegisterInterval)
	defer func() {
		t.Stop()
	}()

	registeTime := 0
	for {
		select {
		case <-t.C:
			err := s.register()
			if err != nil {
				log.Logger.Warnf("service run register err:%s ", err.Error())
				registeTime = registeTime + 1
				if registeTime > s.app.Options().RegisterTTL {
					log.Logger.Errorf("service run register timeout :%s ", err.Error())
				}
			}
			registeTime = 0
		case _, ok := <-exit:
			if !ok {
				return
			}
		}
	}
}

func (s *service) Server() module.Server {
	return s.server
}

func (s *service) String() string {
	return "newtbig_service"
}

func (s *service) Start() error {
	if err := s.server.BeforeStart(); err != nil {
		return err
	}
	if err := s.server.Start(); err != nil {
		return err
	}
	if err := s.register(); err != nil {
		return err
	}
	if err := s.server.AfterStart(); err != nil {
		return err
	}
	return nil
}

func (s *service) Stop() error {
	if err := s.server.BeforeStop(); err != nil {
		log.Logger.Warnf("service Stop BeforeStop err: %s ", err.Error())
	}
	if err := s.deregister(); err != nil {
		log.Logger.Warnf("service Stop deregister err: %s ", err.Error())
	}
	if err := s.server.Stop(); err != nil {
		log.Logger.Warnf("service Stop Stop err: %s ", err.Error())
	}
	if err := s.server.AfterStop(); err != nil {
		log.Logger.Warnf("service Stop AfterStop err: %s ", err.Error())
	}
	return nil
}

func (s *service) Run() error {
	if err := s.Start(); err != nil {
		return err
	}
	ex := make(chan struct{})
	utils.SafeGO(func() {
		s.run(ex)
	})
	select {
	case <-s.ctx.Done():
		log.Logger.Info("service Run app done sign ")
	}
	close(ex)
	return s.Stop()
}
