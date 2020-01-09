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
	RmrEndpoint // xapp endpoint
	Trans       *Transaction
}

func (s *Subscription) String() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return strconv.FormatUint(uint64(s.Seq), 10) + "/" + s.RmrEndpoint.String() + "/" + s.Meid.RanName
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

func (s *Subscription) SetTransaction(trans *Transaction) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.Trans == nil {
		s.Trans = trans
		return true
	}
	return false
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

func (s *Subscription) SubRouteInfo(act Action) SubRouteInfo {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return SubRouteInfo{act, s.RmrEndpoint.Addr, s.RmrEndpoint.Port, s.Seq}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Registry struct {
	register map[uint16]*Subscription
	counter  uint16
	mutex    sync.Mutex
}

// This method should run as a constructor
func (r *Registry) Initialize(seedsn uint16) {
	r.register = make(map[uint16]*Subscription)
	r.counter = seedsn
}

// Reserves and returns the next free sequence number
func (r *Registry) ReserveSubscription(endPoint RmrEndpoint, meid *xapp.RMRMeid) *Subscription {
	// Check is current SequenceNumber valid
	// Allocate next SequenceNumber value and retry N times
	r.mutex.Lock()
	defer r.mutex.Unlock()
	var subs *Subscription = nil
	var retrytimes uint16 = 1000
	for ; subs == nil && retrytimes > 0; retrytimes-- {
		sequenceNumber := r.counter
		if r.counter == 65535 {
			r.counter = 0
		} else {
			r.counter++
		}
		if _, ok := r.register[sequenceNumber]; ok == false {
			r.register[sequenceNumber] = &Subscription{
				Seq:         sequenceNumber,
				Active:      false,
				RmrEndpoint: endPoint,
				Meid:        meid,
				Trans:       nil,
			}
			return r.register[sequenceNumber]
		}
	}
	return nil
}

func (r *Registry) GetSubscription(sn uint16) *Subscription {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	xapp.Logger.Debug("Registry map: %v", r.register)
	if _, ok := r.register[sn]; ok {
		return r.register[sn]
	}
	return nil
}

//This function releases the given id as unused in the register
func (r *Registry) releaseSequenceNumber(sn uint16) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.register[sn]; ok {
		delete(r.register, sn)
		return true
	} else {
		return false
	}
}
