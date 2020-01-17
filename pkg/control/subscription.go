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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strconv"
	"sync"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Subscription struct {
	mutex      sync.Mutex      // Lock
	registry   *Registry       // Registry
	Seq        uint16          // SubsId
	Meid       *xapp.RMRMeid   // Meid/ RanName
	EpList     RmrEndpointList // Endpoints
	DelEpList  RmrEndpointList // Endpoints
	DelSeq     uint64
	TransLock  sync.Mutex                     // Lock transactions, only one executed per time for subs
	TheTrans   *Transaction                   // Ongoing transaction from xapp
	SubReqMsg  *e2ap.E2APSubscriptionRequest  // Subscription information
	SubRespMsg *e2ap.E2APSubscriptionResponse // Subscription information
}

func (s *Subscription) stringImpl() string {
	return "subs(" + strconv.FormatUint(uint64(s.Seq), 10) + "/" + s.Meid.RanName + "/" + s.EpList.String() + ")"
}

func (s *Subscription) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.stringImpl()
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

func (s *Subscription) AddEndpoint(ep *RmrEndpoint) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if ep == nil {
		return fmt.Errorf("AddEndpoint no endpoint given")
	}
	if s.EpList.AddEndpoint(ep) {
		s.DelEpList.DelEndpoint(ep)
		if s.EpList.Size() == 1 {
			return s.updateRouteImpl(CREATE)
		}
		return s.updateRouteImpl(MERGE)
	}
	return nil
}

func (s *Subscription) DelEndpoint(ep *RmrEndpoint) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var err error
	if ep == nil {
		return fmt.Errorf("DelEndpoint no endpoint given")
	}
	if s.EpList.HasEndpoint(ep) == false {
		return fmt.Errorf("DelEndpoint endpoint not found")
	}
	if s.DelEpList.HasEndpoint(ep) == true {
		return fmt.Errorf("DelEndpoint endpoint already under del")
	}
	s.DelEpList.AddEndpoint(ep)
	go s.CleanCheck()
	return err
}

func (s *Subscription) CleanCheck() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.DelSeq++
	// Only one clean ongoing
	if s.DelSeq > 1 {
		return
	}
	var currSeq uint64 = 0
	// Make sure that routes to be deleted
	// are not deleted too fast
	for currSeq < s.DelSeq {
		currSeq = s.DelSeq
		s.mutex.Unlock()
		time.Sleep(5 * time.Second)
		s.mutex.Lock()
	}
	xapp.Logger.Info("DelEndpoint: delete cleaning %s", s.stringImpl())
	if s.EpList.Size() <= s.DelEpList.Size() {
		s.updateRouteImpl(DELETE)
		go s.registry.DelSubscription(s.Seq)
	} else if s.EpList.DelEndpoints(&s.DelEpList) {
		s.updateRouteImpl(MERGE)
	}
	s.DelSeq = 0

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

func (s *Subscription) updateRouteImpl(act Action) error {
	subRouteAction := SubRouteInfo{act, s.EpList, s.Seq}
	err := s.registry.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		return fmt.Errorf("%s %s", s.stringImpl(), err.Error())
	}
	return nil
}
