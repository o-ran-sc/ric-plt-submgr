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

type TransactionKey struct {
	SubID     uint16 // subscription id / sequence number
	TransType Action // action ongoing (CREATE/DELETE etc)
}

type TransactionXappKey struct {
	Addr string // xapp addr
	Port uint16 // xapp port
	Xid  string // xapp xid in req
}

type Transaction struct {
	tracker           *Tracker           // tracker instance
	Key               TransactionKey     // action key
	Xappkey           TransactionXappKey // transaction key
	OrigParams        *xapp.RMRParams    // request orginal params
	RespReceived      bool
	ForwardRespToXapp bool
}

func (t *Transaction) SubRouteInfo() SubRouteInfo {
	return SubRouteInfo{t.Key.TransType, t.Xappkey.Addr, t.Xappkey.Port, t.Key.SubID}
}

/*
Implements a record of ongoing transactions and helper functions to CRUD the records.
*/
type Tracker struct {
	transactionTable     map[TransactionKey]*Transaction
	transactionXappTable map[TransactionXappKey]*Transaction
	mutex                sync.Mutex
}

func (t *Tracker) Init() {
	t.transactionTable = make(map[TransactionKey]*Transaction)
	t.transactionXappTable = make(map[TransactionXappKey]*Transaction)
}

/*
Checks if a tranascation with similar type has been ongoing. If not then creates one.
Returns error if there is similar transatcion ongoing.
*/
func (t *Tracker) TrackTransaction(subID uint16, act Action, addr string, port uint16, params *xapp.RMRParams, respReceived bool, forwardRespToXapp bool) (*Transaction, error) {
	key := TransactionKey{subID, act}
	xappkey := TransactionXappKey{addr, port, params.Xid}
	trans := &Transaction{t, key, xappkey, params, respReceived, forwardRespToXapp}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.transactionTable[key]; ok {
		// TODO: Implement merge related check here. If the key is same but the value is different.
		err := fmt.Errorf("transaction tracker: Similar transaction with sub id %d and type %s is ongoing", key.SubID, key.TransType)
		return nil, err
	}
	if _, ok := t.transactionXappTable[xappkey]; ok {
		// TODO: Implement merge related check here. If the key is same but the value is different.
		err := fmt.Errorf("transaction tracker: Similar transaction with xapp key %v is ongoing", xappkey)
		return nil, err
	}
	t.transactionTable[key] = trans
	t.transactionXappTable[xappkey] = trans
	return trans, nil
}

/*
Retreives the transaction table entry for the given request. Controls that only one response is sent to xapp.
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) RetriveTransaction(subID uint16, act Action) (*Transaction, error) {
	key := TransactionKey{subID, act}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if trans, ok := t.transactionTable[key]; ok {
		if trans.RespReceived == false {
			trans.RespReceived = true
			t.transactionTable[key] = trans
			// This is used to control that only one response action (success response, failure or timer) is excecuted for the transaction
			trans.RespReceived = false
		}
		return trans, nil
	}
	err := fmt.Errorf("transaction record for Subscription ID %d and action %s does not exist", subID, act)
	return nil, err
}

/*
Deletes the transaction table entry for the given request and returns the deleted xapp's address and port for reference.
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) completeTransaction(subID uint16, act Action) (*Transaction, error) {
	key := TransactionKey{subID, act}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if trans, ok1 := t.transactionTable[key]; ok1 {
		if _, ok2 := t.transactionXappTable[trans.Xappkey]; ok2 {
			delete(t.transactionXappTable, trans.Xappkey)
		}
		delete(t.transactionTable, key)
		return trans, nil
	}
	err := fmt.Errorf("transaction record for Subscription ID %d and action %s does not exist", subID, act)
	return nil, err
}

/*
Makes possible to receive response to retransmitted request to BTS
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) RetryTransaction(subID uint16, act Action) (*Transaction, error) {
	key := TransactionKey{subID, act}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if trans, ok := t.transactionTable[key]; ok {
		trans.RespReceived = false
		return trans, nil
	}
	err := fmt.Errorf("transaction record for Subscription ID %d and action %s does not exist", subID, act)
	return nil, err
}
