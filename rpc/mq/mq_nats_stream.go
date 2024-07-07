// Copyright 2020 newtbig Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package mq

import (
	"context"
	"fmt"
	"math"
	"time"

	log "github.com/eleren/newtbig/logging"
	"github.com/eleren/newtbig/msg/framepb"
	"github.com/eleren/newtbig/utils"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/proto"

	"github.com/nats-io/nats.go"
)

func (nc *NatsClient) AddStream(stream string, subjects []string) *nats.StreamInfo {
	nctx := nats.Context(context.Background())
	info, errSI := nc.nats_js.StreamInfo(stream)
	if errSI != nil {
		log.Logger.Warnf(fmt.Sprintf("Nats StreamInfo err:%s ", errSI.Error()))
	}
	if nil == info {
		_, errAS := nc.nats_js.AddStream(&nats.StreamConfig{
			Name:       stream,
			Subjects:   subjects,
			Retention:  nats.WorkQueuePolicy,
			Replicas:   1,
			Discard:    nats.DiscardOld,
			Duplicates: 30 * time.Second,
		}, nctx)
		if errAS != nil {
			log.Logger.Error(fmt.Sprintf("Nats AddStream err:%s ", errAS.Error()))
			return nil
		}
	}
	return info
}

func (nc *NatsClient) InitStreamWithSubjects() {
	stream := nc.AddStream(nc.opts.AppName, nc.streamSubs)
	if stream == nil {
		log.Error("InitStreamWithSubjects AddStream err ! ")
		return
	}

	nctx := nats.Context(context.Background())
	tctx, cancel := context.WithTimeout(nctx, 240000*time.Hour)
	deadlineCtx := nats.Context(tctx)

	results := make(chan int64)
	var totalTime int64
	var totalMessages int64

	log.Logger.Infof("InitStreamWithSubjects sub:%s ", nc.key)
	utils.SafeGO(func() {
		for !nc.isClose {
			select {
			case msg := <-nc.msg_send_chan:
				now := time.Now().UnixMicro()
				ack, err := nc.nats_js.Publish(msg.Subj, msg.Data)
				results <- time.Now().UnixMicro() - now
				if ack != nil {
					log.Logger.Info("InitStreamWithSubjects ack:", ack.Domain, ack.Duplicate, ack.Sequence, ack.Stream)
				}
				if err != nil {
					log.Logger.Error("InitStreamWithSubjects Publish err:", err.Error())

				}
			}
		}
	})

	utils.SafeGO(func() {
		sub, err := nc.CreatNatsSubscriptionPull(nc.sendKey, "group")
		if err != nil {
			log.Logger.Error("InitStreamWithSubjects CreatNatsSubscriptionPull err:", err.Error())
			return
		}
		for !nc.isClose {
			msgs, err := sub.Fetch(1, deadlineCtx)
			if err != nil {
				log.Logger.Error("InitStreamWithSubjects PullSubscribe err:", err.Error())
				return
			}
			m := msgs[0]
			m.Ack(nats.Context(nctx))
			if m.Data == nil {
				log.Logger.Errorf("InitStreamWithSubjects sub:%s data is nil", m.Subject)
				continue
			}
			err = nc.streamHandler(m.Data)
			if err != nil {
				log.Logger.Error("InitStreamWithSubjects handler err:", err.Error())
				continue
			}
		}
	})

	for {
		select {
		case <-deadlineCtx.Done():
			cancel()
			log.Logger.Infof("sent %d messages with average time of %f", totalMessages, math.Round(float64(totalTime/totalMessages)))
			nc.nats_js.DeleteStream(nc.opts.AppName)
			return
		case usec := <-results:
			totalTime += usec
			totalMessages++
		}
	}
}

func (nc *NatsClient) SendMsg(subj string, msg *framepb.Msg) error {
	defer func() {
		utils.Put(msg)
	}()
	data, err1 := proto.Marshal(msg)
	if err1 != nil {
		return err1
	}
	nc.msg_send_chan <- &StreamMsg{Subj: subj, Data: data}
	return nil
}

func (nc *NatsClient) CreatNatsQueueSubscription(subj string, queue string) (*nats.Subscription, error) {
	id := uuid.NewV4().String()
	sub, err := nc.nats_js.QueueSubscribeSync(subj, queue, nats.Durable(id), nats.DeliverNew())
	return sub, err
}

func (nc *NatsClient) CreatNatsSubscription(subj string) (*nats.Subscription, error) {
	id := uuid.NewV4().String()
	sub, err := nc.nats_js.SubscribeSync(subj, nats.Durable(id), nats.DeliverNew())
	return sub, err
}

func (nc *NatsClient) CreatNatsSubscriptionPull(subj string, queue string) (*nats.Subscription, error) {
	sub, err := nc.nats_js.PullSubscribe(subj, queue)
	return sub, err
}

func (nc *NatsClient) CreatNatsKeyValue(bukName string, ttl time.Duration) (nats.KeyValue, error) {
	kv, err := nc.nats_js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: bukName,
		TTL:    ttl,
	})
	return kv, err
}
