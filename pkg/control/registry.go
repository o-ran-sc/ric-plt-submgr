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
	"sync"
)

type Subscription struct {
	Seq       uint16
	Confirmed bool
}

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
func (r *Registry) ReserveSequenceNumber() (uint16, bool) {
	// Check is current SequenceNumber valid
	r.mutex.Lock()
	defer r.mutex.Unlock()
	sequenceNumber := r.counter
	if _, ok := r.register[sequenceNumber]; ok {
		xapp.Logger.Error("Invalid SeqenceNumber sequenceNumber: %v", sequenceNumber)
		return sequenceNumber, false
	}
	r.register[sequenceNumber] = &Subscription{sequenceNumber, false}

	// Allocate next SequenceNumber value
	if r.counter == 65535 {
		r.counter = 0
	} else {
		r.counter++
	}
	return sequenceNumber, true
}

// This function checks the validity of the given subscription id
func (r *Registry) IsValidSequenceNumber(sn uint16) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	xapp.Logger.Debug("Registry map: %v", r.register)
	if _, ok := r.register[sn]; ok {
		return true
	}
	return false
}

// This function sets the give id as confirmed in the register
func (r *Registry) setSubscriptionToConfirmed(sn uint16) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.register[sn].Confirmed = true
}

//This function sets the given id as unused in the register
func (r *Registry) deleteSubscription(sn uint16) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.register[sn].Confirmed = false
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
