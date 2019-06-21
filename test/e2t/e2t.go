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
	/* TODO: removed to being able to integrate with UEMGR
	submgr "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/control"
	*/
	"errors"
)

type E2t struct {
}

func (e E2t ) Consume(mtype, sub_id int, len int, payload []byte) (err error) {
	/* TODO: removed to being able to integrate with UEMGR
	asn1 := submgr.Asn1{}
	message, err := asn1.Decode(payload)
	if err != nil {
		xapp.Logger.Debug("E2T asn1Decoding failure due to "+ err.Error())
		return
	}
	*/
	xapp.Logger.Info("E2T Received Message with RMR Subsriprion ID: %v, Responding...", sub_id)
	err = e.subscriptionResponse(sub_id, payload)
	return
}

func (e E2t ) subscriptionResponse(sub_id int, payload []byte) (err error) {
	/* TODO: removed to being able to integrate with UEMGR
	asn1 := submgr.Asn1{}
	payload, err := asn1.Encode(submgr.RmrPayload{8, sub_id, "E2T: RCO Subscribed"})
	if err != nil {
		return
	}
	*/
	if !xapp.Rmr.Send(12011, sub_id, len(payload), payload) {
		err = errors.New("rmr.Send() failed")	
	}
	return
}

func main() {
	e2t := E2t{}
	xapp.Run(e2t)
}
