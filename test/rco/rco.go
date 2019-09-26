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
	"os"
)

type Rco struct {
	submgr.E2ap
}

var c chan submgr.RmrDatagram = make(chan submgr.RmrDatagram, 1)
var params *xapp.RMRParams

var REQUESTRAWDATA string
var DELETERAWDATA string
var SEEDSN uint16
var DELETESEEDSN uint16

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("rco")
	viper.AllowEmptyEnv(true)
	REQUESTRAWDATA = viper.GetString("rawdata")
	if REQUESTRAWDATA == "" {
		REQUESTRAWDATA = "000003ea7e000500aaaaccccea6300020000ea81000e00045465737400ea6b0003000100"
	}
	DELETERAWDATA = viper.GetString("deleterawdata")
	if DELETERAWDATA == "" {
		DELETERAWDATA = "000002ea7e000500aaaabbbbea6300020000"
	}
	xapp.Logger.Info("Initial RAW DATA: %v", REQUESTRAWDATA)
	xapp.Logger.Info("Initial DELETE RAW DATA: %v", DELETERAWDATA)
	SEEDSN = uint16(viper.GetInt("seed_sn"))
	if SEEDSN == 0 || SEEDSN > 65535 {
		SEEDSN = 12345
	}
	DELETESEEDSN = uint16(viper.GetInt("delete_seed_sn"))
	if DELETESEEDSN == 0 || DELETESEEDSN > 65535 {
		DELETESEEDSN = SEEDSN
	}

	xapp.Logger.Info("Initial SEQUENCE NUMBER: %v", SEEDSN)
}

func (r *Rco) GeneratePayload(sub_id uint16) (payload []byte, err error) {
	skeleton, err := hex.DecodeString(REQUESTRAWDATA)
	if err != nil {
		return make([]byte, 0), errors.New("Unable to decode data provided in RCO_RAWDATA environment variable")
	}
	payload, err = r.SetSubscriptionRequestSequenceNumber(skeleton, sub_id)
	return
}

func (r *Rco) GenerateDeletePayload(sub_id uint16) (payload []byte, err error) {
	skeleton, err := hex.DecodeString(DELETERAWDATA)
	if err != nil {
		return make([]byte, 0), errors.New("Unable to decode data provided in RCO_DELETE RAWDATA environment variable")
	}
	payload, err = r.SetSubscriptionDeleteRequestSequenceNumber(skeleton, sub_id)
	return
}

func (r Rco) Consume(params *xapp.RMRParams) (err error) {
	payload_seq_num, err := r.GetSubscriptionResponseSequenceNumber(params.Payload)
	if err != nil {
		xapp.Logger.Error("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())	
	}
	xapp.Logger.Info("Message Received: RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", params.SubId, payload_seq_num)
	return
}

func (r *Rco) SendRequests() (err error) {
	message, err := r.GeneratePayload(SEEDSN)
	if err != nil {
		xapp.Logger.Debug(err.Error())
		return
	}
	deletemessage, err := r.GenerateDeletePayload(DELETESEEDSN)
	if err != nil {
		xapp.Logger.Debug(err.Error())
		return
	}
	for {
		time.Sleep(2 * time.Second)
		c <- submgr.RmrDatagram{12010, SEEDSN, message}
		SEEDSN++
		time.Sleep(2 * time.Second)
		c <- submgr.RmrDatagram{12020, DELETESEEDSN, deletemessage}
		DELETESEEDSN++
	}
	return
}

func (r *Rco) Run() {
	for {
		message := <-c
		payload_seq_num, err := r.GetSubscriptionRequestSequenceNumber(message.Payload)
		if err != nil {
			xapp.Logger.Debug("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		}
		params.SubId = int(message.SubscriptionId)
		params.Mtype = message.MessageType
		params.PayloadLen = len(message.Payload)
		params.Payload = message.Payload
		xapp.Logger.Info("Sending Message: TYPE: %v | RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v)", message.MessageType, message.SubscriptionId, payload_seq_num)
		xapp.Rmr.Send(params, false)
	}
}

func (r *Rco) sendInvalidTestMessages(){
	for {
		time.Sleep(7 * time.Second)
		c <- submgr.RmrDatagram{10000, 0, make([]byte, 1)}
		time.Sleep(7 * time.Second)
		c <- submgr.RmrDatagram{12010, 0, make([]byte, 1)}
	}
}

func main() {
	rco := Rco{}
	go xapp.Rmr.Start(rco)
	go rco.Run()
	go rco.sendInvalidTestMessages()
	err := rco.SendRequests()
	if err != nil {
		os.Exit(1)
	}
}
