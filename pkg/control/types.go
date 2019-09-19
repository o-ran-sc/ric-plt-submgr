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

type RmrDatagram struct {
	MessageType    int
	SubscriptionId uint16
	Payload        []byte
}

type subRouteInfo struct {
	Command  Action
	Address  string
	Port     uint16
	SubID    uint16
}

type Action int

type Transaction_key struct {
	SubID      uint16
	trans_type Action
}

type Transaction struct {
//	Xapp_address          string
	Xapp_instance_address string
	Xapp_port             uint16
	Ric_sub_req           []byte
}
