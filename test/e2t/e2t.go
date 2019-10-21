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

/*
#include <rmr/RIC_message_types.h>
#include <rmr/rmr.h>

#cgo CFLAGS: -I../
#cgo LDFLAGS: -lrmr_nng -lnng
*/
import "C"

import (
	"encoding/hex"
	"errors"
	submgr "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/control"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/spf13/viper"
	"strconv"
)

type E2t struct {
	submgr.E2ap
}

var c chan xapp.RMRParams = make(chan xapp.RMRParams, 1)

var requestRawData string
var deleteRawData string

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("e2t")
	viper.AllowEmptyEnv(true)
	requestRawData = viper.GetString("rawdata")
	if requestRawData == "" {
		requestRawData = "20c9001d000003ea7e00050000010002ea6300020003ea6c000700ea6d40020004"
	}
	xapp.Logger.Info("Initial RAW Data: %v", requestRawData)
	deleteRawData = viper.GetString("rawdata")
	if deleteRawData == "" {
		deleteRawData = "20ca0012000002ea7e00050000010002ea6300020003"
	}
	xapp.Logger.Info("Initial RAW Data: %v", deleteRawData)
}

func (e *E2t) GenerateRequestPayload(subId uint16) (payload []byte, err error) {
	skeleton, err := hex.DecodeString(requestRawData)
	if err != nil {
		return make([]byte, 0), errors.New("unable to decode data provided in \"RCO_RAWDATA\" environment variable")
	}
	payload, err = e.SetSubscriptionResponseSequenceNumber(skeleton, subId)
	return
}

func (e *E2t) GenerateDeletePayload(subId uint16) (payload []byte, err error) {
	skeleton, err := hex.DecodeString(deleteRawData)
	if err != nil {
		return make([]byte, 0), errors.New("unable to decode data provided in \"RCO_RAWDATA\" environment variable")
	}
	payload, err = e.SetSubscriptionDeleteResponseSequenceNumber(skeleton, subId)
	return
}

func (e E2t) Consume(rp *xapp.RMRParams) (err error) {
	switch rp.Mtype {
	case C.RIC_SUB_REQ:
		err = e.handleSubscriptionRequest(rp)
	case C.RIC_SUB_DEL_REQ:
		err = e.handleSubscriptionDeleteRequest(rp)
	default:
		err = errors.New("Message Type " + strconv.Itoa(rp.Mtype) + " is discarded")
		xapp.Logger.Error("Unknown message type: %v", err)
	}
	return
}

func (e E2t) handleSubscriptionRequest(request *xapp.RMRParams) (err error) {
	payloadSeqNum, err := e.GetSubscriptionRequestSequenceNumber(request.Payload)
	if err != nil {
		xapp.Logger.Error("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
	}
	xapp.Logger.Info("Subscription Request Received: RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", payloadSeqNum, payloadSeqNum)
	payload, err := e.GenerateRequestPayload(payloadSeqNum)
	if err != nil {
		xapp.Logger.Debug(err.Error())
		return
	}
	request.Payload = payload
	request.Mtype = 12011
	c <- *request
	return
}

func (e E2t) handleSubscriptionDeleteRequest(request *xapp.RMRParams) (err error) {
	payloadSeqNum, err := e.GetSubscriptionDeleteRequestSequenceNumber(request.Payload)
	if err != nil {
		xapp.Logger.Error("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
	}
	xapp.Logger.Info("Subscription Delete Request Received: RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", payloadSeqNum, payloadSeqNum)
	payload, err := e.GenerateDeletePayload(payloadSeqNum)
	if err != nil {
		xapp.Logger.Debug(err.Error())
		return
	}
	request.Payload = payload
	request.Mtype = 12021
	c <- *request
	return
}

func (e *E2t) Run() {
	for {
		message := <-c
		var payloadSeqNum uint16
		var err error
		if message.Mtype == 12011 {
			payloadSeqNum, err = e.GetSubscriptionResponseSequenceNumber(message.Payload)
		} else if message.Mtype == 12021 {
			payloadSeqNum, err = e.GetSubscriptionDeleteResponseSequenceNumber(message.Payload)
		} else {
			err = errors.New("OH MY GOD")
		}
		if err != nil {
			xapp.Logger.Debug("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		}
		xapp.Logger.Info("Sending Message: TYPE: %v | RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v)", message.Mtype, message.SubId, payloadSeqNum)
		xapp.Rmr.Send(&message, true)
	}
}

func main() {
	e2t := E2t{}
	go e2t.Run()
	xapp.Run(e2t)
}
