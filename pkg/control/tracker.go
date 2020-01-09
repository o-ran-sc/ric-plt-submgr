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
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
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
	tracker           *Tracker // tracker instance
	Subs              *Subscription
	RmrEndpoint       RmrEndpoint
	Xid               string          // xapp xid in req
	OrigParams        *xapp.RMRParams // request orginal params
	RespReceived      bool
	ForwardRespToXapp bool
	mutex             sync.Mutex
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

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Tracker struct {
	transactionXappTable map[TransactionXappKey]*Transaction
	mutex                sync.Mutex
}

func (t *Tracker) Init() {
	t.transactionXappTable = make(map[TransactionXappKey]*Transaction)
}

func (t *Tracker) TrackTransaction(subs *Subscription, endpoint RmrEndpoint, params *xapp.RMRParams, respReceived bool, forwardRespToXapp bool) (*Transaction, error) {

	trans := &Transaction{
		tracker:           nil,
		Subs:              nil,
		RmrEndpoint:       endpoint,
		Xid:               params.Xid,
		OrigParams:        params,
		RespReceived:      respReceived,
		ForwardRespToXapp: forwardRespToXapp,
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	xappkey := TransactionXappKey{endpoint, params.Xid}
	if _, ok := t.transactionXappTable[xappkey]; ok {
		err := fmt.Errorf("Tracker: Similar transaction with xappkey %s is ongoing, transaction %s not created ", xappkey, trans)
		return nil, err
	}

	if subs.SetTransaction(trans) == false {
		othTrans := subs.GetTransaction()
		err := fmt.Errorf("Tracker: Subscription %s got already transaction ongoing: %s, transaction %s not created", subs, othTrans, trans)
		return nil, err
	}
	trans.Subs = subs
	if (trans.Subs.RmrEndpoint.Addr != trans.RmrEndpoint.Addr) || (trans.Subs.RmrEndpoint.Port != trans.RmrEndpoint.Port) {
		err := fmt.Errorf("Tracker: Subscription endpoint %s mismatch with trans: %s", subs, trans)
		trans.Subs.UnSetTransaction(nil)
		return nil, err
	}

	trans.tracker = t
	t.transactionXappTable[xappkey] = trans
	return trans, nil
}

func (t *Tracker) UnTrackTransaction(xappKey TransactionXappKey) (*Transaction, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if trans, ok2 := t.transactionXappTable[xappKey]; ok2 {
		delete(t.transactionXappTable, xappKey)
		return trans, nil
	}
	return nil, fmt.Errorf("Tracker: No record for xappkey %s", xappKey)
}
