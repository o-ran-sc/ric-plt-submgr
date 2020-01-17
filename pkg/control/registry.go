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
	"sync"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Registry struct {
	mutex       sync.Mutex
	register    map[uint16]*Subscription
	subIds      []uint16
	rtmgrClient *RtmgrClient
}

// This method should run as a constructor
func (r *Registry) Initialize() {
	r.register = make(map[uint16]*Subscription)
	var i uint16
	for i = 0; i < 65535; i++ {
		r.subIds = append(r.subIds, i+1)
	}
}

// Reserves and returns the next free sequence number
func (r *Registry) AssignToSubscription(trans *Transaction, subReqMsg *e2ap.E2APSubscriptionRequest) (*Subscription, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if len(r.subIds) > 0 {
		sequenceNumber := r.subIds[0]
		r.subIds = r.subIds[1:]
		if _, ok := r.register[sequenceNumber]; ok == false {
			subs := &Subscription{
				registry: r,
				Seq:      sequenceNumber,
				Meid:     trans.Meid,
			}
			err := subs.AddEndpoint(trans.GetEndpoint())
			if err != nil {
				return nil, err
			}
			subs.SubReqMsg = subReqMsg

			r.register[sequenceNumber] = subs
			xapp.Logger.Debug("Registry: Create %s", subs.String())
			xapp.Logger.Debug("Registry: substable=%v", r.register)
			return subs, nil
		}
	}
	return nil, fmt.Errorf("Registry: Failed to reserves subscription")
}

func (r *Registry) GetSubscription(sn uint16) *Subscription {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.register[sn]; ok {
		return r.register[sn]
	}
	return nil
}

func (r *Registry) GetSubscriptionFirstMatch(ids []uint16) (*Subscription, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, id := range ids {
		if _, ok := r.register[id]; ok {
			return r.register[id], nil
		}
	}
	return nil, fmt.Errorf("No valid subscription found with ids %v", ids)
}

func (r *Registry) DelSubscription(sn uint16) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.register[sn]; ok {
		subs := r.register[sn]
		xapp.Logger.Debug("Registry: Delete %s", subs.String())
		r.subIds = append(r.subIds, sn)
		delete(r.register, sn)
		xapp.Logger.Debug("Registry: substable=%v", r.register)
		return true
	}
	return false
}
