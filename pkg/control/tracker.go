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
	"sync"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Tracker struct {
	mutex                sync.Mutex
	transactionXappTable map[TransactionXappKey]*Transaction
}

func (t *Tracker) Init() {
	t.transactionXappTable = make(map[TransactionXappKey]*Transaction)
}

func (t *Tracker) TrackTransaction(subs *Subscription, endpoint RmrEndpoint, params *RMRParams, respReceived bool, forwardRespToXapp bool) (*Transaction, error) {

	trans := &Transaction{
		tracker:           nil,
		Subs:              nil,
		RmrEndpoint:       endpoint,
		Mtype:             params.Mtype,
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

	err := subs.SetTransaction(trans)
	if err != nil {
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
