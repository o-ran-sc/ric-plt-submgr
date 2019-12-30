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
	SubID     uint16
	transType Action
}

type TransactionXappKey struct {
	XappInstanceAddress string
	XappPort            uint16
	Xid                 string
}

type Transaction struct {
	XappInstanceAddress string
	XappPort            uint16
	OrigParams          *xapp.RMRParams
}

type transactionData struct {
	key     TransactionKey
	xappkey TransactionXappKey
	xact    Transaction
}

/*
Implements a record of ongoing transactions and helper functions to CRUD the records.
*/
type Tracker struct {
	transactionTable     map[TransactionKey]*transactionData
	transactionXappTable map[TransactionXappKey]*transactionData
	mutex                sync.Mutex
}

func (t *Tracker) Init() {
	t.transactionTable = make(map[TransactionKey]*transactionData)
	t.transactionXappTable = make(map[TransactionXappKey]*transactionData)
}

/*
Checks if a tranascation with similar type has been ongoing. If not then creates one.
Returns error if there is similar transatcion ongoing.
*/
func (t *Tracker) TrackTransaction(subID uint16, act Action, xact Transaction) error {
	key := TransactionKey{subID, act}
	xappkey := TransactionXappKey{xact.XappInstanceAddress, xact.XappPort, xact.OrigParams.Xid}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.transactionTable[key]; ok {
		// TODO: Implement merge related check here. If the key is same but the value is different.
		err := fmt.Errorf("transaction tracker: Similar transaction with sub id %d and type %s is ongoing", key.SubID, key.transType)
		return err
	}
	if _, ok := t.transactionXappTable[xappkey]; ok {
		// TODO: Implement merge related check here. If the key is same but the value is different.
		err := fmt.Errorf("transaction tracker: Similar transaction with xapp key %v is ongoing", xappkey)
		return err
	}
	xdata := &transactionData{key, xappkey, xact}
	t.transactionTable[key] = xdata
	t.transactionXappTable[xappkey] = xdata
	return nil
}

/*
Retreives the transaction table entry for the given request.
Returns error in case the transaction cannot be found.
func (t *Tracker) UpdateTransaction(subID uint16, act Action, xact Transaction) error {
	key := TransactionKey{subID, act}
	xappkey := TransactionXappKey{xact.XappInstanceAddress, xact.XappPort, xact.OrigParams.Xid}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	if _, ok := t.transactionTable[key]; ok {
		// TODO: Implement merge related check here. If the key is same but the value is different.
		err := fmt.Errorf("transaction tracker: Similar transaction with sub id %d and type %v is ongoing", key.SubID, key.transType)
		return err
	}
	xdata := &transactionData{key, xappkey, xact}
	t.transactionTable[key] = xdata
	return nil
}
*/

/*
Retreives the transaction table entry for the given request.
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) RetriveTransaction(subID uint16, act Action) (Transaction, error) {
	key := TransactionKey{subID, act}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if xdata, ok := t.transactionTable[key]; ok {
		return xdata.xact, nil
	}
	err := fmt.Errorf("transaction record for Subscription ID %d and action %s does not exist", subID, act)
	return Transaction{}, err
}

/*
Deletes the transaction table entry for the given request and returns the deleted xapp's address and port for reference.
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) completeTransaction(subID uint16, act Action) (Transaction, error) {
	key := TransactionKey{subID, act}
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if xdata, ok1 := t.transactionTable[key]; ok1 {
		if _, ok2 := t.transactionXappTable[xdata.xappkey]; ok2 {
			delete(t.transactionXappTable, xdata.xappkey)
		}
		delete(t.transactionTable, key)
		return xdata.xact, nil
	}
	err := fmt.Errorf("transaction record for Subscription ID %d and action %s does not exist", subID, act)
	return Transaction{}, err
}
