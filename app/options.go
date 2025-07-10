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
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/module"
	"github.com/eleren/newtbig/process"
	"github.com/spf13/viper"
)

func newOptions(opts ...module.Option) module.Options {
	var wdPath, port, confName *string
	var serverName *string
	var err error
	opt := module.Options{
		Process:           process.NewProcessor(),
		Metadata:          make(map[string]string),
		Debug:             false,
		Parse:             true,
		ConfName:          "app",
		NetType:           "tcp",
		MaxPacketSize:     int64(4096),
		MaxMsgChanLen:     uint32(128),
		HeartBeatInterval: 10,
		ExitWaitTTL:       time.Duration(10),
		RegisterInterval:  time.Duration(3),
		RegisterTTL:       5,
		RPCExpired:        time.Duration(5),
		RPCMaxCoroutine:   0,
		RpmLimit:          20,
		MaxConn:           100000,
		ReadDeadline:      60,
		WriteDeadline:     60,
		NetHeadSize:       uint16(2),
		ByteOrder:         true,
		DisLen:            10000,
		DisRep:            5,
		HBBack:            false,
		QMsgLen:           10000,
	}

	for _, o := range opts {
		o(&opt)
	}

	if opt.Parse {
		wdPath = flag.String("wdir", "", "Server work directory")
		confName = flag.String("conf", "", "Server configuration file name")
		serverName = flag.String("name", "", "Server Name ")
		port = flag.String("port", "", "server listen port")
		flag.Parse()
	}

	if *wdPath == "" {
		opt.WorkDir, err = os.Getwd()
		if err != nil {
			fmt.Println("get cur workspace err:", err.Error())
			file, _ := exec.LookPath(os.Args[0])
			ApplicationPath, _ := filepath.Abs(file)
			opt.WorkDir, _ = filepath.Split(ApplicationPath)
		}
	} else {
		opt.WorkDir = *wdPath
		_, err = os.Open(opt.WorkDir)
		if err != nil {
			panic(err)
		}
		os.Chdir(opt.WorkDir)
	}
	fmt.Println("options work direction :", opt.WorkDir)

	if *confName != "" {
		opt.ConfName = *confName
	}
	confPath := fmt.Sprintf("%s/config", opt.WorkDir)
	confFile := fmt.Sprintf("%s/%s.yaml", confPath, opt.ConfName)
	_, err = os.Open(confFile)
	if err != nil {
		panic(fmt.Sprintf("options config path error %v", err))
	}

	viper.AddConfigPath(confPath)
	viper.SetConfigName(opt.ConfName)
	err = viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("options read config error config file: %s \n", err))
	}

	logsPath := viper.GetString("app.logPath")
	if logsPath == "" {
		logsPath = fmt.Sprintf("%s/%s", opt.WorkDir, "logs/")
	}
	_, err = os.Open(logsPath)
	if err != nil {
		err := os.Mkdir(logsPath, os.ModePerm) //
		if err != nil {
			panic(fmt.Sprintf("options make logs dir err: %s \n", err))
		}
	}

	opt.ETCDHosts = viper.GetStringSlice("app.etcd.hosts")
	if len(opt.ETCDHosts) == 0 {
		panic("etcd config err: hosts is nil   \n")
	}

	opt.NatsHosts = viper.GetStringSlice("app.nats.hosts")
	if len(opt.NatsHosts) == 0 {
		panic("nats config err: hosts is nil  \n ")
	}

	opt.ETCDUser = viper.GetString("app.etcd.user")
	opt.ETCDPassword = viper.GetString("app.etcd.pwd")
	certFile := viper.GetString("app.etcd.certFile")
	if certFile != "" {
		opt.ETCDCertFile = fmt.Sprintf("%s/config/tls/%s", opt.WorkDir, certFile)
	}
	keyFile := viper.GetString("app.etcd.keyFile")
	if keyFile != "" {
		opt.ETCDKeyFile = fmt.Sprintf("%s/config/tls/%s", opt.WorkDir, keyFile)
	}

	opt.AppName = viper.GetString("app.name")
	opt.ServerType = viper.GetString("app.serverType")
	opt.ServerName = viper.GetString("app.serverName")
	opt.ServerIP = viper.GetString("app.ip")
	opt.ServerIPOut = viper.GetString("app.ipOut")
	opt.ServerPort = viper.GetString("app.port")
	opt.ServerConTypes = viper.GetStringSlice("app.conTypes")
	opt.RunMode = viper.GetString("app.runMode")
	disLen := viper.GetUint32("app.disLen")
	if disLen > 0 {
		opt.DisLen = disLen
	}
	disRep := viper.GetUint32("app.disRep")
	if disRep > 0 {
		opt.DisRep = disRep
	}
	qMsgLen := viper.GetUint32("app.qMsgLen")
	if qMsgLen > 0 {
		opt.QMsgLen = qMsgLen
	}
	if opt.RunMode == "debug" {
		opt.Debug = true
	} else if opt.RunMode == "release" || opt.RunMode == "dev" {
		opt.Debug = false
	}

	if *port != "" {
		opt.ServerPort = *port
	}

	if *serverName != "" {
		opt.ServerName = *serverName
	}

	logging.InitLog(logsPath, fmt.Sprintf("%s_%s_%s", opt.AppName, opt.ServerType, opt.ServerName), opt.Debug)
	return opt
}
