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

import (
	"bytes"
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strconv"
	"strings"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrDatagram struct {
	MessageType    int
	SubscriptionId uint16
	Payload        []byte
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type SubRouteInfo struct {
	Command Action
	Address string
	Port    uint16
	SubID   uint16
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrEndpoint struct {
	Addr string // xapp addr
	Port uint16 // xapp port
}

func (endpoint RmrEndpoint) String() string {
	return endpoint.Get()
}

func (endpoint *RmrEndpoint) GetAddr() string {
	return endpoint.Addr
}

func (endpoint *RmrEndpoint) GetPort() uint16 {
	return endpoint.Port
}

func (endpoint *RmrEndpoint) Get() string {
	return endpoint.Addr + ":" + strconv.FormatUint(uint64(endpoint.Port), 10)
}

func (endpoint *RmrEndpoint) Set(src string) bool {
	elems := strings.Split(src, ":")
	if len(elems) == 2 {
		srcAddr := elems[0]
		srcPort, err := strconv.ParseUint(elems[1], 10, 16)
		if err == nil {
			endpoint.Addr = srcAddr
			endpoint.Port = uint16(srcPort)
			return true
		}
	}
	return false
}

func NewRmrEndpoint(src string) *RmrEndpoint {
	ep := &RmrEndpoint{}
	if ep.Set(src) == false {
		return nil
	}
	return ep
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Action int

func (act Action) String() string {
	actions := [...]string{
		"CREATE",
		"MERGE",
		"NONE",
		"DELETE",
	}

	if act < CREATE || act > DELETE {
		return "UNKNOWN"
	}
	return actions[act]
}

//-----------------------------------------------------------------------------
// To add own method for rmrparams
//-----------------------------------------------------------------------------
type RMRParams struct {
	*xapp.RMRParams
}

func (params *RMRParams) String() string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "Src=%s Mtype=%s(%d) SubId=%v Xid=%s Meid=%v", params.Src, xapp.RicMessageTypeToName[params.Mtype], params.Mtype, params.SubId, params.Xid, params.Meid)
	return b.String()
}
