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

func (t *Tracker) TrackTransaction(
	endpoint *RmrEndpoint,
	xid string,
	meid *xapp.RMRMeid,
	respReceived bool,
	forwardRespToXapp bool) (*Transaction, error) {

	if endpoint == nil {
		err := fmt.Errorf("Tracker: No valid endpoint given")
		return nil, err
	}

	trans := &Transaction{
		tracker:           nil,
		Subs:              nil,
		RmrEndpoint:       *endpoint,
		Xid:               xid,
		Meid:              meid,
		RespReceived:      respReceived,
		ForwardRespToXapp: forwardRespToXapp,
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()

	xappkey := TransactionXappKey{*endpoint, xid}
	if othtrans, ok := t.transactionXappTable[xappkey]; ok {
		err := fmt.Errorf("Tracker: %s is ongoing, %s not created ", othtrans, trans)
		return nil, err
	}

	trans.tracker = t
	t.transactionXappTable[xappkey] = trans
	xapp.Logger.Info("Tracker: Create %s", trans.String())
	xapp.Logger.Debug("Tracker: transtable=%v", t.transactionXappTable)
	return trans, nil
}

func (t *Tracker) UnTrackTransaction(xappKey TransactionXappKey) (*Transaction, error) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if trans, ok2 := t.transactionXappTable[xappKey]; ok2 {
		xapp.Logger.Info("Tracker: Delete %s", trans.String())
		delete(t.transactionXappTable, xappKey)
		xapp.Logger.Debug("Tracker: transtable=%v", t.transactionXappTable)
		return trans, nil
	}
	return nil, fmt.Errorf("Tracker: No record %s", xappKey)
}
