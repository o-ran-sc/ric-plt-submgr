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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/packer"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strconv"
	"sync"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type TransactionBase struct {
	mutex     sync.Mutex         //
	Seq       uint64             //
	tracker   *Tracker           //tracker instance
	Meid      *xapp.RMRMeid      //meid transaction related
	Mtype     int                //Encoded message type to be send
	Payload   *packer.PackedData //Encoded message to be send
	EventChan chan interface{}
}

func (t *TransactionBase) SendEvent(event interface{}, waittime time.Duration) (bool, bool) {
	if waittime > 0 {
		select {
		case t.EventChan <- event:
			return true, false
		case <-time.After(waittime):
			return false, true
		}
		return false, false
	}
	t.EventChan <- event
	return true, false
}

func (t *TransactionBase) WaitEvent(waittime time.Duration) (interface{}, bool) {
	if waittime > 0 {
		select {
		case event := <-t.EventChan:
			return event, false
		case <-time.After(waittime):
			return nil, true
		}
	}
	event := <-t.EventChan
	return event, false
}

func (t *TransactionBase) GetMtype() int {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.Mtype
}

func (t *TransactionBase) GetMeid() *xapp.RMRMeid {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.Meid != nil {
		return t.Meid
	}
	return nil
}

func (t *TransactionBase) GetPayload() *packer.PackedData {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	return t.Payload
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type TransactionXappKey struct {
	RmrEndpoint
	Xid string // xapp xid in req
}

func (key *TransactionXappKey) String() string {
	return "transkey(" + key.RmrEndpoint.String() + "/" + key.Xid + ")"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Transaction struct {
	TransactionBase                     //
	XappKey         *TransactionXappKey //
}

func (t *Transaction) String() string {
	var transkey string = "transkey(N/A)"
	if t.XappKey != nil {
		transkey = t.XappKey.String()
	}
	return "trans(" + strconv.FormatUint(uint64(t.Seq), 10) + "/" + t.Meid.RanName + "/" + transkey + ")"
}

func (t *Transaction) GetEndpoint() *RmrEndpoint {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.XappKey != nil {
		return &t.XappKey.RmrEndpoint
	}
	return nil
}

func (t *Transaction) GetXid() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.XappKey != nil {
		return t.XappKey.Xid
	}
	return ""
}

func (t *Transaction) GetSrc() string {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.XappKey != nil {
		return t.XappKey.RmrEndpoint.String()
	}
	return ""
}

func (t *Transaction) Release() {
	t.mutex.Lock()
	xapp.Logger.Debug("Transaction: Release %s", t.String())
	tracker := t.tracker
	xappkey := t.XappKey
	t.tracker = nil
	t.mutex.Unlock()

	if tracker != nil && xappkey != nil {
		tracker.UnTrackTransaction(*xappkey)
	}
}
