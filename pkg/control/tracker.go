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
)

/*
Implements a record of ongoing transactions and helper functions to CRUD the records.
*/
type Tracker struct {
	transaction_table map[Transaction_key]Transaction
}

func (t *Tracker) Init() {
	t.transaction_table = make(map[Transaction_key]Transaction)
}

/*
Checks if a tranascation with similar type has been ongoing. If not then creates one.
Returns error if there is similar transatcion ongoing.
*/
func (t *Tracker) Track_transaction(key Transaction_key, xact Transaction) error{
	if _, ok := t.transaction_table[key]; ok {
		// TODO: Implement merge related check here. If the key is same but the value is different.
		err := fmt.Errorf("Transaction tracker: Similar transaction with sub id %d and type %s is ongoing", key.SubID, key.trans_type )
		return err
	}
	t.transaction_table[key] = xact
	return nil
}

/*
Retreives the transaction table entry for the given request.
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) Update_transaction(SubID uint16, trans_type Action, xact Transaction) error{
	key := Transaction_key{SubID, trans_type}
	if _, ok := t.transaction_table[key]; ok {
		// TODO: Implement merge related check here. If the key is same but the value is different.
		err := fmt.Errorf("Transaction tracker: Similar transaction with sub id %d and type %s is ongoing", key.SubID, key.trans_type )
		return err
	}
	t.transaction_table[key] = xact
	return nil
}

/*
Retreives the transaction table entry for the given request.
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) Retrive_transaction(subID uint16, act Action) (Transaction, error){
	key := Transaction_key{subID, act}
	var xact Transaction
	if xact, ok := t.transaction_table[key]; ok {
		return xact, nil
	}
	err := fmt.Errorf("Tranaction record for Subscription ID %d and action %s does not exist", subID, act)
	return xact, err
}

/*
Deletes the transaction table entry for the given request and returns the deleted xapp's address and port for reference.
Returns error in case the transaction cannot be found.
*/
func (t *Tracker) complete_transaction(subID uint16, act Action) (Transaction, error){
	key := Transaction_key{subID, act}
	var empty_transaction Transaction
	if xact, ok := t.transaction_table[key]; ok {
		delete(t.transaction_table, key)
		return xact, nil
	}
	err := fmt.Errorf("Tranaction record for Subscription ID %d and action %s does not exist", subID, act)
	return empty_transaction, err
}
