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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststube2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strconv"
	"testing"
	"time"
)

const (
	subsResponse = 1
	subsFailure  = 2
	noResponse   = 3
)

type Mock struct {
	subsDB             map[string]interface{}
	register           map[uint32]*Subscription
	subIds             []uint32
	lastAllocatedSubId uint32
}

var mock *Mock

func CreateMock() *Mock {
	fmt.Println("Test CreateMock()")
	mock = new(Mock)
	mock.ResetTestSttings()
	return mock
}

func (m *Mock) ResetTestSttings() {
	m.subsDB = make(map[string]interface{}, 0)
	m.register = make(map[uint32]*Subscription)
	var i uint32
	for i = 1; i < 65535; i++ {
		m.subIds = append(mock.subIds, i)
	}
}

func (m *Mock) AllocNextSubId() uint32 {
	m.lastAllocatedSubId = mock.subIds[0]
	return m.lastAllocatedSubId
}

func TestWait(t *testing.T) {
	// Wait to test settings to complete
	<-time.After(1 * time.Second)
}

func GetSubscription(t *testing.T, e2SubId uint32, responseType int, srcEndPoint, ranName string, xId string) *Subscription {
	t.Log("TEST: Getting subscription")

	subs := &Subscription{}

	// Create unpacked e2SubReqMsg
	subReqParams := &teststube2ap.E2StubSubsReqParams{}
	subReqParams.Init()

	meid := xapp.RMRMeid{}
	meid.RanName = ranName

	params := &xapp.RMRParams{}
	params.Src = srcEndPoint
	params.Xid = xId
	params.Meid = &meid

	// Create xApp transaction
	trans := mainCtrl.c.tracker.NewXappTransaction(xapp.NewRmrEndpoint(params.Src), params.Xid, subReqParams.Req.RequestId.InstanceId, params.Meid)
	if trans == nil {
		t.Errorf("TEST: %s", idstring(fmt.Errorf("transaction not created"), params))
		return nil
	}

	// Allocate E2 instanceId/subId
	subReqParams.Req.RequestId.InstanceId = e2SubId

	subs.ReqId.Id = 123
	subs.ReqId.InstanceId = subReqParams.Req.RequestId.InstanceId
	subs.Meid = &meid
	subs.EpList.AddEndpoint(trans.GetEndpoint())
	subs.SubReqMsg = subReqParams.Req
	// subs.SubRFMsg contains received/cached SubscriptionResponse or SubscriptionFailure, nil in no response received
	if responseType == subsResponse {
		subs.SubRFMsg = GetSubsResponse(t, subReqParams.Req)
		subs.valid = true
	} else if responseType == subsFailure {
		subs.SubRFMsg = GetSubsFailure(t, subReqParams.Req)
		subs.valid = false
	} else if responseType == noResponse {
		subs.SubRFMsg = nil
		subs.valid = false
	}
	return subs
}

func GetSubsResponse(t *testing.T, req *e2ap.E2APSubscriptionRequest) interface{} {
	t.Log("TEST: Getting ricSubscriptionResponse")

	// Create e2SubRespMsg
	var e2asnpacker e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()
	e2SubsResp := e2asnpacker.NewPackerSubscriptionResponse()

	resp := &e2ap.E2APSubscriptionResponse{}
	resp.RequestId.Id = 123
	resp.RequestId.InstanceId = req.RequestId.InstanceId
	resp.FunctionId = req.FunctionId

	resp.ActionAdmittedList.Items = make([]e2ap.ActionAdmittedItem, len(req.ActionSetups))
	for index := int(0); index < len(req.ActionSetups); index++ {
		resp.ActionAdmittedList.Items[index].ActionId = req.ActionSetups[index].ActionId
	}

	for index := uint64(0); index < 1; index++ {
		item := e2ap.ActionNotAdmittedItem{}
		item.ActionId = index
		item.Cause.Content = 1
		item.Cause.Value = 1
		resp.ActionNotAdmittedList.Items = append(resp.ActionNotAdmittedList.Items, item)
	}

	packerr, packedMsg := e2SubsResp.Pack(resp)
	if packerr != nil {
		t.Errorf("TEST: packing SubscriptionResponse failed. %s", packerr.Error())
		return nil
	}
	t.Logf("TEST: %s", e2SubsResp.String())
	return packedMsg
}

func GetSubsFailure(t *testing.T, req *e2ap.E2APSubscriptionRequest) interface{} {
	t.Log("TEST: Getting ricSubscriptionFailure")

	var e2asnpacker e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()
	e2SubsFail := e2asnpacker.NewPackerSubscriptionFailure()

	fail := &e2ap.E2APSubscriptionFailure{}
	fail.RequestId.Id = req.RequestId.Id
	fail.RequestId.InstanceId = req.RequestId.InstanceId
	fail.FunctionId = req.FunctionId

	packerr, packedMsg := e2SubsFail.Pack(fail)
	if packerr != nil {
		t.Errorf("TEST: packing SubscriptionFailure failed. %s", packerr.Error())
		return nil
	}
	t.Logf("TEST: %s", e2SubsFail.String())
	return packedMsg
}

func PrintSubscriptionData(t *testing.T, subs *Subscription) {
	t.Log("TEST: subscription data")
	t.Logf("TEST: subs.mutex = %v", subs.mutex)
	t.Logf("TEST: subs.ReqId.InstanceId = %v", subs.ReqId.InstanceId)
	t.Logf("TEST: subs.ReqId.Id = %v", subs.ReqId.Id)
	t.Logf("TEST: subs.EpList = %v", subs.EpList)
	t.Logf("TEST: subs.Meid.RanName = %v", subs.Meid.RanName)
	t.Logf("TEST: subs.SubReqMsg = %v", subs.SubReqMsg.String())
	t.Logf("TEST: subs.valid = %v", subs.valid)
	if subs.SubRFMsg != nil {
		switch typeofSubsMessage(subs.SubRFMsg) {
		case "SubResp":
			//subReqMap := subs.SubRFMsg.(*e2ap.PackedData)
			//t.Logf("TEST: len(subs.SubRFMsg) = %v", len(subReq.Buf))
			//t.Logf("TEST: subs.SubRFMsg(subReq) = %v", subReq.Buf)
		case "SubFail":
			//subFailMap := subs.SubRFMsg.(*e2ap.PackedData)
			//t.Logf("TEST: len(subs.SubRFMsg) = %v", len(subFail.Buf))
			//t.Logf("TEST: subs.SubRFMsg(subFail) = %v", subFail.Buf)
		}
	}
}

func TestWriteSubscriptionToSdl(t *testing.T) {
	t.Log("TestWriteSubscriptionToSdl")

	// Write one subscription
	subId := mock.AllocNextSubId()
	subs := GetSubscription(t, subId, subsResponse, "localhost:13560", "RAN_NAME_1", "123456")
	PrintSubscriptionData(t, subs)
	t.Logf("TEST: Writing subId = %v\n", subId)
	err := mainCtrl.c.WriteSubscriptionToSdl(subId, subs)
	if err != nil {
		t.Errorf("TEST: %v", err)
		return
	}
}

func TestReadSubscriptionFromSdl(t *testing.T) {
	t.Log("TestReadSubscriptionFromSdl")

	subId := mock.lastAllocatedSubId
	t.Logf("Reading subId = %v\n", subId)
	subs, err := mainCtrl.c.ReadSubscriptionFromSdl(subId)
	if err != nil {
		t.Errorf("TEST: %v", err)
		return
	}
	PrintSubscriptionData(t, subs)
	assert.Equal(t, mock.register[subId].SubReqMsg, subs.SubReqMsg)
}

func TestRemoveSubscriptionFromSdl(t *testing.T) {
	t.Log("TestRemoveSubscriptionFromSdl")

	subId := mock.lastAllocatedSubId
	err := mainCtrl.c.RemoveSubscriptionFromSdl(subId)
	if err != nil {
		t.Errorf("TEST: %v", err)
		return
	}
	delete(mock.register, subId)
	mock.subIds = append(mock.subIds, subId)
	t.Logf("TEST: subscription removed from db. subId = %v", subId)
}

func TestReadNotExistingSubscriptionFromSdl(t *testing.T) {
	t.Log("TestReadNotExistingSubscriptionFromSdl")

	var subId uint32 = 0
	subs, err := mainCtrl.c.ReadSubscriptionFromSdl(subId)
	if err != nil {
		t.Logf("TEST: subscription not found from db. subId = %v", subId)
		return
	}
	t.Errorf("TEST: subscription read from db. %v", subs.String())
	PrintSubscriptionData(t, subs)
}

func TestReadNotExistingSubscriptionFromSdl2(t *testing.T) {
	t.Log("TestReadNotExistingSubscriptionFromSdl")

	var subId uint32 = 7
	subs, err := mainCtrl.c.ReadSubscriptionFromSdl(subId)
	if err != nil {
		t.Logf("TEST: subscription not found from db. subId = %v", subId)
		return
	}
	t.Errorf("TEST: subscription read from db. %v", subs.String())
	PrintSubscriptionData(t, subs)
}

func TestRemoveNotExistingSubscriptionFromSdl(t *testing.T) {
	t.Log("TestRemoveNotExistingSubscriptionFromSdl")

	var subId uint32 = 0
	err := mainCtrl.c.RemoveSubscriptionFromSdl(subId)
	if err != nil {
		t.Logf("TEST: %v", err)
		return
	}
	t.Logf("TEST: subscription removed from db. subId = %v", subId)
}

func TestWriteSubscriptionsToSdl(t *testing.T) {
	t.Log("TestWriteSubscriptionsToSdl")

	// Write 1st subscription
	subId := mock.AllocNextSubId()
	t.Logf("TEST: Writing subId = %v\n", subId)
	subs := GetSubscription(t, subId, subsResponse, "localhost:13560", "RAN_NAME_1", "123456")
	PrintSubscriptionData(t, subs)
	err := mainCtrl.c.WriteSubscriptionToSdl(subId, subs)
	if err != nil {
		t.Errorf("TEST: %v", err)
		return
	}
	t.Logf("TEST: subscription written in db = %v", subs.String())

	// Write 2nd subscription
	subId = mock.AllocNextSubId()
	t.Logf("TEST:Writing subId = %v\n", subId)
	subs = GetSubscription(t, subId, subsFailure, "localhost:13560", "RAN_NAME_2", "123457")
	PrintSubscriptionData(t, subs)
	err = mainCtrl.c.WriteSubscriptionToSdl(subId, subs)
	if err != nil {
		t.Errorf("TEST: %v", err)
		return
	}
	t.Logf("TEST: subscription written in db = %v", subs.String())

	// Write 3rd subscription
	subId = mock.AllocNextSubId()
	t.Logf("TEST:Writing subId = %v\n", subId)
	subs = GetSubscription(t, subId, noResponse, "localhost:13560", "RAN_NAME_3", "123458")
	PrintSubscriptionData(t, subs)
	err = mainCtrl.c.WriteSubscriptionToSdl(subId, subs)
	if err != nil {
		t.Errorf("TEST: %v", err)
		return
	}
	t.Logf("TEST: subscription written in db = %v", subs.String())
}

func TestReadSubscriptionsFromSdl(t *testing.T) {
	t.Log("TestReadSubscriptionsFromSdl")

	// Subscription with subId 1 was added and and removed above. Then subscriptions with subIds 2, 3 and 4 was added
	// Db subscriptions should now contain subIDs 2, 3 and 4
	var subId uint32
	for subId = 2; subId <= 4; subId++ {
		subs, err := mainCtrl.c.ReadSubscriptionFromSdl(subId)
		if err != nil {
			t.Errorf("TEST: %v", err)
			return
		}
		PrintSubscriptionData(t, subs)
	}
}

func TestReadAllSubscriptionsFromSdl(t *testing.T) {
	t.Log("TestReadAllSubscriptionsFromSdl")

	// This test cases simulates submgr restart. SubIds and subscriptions are restored from db
	// after initializing mock.subIds and mock.register
	var err error
	//beforeRestartRegister := mock.register
	mock.subIds, mock.register, err = mainCtrl.c.ReadAllSubscriptionsFromSdl()
	if err != nil {
		t.Errorf("TEST: %v", err)
		return
	}
	for _, subs := range mock.register {
		PrintSubscriptionData(t, subs)
	}
	// SubIds slices before and after restart can't be directly compared as original slice is not stored
	// in the db. SubId values 1, 2, 3, 4 are already removed from the beginning of subIds slice above
	// so far. Next free subId is 5 in the beginning of mock.subIds slice. The db contains now however only
	// 3 subscriptions with subIds 2, 3 and 4, so only subId values 2, 3, 4 are removed from the returned
	// subIds slice and there next free value is 1
	assert.Equal(t, uint32(0x1), mock.subIds[0])
	// Data compare for subscriptions could be done field by field. Pointers in the struct can't be compared directly
	//assert.Equal(t, mock.register..., beforeRestartRegister...)
}

func TestRemoveAllSubscriptionsFromSdl(t *testing.T) {
	t.Log("TestRemoveAllSubscriptionsFromSdl")

	err := mainCtrl.c.RemoveAllSubscriptionsFromSdl()
	if err != nil {
		t.Errorf("TEST: %v", err)
		return
	}
	t.Log("TEST: All subscription removed from db")
}

func TestReadAllSubscriptionsFromSdl2(t *testing.T) {
	t.Log("TestReadAllSubscriptionsFromSdl2")

	// This test cases simulates submgr startup. SubIds and subscriptions are restored from empty db
	// after initializing mock.subIds and mock.register
	var err error
	mock.subIds, mock.register, err = mainCtrl.c.ReadAllSubscriptionsFromSdl()
	if err != nil {
		t.Errorf("TEST: %v", err)
		return
	}
	for _, subs := range mock.register {
		PrintSubscriptionData(t, subs)
	}
	assert.Equal(t, len(mock.subIds), 65534)
	assert.Equal(t, len(mock.register), 0)
}

func (m *Mock) Set(pairs ...interface{}) error {
	var key string
	var val interface{}

	for _, v := range pairs {
		reflectType := reflect.TypeOf(v)
		switch reflectType.Kind() {
		case reflect.Slice:
			val = v
		default:
			switch v.(type) {
			case string:
				key = v.(string)
			default:
				return fmt.Errorf("Set() error: Unexpected type")
			}
		}
	}

	if key != "" {
		m.subsDB[key] = val
		subId := m.subIds[0]
		jsonSubscription := val.([]uint8)
		subscription := &Subscription{}
		err := json.Unmarshal([]byte(jsonSubscription), subscription)
		if err != nil {
			return fmt.Errorf("Set() () json.unmarshal error: %v\n", err)
		}
		m.register[subId] = subscription
		m.subIds = m.subIds[1:]
	} else {
		return fmt.Errorf("Set() error: key == %v\n", key)
	}
	return nil
}

func (m *Mock) Get(keys []string) (map[string]interface{}, error) {
	retMap := make(map[string]interface{})
	if len(keys) == 0 {
		return nil, fmt.Errorf("Get() error: len(key) == 0\n")
	}

	key := keys[0]
	if key != "" {
		retMap[key] = m.subsDB[key]
	} else {
		return nil, fmt.Errorf("Get() error: key == %v\n", key)
	}
	return retMap, nil
}

func (m *Mock) GetAll() ([]string, error) {
	retString := []string{}
	for _, subs := range m.subsDB {
		retString = append(retString, string(subs.([]uint8)))
	}
	return retString, nil
}

func (m *Mock) Remove(keys []string) error {
	if len(keys) == 0 {
		return fmt.Errorf("Remove() error: len(key) == 0\n")
	}
	subId64, err := strconv.ParseUint(keys[0], 10, 64)
	if err != nil {
		return fmt.Errorf("Remove() ParseUint() error: %v\n", err)
	}
	subId := uint32(subId64)
	delete(m.subsDB, keys[0])
	delete(m.register, subId)
	m.subIds = append(m.subIds, subId)
	return nil
}

func (m *Mock) RemoveAll() error {
	for key := range m.subsDB {
		subId64, err := strconv.ParseUint(key, 10, 64)
		if err != nil {
			return fmt.Errorf("RemoveAll() ParseUint() error: %v\n", err)
		}
		subId := uint32(subId64)
		delete(m.subsDB, key)
		delete(m.register, subId)
		m.subIds = append(m.subIds, subId)
	}
	return nil
}
