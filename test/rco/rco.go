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
	"os"
	"strconv"
	"time"
)

type Rco struct {
	submgr.E2ap
}

var c = make(chan submgr.RmrDatagram, 1)
var params xapp.RMRParams

var requestRawData string
var deleteRawData string
var seedSN uint16
var deleteSeedSN uint16

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("rco")
	viper.AllowEmptyEnv(true)
	requestRawData = viper.GetString("rawdata")
	if requestRawData == "" {
		requestRawData = "00c90020000003ea7e00050000010002ea6300020003ea81000a000000ea6b4003000440"
	}
	deleteRawData = viper.GetString("deleterawdata")
	if deleteRawData == "" {
		deleteRawData = "00ca0012000002ea7e00050000010002ea6300020003"
	}
	xapp.Logger.Info("Initial RAW DATA: %v", requestRawData)
	xapp.Logger.Info("Initial DELETE RAW DATA: %v", deleteRawData)
	seedSN = uint16(viper.GetInt("seed_sn"))
	if seedSN == 0 || seedSN > 65535 {
		seedSN = 12345
	}
	deleteSeedSN = uint16(viper.GetInt("delete_seed_sn"))
	if deleteSeedSN == 0 || deleteSeedSN > 65535 {
		deleteSeedSN = seedSN
	}

	xapp.Logger.Info("Initial SEQUENCE NUMBER: %v", seedSN)
}

func (r *Rco) GeneratePayload(subId uint16) (payload []byte, err error) {
	skeleton, err := hex.DecodeString(requestRawData)
	if err != nil {
		return make([]byte, 0), errors.New("nable to decode data provided in \"RCO_RAWDATA\" environment variable")
	}
	payload, err = r.SetSubscriptionRequestSequenceNumber(skeleton, subId)
	return
}

func (r *Rco) GenerateDeletePayload(subId uint16) (payload []byte, err error) {
	skeleton, err := hex.DecodeString(deleteRawData)
	if err != nil {
		return make([]byte, 0), errors.New("unable to decode data provided in \"RCO_DELETERAWDATA\" environment variable")
	}
	payload, err = r.SetSubscriptionDeleteRequestSequenceNumber(skeleton, subId)
	return
}

func (r Rco) Consume(params *xapp.RMRParams) (err error) {
	switch params.Mtype {
	case xapp.RICMessageTypes["RIC_SUB_RESP"]:
		payloadSeqNum, err := r.GetSubscriptionResponseSequenceNumber(params.Payload)
		if err != nil {
			xapp.Logger.Error("SUBRESP: Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		}
		xapp.Logger.Info("Subscription Response Message Received: RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", params.SubId, payloadSeqNum)
		return err
	case xapp.RICMessageTypes["RIC_SUB_DEL_RESP"]:
		payloadSeqNum, err := r.GetSubscriptionDeleteResponseSequenceNumber(params.Payload)
		if err != nil {
			xapp.Logger.Error("DELRESP: Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		}
		xapp.Logger.Info("Subscription Delete Response Message Received: RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", params.SubId, payloadSeqNum)
		return err
	default:
		err = errors.New("Message Type " + strconv.Itoa(params.Mtype) + " is discarded")
		xapp.Logger.Error("Unknown message type: %v", err)
		return
	}
}

func (r *Rco) SendRequests() (err error) {
	message, err := r.GeneratePayload(seedSN)
	if err != nil {
		xapp.Logger.Debug(err.Error())
		return
	}
	deletemessage, err := r.GenerateDeletePayload(deleteSeedSN)
	if err != nil {
		xapp.Logger.Debug(err.Error())
		return
	}
	for {
		time.Sleep(5 * time.Second)
		c <- submgr.RmrDatagram{MessageType: 12010, SubscriptionId: seedSN, Payload: message}
		seedSN++
		time.Sleep(5 * time.Second)
		c <- submgr.RmrDatagram{MessageType: 12020, SubscriptionId: deleteSeedSN, Payload: deletemessage}
		deleteSeedSN++
	}
}

func (r *Rco) Run() {
	for {
		message := <-c
		payloadSeqNum, err := r.GetSubscriptionRequestSequenceNumber(message.Payload)
		if err != nil {
			xapp.Logger.Debug("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		}
		params.SubId = int(message.SubscriptionId)
		params.Mtype = message.MessageType
		params.PayloadLen = len(message.Payload)
		params.Payload = message.Payload
		xapp.Logger.Info("Sending Message: TYPE: %v | RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v)", message.MessageType, message.SubscriptionId, payloadSeqNum)
		xapp.Rmr.Send(&params, false)
	}
}

func (r *Rco) sendInvalidTestMessages() {
	for {
		time.Sleep(7 * time.Second)
		c <- submgr.RmrDatagram{MessageType: 10000, SubscriptionId: 0, Payload: make([]byte, 1)}
		time.Sleep(7 * time.Second)
		c <- submgr.RmrDatagram{MessageType: 12010, SubscriptionId: 0, Payload: make([]byte, 1)}
	}
}

func main() {
	rco := Rco{}
	go xapp.Rmr.Start(rco)
	go rco.Run()
	go rco.sendInvalidTestMessages()
	err := rco.SendRequests()
	if err != nil {
		xapp.Logger.Info("Error: %v", err)
		os.Exit(1)
	}
}
