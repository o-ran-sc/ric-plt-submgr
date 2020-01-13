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
	mutex  sync.Mutex
	Seq    uint16
	Active bool
	//
	Meid        *xapp.RMRMeid
	RmrEndpoint // xapp endpoint. Now only one xapp can have relation to single subscription. To be changed in merge
	Trans       *Transaction
}

func (s *Subscription) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return strconv.FormatUint(uint64(s.Seq), 10) + "/" + s.RmrEndpoint.String() + "/" + s.Meid.RanName
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

func (s *Subscription) SetTransaction(trans *Transaction) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	subString := strconv.FormatUint(uint64(s.Seq), 10) + "/" + s.RmrEndpoint.String() + "/" + s.Meid.RanName

	if (s.RmrEndpoint.Addr != trans.RmrEndpoint.Addr) || (s.RmrEndpoint.Port != trans.RmrEndpoint.Port) {
		return fmt.Errorf("Subscription: %s endpoint mismatch with trans: %s", subString, trans)
	}
	if s.Trans != nil {
		return fmt.Errorf("Subscription: %s trans %s exist, can not register %s", subString, s.Trans, trans)
	}
	trans.Subs = s
	s.Trans = trans
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

func (s *Subscription) UpdateRoute(act Action, rtmgrClient *RtmgrClient) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	xapp.Logger.Info("Subscription: Starting routing manager route add. SubId: %d, RmrEndpoint: %s", s.Seq, s.RmrEndpoint)
	subRouteAction := SubRouteInfo{act, s.RmrEndpoint.Addr, s.RmrEndpoint.Port, s.Seq}
	err := rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	if err != nil {
		return fmt.Errorf("Subscription: Failed to add route. SubId: %d, RmrEndpoint: %s", s.Seq, s.RmrEndpoint)
	}
	return nil
}
