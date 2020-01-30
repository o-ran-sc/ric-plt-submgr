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
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type Registry struct {
	mutex       sync.Mutex
	register    map[uint32]*Subscription
	subIds      []uint32
	rtmgrClient *RtmgrClient
}

func (r *Registry) Initialize() {
	r.register = make(map[uint32]*Subscription)
	var i uint32
	for i = 0; i < 65535; i++ {
		r.subIds = append(r.subIds, i+1)
	}
}

func (r *Registry) allocateSubs(trans *Transaction, subReqMsg *e2ap.E2APSubscriptionRequest) (*Subscription, error) {
	if len(r.subIds) > 0 {
		sequenceNumber := r.subIds[0]
		r.subIds = r.subIds[1:]
		if _, ok := r.register[sequenceNumber]; ok == true {
			r.subIds = append(r.subIds, sequenceNumber)
			return nil, fmt.Errorf("Registry: Failed to reserve subscription exists")
		}
		subs := &Subscription{
			registry:  r,
			Meid:      trans.Meid,
			SubReqMsg: subReqMsg,
			valid:     true,
		}
		subs.ReqId.Id = 123
		subs.ReqId.Seq = sequenceNumber

		if subs.EpList.AddEndpoint(trans.GetEndpoint()) == false {
			r.subIds = append(r.subIds, subs.ReqId.Seq)
			return nil, fmt.Errorf("Registry: Endpoint existing already in subscription")
		}

		return subs, nil
	}
	return nil, fmt.Errorf("Registry: Failed to reserve subscription no free ids")
}

func (r *Registry) findExistingSubs(trans *Transaction, subReqMsg *e2ap.E2APSubscriptionRequest) *Subscription {

	for _, subs := range r.register {
		if subs.IsMergeable(trans, subReqMsg) {

			//
			// check if there has been race conditions
			//
			subs.mutex.Lock()
			//subs has been set to invalid
			if subs.valid == false {
				subs.mutex.Unlock()
				continue
			}
			// If size is zero, entry is to be deleted
			if subs.EpList.Size() == 0 {
				subs.mutex.Unlock()
				continue
			}
			// try to add to endpointlist.
			if subs.EpList.AddEndpoint(trans.GetEndpoint()) == false {
				subs.mutex.Unlock()
				continue
			}
			subs.mutex.Unlock()

			xapp.Logger.Debug("Registry: Mergeable subs found %s for %s", subs.String(), trans.String())
			return subs
		}
	}
	return nil
}

func (r *Registry) AssignToSubscription(trans *Transaction, subReqMsg *e2ap.E2APSubscriptionRequest) (*Subscription, error) {
	var err error
	var newAlloc bool
	r.mutex.Lock()
	defer r.mutex.Unlock()

	subs := r.findExistingSubs(trans, subReqMsg)

	if subs == nil {
		subs, err = r.allocateSubs(trans, subReqMsg)
		if err != nil {
			return nil, err
		}
		newAlloc = true
	}

	//
	// Add to subscription
	//
	subs.mutex.Lock()
	defer subs.mutex.Unlock()

	epamount := subs.EpList.Size()

	r.mutex.Unlock()
	//
	// Subscription route updates
	//
	if epamount == 1 {
		subRouteAction := SubRouteInfo{subs.EpList, uint16(subs.ReqId.Seq)}
		err = r.rtmgrClient.SubscriptionRequestCreate(subRouteAction)
	} else {
		subRouteAction := SubRouteInfo{subs.EpList, uint16(subs.ReqId.Seq)}
		err = r.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	}
	r.mutex.Lock()

	if err != nil {
		if newAlloc {
			r.subIds = append(r.subIds, subs.ReqId.Seq)
		}
		return nil, err
	}

	if newAlloc {
		r.register[subs.ReqId.Seq] = subs
	}
	xapp.Logger.Debug("Registry: Create %s", subs.String())
	xapp.Logger.Debug("Registry: substable=%v", r.register)
	return subs, nil
}

// TODO: Needs better logic when there is concurrent calls
func (r *Registry) RemoveFromSubscription(subs *Subscription, trans *Transaction, waitRouteClean time.Duration) error {

	r.mutex.Lock()
	defer r.mutex.Unlock()
	subs.mutex.Lock()
	defer subs.mutex.Unlock()

	delStatus := subs.EpList.DelEndpoint(trans.GetEndpoint())
	epamount := subs.EpList.Size()

	//
	// If last endpoint remove from register map
	//
	if epamount == 0 {
		if _, ok := r.register[subs.ReqId.Seq]; ok {
			xapp.Logger.Debug("Registry: Delete %s", subs.String())
			delete(r.register, subs.ReqId.Seq)
			xapp.Logger.Debug("Registry: substable=%v", r.register)
		}
	}
	r.mutex.Unlock()

	//
	// Wait some time before really do route updates
	//
	if waitRouteClean > 0 {
		subs.mutex.Unlock()
		time.Sleep(waitRouteClean)
		subs.mutex.Lock()
	}

	xapp.Logger.Info("Registry: Cleaning %s", subs.String())

	//
	// Subscription route updates
	//
	if delStatus {
		if epamount == 0 {
			tmpList := RmrEndpointList{}
			tmpList.AddEndpoint(trans.GetEndpoint())
			subRouteAction := SubRouteInfo{tmpList, uint16(subs.ReqId.Seq)}
			r.rtmgrClient.SubscriptionRequestDelete(subRouteAction)
		} else {
			subRouteAction := SubRouteInfo{subs.EpList, uint16(subs.ReqId.Seq)}
			r.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
		}
	}

	r.mutex.Lock()
	//
	// If last endpoint free seq nro
	//
	if epamount == 0 {
		r.subIds = append(r.subIds, subs.ReqId.Seq)
	}

	return nil
}

func (r *Registry) GetSubscription(sn uint32) *Subscription {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.register[sn]; ok {
		return r.register[sn]
	}
	return nil
}

func (r *Registry) GetSubscriptionFirstMatch(ids []uint32) (*Subscription, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, id := range ids {
		if _, ok := r.register[id]; ok {
			return r.register[id], nil
		}
	}
	return nil, fmt.Errorf("No valid subscription found with ids %v", ids)
}
