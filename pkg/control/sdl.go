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
	"encoding/json"
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	sdl "gerrit.o-ran-sc.org/r/ric-plt/sdlgo"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strconv"
)

type SubscriptionInfo struct {
	Valid       bool
	ReqId       RequestId
	Meid        xapp.RMRMeid
	EpList      xapp.RmrEndpointList
	SubReqMsg   e2ap.E2APSubscriptionRequest
	SubRFMsg    interface{}
	SubRespRcvd bool
}

func CreateSdl() Sdlnterface {
	return sdl.NewSdlInstance("submgr", sdl.NewDatabase())
}

func (c *Control) WriteSubscriptionToSdl(subId uint32, subs *Subscription) error {

	var subscriptionInfo SubscriptionInfo
	subscriptionInfo.Valid = subs.valid
	subscriptionInfo.SubRespRcvd = subs.SubRespRcvd
	subscriptionInfo.ReqId = subs.ReqId
	subscriptionInfo.Meid = *subs.Meid
	subscriptionInfo.EpList = subs.EpList
	subscriptionInfo.SubReqMsg = *subs.SubReqMsg
	subscriptionInfo.SubRFMsg = subs.SubRFMsg

	jsonData, err := json.Marshal(subscriptionInfo)
	if err != nil {
		return fmt.Errorf("SDL: WriteSubscriptionToSdl() json.Marshal error: %s", err)
	}

	err = c.db.Set(strconv.FormatUint(uint64(subId), 10), jsonData)
	if err != nil {
		return fmt.Errorf("SDL: WriteSubscriptionToSdl(): %s", err)
	} else {
		xapp.Logger.Debug("SDL: Subscription written in db.  subId = %v", subId)
	}
	return nil
}

func (c *Control) ReadSubscriptionFromSdl(subId uint32) (*Subscription, error) {

	// This function is now just for testing purpose
	key := strconv.FormatUint(uint64(subId), 10)
	retMap, err := c.db.Get([]string{key})
	if err != nil {
		return nil, fmt.Errorf("SDL: ReadSubscriptionFromSdl(): %s", err)
	} else {
		xapp.Logger.Debug("SDL: Subscription read from db.  subId = %v", subId)
	}

	subs := &Subscription{}
	for _, iSubscriptionInfo := range retMap {

		if iSubscriptionInfo == nil {
			return nil, fmt.Errorf("SDL: ReadSubscriptionFromSdl() subscription not found. subId = %v\n", subId)
		}

		subscriptionInfo := &SubscriptionInfo{}
		jsonSubscriptionInfo := iSubscriptionInfo.(string)
		err := json.Unmarshal([]byte(jsonSubscriptionInfo), subscriptionInfo)
		if err != nil {
			return nil, fmt.Errorf("SDL: ReadSubscriptionFromSdl() json.unmarshal error: %s\n", err.Error())
		}

		subs.valid = subscriptionInfo.Valid
		subs.ReqId = subscriptionInfo.ReqId
		meid := xapp.RMRMeid{}
		meid = subscriptionInfo.Meid
		subs.Meid = &meid
		subs.EpList = subscriptionInfo.EpList
		subs.TheTrans = nil
		subReq := e2ap.E2APSubscriptionRequest{}
		subReq = subscriptionInfo.SubReqMsg
		subs.SubReqMsg = &subReq
		subs.SubRFMsg = subscriptionInfo.SubRFMsg
		subs.SubRespRcvd = subscriptionInfo.SubRespRcvd
	}
	return subs, nil
}

func (c *Control) RemoveSubscriptionFromSdl(subId uint32) error {

	key := strconv.FormatUint(uint64(subId), 10)
	err := c.db.Remove([]string{key})
	if err != nil {
		return fmt.Errorf("SDL: RemoveSubscriptionfromSdl(): %s\n", err.Error())
	} else {
		xapp.Logger.Debug("SDL: Subscription removed from db. subId = %v", subId)
	}
	return nil
}

func (c *Control) ReadAllSubscriptionsFromSdl() ([]uint32, map[uint32]*Subscription, error) {

	// Read all subscriptionInfos
	var subIds []uint32
	var i uint32
	for i = 1; i < 65535; i++ {
		subIds = append(subIds, i)
	}

	retMap := make(map[uint32]*Subscription)
	// Get all keys
	keys, err := c.db.GetAll()
	if err != nil {
		return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl(), GetAll(). Error while reading keys from DBAAS %s\n", err.Error())
	}

	if len(keys) == 0 {
		return subIds, retMap, nil
	}

	// Get all subscriptionInfos
	iSubscriptionMap, err := c.db.Get(keys)
	if err != nil {
		return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl(), Get():  Error while reading subscriptions from DBAAS %s\n", err.Error())
	}

	for _, iSubscriptionInfo := range iSubscriptionMap {

		if iSubscriptionInfo == nil {
			return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl() iSubscriptionInfo = nil\n")
		}

		subscriptionInfo := &SubscriptionInfo{}
		jsonSubscriptionInfo := iSubscriptionInfo.(string)
		err := json.Unmarshal([]byte(jsonSubscriptionInfo), subscriptionInfo)
		if err != nil {
			return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl() json.unmarshal error: %s\n", err.Error())
		}

		subs := &Subscription{}
		subs.registry = c.registry
		subs.valid = subscriptionInfo.Valid
		subs.ReqId = subscriptionInfo.ReqId
		meid := xapp.RMRMeid{}
		meid = subscriptionInfo.Meid
		subs.Meid = &meid
		subs.EpList = subscriptionInfo.EpList
		subs.TheTrans = nil
		subReq := e2ap.E2APSubscriptionRequest{}
		subReq = subscriptionInfo.SubReqMsg
		subs.SubReqMsg = &subReq
		subs.SubRFMsg = subscriptionInfo.SubRFMsg
		subs.SubRespRcvd = subscriptionInfo.SubRespRcvd

		if int(subscriptionInfo.ReqId.InstanceId) >= len(subIds) {
			return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl() index is out of range. Index is %d with slice length %d", subscriptionInfo.ReqId.InstanceId, len(subIds))
		}
		retMap[subscriptionInfo.ReqId.InstanceId] = subs

		// Remove subId from free subIds. Original slice is modified here!
		subIds, err = removeNumber(subIds, subscriptionInfo.ReqId.InstanceId)
		if err != nil {
			return nil, nil, fmt.Errorf("SDL: ReadAllSubscriptionsFromSdl() error: %s\n", err.Error())
		}
	}
	return subIds, retMap, nil
}

func removeNumber(s []uint32, removedNum uint32) ([]uint32, error) {
	for i, num := range s {
		if removedNum == uint32(num) {
			s = append(s[:i], s[i+1:]...)
			return s[:len(s)], nil
		}
	}
	return nil, fmt.Errorf("SDL: To be removed number not in the slice. removedNum: %v", removedNum)
}
func (c *Control) RemoveAllSubscriptionsFromSdl() error {

	err := c.db.RemoveAll()
	if err != nil {
		return fmt.Errorf("SDL: RemoveAllSubscriptionsFromSdl(): %s\n", err.Error())
	} else {
		xapp.Logger.Debug("SDL: All subscriptions removed from db")
	}
	return nil
}
