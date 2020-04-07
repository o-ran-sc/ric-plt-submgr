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
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/xapptweaks"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
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

func (r *Registry) QueryHandler() (models.SubscriptionList, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	resp := models.SubscriptionList{}
	for _, subs := range r.register {
		subs.mutex.Lock()
		resp = append(resp, &models.SubscriptionData{SubscriptionID: int64(subs.ReqId.InstanceId), Meid: subs.Meid.RanName, Endpoint: subs.EpList.StringList()})
		subs.mutex.Unlock()
	}
	return resp, nil
}

func (r *Registry) allocateSubs(trans *TransactionXapp, subReqMsg *e2ap.E2APSubscriptionRequest) (*Subscription, error) {
	if len(r.subIds) > 0 {
		subId := r.subIds[0]
		r.subIds = r.subIds[1:]
		if _, ok := r.register[subId]; ok == true {
			r.subIds = append(r.subIds, subId)
			return nil, fmt.Errorf("Registry: Failed to reserve subscription exists")
		}
		subs := &Subscription{
			registry:  r,
			Meid:      trans.Meid,
			SubReqMsg: subReqMsg,
			valid:     true,
		}
		subs.ReqId.Id = 123
		subs.ReqId.InstanceId = subId

		if subs.EpList.AddEndpoint(trans.GetEndpoint()) == false {
			r.subIds = append(r.subIds, subs.ReqId.InstanceId)
			return nil, fmt.Errorf("Registry: Endpoint existing already in subscription")
		}

		return subs, nil
	}
	return nil, fmt.Errorf("Registry: Failed to reserve subscription no free ids")
}

func (r *Registry) findExistingSubs(trans *TransactionXapp, subReqMsg *e2ap.E2APSubscriptionRequest) *Subscription {

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

func (r *Registry) AssignToSubscription(trans *TransactionXapp, subReqMsg *e2ap.E2APSubscriptionRequest) (*Subscription, error) {
	var err error
	var newAlloc bool
	r.mutex.Lock()
	defer r.mutex.Unlock()

	//
	// Check validity of subscription action types
	//
	actionType, err := r.CheckActionTypes(subReqMsg)
	if err != nil {
		xapp.Logger.Debug("CREATE %s", err)
		return nil, err
	}

	//
	// Find possible existing Policy subscription
	//
	if actionType == e2ap.E2AP_ActionTypePolicy {
		if subs, ok := r.register[trans.GetSubId()]; ok {
			xapp.Logger.Debug("CREATE %s. Existing subscription for Policy found", subs.String())
			subs.SetCachedResponse(nil, true)
			return subs, nil
		}
	}

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
		subRouteAction := SubRouteInfo{subs.EpList, uint16(subs.ReqId.InstanceId)}
		err = r.rtmgrClient.SubscriptionRequestCreate(subRouteAction)
	} else {
		subRouteAction := SubRouteInfo{subs.EpList, uint16(subs.ReqId.InstanceId)}
		err = r.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
	}
	r.mutex.Lock()

	if err != nil {
		if newAlloc {
			r.subIds = append(r.subIds, subs.ReqId.InstanceId)
		}
		return nil, err
	}

	if newAlloc {
		r.register[subs.ReqId.InstanceId] = subs
	}
	xapp.Logger.Debug("CREATE %s", subs.String())
	xapp.Logger.Debug("Registry: substable=%v", r.register)
	return subs, nil
}

func (r *Registry) CheckActionTypes(subReqMsg *e2ap.E2APSubscriptionRequest) (uint64, error) {
	var reportFound bool = false
	var policyFound bool = false

	for _, acts := range subReqMsg.ActionSetups {
		if acts.ActionType == e2ap.E2AP_ActionTypeReport {
			reportFound = true
		}
		if acts.ActionType == e2ap.E2AP_ActionTypePolicy {
			policyFound = true
		}
	}
	if reportFound == true && policyFound == true {
		return e2ap.E2AP_ActionTypeInvalid, fmt.Errorf("Report and Policy in same RICactions-ToBeSetup-List")
	}
	if reportFound == true {
		return e2ap.E2AP_ActionTypeReport, nil
	}
	if policyFound == true {
		return e2ap.E2AP_ActionTypePolicy, nil
	}
	return e2ap.E2AP_ActionTypeInvalid, fmt.Errorf("Invalid action type in RICactions-ToBeSetup-List")
}

// TODO: Works with concurrent calls, but check if can be improved
func (r *Registry) RemoveFromSubscription(subs *Subscription, trans *TransactionXapp, waitRouteClean time.Duration) error {

	r.mutex.Lock()
	defer r.mutex.Unlock()
	subs.mutex.Lock()
	defer subs.mutex.Unlock()

	delStatus := subs.EpList.DelEndpoint(trans.GetEndpoint())
	epamount := subs.EpList.Size()
	subId := subs.ReqId.InstanceId

	if delStatus == false {
		return nil
	}

	go func() {
		if waitRouteClean > 0 {
			time.Sleep(waitRouteClean)
		}

		subs.mutex.Lock()
		defer subs.mutex.Unlock()
		xapp.Logger.Info("CLEAN %s", subs.String())

		if epamount == 0 {
			//
			// Subscription route delete
			//
			tmpList := xapptweaks.RmrEndpointList{}
			tmpList.AddEndpoint(trans.GetEndpoint())
			subRouteAction := SubRouteInfo{tmpList, uint16(subId)}
			r.rtmgrClient.SubscriptionRequestDelete(subRouteAction)

			//
			// Subscription release
			//
			r.mutex.Lock()
			defer r.mutex.Unlock()

			if _, ok := r.register[subId]; ok {
				xapp.Logger.Debug("RELEASE %s", subs.String())
				delete(r.register, subId)
				xapp.Logger.Debug("Registry: substable=%v", r.register)
			}
			r.subIds = append(r.subIds, subId)

		} else if subs.EpList.Size() > 0 {
			//
			// Subscription route updates
			//
			subRouteAction := SubRouteInfo{subs.EpList, uint16(subId)}
			r.rtmgrClient.SubscriptionRequestUpdate(subRouteAction)
		}

	}()

	return nil
}

func (r *Registry) GetSubscription(subId uint32) *Subscription {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.register[subId]; ok {
		return r.register[subId]
	}
	return nil
}

func (r *Registry) GetSubscriptionFirstMatch(subIds []uint32) (*Subscription, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, subId := range subIds {
		if _, ok := r.register[subId]; ok {
			return r.register[subId], nil
		}
	}
	return nil, fmt.Errorf("No valid subscription found with subIds %v", subIds)
}
