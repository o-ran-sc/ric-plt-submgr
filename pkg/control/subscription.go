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
type Subscription struct {
	mutex    sync.Mutex
	registry *Registry
	Seq      uint16
	Active   bool
	//
	Meid   *xapp.RMRMeid
	EpList RmrEndpointList
	Trans  *Transaction
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

func (s *Subscription) Confirmed() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Active = true
}

func (s *Subscription) UnConfirmed() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.Active = false
}

func (s *Subscription) IsConfirmed() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.Active
}

func (s *Subscription) IsEndpoint(ep *RmrEndpoint) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.EpList.HasEndpoint(ep)
}

func (s *Subscription) SetTransaction(trans *Transaction) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.Trans != nil {
		return fmt.Errorf("subs(%s) trans(%s) exist, can not register trans(%s)", s.stringImpl(), s.Trans, trans)
	}
	trans.Subs = s
	s.Trans = trans

	if len(s.EpList.Endpoints) == 0 {
		s.EpList.Endpoints = append(s.EpList.Endpoints, trans.RmrEndpoint)
		return s.updateRouteImpl(CREATE)
	} else if s.EpList.HasEndpoint(&trans.RmrEndpoint) == false {
		s.EpList.Endpoints = append(s.EpList.Endpoints, trans.RmrEndpoint)
		return s.updateRouteImpl(MERGE)
	}
	return nil
}

func (s *Subscription) UnSetTransaction(trans *Transaction) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if trans == nil || trans == s.Trans {
		s.Trans = nil
		return true
	}
	return false
}

func (s *Subscription) GetTransaction() *Transaction {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.Trans
}

func (s *Subscription) updateRouteImpl(act Action) error {
	subRouteAction := SubRouteInfo{act, s.EpList, s.Seq}
	err := s.registry.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		return fmt.Errorf("subs(%s) %s", s.stringImpl(), err.Error())
	}
	return nil
}

func (s *Subscription) UpdateRoute(act Action) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.updateRouteImpl(act)
}

func (s *Subscription) Release() {
	s.registry.DelSubscription(s.Seq)
	err := s.UpdateRoute(DELETE)
	if err != nil {
		xapp.Logger.Error("%s", err.Error())
	}
}
