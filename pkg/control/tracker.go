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

var trackerMutex = &sync.Mutex{}

/*
Implements a record of ongoing transactions and helper functions to CRUD the records.
*/
type Tracker struct {
	transactionTable map[TransactionKey]Transaction
}

func (t *Tracker) Init() {
	t.transactionTable = make(map[TransactionKey]Transaction)
}

/*
Checks if a tranascation with similar type has been ongoing. If not then creates one.
Returns error if there is similar transatcion ongoing.
*/
func (t *Tracker) TrackTransaction(key TransactionKey, xact Transaction) error {
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	if _, ok := t.transactionTable[key]; ok {
		// TODO: Implement merge related check here. If the key is same but the value is different.
		err := fmt.Errorf("transaction tracker: Similar transaction with sub id %d and type %s is ongoing", key.SubID, key.transType)
		return err
	}
	t.transactionTable[key] = xact
	return nil
}

/*
Retreives the transaction table entry for the given request.
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) UpdateTransaction(SubID uint16, transType Action, xact Transaction) error {
	key := TransactionKey{SubID, transType}
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	if _, ok := t.transactionTable[key]; ok {
		// TODO: Implement merge related check here. If the key is same but the value is different.
		err := fmt.Errorf("transaction tracker: Similar transaction with sub id %d and type %v is ongoing", key.SubID, key.transType)
		return err
	}
	t.transactionTable[key] = xact
	return nil
}

/*
Retreives the transaction table entry for the given request.
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) RetriveTransaction(subID uint16, act Action) (Transaction, error) {
	key := TransactionKey{subID, act}
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	var xact Transaction
	if xact, ok := t.transactionTable[key]; ok {
		return xact, nil
	}
	err := fmt.Errorf("transaction record for Subscription ID %d and action %s does not exist", subID, act)
	return xact, err
}

/*
Deletes the transaction table entry for the given request and returns the deleted xapp's address and port for reference.
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) completeTransaction(subID uint16, act Action) (Transaction, error) {
	key := TransactionKey{subID, act}
	var emptyTransaction Transaction
	trackerMutex.Lock()
	defer trackerMutex.Unlock()
	if xact, ok := t.transactionTable[key]; ok {
		delete(t.transactionTable, key)
		return xact, nil
	}
	err := fmt.Errorf("transaction record for Subscription ID %d and action %s does not exist", subID, act)
	return emptyTransaction, err
}
