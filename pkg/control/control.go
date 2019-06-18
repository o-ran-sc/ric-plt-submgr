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
  "gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
  "errors"
  "strconv"
)

type Control struct {
  e2ap *E2ap
  registry *Registry
}

func NewControl() Control {
  return Control{new(E2ap),new(Registry)}
}

func (c *Control) Run() {
  xapp.Run(c)
}

func (c *Control) Consume(mtype, sub_id int, len int, payload []byte) (err error) {
  switch mtype {
  case C.RIC_SUB_REQ:
    err = c.handleSubscriptionRequest(&RmrDatagram{mtype, sub_id, payload})
  case C.RIC_SUB_RESP:
    err = c.handleSubscriptionResponse(&RmrDatagram{mtype, sub_id, payload})
  default:
    err = errors.New("Message Type "+strconv.Itoa(mtype)+" discarded")
  }
  return
}

func (c *Control) rmrSend(datagram *RmrDatagram) (err error) {
  if !xapp.Rmr.Send(datagram.MessageType, datagram.SubscriptionId, len(datagram.Payload), datagram.Payload) {
    err = errors.New("rmr.Send() failed")
  }
  return
}

func (c *Control) handleSubscriptionRequest(datagram *RmrDatagram) ( err error) {
  content, err := c.e2ap.GetPayloadContent(datagram.Payload)
  xapp.Logger.Info("Subscription Request received: %v", content)
  new_sub_id := c.registry.GetSubscriptionId()
  payload, err := c.e2ap.SetSubscriptionSequenceNumber(datagram.Payload, new_sub_id)
  if err != nil {
    xapp.Logger.Error("Unable to set Subscription Sequence Number in Payload due to: "+ err.Error())
    return
  }
  xapp.Logger.Info("New Subscription Accepted, Forwarding to E2T")
  c.rmrSend(&RmrDatagram{C.RIC_SUB_REQ , new_sub_id, payload})
  return
}

func (c *Control) handleSubscriptionResponse(datagram *RmrDatagram) ( err error) {
  content, err := c.e2ap.GetPayloadContent(datagram.Payload)
  xapp.Logger.Info("Subscription Response received: %v", content)
  return
}