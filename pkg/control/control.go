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

package control

/*
#include <rmr/RIC_message_types.h>

#cgo CFLAGS: -I../
#cgo LDFLAGS: -lrmr_nng -lnng
*/
import "C"

import (
	"errors"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/spf13/viper"
	"math/rand"
	"strconv"
	"time"
)

type Control struct {
	e2ap     *E2ap
	registry *Registry
}

var SEEDSN uint16

func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("submgr")
	viper.AllowEmptyEnv(true)
	SEEDSN = uint16(viper.GetInt("seed_sn"))
	if SEEDSN == 0 {
		rand.Seed(time.Now().UnixNano())
		SEEDSN = uint16(rand.Intn(65535))
	}
	if SEEDSN > 65535 {
		SEEDSN = 0
	}
	xapp.Logger.Info("SUBMGR: Initial Sequence Number: %v", SEEDSN)
}

func NewControl() Control {
	registry := new(Registry)
	registry.Initialize(SEEDSN)
	return Control{new(E2ap), registry}
}

func (c *Control) Run() {
	xapp.Run(c)
}

func (c *Control) Consume(mtype, sub_id int, len int, payload []byte) (err error) {
	switch mtype {
	case C.RIC_SUB_REQ:
		err = c.handleSubscriptionRequest(&RmrDatagram{mtype, uint16(sub_id), payload})
	case C.RIC_SUB_RESP:
		err = c.handleSubscriptionResponse(&RmrDatagram{mtype, uint16(sub_id), payload})
	default:
		err = errors.New("Message Type " + strconv.Itoa(mtype) + " is discarded")
	}
	return
}

func (c *Control) rmrSend(datagram *RmrDatagram) (err error) {
	if !xapp.Rmr.Send(datagram.MessageType, int(datagram.SubscriptionId), len(datagram.Payload), datagram.Payload) {
		err = errors.New("rmr.Send() failed")
	}
	return
}

func (c *Control) handleSubscriptionRequest(datagram *RmrDatagram) (err error) {
	payload_seq_num, err := c.e2ap.GetSubscriptionRequestSequenceNumber(datagram.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Subscription Request Received. RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", datagram.SubscriptionId, payload_seq_num)
	new_sub_id := c.registry.ReserveSequenceNumber()
	payload, err := c.e2ap.SetSubscriptionRequestSequenceNumber(datagram.Payload, new_sub_id)
	if err != nil {
		err = errors.New("Unable to set Subscription Sequence Number in Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Generated ID: %v. Forwarding to E2 Termination...", int(new_sub_id))
	c.rmrSend(&RmrDatagram{C.RIC_SUB_REQ, new_sub_id, payload})
	return
}

func (c *Control) handleSubscriptionResponse(datagram *RmrDatagram) (err error) {
	payload_seq_num, err := c.e2ap.GetSubscriptionResponseSequenceNumber(datagram.Payload)
	if err != nil {
		err = errors.New("Unable to get Subscription Sequence Number from Payload due to: " + err.Error())
		return
	}
	xapp.Logger.Info("Subscription Response Received. RMR SUBSCRIPTION_ID: %v | PAYLOAD SEQUENCE_NUMBER: %v", datagram.SubscriptionId, payload_seq_num)
	if !c.registry.IsValidSequenceNumber(payload_seq_num) {
		err = errors.New("Unknown Subscription ID: " + strconv.Itoa(int(payload_seq_num)) + " in Subscritpion Response. Message discarded.")
		return
	}
	c.registry.setSubscriptionToConfirmed(payload_seq_num)
	xapp.Logger.Info("Subscription Response Registered. Forwarding to Requestor...")
	c.rmrSend(&RmrDatagram{C.RIC_SUB_RESP, payload_seq_num, datagram.Payload})
	return
}
