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
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	submgr "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/control"
	"time"
	"os"
)

type Rco struct {
}

var	c chan submgr.RmrDatagram = make(chan submgr.RmrDatagram, 1)

func (m Rco) Consume(mtype, sub_id int, len int, payload []byte) (err error) {
	return
}

func (r *Rco) send(datagram submgr.RmrDatagram) {
	xapp.Rmr.Send(datagram.MessageType, datagram.SubscriptionId, len(datagram.Payload), datagram.Payload)
}

func (r *Rco) Run() {
	for {
		message := <- c
		xapp.Logger.Info("RCO Message - Type=%v SubID=%v", message.MessageType, message.SubscriptionId)
		r.send(message)
	}
}

func main() {
	rco := Rco{}
	go xapp.Rmr.Start(rco)
	go rco.Run()
	asn1 := submgr.Asn1{}
	message, err := asn1.Encode(submgr.RmrPayload{8, 1111, "RCO: Subscription Request"})
	if err != nil {
		xapp.Logger.Debug(err.Error())
		os.Exit(1)
	}
	doSubscribe := true
	for {
		time.Sleep(2 * time.Second)
		if doSubscribe {
			c <- submgr.RmrDatagram{12010, 9999, message}
			doSubscribe = false
		} else {
			c <- submgr.RmrDatagram{10000, 9999, make([]byte,0)}
			doSubscribe = true
		}
	}
}