/*
==================================================================================
  Copyright (c) 2019 AT&T Intellectual Property.
  Copyright (c) 2019 Nokia

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
==================================================================================
*/

package main

import (
	"encoding/hex"
	"errors"
	submgr "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/control"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/spf13/viper"
	"time"
)


type E2t struct {
	submgr.E2ap
}

var c chan submgr.RmrDatagram = make(chan submgr.RmrDatagram, 1)

var RAWDATA string

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("e2t")
	viper.AllowEmptyEnv(true)
	RAWDATA = viper.GetString("rawdata")
	if RAWDATA == "" {
		RAWDATA = "000001ea7e000500aaaabbbb"
	}
	xapp.Logger.Info("Initial RAW Data: %v", RAWDATA)
}

func (e *E2t) GeneratePayload(sub_id uint16) (payload []byte, err error) {
	skeleton, err := hex.DecodeString(RAWDATA)
	if err != nil {
		return make([]byte, 0), errors.New("Unable to decode data provided in RCO_RAWDATA environment variable")
	}
	payload, err = e.SetSubscriptionResponseSequenceNumber(skeleton, sub_id)
	return
}

func (e E2t) Consume(mtype, sub_id int, len int, payload []byte) (err error) {
	payload_seq_num, err := e.GetSubscriptionRequestSequenceNumber(payload)
	if err != nil {
		xapp.Logger.Error("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())	
	}	
	xapp.Logger.Info("Message Received: RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", sub_id, payload_seq_num)
	err = e.sendSubscriptionResponse(uint16(sub_id))
	return
}

func (e *E2t) sendSubscriptionResponse(sub_id uint16) (err error) {
	payload, err := e.GeneratePayload(sub_id)
	if err != nil {
		xapp.Logger.Debug(err.Error())
		return
	}
	c  <- submgr.RmrDatagram{12011, sub_id, payload}
	return
}

func (e *E2t) Run() {
	for {
		message := <-c
		payload_seq_num, err := e.GetSubscriptionResponseSequenceNumber(message.Payload)
		if err != nil {
			xapp.Logger.Debug("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())	
		}
		xapp.Logger.Info("Sending Message: TYPE: %v | RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v)", message.MessageType, message.SubscriptionId, payload_seq_num)
		xapp.Rmr.Send(message.MessageType, int(message.SubscriptionId), len(message.Payload), message.Payload)
	}
}

func (e *E2t) sendInvalidTestMessages() {
	payload, err := e.GeneratePayload(0)
	if err != nil {
		return
	}
	for {
		time.Sleep(7 * time.Second)
		c <- submgr.RmrDatagram{12011, 0, payload}
		time.Sleep(7 * time.Second)
		c <- submgr.RmrDatagram{12011, 0, make([]byte, 1)}
	}
}

func main() {
	e2t := E2t{}
	go e2t.Run()
	go e2t.sendInvalidTestMessages()
	xapp.Run(e2t)
}
