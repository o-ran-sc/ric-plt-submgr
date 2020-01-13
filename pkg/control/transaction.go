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
	"strconv"
	"sync"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type TransactionXappKey struct {
	RmrEndpoint
	Xid string // xapp xid in req
}

func (key *TransactionXappKey) String() string {
	return key.RmrEndpoint.String() + "/" + key.Xid
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Transaction struct {
	mutex             sync.Mutex
	tracker           *Tracker // tracker instance
	Subs              *Subscription
	RmrEndpoint       RmrEndpoint
	Mtype             int
	Xid               string     // xapp xid in req
	OrigParams        *RMRParams // request orginal params
	RespReceived      bool
	ForwardRespToXapp bool
}

func (t *Transaction) String() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	var subId string = "?"
	if t.Subs != nil {
		subId = strconv.FormatUint(uint64(t.Subs.Seq), 10)
	}
	return subId + "/" + t.RmrEndpoint.String() + "/" + t.Xid
}

func (t *Transaction) GetXid() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.Xid
}

func (t *Transaction) GetMtype() int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.Mtype
}

func (t *Transaction) GetSrc() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.RmrEndpoint.String()
}

func (t *Transaction) CheckResponseReceived() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.RespReceived == false {
		t.RespReceived = true
		return false
	}
	return true
}

func (t *Transaction) RetryTransaction() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.RespReceived = false
}

func (t *Transaction) Release() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.Subs != nil {
		t.Subs.UnSetTransaction(t)
	}
	if t.tracker != nil {
		xappkey := TransactionXappKey{t.RmrEndpoint, t.Xid}
		t.tracker.UnTrackTransaction(xappkey)
	}
	t.Subs = nil
	t.tracker = nil
}
