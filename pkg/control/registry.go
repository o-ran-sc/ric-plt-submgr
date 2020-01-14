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

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Registry struct {
	mutex       sync.Mutex
	register    map[uint16]*Subscription
	counter     uint16
	rtmgrClient *RtmgrClient
}

// This method should run as a constructor
func (r *Registry) Initialize(seedsn uint16) {
	r.register = make(map[uint16]*Subscription)
	r.counter = seedsn
}

// Reserves and returns the next free sequence number
func (r *Registry) ReserveSubscription(endPoint *RmrEndpoint, meid *xapp.RMRMeid) (*Subscription, error) {
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
			subs := &Subscription{
				Seq:         sequenceNumber,
				Active:      false,
				RmrEndpoint: *endPoint,
				Meid:        meid,
				Trans:       nil,
			}
			r.register[sequenceNumber] = subs

			// Update routing
			r.mutex.Unlock()
			err := subs.UpdateRoute(CREATE, r.rtmgrClient)
			r.mutex.Lock()
			if err != nil {
				if _, ok := r.register[sequenceNumber]; ok {
					delete(r.register, sequenceNumber)
				}
				return nil, err
			}
			return subs, nil
		}
	}
	return nil, fmt.Errorf("Registry: Failed to reserves subcription. RmrEndpoint: %s, Meid: %s", endPoint, meid.RanName)
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

func (r *Registry) DelSubscription(sn uint16) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.register[sn]; ok {
		subs := r.register[sn]
		delete(r.register, sn)

		// Update routing
		r.mutex.Unlock()
		err := subs.UpdateRoute(DELETE, r.rtmgrClient)
		r.mutex.Lock()
		if err != nil {
			xapp.Logger.Error("Registry: Failed to del route. SubId: %d, RmrEndpoint: %s", subs.Seq, subs.RmrEndpoint)
		}
		return true
	}
	return false
}
