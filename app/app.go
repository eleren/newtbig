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
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/eleren/newtbig/dispatcher"
	"github.com/eleren/newtbig/rpc"
	"github.com/eleren/newtbig/server"
	"github.com/eleren/newtbig/service"
	"github.com/eleren/newtbig/utils"

	"github.com/eleren/newtbig/etcdclient"
	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
)

var (
	appOnce sync.Once
	app     *DefaultApp
)

func NewApp(opts ...module.Option) module.App {
	appOnce.Do(func() {
		options := newOptions(opts...)
		app = new(DefaultApp)
		app.isGate = false
		app.opts = options
		app.opts.Serv = server.NewServer(app)
		app.dspatcher = dispatcher.NewDispatcher(app)

	})
	return app
}

type DefaultApp struct {
	module.App
	context.Context
	serStoped  chan bool
	opts       module.Options
	exit       context.CancelFunc
	etcdClient *etcdclient.ETCDClient
	dspatcher  module.Dispatcher
	rpc        module.RPC
	service    module.Service
	registe    func(app module.App)
	isGate     bool
}

func (app *DefaultApp) Run(opts ...module.Option) error {
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("app run recover err stack :%v", r)
		}
	}()
	log.Logger.Info("app starting  ...")

	err := app.OnInit()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	app.exit = cancel
	app.serStoped = make(chan bool)
	app.service = service.NewService(ctx, app)

	utils.SafeGO(func() {
		err := app.service.Run()
		if err != nil {
			log.Logger.Warnf("service run fail id : %s error :%s", app.GetAPPID(), err.Error())
		}
		close(app.serStoped)
	})

	log.Logger.Info("app started :", app.opts.Version)
	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	sig := <-c
	//日志记录保存
	log.Logger.Sync()

	timeout := time.NewTimer(time.Second * app.opts.ExitWaitTTL)
	wait := make(chan struct{})
	utils.SafeGO(func() {
		app.OnDestroy()
		close(wait)
	})
	select {
	case <-timeout.C:
		log.Logger.Warnf(" close timeout signal: %v", sig)
	case _, ok := <-wait:
		if !ok {
			log.Logger.Infof(" closing done signal: %v ", sig)
		}

	}

	return nil
}

func (app *DefaultApp) UpdateOptions(opts ...module.Option) error {
	for _, o := range opts {
		o(&app.opts)
	}
	return nil
}

func (app *DefaultApp) Options() module.Options {
	return app.opts
}

func (app *DefaultApp) OnInit() error {
	hostname, _ := os.Hostname()
	app.opts.HostName = hostname
	app.opts.PID = fmt.Sprintf("%v", os.Getpid())
	log.Logger.Debug("etcd init host:", app.Options().ETCDHosts)
	log.Logger.Debug("etcd init tls:", app.Options().ETCDCertFile, app.Options().ETCDKeyFile)
	app.etcdClient = &etcdclient.ETCDClient{}
	err := app.etcdClient.Init(app.Options().ETCDHosts, app.Options().ETCDCertFile, app.Options().ETCDKeyFile)
	if err != nil {
		return err
	}
	log.Logger.Infof("etcd init suc :%v", app.Options().ETCDHosts)
	if app.Options().Process == nil {
		return fmt.Errorf("process is nill err !")
	}
	app.Options().Process.Init(app)
	if app.registe != nil {
		app.registe(app)
	}
	app.rpc, err = rpc.NewRPCClient(app)
	if err != nil {
		return fmt.Errorf("DefaultApp NewRPCClient err:%s", err.Error())
	}

	err = app.rpc.Start()
	if err != nil {
		return err
	}
	err = app.dspatcher.Start()
	if err != nil {
		return err
	}
	return nil
}

func (app *DefaultApp) OnDestroy() error {
	log.Logger.Info("app ondestroy ")
	defer func() {
		if r := recover(); r != nil {
			log.Logger.Errorf("app ondestroy recover err stack :%v", r)
		}
	}()
	log.Logger.Info("app ondestroy service")
	app.exit()
	select {
	case <-app.serStoped:
	}
	log.Logger.Info("app ondestroy app")
	if err := app.rpc.Stop(); err != nil {
		log.Logger.Errorf("app stop rpc err :%s", err.Error())
	}
	if err := app.dspatcher.Stop(); err != nil {
		log.Logger.Errorf("app stop dis err :%s", err.Error())
	}
	if err := app.etcdClient.GetClient().Close(); err != nil {
		log.Logger.Errorf("app stop etcd err :%s", err.Error())
	}

	return nil
}

func (app *DefaultApp) GetAPPID() string {
	return app.opts.ServerType + "/" + app.opts.ServerName
}

func (app *DefaultApp) WorkDir() string {
	return app.opts.WorkDir
}

func (app *DefaultApp) GetEClient() *etcdclient.ETCDClient {
	return app.etcdClient
}

func (app *DefaultApp) GetRPC() module.RPC {
	return app.rpc
}

func (app *DefaultApp) IsGate() bool {
	return app.isGate
}

func (app *DefaultApp) GetDispatcher() module.Dispatcher {
	return app.dspatcher
}

func (app *DefaultApp) Register(msgID uint32, msgHandler module.DisposeFunc) {
	app.opts.Process.Register(msgID, msgHandler)
}

func (app *DefaultApp) RegisteSer(id uint32, ser string) error {
	if app.dspatcher == nil {
		err := fmt.Errorf("dispatcher is nil ")
		log.Logger.Error(err.Error())
		return err
	}
	return app.dspatcher.RegisteSer(id, ser)
}

func (app *DefaultApp) OnRegiste(_func func(app module.App)) error {
	app.registe = _func
	return nil
}
