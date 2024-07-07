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
package etcdclient

import (
	"context"
	"errors"
	"time"

	log "github.com/eleren/newtbig/logging"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ETCDClient struct {
	api    clientv3.KV
	config clientv3.Config
	client *clientv3.Client
}

func (ec *ETCDClient) Init(host []string) error {
	ec.config = clientv3.Config{
		Endpoints:   host,
		DialTimeout: 5 * time.Second,
	}

	var err error
	ec.client, err = clientv3.New(ec.config)
	if err != nil {
		return err
	}
	ec.api = clientv3.NewKV(ec.client)
	if ec.api == nil {
		return errors.New("etcd api is nil err")
	}

	ctx, _ := context.WithTimeout(context.Background(), time.Duration(3*time.Second))
	_, err = ec.api.Put(ctx, "newtbig_etcd_test1_key", "newtbig_etcd_test_value")
	if err != nil {
		log.Logger.Error("etcd test err:", err.Error())
		return err
	}

	return nil
}

func (ec *ETCDClient) KeysAPI() clientv3.KV {
	return ec.api
}

func (ec *ETCDClient) GetClient() *clientv3.Client {
	return ec.client
}

func (ec *ETCDClient) Watch(key string) (clientv3.WatchChan, error) {
	getResp, err := ec.api.Get(context.TODO(), key)
	if err != nil {
		return nil, err
	}

	watchStartRevision := getResp.Header.Revision + 1
	if len(getResp.Kvs) != 0 {
		log.Logger.Infof("etcd watch :%s  vision :%d  value:%v", key, watchStartRevision, getResp.Kvs)
	}

	watcher := clientv3.NewWatcher(ec.client)
	watchRespChan := watcher.Watch(context.Background(), key, clientv3.WithRev(watchStartRevision), clientv3.WithPrefix())
	return watchRespChan, nil

}
