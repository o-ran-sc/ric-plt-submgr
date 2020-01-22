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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strconv"
	"sync"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Subscription struct {
	mutex      sync.Mutex                     // Lock
	registry   *Registry                      // Registry
	Seq        uint16                         // SubsId
	Meid       *xapp.RMRMeid                  // Meid/ RanName
	EpList     RmrEndpointList                // Endpoints
	TransLock  sync.Mutex                     // Lock transactions, only one executed per time for subs
	TheTrans   *Transaction                   // Ongoing transaction from xapp
	SubReqMsg  *e2ap.E2APSubscriptionRequest  // Subscription information
	SubRespMsg *e2ap.E2APSubscriptionResponse // Subscription information
}

func (s *Subscription) String() string {
	return "subs(" + strconv.FormatUint(uint64(s.Seq), 10) + "/" + s.Meid.RanName + "/" + s.EpList.String() + ")"
}

func (s *Subscription) GetSubId() uint16 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.Seq
}

func (s *Subscription) GetMeid() *xapp.RMRMeid {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.Meid != nil {
		return s.Meid
	}
	return nil
}

func (s *Subscription) IsTransactionReserved() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.TheTrans != nil {
		return true
	}
	return false

}

func (s *Subscription) GetTransaction() *Transaction {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.TheTrans
}

func (s *Subscription) WaitTransactionTurn(trans *Transaction) {
	s.TransLock.Lock()
	s.mutex.Lock()
	s.TheTrans = trans
	s.mutex.Unlock()
}

func (s *Subscription) ReleaseTransactionTurn(trans *Transaction) {
	s.mutex.Lock()
	if trans != nil && trans == s.TheTrans {
		s.TheTrans = nil
	}
	s.mutex.Unlock()
	s.TransLock.Unlock()
}
