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
	"testing"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststube2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/stretchr/testify/assert"
)

func TestSuiteSetup(t *testing.T) {
	// The effect of this call shall endure thgough the UT suite!
	// If this causes any issues, the previout interface can be restored
	// like this:git log
	// SetPackerIf(e2ap_wrapper.NewAsn1E2APPacker())

	SetPackerIf(e2ap_wrapper.NewUtAsn1E2APPacker())
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndDeleteOkWithE2apUtWrapper
//
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|  // The order of these events may vary
//     |                 |              |             |
//     |     RESTSubResp |              |             |  // The order of these events may vary
//     |<----------------|              |             |
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|  // The order of these events may vary
//     |                 |              |             |
//     |                 | SubReq       |             |
//     |                 |------------->|             |  // The order of these events may vary
//     |                 |              |             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |              |             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             |
//     |                 |              |             |
//     |                 |              |             |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndDeleteOkWithE2apUtWrapper(t *testing.T) {

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, nil)

	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	waitSubsCleanup(t, e2SubsId, 10)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndE1apDeleteReqPackingError
//
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|  // The order of these events may vary
//     |                 |              |             |
//     |     RESTSubResp |              |             |  // The order of these events may vary
//     |<----------------|              |             |
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|  // The order of these events may vary
//     |                 |              |             |
//     |                 | SubReq       |             |
//     |                 |------------->|             |  // The order of these events may vary
//     |                 |              |             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |                 |              |             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     |                 |              |             |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndE1apDeleteReqPackingError(t *testing.T) {

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, nil)

	e2ap_wrapper.AllowE2apToProcess(e2ap_wrapper.SUB_DEL_REQ, false)
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	defer e2ap_wrapper.AllowE2apToProcess(e2ap_wrapper.SUB_DEL_REQ, true)

	waitSubsCleanup(t, e2SubsId, 10)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndE1apDeleteRespUnpackingError
//
//   stub                             stub          stub
// +-------+        +---------+    +---------+   +---------+
// | xapp  |        | submgr  |    | e2term  |   |  rtmgr  |
// +-------+        +---------+    +---------+   +---------+
//     |                 |              |             |
//     | RESTSubReq      |              |             |
//     |---------------->|              |             |
//     |                 | RouteCreate  |             |
//     |                 |--------------------------->|  // The order of these events may vary
//     |                 |              |             |
//     |     RESTSubResp |              |             |  // The order of these events may vary
//     |<----------------|              |             |
//     |                 | RouteResponse|             |
//     |                 |<---------------------------|  // The order of these events may vary
//     |                 |              |             |
//     |                 | SubReq       |             |
//     |                 |------------->|             |  // The order of these events may vary
//     |                 |              |             |
//     |                 |      SubResp |             |
//     |                 |<-------------|             |
//     |      RESTNotif1 |              |             |
//     |<----------------|              |             |
//     |                 |              |             |
//     | RESTSubDelReq   |              |             |
//     |---------------->|              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |              |             |
//     |   RESTSubDelResp|              |             |
//     |<----------------|              |             | // The order of these events may vary
//     |                 |              |             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             | // 1.st NOK
//     |                 |              |             |
//     |                 | SubDelReq    |             |
//     |                 |------------->|             |
//     |                 |              |             |
//     |                 |   SubDelResp |             |
//     |                 |<-------------|             | // 2.nd NOK
//
//-----------------------------------------------------------------------------

func TestRESTSubReqAndE1apDeleteRespUnpackingError(t *testing.T) {

	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, nil)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	e2ap_wrapper.AllowE2apToProcess(e2ap_wrapper.SUB_DEL_RESP, false)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	defer e2ap_wrapper.AllowE2apToProcess(e2ap_wrapper.SUB_DEL_RESP, true)

	waitSubsCleanup(t, e2SubsId, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndRouteNok
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    |  rtmgr  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | RouteCreate  |
//     |              |------------->|
//     |              |              |
//     |              | RouteCreate  |
//     |              |  status:400  |
//     |              |<-------------|
//     |              |              |
//     |       [SUBS INT DELETE]     |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqAndRouteNok(t *testing.T) {
	CaseBegin("TestSubReqAndRouteNok")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cRouteCreateFail, 1},
	})

	waiter := rtmgrHttp.AllocNextEvent(false)
	newSubsId := mainCtrl.get_registry_next_subid(t)
	xappConn1.SendSubsReq(t, nil, nil)
	waiter.WaitResult(t)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, newSubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	<-time.After(1 * time.Second)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqAndRouteUpdateNok

//   stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    |  rtmgr  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |        [SUBS CREATE]       |              |
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     | SubReq (mergeable)         |              |
//     |--------------------------->|              |              |
//     |             |              |              |
//     |             |              | RouteUpdate  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | RouteUpdate  |
//     |             |              |  status:400  |
//     |             |              |<-------------|
//     |             |              |              |
//     |       [SUBS INT DELETE]    |              |
//     |             |              |              |
//     |             |              |              |
//     |        [SUBS DELETE]       |              |
//     |             |              |              |

func TestSubReqAndRouteUpdateNok(t *testing.T) {
	CaseBegin("TestSubReqAndRouteUpdateNok")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cRouteCreateUpdateFail, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	waiter := rtmgrHttp.AllocNextEvent(false)
	newSubsId := mainCtrl.get_registry_next_subid(t)
	xappConn2.SendSubsReq(t, nil, nil)
	waiter.WaitResult(t)

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, newSubsId, 10)
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubDelReqAndRouteDeleteNok
//
//   stub                          stub
// +-------+     +---------+    +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |    |  rtmgr  |
// +-------+     +---------+    +---------+    +---------+
//     |              |              |              |
//     |         [SUBS CREATE]       |              |
//     |              |              |              |
//     |              |              |              |
//     |              |              |              |
//     | SubDelReq    |              |              |
//     |------------->|              |              |
//     |              |  SubDelReq   |              |
//     |              |------------->|              |
//     |              |  SubDelRsp   |              |
//     |              |<-------------|              |
//     |  SubDelRsp   |              |              |
//     |<-------------|              |              |
//     |              | RouteDelete  |              |
//     |              |---------------------------->|
//     |              |              |              |
//     |              | RouteDelete  |              |
//     |              |  status:400  |              |
//     |              |<----------------------------|
//     |              |              |              |
func TestSubDelReqAndRouteDeleteNok(t *testing.T) {
	CaseBegin("TestSubDelReqAndRouteDeleteNok")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cRouteDeleteFail, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	waiter := rtmgrHttp.AllocNextEvent(false)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	waiter.WaitResult(t)

	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubMergeDelAndRouteUpdateNok
//   stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |             |              |    SubResp1  |
//     |             |              |<-------------|
//     |             |    SubResp1  |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |          SubResp2          |              |
//     |<---------------------------|              |
//     |             |              |              |
//     |             | SubDelReq 1  |              |
//     |             |------------->|              |
//     |             |              | RouteUpdate  |
//     |             |              |-----> rtmgr  |
//     |             |              |              |
//     |             |              | RouteUpdate  |
//     |             |              |  status:400  |
//     |             |              |<----- rtmgr  |
//     |             |              |              |
//     |             | SubDelResp 1 |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |         SubDelReq 2        |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |         SubDelResp 2       |              |
//     |<---------------------------|              |
//
//-----------------------------------------------------------------------------
func TestSubMergeDelAndRouteUpdateNok(t *testing.T) {
	CaseBegin("TestSubMergeDelAndRouteUpdateNok")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cRouteDeleteUpdateFail, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 2},
	})

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)
	e2SubsId2 := xappConn2.RecvSubsResp(t, cretrans2)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId1))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560", "localhost:13660"})

	//Del1
	waiter := rtmgrHttp.AllocNextEvent(false)
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	waiter.WaitResult(t)

	xappConn1.RecvSubsDelResp(t, deltrans1)

	//Del2
	deltrans2 := xappConn2.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn2.RecvSubsDelResp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOk
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOk(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOk")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqRetransmission
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |  SubReq      |              |
//     | (retrans)    |              |
//     |------------->|              |
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |         [SUBS DELETE]       |
//     |              |              |
//
//-----------------------------------------------------------------------------
func TestSubReqRetransmission(t *testing.T) {
	CaseBegin("TestSubReqRetransmission")

	//Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	seqBef := mainCtrl.get_msgcounter(t)
	xappConn1.SendSubsReq(t, nil, cretrans) //Retransmitted SubReq
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	// hack as there is no real way to see has message be handled.
	// Previuos counter check just tells that is has been received by submgr
	// --> artificial delay
	<-time.After(1 * time.Second)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	//Subs Delete
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubDelReqRetransmission
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     | SubDelReq    |              |
//     | (same sub)   |              |
//     | (same xid)   |              |
//     |------------->|              |
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubDelReqRetransmission(t *testing.T) {
	CaseBegin("TestSubDelReqRetransmission")

	//Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	//Subs Delete
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	seqBef := mainCtrl.get_msgcounter(t)
	xappConn1.SendSubsDelReq(t, deltrans, e2SubsId) //Retransmitted SubDelReq
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	// hack as there is no real way to see has message be handled.
	// Previuos counter check just tells that is has been received by submgr
	// --> artificial delay
	<-time.After(1 * time.Second)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubDelReqCollision
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     | SubDelReq 1  |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq 1  |
//     |              |------------->|
//     |              |              |
//     | SubDelReq 2  |              |
//     | (same sub)   |              |
//     | (diff xid)   |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelResp 1 |
//     |              |<-------------|
//     |              |              |
//     | SubDelResp 1 |              |
//     |<-------------|              |
//     |              |              |
//     | SubDelResp 2 |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubDelReqCollision(t *testing.T) {
	CaseBegin("TestSubDelReqCollision")

	//Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	//Subs Delete
	xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)

	// Subs Delete colliding
	seqBef := mainCtrl.get_msgcounter(t)
	deltranscol2 := xappConn1.NewRmrTransactionId("", "RAN_NAME_1")
	xappConn1.SendSubsDelReq(t, deltranscol2, e2SubsId) //Colliding SubDelReq
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	// hack as there is no real way to see has message be handled.
	// Previuos counter check just tells that is has been received by submgr
	// --> artificial delay
	<-time.After(1 * time.Second)

	// Del resp for first and second
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// don't care in which order responses are received
	xappConn1.RecvSubsDelResp(t, nil)
	xappConn1.RecvSubsDelResp(t, nil)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkTwoParallel
//
//   stub       stub                          stub
// +-------+  +-------+     +---------+    +---------+
// | xapp  |  | xapp  |     | submgr  |    | e2term  |
// +-------+  +-------+     +---------+    +---------+
//     |          |              |              |
//     |          |              |              |
//     |          |              |              |
//     |          | SubReq1      |              |
//     |          |------------->|              |
//     |          |              |              |
//     |          |              | SubReq1      |
//     |          |              |------------->|
//     |          |              |              |
//     |       SubReq2           |              |
//     |------------------------>|              |
//     |          |              |              |
//     |          |              | SubReq2      |
//     |          |              |------------->|
//     |          |              |              |
//     |          |              |    SubResp1  |
//     |          |              |<-------------|
//     |          |    SubResp1  |              |
//     |          |<-------------|              |
//     |          |              |              |
//     |          |              |    SubResp2  |
//     |          |              |<-------------|
//     |       SubResp2          |              |
//     |<------------------------|              |
//     |          |              |              |
//     |          |        [SUBS 1 DELETE]      |
//     |          |              |              |
//     |          |        [SUBS 2 DELETE]      |
//     |          |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqAndSubDelOkTwoParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkTwoParallel")

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()

	rparams2.Req.EventTriggerDefinition.Data.Length = 1
	rparams2.Req.EventTriggerDefinition.Data.Data = make([]uint8, rparams2.Req.EventTriggerDefinition.Data.Length)
	rparams2.Req.EventTriggerDefinition.Data.Data[0] = 2

	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)

	//Resp1
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Resp2
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn2.RecvSubsResp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	xappConn1.RecvSubsDelResp(t, deltrans1)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn2.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn2.RecvSubsDelResp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSameSubsDiffRan
// Same subscription to different RANs
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |              |              |
//     |              |              |
//     | SubReq(r1)   |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq(r1)   |
//     |              |------------->|
//     |              |              |
//     |              | SubResp(r1)  |
//     |              |<-------------|
//     |              |              |
//     | SubResp(r1)  |              |
//     |<-------------|              |
//     |              |              |
//     | SubReq(r2)   |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq(r2)   |
//     |              |------------->|
//     |              |              |
//     |              | SubResp(r2)  |
//     |              |<-------------|
//     |              |              |
//     | SubResp(r2)  |              |
//     |<-------------|              |
//     |              |              |
//     |       [SUBS r1 DELETE]      |
//     |              |              |
//     |       [SUBS r2 DELETE]      |
//     |              |              |
//
//-----------------------------------------------------------------------------
func TestSameSubsDiffRan(t *testing.T) {
	CaseBegin("TestSameSubsDiffRan")

	//Req1
	cretrans1 := xappConn1.NewRmrTransactionId("", "RAN_NAME_1")
	xappConn1.SendSubsReq(t, nil, cretrans1)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Req2
	cretrans2 := xappConn1.NewRmrTransactionId("", "RAN_NAME_2")
	xappConn1.SendSubsReq(t, nil, cretrans2)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.RecvSubsResp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.NewRmrTransactionId("", "RAN_NAME_1")
	xappConn1.SendSubsDelReq(t, deltrans1, e2SubsId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	xappConn1.RecvSubsDelResp(t, deltrans1)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn1.NewRmrTransactionId("", "RAN_NAME_2")
	xappConn1.SendSubsDelReq(t, deltrans2, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn1.RecvSubsDelResp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqRetryInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |         [SUBS DELETE]       |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqRetryInSubmgr(t *testing.T) {
	CaseBegin("TestSubReqRetryInSubmgr start")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	// Xapp: Send SubsReq
	cretrans := xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive 1st SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 2nd SubsReq and send SubsResp
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)

	// Xapp: Receive SubsResp
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqTwoRetriesNoRespSubDelRespInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//
//-----------------------------------------------------------------------------
func TestSubReqRetryNoRespSubDelRespInSubmgr(t *testing.T) {
	CaseBegin("TestSubReqTwoRetriesNoRespSubDelRespInSubmgr start")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
	})

	// Xapp: Send SubsReq
	xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive 1st SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 2nd SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Send receive SubsDelReq and send SubsResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqTwoRetriesNoRespAtAllInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqTwoRetriesNoRespAtAllInSubmgr(t *testing.T) {
	CaseBegin("TestSubReqTwoRetriesNoRespAtAllInSubmgr start")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 2},
	})

	// Xapp: Send SubsReq
	xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive 1st SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 2nd SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq
	delreq, _ := e2termConn1.RecvSubsDelReq(t)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq.RequestId.InstanceId, 15)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqSubFailRespInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubFail |
//     |              |<-------------|
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |      SubFail |              |
//     |<-------------|              |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqSubFailRespInSubmgr(t *testing.T) {
	CaseBegin("TestSubReqSubFailRespInSubmgr start")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cSubFailToXapp, 1},
	})

	// Xapp: Send SubsReq
	cretrans := xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive SubsReq and send SubsFail (first)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// E2t: Receive SubsDelReq and send SubsDelResp (internal first)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// Xapp: Receive SubsFail
	e2SubsId := xappConn1.RecvSubsFail(t, cretrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubDelReqRetryInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubDelReqRetryInSubmgr(t *testing.T) {

	CaseBegin("TestSubDelReqRetryInSubmgr start")

	// Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Subs Delete
	// Xapp: Send SubsDelReq
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Xapp: Receive SubsDelResp
	xappConn1.RecvSubsDelResp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubDelReqTwoRetriesNoRespInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubDelReqTwoRetriesNoRespInSubmgr(t *testing.T) {

	CaseBegin("TestSubDelReTwoRetriesNoRespInSubmgr start")

	// Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Subs Delete
	// Xapp: Send SubsDelReq
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// Xapp: Receive SubsDelResp
	xappConn1.RecvSubsDelResp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubDelReqSubDelFailRespInSubmgr
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |         [SUBS CREATE]       |
//     |              |              |
//     |              |              |
//     |  SubDelReq   |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelFail |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubDelReqSubDelFailRespInSubmgr(t *testing.T) {
	CaseBegin("TestSubReqSubDelFailRespInSubmgr start")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelFailFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	// Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Xapp: Send SubsDelReq
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)

	// E2t: Send receive SubsDelReq and send SubsDelFail
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Xapp: Receive SubsDelResp
	xappConn1.RecvSubsDelResp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameAction
//
//   stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |             |              |    SubResp1  |
//     |             |              |<-------------|
//     |             |    SubResp1  |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |          SubResp2          |              |
//     |<---------------------------|              |
//     |             |              |              |
//     |             | SubDelReq 1  |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             | SubDelResp 1 |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |         SubDelReq 2        |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |         SubDelResp 2       |              |
//     |<---------------------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOkSameAction(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkSameAction")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 2},
		Counter{cMergedSubscriptions, 1},
		Counter{cUnmergedSubscriptions, 1},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 2},
	})

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)
	//crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)
	//e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn2.RecvSubsResp(t, cretrans2)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId1))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560", "localhost:13660"})

	//Del1
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	//e2termConn1.RecvSubsDelReq(t)
	//e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	xappConn1.RecvSubsDelResp(t, deltrans1)
	//Wait that subs is cleaned
	//mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn2.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn2.RecvSubsDelResp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameActionParallel
//
//   stub          stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |    SubResp1  |
//     |             |              |<-------------|
//     |             |    SubResp1  |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |          SubResp2          |              |
//     |<---------------------------|              |
//     |             |              |              |
//     |             | SubDelReq 1  |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             | SubDelResp 1 |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |         SubDelReq 2        |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |         SubDelResp 2       |              |
//     |<---------------------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOkSameActionParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkSameActionParallel")

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)

	//Resp1
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Resp2
	e2SubsId2 := xappConn2.RecvSubsResp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	xappConn1.RecvSubsDelResp(t, deltrans1)

	//Del2
	deltrans2 := xappConn2.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn2.RecvSubsDelResp(t, deltrans2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelNokSameActionParallel
//
//   stub          stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |    SubFail1  |
//     |             |              |<-------------|
//     |             |              |              |
//     |             |              | SubDelReq    |
//     |             |              |------------->|
//     |             |              |   SubDelResp |
//     |             |              |<-------------|
//     |             |              |              |
//     |             |    SubFail1  |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |          SubFail2          |              |
//     |<---------------------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelNokSameActionParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelNokSameActionParallel")

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)

	// E2t: Receive SubsReq (first)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	subepcnt2 := mainCtrl.get_subs_entrypoint_cnt(t, crereq1.RequestId.InstanceId)
	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, crereq1.RequestId.InstanceId, subepcnt2, 10)

	// E2t: send SubsFail (first)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// E2t: internal delete
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	//Fail1
	e2SubsId1 := xappConn1.RecvSubsFail(t, cretrans1)
	//Fail2
	xappConn2.RecvSubsFail(t, cretrans2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 15)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelNoAnswerSameActionParallel
//
//   stub          stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |             | SubReq2      |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              |              |
//     |             |              | SubDelReq    |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              |   SubDelResp |
//     |             |              |<-------------|
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelNoAnswerSameActionParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelNoAnswerSameActionParallel")

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	xappConn1.SendSubsReq(t, rparams1, nil)

	crereq1, _ := e2termConn1.RecvSubsReq(t)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	subepcnt2 := mainCtrl.get_subs_entrypoint_cnt(t, crereq1.RequestId.InstanceId)
	xappConn2.SendSubsReq(t, rparams2, nil)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, crereq1.RequestId.InstanceId, subepcnt2, 10)

	//Req1 (retransmitted)
	e2termConn1.RecvSubsReq(t)

	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq1.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 15)
}

//-----------------------------  Policy cases ---------------------------------
//-----------------------------------------------------------------------------
// TestSubReqPolicyAndSubDelOk
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqPolicyAndSubDelOk(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOk")

	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	rparams1.Req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypePolicy
	cretrans := xappConn1.SendSubsReq(t, rparams1, nil)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqPolicyChangeAndSubDelOk
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubReqPolicyChangeAndSubDelOk(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOk")

	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	rparams1.Req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypePolicy
	cretrans := xappConn1.SendSubsReq(t, rparams1, nil)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	//Policy change
	rparams1.Req.RequestId.InstanceId = e2SubsId
	rparams1.Req.ActionSetups[0].SubsequentAction.TimetoWait = e2ap.E2AP_TimeToWaitW200ms
	xappConn1.SendSubsReq(t, rparams1, cretrans)

	crereq, cremsg = e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId = xappConn1.RecvSubsResp(t, cretrans)
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkTwoE2termParallel
//
//   stub                          stub           stub
// +-------+     +---------+    +---------+    +---------+
// | xapp  |     | submgr  |    | e2term1 |    | e2term2 |
// +-------+     +---------+    +---------+    +---------+
//     |              |              |              |
//     |              |              |              |
//     |              |              |              |
//     | SubReq1      |              |              |
//     |------------->|              |              |
//     |              |              |              |
//     |              | SubReq1      |              |
//     |              |------------->|              |
//     |              |              |              |
//     | SubReq2      |              |              |
//     |------------->|              |              |
//     |              |              |              |
//     |              | SubReq2      |              |
//     |              |---------------------------->|
//     |              |              |              |
//     |              |    SubResp1  |              |
//     |              |<-------------|              |
//     |    SubResp1  |              |              |
//     |<-------------|              |              |
//     |              |    SubResp2  |              |
//     |              |<----------------------------|
//     |    SubResp2  |              |              |
//     |<-------------|              |              |
//     |              |              |              |
//     |        [SUBS 1 DELETE]      |              |
//     |              |              |              |
//     |        [SUBS 2 DELETE]      |              |
//     |              |              |              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOkTwoE2termParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkTwoE2termParallel")

	//Req1
	cretrans1 := xappConn1.NewRmrTransactionId("", "RAN_NAME_1")
	xappConn1.SendSubsReq(t, nil, cretrans1)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	cretrans2 := xappConn1.NewRmrTransactionId("", "RAN_NAME_11")
	xappConn1.SendSubsReq(t, nil, cretrans2)
	crereq2, cremsg2 := e2termConn2.RecvSubsReq(t)

	//Resp1
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Resp2
	e2termConn2.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.RecvSubsResp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	xappConn1.RecvSubsDelResp(t, deltrans1)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn1.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn2.RecvSubsDelReq(t)
	e2termConn2.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn1.RecvSubsDelResp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	e2termConn2.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqInsertAndSubDelOk
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |              |              |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqInsertAndSubDelOk(t *testing.T) {
	CaseBegin("TestInsertSubReqAndSubDelOk")

	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	rparams1.Req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypeInsert
	cretrans := xappConn1.SendSubsReq(t, rparams1, nil)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqRetransmissionWithSameSubIdDiffXid
//
// This case simulates case where xApp restarts and starts sending same
// subscription requests which have already subscribed successfully

//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |  SubReq      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     | xApp restart |              |
//     |              |              |
//     |  SubReq      |              |
//     | (retrans with same xApp generated subid but diff xid)
//     |------------->|              |
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |         [SUBS DELETE]       |
//     |              |              |
//
//-----------------------------------------------------------------------------
func TestSubReqRetransmissionWithSameSubIdDiffXid(t *testing.T) {
	CaseBegin("TestSubReqRetransmissionWithSameSubIdDiffXid")

	//Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// xApp restart here
	// --> artificial delay
	<-time.After(1 * time.Second)

	//Subs Create
	cretrans = xappConn1.SendSubsReq(t, nil, nil) //Retransmitted SubReq
	e2SubsId = xappConn1.RecvSubsResp(t, cretrans)

	//Subs Delete
	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqNokAndSubDelOkWithRestartInMiddle
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |                        <----|
//     |                             |
//     |        Submgr restart       |
//     |                             |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqNokAndSubDelOkWithRestartInMiddle(t *testing.T) {
	CaseBegin("TestSubReqNokAndSubDelOkWithRestartInMiddle")

	// Remove possible existing subscrition
	mainCtrl.removeExistingSubscriptions(t)

	mainCtrl.SetResetTestFlag(t, true) // subs.DoNotWaitSubResp will be set TRUE for the subscription
	xappConn1.SendSubsReq(t, nil, nil)
	e2termConn1.RecvSubsReq(t)
	mainCtrl.SetResetTestFlag(t, false)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})
	e2SubsId := uint32(resp[0].SubscriptionID)
	t.Logf("e2SubsId = %v", e2SubsId)

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkWithRestartInMiddle
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | SubReq       |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |              |<-------------|
//     |              |              |
//     |      SubResp |              |
//     |<-------------|              |
//     |              |              |
//     |                             |
//     |        Submgr restart       |
//     |                             |
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------

func TestSubReqAndSubDelOkWithRestartInMiddle(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkWithRestartInMiddle")

	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Check subscription
	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// Check that subscription is restored correctly after restart
	resp, _ = xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560"})

	deltrans := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	xappConn1.RecvSubsDelResp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameActionWithRestartsInMiddle
//
//   stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |             |              |    SubResp1  |
//     |             |              |<-------------|
//     |             |    SubResp1  |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |                                           |
//     |              submgr restart               |
//     |                                           |
//     |             |              |              |
//     |             |              |              |
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |          SubResp2          |              |
//     |<---------------------------|              |
//     |             |              |              |
//     |             | SubDelReq 1  |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             | SubDelResp 1 |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |             |              |              |
//     |                                           |
//     |              submgr restart               |
//     |                                           |
//     |             |              |              |
//     |         SubDelReq 2        |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |         SubDelResp 2       |              |
//     |<---------------------------|              |
//
//-----------------------------------------------------------------------------

func TestSubReqAndSubDelOkSameActionWithRestartsInMiddle(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkSameActionWithRestartsInMiddle")

	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.RecvSubsResp(t, cretrans1)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	cretrans2 := xappConn2.SendSubsReq(t, rparams2, nil)
	e2SubsId2 := xappConn2.RecvSubsResp(t, cretrans2)

	// Check subscription
	resp, _ := xapp.Subscription.QuerySubscriptions() ////////////////////////////////
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId1))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560", "localhost:13660"})

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// Check that subscription is restored correctly after restart
	resp, _ = xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId1))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].ClientEndpoint, []string{"localhost:13560", "localhost:13660"})

	//Del1
	deltrans1 := xappConn1.SendSubsDelReq(t, nil, e2SubsId1)
	xapp.Logger.Debug("xappConn1.RecvSubsDelResp")
	xappConn1.RecvSubsDelResp(t, deltrans1)
	xapp.Logger.Debug("xappConn1.RecvSubsDelResp received")

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	//Del2
	deltrans2 := xappConn2.SendSubsDelReq(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)

	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn2.RecvSubsDelResp(t, deltrans2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//*****************************************************************************
//  REST interface test cases
//*****************************************************************************

//-----------------------------------------------------------------------------
// Test debug GET and POST requests
//
//   curl
// +-------+     +---------+
// | user  |     | submgr  |
// +-------+     +---------+
//     |              |
//     | GET/POST Req |
//     |------------->|
//     |         Resp |
//     |<-------------|
//     |              |
func TestGetSubscriptions(t *testing.T) {

	mainCtrl.sendGetRequest(t, "localhost:8088", "/ric/v1/subscriptions")
}

func TestGetSymptomData(t *testing.T) {

	mainCtrl.sendGetRequest(t, "localhost:8080", "/ric/v1/symptomdata")
}

func TestPostdeleteSubId(t *testing.T) {

	mainCtrl.sendPostRequest(t, "localhost:8080", "/ric/v1/test/deletesubid=1")
}

func TestPostEmptyDb(t *testing.T) {

	mainCtrl.sendPostRequest(t, "localhost:8080", "/ric/v1/test/emptydb")
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndRouteNok
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | rtmgr   |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | RouteCreate  |
//     |                 |------------->|
//     |                 |              |
//     |                 | RouteCreate  |
//     |                 |  status:400  |
//     |                 |(Bad request) |
//     |                 |<-------------|
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     |          [SUBS INT DELETE]     |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndRouteNok(t *testing.T) {
	CaseBegin("TestRESTSubReqAndRouteNok")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cRouteCreateFail, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	const subReqCount int = 1
	const parameterSet = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	// Add delay for rtmgt HTTP handling so that HTTP response is received before notify on XAPP side
	waiter := rtmgrHttp.AllocNextSleep(50, false)
	newSubsId := mainCtrl.get_registry_next_subid(t)

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xappConn1.ExpectRESTNotification(t, restSubId)
	waiter.WaitResult(t)

	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, newSubsId, 10)
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
}

func TestRESTSubReqAndRouteUpdateNok(t *testing.T) {
	CaseBegin("TestSubReqAndRouteUpdateNok")

	//Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 2},
		Counter{cRouteCreateUpdateFail, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	// xapp2 ROUTE creation shall fail with  400 from rtmgr -> submgr
	waiter := rtmgrHttp.AllocNextEvent(false)
	newSubsId := mainCtrl.get_registry_next_subid(t)
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_1")
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for second subscriber : %v", restSubId2)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	waiter.WaitResult(t)
	// e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2) - TOD: missing delete
	xappConn2.WaitRESTNotification(t, restSubId2)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	mainCtrl.wait_subs_clean(t, newSubsId, 10)
	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

func TestRESTSubDelReqAndRouteDeleteNok(t *testing.T) {
	CaseBegin("TestRESTSubDelReqAndRouteDeleteNok")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cRouteDeleteFail, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	waiter := rtmgrHttp.AllocNextEvent(false)
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	waiter.WaitResult(t)

	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

func TestRESTSubMergeDelAndRouteUpdateNok(t *testing.T) {
	CaseBegin("TestRESTSubMergeDelAndRouteUpdateNok")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cRouteDeleteUpdateFail, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 2},
	})

	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})
	restSubId2, e2SubsId2 := createXapp2MergedSubscription(t, "RAN_NAME_1")

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560", "localhost:13660"})

	//Del1, this shall fail on rtmgr side
	waiter := rtmgrHttp.AllocNextEvent(false)
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	waiter.WaitResult(t)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13660"})

	//Del2
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqRetransmission
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq1     |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq1      |
//     |                 |------------->|
//     |                 |              |
//     | RESTSubReq2     |              |
//     | (retrans)       |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubReq2      |
//     |                 |------------->|
//     |    RESTSubResp2 |              |
//     |<----------------|              |
//     |                 |     SubResp1 |
//     |                 |<-------------|
//     |      RESTNotif1 |              |
//     |<----------------|              |
//     |                 |     SubResp1 |
//     |                 |<-------------|
//     |      RESTNotif2 |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqRetransmission(t *testing.T) {
	CaseBegin("TestRESTSubReqRetransmission")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})
	// Retry/duplicate will get the same way as the first request.  Submgr cannot detect duplicate RESTRequests
	// Contianed duplicate messages from same xapp will not be merged. Here we use xappConn2 to simulate sending
	// second request from same xapp as doing it from xappConn1 would not work as notification would not be received

	// Subs Create
	const subReqCount int = 1
	const parameterSet = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	// In order to force both XAPP's to create their own subscriptions, force rtmgr to block a while so that 2nd create
	// gets into execution before the rtmgrg responds for the first one.
	waiter := rtmgrHttp.AllocNextSleep(10, true)
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params)
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)

	waiter.WaitResult(t)

	xappConn1.WaitListedRestNotifications(t, []string{restSubId1, restSubId2})

	// Depending one goroutine scheduling order, we cannot say for sure which xapp reaches e2term first. Thus
	// the order is not significant he6re.
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	crereq, cremsg = e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)

	e2SubsIdA := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Info("TEST: 1.st XAPP notification received e2SubsId=%v", e2SubsIdA)
	e2SubsIdB := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Info("TEST: 2.nd XAPP notification received e2SubsId=%v", e2SubsIdB)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// Del2
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)

	mainCtrl.wait_multi_subs_clean(t, []uint32{e2SubsIdA.E2SubsId, e2SubsIdB.E2SubsId}, 10)

	waitSubsCleanup(t, e2SubsIdB.E2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

func TestRESTSubDelReqRetransmission(t *testing.T) {
	CaseBegin("TestRESTSubDelReqRetransmission")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	//Subs Delete
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	seqBef := mainCtrl.get_msgcounter(t)
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqDelReq
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |  RESTSubDelResp |              |
//     |     unsuccess   |              |
//     |<----------------|              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |      RESTNotif1 |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqDelReq(t *testing.T) {
	CaseBegin("TestRESTSubReqDelReq")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	const subReqCount int = 1
	const parameterSet = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	// Del. This will fail as processing of the subscription
	// is still ongoing in submgr. Deletion is not allowed before
	// subscription creation has been completed.
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	// Retry del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)

}

func TestRESTSubDelReqCollision(t *testing.T) {
	CaseBegin("TestRESTSubDelReqCollision - not relevant for REST API")
}

func TestRESTSubReqAndSubDelOkTwoParallel(t *testing.T) {
	CaseBegin("TestRESTSubReqAndSubDelOkTwoParallel")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	//Req1
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send 1st REST subscriber request for subscriberId : %v", restSubId1)

	//Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send 2nd REST subscriber request for subscriberId : %v", restSubId2)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)

	//XappConn1 receives both of the  responses
	xappConn1.WaitListedRestNotifications(t, []string{restSubId1, restSubId2})

	//Resp1
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	//Resp2
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)

	e2SubsIdA := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Info("TEST: 1.st XAPP notification received e2SubsId=%v", e2SubsIdA)
	e2SubsIdB := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Info("TEST: 2.nd XAPP notification received e2SubsId=%v", e2SubsIdB)

	//Del1
	deleteSubscription(t, xappConn1, e2termConn1, &restSubId1)
	//Del2
	deleteSubscription(t, xappConn2, e2termConn1, &restSubId2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsIdA.E2SubsId, 10)
	waitSubsCleanup(t, e2SubsIdB.E2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)

}

func TestRESTSameSubsDiffRan(t *testing.T) {
	CaseBegin("TestRESTSameSubsDiffRan")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send 1st REST subscriber request for subscriberId : %v", restSubId1)

	params = xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_2")
	restSubId2, e2SubsId2 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send 2nd REST subscriber request for subscriberId : %v", restSubId2)

	//Del1
	deleteSubscription(t, xappConn1, e2termConn1, &restSubId1)
	//Del2
	deleteSubscription(t, xappConn1, e2termConn1, &restSubId2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)

}

func TestRESTSubReqRetryInSubmgr(t *testing.T) {
	CaseBegin("TestRESTSubReqRetryInSubmgr start")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	xapp.Logger.Info("Send REST subscriber request for subscriber : %v", restSubId)

	// Catch the first message and ignore it
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xapp.Logger.Info("Ignore REST subscriber request for subscriber : %v", restSubId)

	// The second request is being handled normally
	crereq, cremsg = e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)

}

//-----------------------------------------------------------------------------
// TestRESTSubReqTwoRetriesNoRespSubDelRespInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |       RESTNotif |              |
//     |       unsuccess |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqRetryNoRespSubDelRespInSubmgr(t *testing.T) {
	CaseBegin("TestRESTSubReqTwoRetriesNoRespSubDelRespInSubmgr start")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Info("Ignore 1st REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Info("Ignore 2nd REST subscriber request for subscriber : %v", restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	// e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)	- TODO:  Should we delete this?
	xappConn1.WaitRESTNotification(t, restSubId)

	// Wait that subs is cleaned
	waitSubsCleanup(t, delreq.RequestId.InstanceId, 10)

	mainCtrl.VerifyCounterValues(t)
}

func TestREST2eTermNotRespondingToSubReq(t *testing.T) {
	CaseBegin("TestREST2eTermNotRespondingToSubReq start")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 2},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Info("Ignore 1st REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Info("Ignore 2nd REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsDelReq(t)
	xapp.Logger.Info("Ignore 1st INTERNAL delete request for subscriber : %v", restSubId)

	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.RecvSubsDelReq(t)
	xapp.Logger.Info("Ignore 2nd INTERNAL delete request for subscriber : %v", restSubId)

	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)

}

//-----------------------------------------------------------------------------
// TestRESTSubReqTwoRetriesNoRespSubDelRespInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |       RESTNotif |              |
//     |       unsuccess |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqTwoRetriesNoRespAtAllInSubmgr(t *testing.T) {
	CaseBegin("TestRESTSubReqTwoRetriesNoRespAtAllInSubmgr start")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubReReqToE2, 1},
		Counter{cSubReqTimerExpiry, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelReqTimerExpiry, 2},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Info("Ignore 1st REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)
	xapp.Logger.Info("Ignore 2nd REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsDelReq(t)
	xapp.Logger.Info("Ignore 1st INTERNAL delete request for subscriber : %v", restSubId)

	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.RecvSubsDelReq(t)
	xapp.Logger.Info("Ignore 2nd INTERNAL delete request for subscriber : %v", restSubId)

	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)

	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqSubFailRespInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail |
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |       unsuccess |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqSubFailRespInSubmgr(t *testing.T) {
	CaseBegin("TestRESTSubReqSubFailRespInSubmgr")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
	})

	const subReqCount int = 1
	const parameterSet = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// REST subscription sill there to be deleted
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)

}

//-----------------------------------------------------------------------------
// TestRESTSubDelReqRetryInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubDelReqRetryInSubmgr(t *testing.T) {
	CaseBegin("TestRESTSubDelReqRetryInSubmgr")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})
	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubDelReqTwoRetriesNoRespInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubDelReqTwoRetriesNoRespInSubmgr(t *testing.T) {
	CaseBegin("TestRESTSubDelReTwoRetriesNoRespInSubmgr")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelReReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubDelReqSubDelFailRespInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail |
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubDelReqSubDelFailRespInSubmgr(t *testing.T) {
	CaseBegin("TestRESTSubDelReqSubDelFailRespInSubmgr")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelFailFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Send receive SubsDelReq and send SubsDelFail
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelOkSameAction
//
//   stub                             stub
// +-------+     +-------+        +---------+    +---------+
// | xapp2 |     | xapp1 |        | submgr  |    | e2term  |
// +-------+     +-------+        +---------+    +---------+
//     |             |                 |              |
//     |             | RESTSubReq1     |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             |    RESTSubResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     |             |                 | SubReq1      |
//     |             |                 |------------->|
//     |             |                 |    SubResp1  |
//     |             |                 |<-------------|
//     |             |      RESTNotif1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     | RESTSubReq2                   |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |                  RESTSubResp2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             |      RESTNotif2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             | RESTSubDelReq1  |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             | RESTSubDelResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     | RESTSubDelReq2                |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             |                 | SubDelReq2   |
//     |             |                 |------------->|
//     |             |                 |              |
//     |             |                 |  SubDelResp2 |
//     |             |                 |<-------------|
//     |             |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndSubDelOkSameAction(t *testing.T) {
	CaseBegin("TestRESTSubReqAndSubDelOkSameAction")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 2},
		Counter{cMergedSubscriptions, 1},
		Counter{cUnmergedSubscriptions, 1},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_1")

	xapp.Subscription.SetResponseCB(xappConn2.SubscriptionRespHandler)
	xappConn2.WaitRESTNotificationForAnySubscriptionId(t)
	waiter := rtmgrHttp.AllocNextSleep(10, true)
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	waiter.WaitResult(t)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId2)
	e2SubsId2 := <-xappConn2.RESTNotification
	xapp.Logger.Info("REST notification received e2SubsId=%v", e2SubsId2)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560", "localhost:13660"})

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)

	// Del2
	deleteXapp2Subscription(t, &restSubId2)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameActionParallel
//
//   stub          stub                          stub
// +-------+     +-------+     +---------+    +---------+
// | xapp2 |     | xapp1 |     | submgr  |    | e2term  |
// +-------+     +-------+     +---------+    +---------+
//     |             |              |              |
//     |             |              |              |
//     |             |              |              |
//     |             | SubReq1      |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             |              | SubReq1      |
//     |             |              |------------->|
//     |          SubReq2           |              |
//     |--------------------------->|              |
//     |             |              |    SubResp1  |
//     |             |              |<-------------|
//     |             |    SubResp1  |              |
//     |             |<-------------|              |
//     |             |              | SubReq2      |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              |    SubResp2  |
//     |             |              |<-------------|
//     |          SubResp2          |              |
//     |<---------------------------|              |
//     |             |              |              |
//     |             | SubDelReq 1  |              |
//     |             |------------->|              |
//     |             |              |              |
//     |             | SubDelResp 1 |              |
//     |             |<-------------|              |
//     |             |              |              |
//     |         SubDelReq 2        |              |
//     |--------------------------->|              |
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |             |              | SubDelReq 2  |
//     |             |              |------------->|
//     |             |              |              |
//     |         SubDelResp 2       |              |
//     |<---------------------------|              |
//
func TestRESTSubReqAndSubDelOkSameActionParallel(t *testing.T) {
	CaseBegin("TestRESTSubReqAndSubDelOkSameActionParallel")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	params2 := xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId2 := xappConn2.SendRESTSubsReq(t, params2)

	xappConn1.ExpectRESTNotification(t, restSubId1)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.WaitRESTNotification(t, restSubId1)

	xappConn2.ExpectRESTNotification(t, restSubId2)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	// Del2
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)

	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelNoAnswerSameActionParallel
//
//   stub          stub                             stub
// +-------+     +-------+        +---------+    +---------+
// | xapp2 |     | xapp1 |        | submgr  |    | e2term  |
// +-------+     +-------+        +---------+    +---------+
//     |             |                 |              |
//     |             |                 |              |
//     |             |                 |              |
//     |             | RESTSubReq1     |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             |    RESTSubResp1 |              |
//     |             |<----------------|              |
//     |             |                 | SubReq1      |
//     |             |                 |------------->|
//     | RESTSubReq2                   |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//     |             |                 | SubReq1      |
//     |             |                 |------------->|
//     |             |                 |              |
//     |             |                 |              |
//     |             |                 | SubDelReq    |
//     |             |                 |------------->|
//     |             |                 |              |
//     |             |                 |   SubDelResp |
//     |             |                 |<-------------|
//     |             |      RESTNotif1 |              |
//     |             |       unsuccess |              |
//     |             |<----------------|              |
//     |                    RESTNotif2 |              |
//     |             |       unsuccess |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             | RESTSubDelReq1  |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             | RESTSubDelResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     | RESTSubDelReq2                |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndSubDelNoAnswerSameActionParallel(t *testing.T) {
	CaseBegin("TestRESTSubReqAndSubDelNoAnswerSameActionParallel")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 2},
	})

	const subReqCount int = 1
	const parameterSet = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	// Req1
	params1 := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params1)
	crereq1, _ := e2termConn1.RecvSubsReq(t)

	// Req2
	subepcnt2 := mainCtrl.get_subs_entrypoint_cnt(t, crereq1.RequestId.InstanceId)
	params2 := xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params2.SetMeid("RAN_NAME_1")
	restSubId2 := xappConn2.SendRESTSubsReq(t, params2)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, crereq1.RequestId.InstanceId, subepcnt2, 10)

	//Req1 (retransmitted)
	e2termConn1.RecvSubsReq(t)

	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)

	xappConn1.WaitListedRestNotifications(t, []string{restSubId1, restSubId2})
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	e2SubsIdA := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Info("TEST: 1.st XAPP notification received e2SubsId=%v", e2SubsIdA)
	e2SubsIdB := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Info("TEST: 2.nd XAPP notification received e2SubsId=%v", e2SubsIdB)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)

	// Del2
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)

	mainCtrl.wait_multi_subs_clean(t, []uint32{e2SubsIdA.E2SubsId, e2SubsIdB.E2SubsId}, 10)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsIdA.E2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelNokSameActionParallel
//
//   stub          stub                             stub
// +-------+     +-------+        +---------+    +---------+
// | xapp2 |     | xapp1 |        | submgr  |    | e2term  |
// +-------+     +-------+        +---------+    +---------+
//     |             |                 |              |
//     |             |                 |              |
//     |             |                 |              |
//     |             | RESTSubReq1     |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             |    RESTSubResp1 |              |
//     |             |<----------------|              |
//     |             |                 | SubReq1      |
//     |             |                 |------------->|
//     | RESTSubReq2                   |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//     |             |                 |    SubFail1  |
//     |             |                 |<-------------|
//     |             |                 |              |
//     |             |      RESTNotif1 |              |
//     |             |       unsuccess |              |
//     |             |<----------------|              |
//     |                    RESTNotif2 |              |
//     |             |       unsuccess |              |
//     |<------------------------------|              |
//     |             |                 | SubDelReq    |
//     |             |                 |------------->|
//     |             |                 |   SubDelResp |
//     |             |                 |<-------------|
//     |             |                 |              |
//     |             | RESTSubDelReq1  |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             | RESTSubDelResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     | RESTSubDelReq2                |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndSubDelNokSameActionParallel(t *testing.T) {
	CaseBegin("TestRESTSubReqAndSubDelNokSameActionParallel")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubFailFromE2, 1},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 2},
	})

	const subReqCount int = 1
	const parameterSet = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	// Req1
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	// Req2
	subepcnt2 := mainCtrl.get_subs_entrypoint_cnt(t, crereq1.RequestId.InstanceId)
	params2 := xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params2.SetMeid("RAN_NAME_1")
	restSubId2 := xappConn2.SendRESTSubsReq(t, params2)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, crereq1.RequestId.InstanceId, subepcnt2, 10)

	// E2t: send SubsFail (first)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// E2t: internal delete
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	xappConn1.WaitListedRestNotifications(t, []string{restSubId1, restSubId2})
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	e2SubsIdA := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Info("TEST: 1.st XAPP notification received e2SubsId=%v", e2SubsIdA)
	e2SubsIdB := <-xappConn1.ListedRESTNotifications
	xapp.Logger.Info("TEST: 2.nd XAPP notification received e2SubsId=%v", e2SubsIdB)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)

	// Del2
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsIdA.E2SubsId, 10)
	waitSubsCleanup(t, e2SubsIdB.E2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

func TestRESTSubReqPolicyAndSubDelOk(t *testing.T) {
	CaseBegin("TestRESTSubReqPolicyAndSubDelOk")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	const subReqCount int = 1
	const parameterSet = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount, actionDefinitionPresent, policyParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST Policy subscriber request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("REST notification received e2SubsId=%v", e2SubsId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqPolicyChangeAndSubDelOk
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqPolicyChangeAndSubDelOk(t *testing.T) {
	CaseBegin("TestRESTSubReqPolicyAndSubDelOk")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	const subReqCount int = 1
	const parameterSet = 1
	const actionDefinitionPresent bool = true
	const policyParamCount int = 1

	// Req
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount, actionDefinitionPresent, policyParamCount)
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Policy change
	instanceId := int64(e2SubsId)
	// GetRESTSubsReqPolicyParams sets some coutners on tc side.
	params = xappConn1.GetRESTSubsReqPolicyParams(subReqCount, actionDefinitionPresent, policyParamCount)
	params.SubsReqParams.SubscriptionDetails[0].InstanceID = &instanceId
	params.SetTimeToWait("w200ms")
	restSubId, e2SubsId = createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelOkTwoE2termParallel
//
//   stub                             stub           stub
// +-------+        +---------+    +---------+    +---------+
// | xapp  |        | submgr  |    | e2term1 |    | e2term2 |
// +-------+        +---------+    +---------+    +---------+
//     |                 |              |              |
//     |                 |              |              |
//     |                 |              |              |
//     | RESTSubReq1     |              |              |
//     |---------------->|              |              |
//     |                 |              |              |
//     |    RESTSubResp1 |              |              |
//     |<----------------|              |              |
//     |                 | SubReq1      |              |
//     |                 |------------->|              |
//     |                 |              |              |
//     | RESTSubReq2     |              |              |
//     |---------------->|              |              |
//     |                 |              |              |
//     |    RESTSubResp2 |              |              |
//     |<----------------|              |              |
//     |                 | SubReq2      |              |
//     |                 |---------------------------->|
//     |                 |              |              |
//     |                 |    SubResp1  |              |
//     |                 |<-------------|              |
//     |      RESTNotif1 |              |              |
//     |<----------------|              |              |
//     |                 |    SubResp2  |              |
//     |                 |<----------------------------|
//     |      RESTNotif2 |              |              |
//     |<----------------|              |              |
//     |                 |              |              |
//     |           [SUBS 1 DELETE]      |              |
//     |                 |              |              |
//     |           [SUBS 2 DELETE]      |              |
//     |                 |              |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndSubDelOkTwoE2termParallel(t *testing.T) {
	CaseBegin("TestRESTSubReqAndSubDelOkTwoE2termParallel")

	// Init counter check
	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	const subReqCount int = 1
	const parameterSet = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	// Req1
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId1 := xappConn1.SendRESTSubsReq(t, params)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_11")
	// Here we use xappConn2 to simulate sending second request from same xapp as doing it from xappConn1
	// would not work as notification would not be received
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	crereq2, cremsg2 := e2termConn2.RecvSubsReq(t)

	// Resp1
	xappConn1.ExpectRESTNotification(t, restSubId1)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.WaitRESTNotification(t, restSubId1)
	xapp.Logger.Info("TEST: REST notification received e2SubsId1=%v", e2SubsId1)

	// Resp2
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn2.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)
	xapp.Logger.Info("TEST: REST notification received e2SubsId2=%v", e2SubsId2)

	// Delete1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	// Delete2
	xappConn1.SendRESTSubsDelReq(t, &restSubId2)
	delreq2, delmsg2 := e2termConn2.RecvSubsDelReq(t)
	e2termConn2.SendSubsDelResp(t, delreq2, delmsg2)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAsn1EncodeFail
//
// In this case submgr send RICSubscriptionDeleteRequest after encode failure which should not happen!
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |  RESTSubDelResp |              |
//     |     unsuccess   |              |
//     |<----------------|              |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAsn1EncodeFail(t *testing.T) {
	CaseBegin("TestRESTSubReqAsn1EncodeFail")

	xapp.Logger.Info("Xapp-frame, v0.8.1 sufficient REST API validation")

}

//-----------------------------------------------------------------------------
// TestRESTSubReqInsertAndSubDelOk
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |       ...       |     ...      |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |   RESTSubDelResp|              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqInsertAndSubDelOk(t *testing.T) {
	CaseBegin("TestRESTInsertSubReqAndSubDelOk")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	const subReqCount int = 1
	const parameterSet int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetSubActionTypes("insert")

	// Req
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)
	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqNokAndSubDelOkWithRestartInMiddle
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     | RESTSubReq   |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq       |
//     |              |------------->|
//     |              |              |
//     |              |      SubResp |
//     |                        <----|
//     |                             |
//     |        Submgr restart       |
//     |                             |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqNokAndSubDelOkWithRestartInMiddle(t *testing.T) {
	CaseBegin("TestRESTSubReqNokAndSubDelOkWithRestartInMiddle")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
	})

	const subReqCount int = 1
	const parameterSet = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	// Remove possible existing subscription
	mainCtrl.removeExistingSubscriptions(t)

	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)

	//Req
	mainCtrl.SetResetTestFlag(t, true) // subs.DoNotWaitSubResp will be set TRUE for the subscription
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriber : %v", restSubId)

	e2termConn1.RecvSubsReq(t)

	mainCtrl.SetResetTestFlag(t, false)

	mainCtrl.SimulateRestart(t)
	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	//Del
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelOkWithRestartInMiddle
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RESTSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     |                                |
//     |           Submgr restart       |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndSubDelOkWithRestartInMiddle(t *testing.T) {
	CaseBegin("TestRESTSubReqAndSubDelOkWithRestartInMiddle")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 1},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespToXapp, 1},
	})

	// Remove possible existing subscription
	mainCtrl.removeExistingSubscriptions(t)

	var params *teststube2ap.RESTSubsReqParams = nil

	// Create subscription
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send REST subscriber request for subscriber : %v", restSubId)

	// Check subscription
	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	// When SDL support for the REST Interface is added
	// the submgr restart statement below should be removed
	// from the comment.

	//	mainCtrl.SimulateRestart(t)
	//	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// Check subscription
	queryXappSubscription(t, int64(e2SubsId), "RAN_NAME_1", []string{"localhost:13560"})

	// Delete subscription
	deleteSubscription(t, xappConn1, e2termConn1, &restSubId)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelOkSameActionWithRestartsInMiddle
//
//   stub                             stub
// +-------+     +-------+        +---------+    +---------+
// | xapp2 |     | xapp1 |        | submgr  |    | e2term  |
// +-------+     +-------+        +---------+    +---------+
//     |             |                 |              |
//     |             | RESTSubReq1     |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             |    RESTSubResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     |             |                 | SubReq1      |
//     |             |                 |------------->|
//     |             |                 |    SubResp1  |
//     |             |                 |<-------------|
//     |             |      RESTNotif1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     | RESTSubReq2                   |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |                  RESTSubResp2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             |      RESTNotif2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             |           Submgr restart       |
//     |             |                 |              |
//     |             | RESTSubDelReq1  |              |
//     |             |---------------->|              |
//     |             |                 |              |
//     |             | RESTSubDelResp1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     |             |           Submgr restart       |
//     |             |                 |              |
//     | RESTSubDelReq2                |              |
//     |------------------------------>|              |
//     |             |                 |              |
//     |               RESTSubDelResp2 |              |
//     |<------------------------------|              |
//     |             |                 |              |
//     |             |                 | SubDelReq2   |
//     |             |                 |------------->|
//     |             |                 |              |
//     |             |                 |  SubDelResp2 |
//     |             |                 |<-------------|
//     |             |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqAndSubDelOkSameActionWithRestartsInMiddle(t *testing.T) {
	CaseBegin("TestRESTSubReqAndSubDelOkSameActionWithRestartsInMiddle")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 1},
		Counter{cSubRespFromE2, 1},
		Counter{cSubRespToXapp, 2},
		Counter{cMergedSubscriptions, 1},
		Counter{cUnmergedSubscriptions, 1},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 1},
		Counter{cSubDelRespFromE2, 1},
		Counter{cSubDelRespToXapp, 2},
	})

	// Remove possible existing subscription
	mainCtrl.removeExistingSubscriptions(t)

	var params *teststube2ap.RESTSubsReqParams = nil

	// Create subscription 1
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send REST subscriber request for subscriber 1 : %v", restSubId1)

	// Create subscription 2 with same action
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_1")
	xapp.Subscription.SetResponseCB(xappConn2.SubscriptionRespHandler)
	xappConn2.WaitRESTNotificationForAnySubscriptionId(t)
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId2)
	e2SubsId2 := <-xappConn2.RESTNotification
	xapp.Logger.Info("REST notification received e2SubsId=%v", e2SubsId2)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560", "localhost:13660"})

	// When SDL support for the REST Interface is added
	// the submgr restart statement below should be removed
	// from the comment.

	//	mainCtrl.SimulateRestart(t)
	//	xapp.Logger.Debug("mainCtrl.SimulateRestart done")

	// Delete subscription 1, and wait until it has removed the first endpoint
	subepcnt := mainCtrl.get_subs_entrypoint_cnt(t, e2SubsId1)
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, e2SubsId1, subepcnt, 10)

	// When SDL support for the REST Interface is added
	// the submgr restart statement below should be removed
	// from the comment.

	//	mainCtrl.SimulateRestart(t)
	//	xapp.Logger.Debug("mainCtrl.SimulateRestart done")
	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13660"})

	// Delete subscription 2
	deleteXapp2Subscription(t, &restSubId2)

	//Wait that subs is cleaned
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)
}

//-----------------------------------------------------------------------------
// TestRESTReportSubReqAndSubDelOk
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     |<----------------|              |
//     |       ...       |     ...      |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |   RESTSubDelResp|              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------
func TestRESTReportSubReqAndSubDelOk(t *testing.T) {
	CaseBegin("TestRESTReportSubReqAndSubDelOk")
	subReqCount := 1
	parameterSet := 1 // E2SM-gNB-X2
	actionDefinitionPresent := true
	actionParamCount := 1
	testIndex := 1
	RESTReportSubReqAndSubDelOk(t, subReqCount, parameterSet, actionDefinitionPresent, actionParamCount, testIndex)
}

func RESTReportSubReqAndSubDelOk(t *testing.T, subReqCount int, parameterSet int, actionDefinitionPresent bool, actionParamCount int, testIndex int) {
	xapp.Logger.Info("TEST: TestRESTReportSubReqAndSubDelOk with parameter set %v", testIndex)

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	var e2SubsId []uint32
	for i := 0; i < subReqCount; i++ {
		crereq, cremsg := e2termConn1.RecvSubsReq(t)
		xappConn1.ExpectRESTNotification(t, restSubId)

		e2termConn1.SendSubsResp(t, crereq, cremsg)
		instanceId := xappConn1.WaitRESTNotification(t, restSubId)
		xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", instanceId)
		e2SubsId = append(e2SubsId, instanceId)
		resp, _ := xapp.Subscription.QuerySubscriptions()
		assert.Equal(t, resp[i].SubscriptionID, (int64)(instanceId))
		assert.Equal(t, resp[i].Meid, "RAN_NAME_1")
		assert.Equal(t, resp[i].ClientEndpoint, []string{"localhost:13560"})

	}

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	for i := 0; i < subReqCount; i++ {
		delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
		e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	}

	// Wait that subs is cleaned
	for i := 0; i < subReqCount; i++ {
		mainCtrl.wait_subs_clean(t, e2SubsId[i], 10)
	}

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

/*
func TestRESTPolicySubReqAndSubDelOk(t *testing.T) {
	CaseBegin("TestRESTPolicySubReqAndSubDelOk")

	subReqCount := 2
	actionDefinitionPresent := true
	policyParamCount := 1
	testIndex := 1
	RESTPolicySubReqAndSubDelOk(t, subReqCount, actionDefinitionPresent, policyParamCount, testIndex)

	subReqCount = 19
	actionDefinitionPresent = false
	policyParamCount = 0
	testIndex = 2
	RESTPolicySubReqAndSubDelOk(t, subReqCount, actionDefinitionPresent, policyParamCount, testIndex)
}
*/
func RESTPolicySubReqAndSubDelOk(t *testing.T, subReqCount int, actionDefinitionPresent bool, policyParamCount int, testIndex int) {
	xapp.Logger.Info("TEST: TestRESTPolicySubReqAndSubDelOk with parameter set %v", testIndex)

	// Req
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount, actionDefinitionPresent, policyParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	//params := xappConn1.GetRESTSubsReqPolicyParams1(subReqCount, actionDefinitionPresent, policyParamCount)
	//restSubId := xappConn1.SendRESTPolicySubsReq(t, params)

	var e2SubsId []uint32
	for i := 0; i < subReqCount; i++ {
		crereq, cremsg := e2termConn1.RecvSubsReq(t)
		xappConn1.ExpectRESTNotification(t, restSubId)
		e2termConn1.SendSubsResp(t, crereq, cremsg)
		instanceId := xappConn1.WaitRESTNotification(t, restSubId)
		xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", instanceId)
		e2SubsId = append(e2SubsId, instanceId)
	}

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	for i := 0; i < subReqCount; i++ {
		delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
		e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	}

	// Wait that subs is cleaned
	for i := 0; i < subReqCount; i++ {
		mainCtrl.wait_subs_clean(t, e2SubsId[i], 10)
	}
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

func TestRESTTwoPolicySubReqAndSubDelOk(t *testing.T) {

	subReqCount := 2

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount, actionDefinitionPresent, policyParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	e2SubsIds := sendAndReceiveMultipleE2SubReqs(t, subReqCount, xappConn1, e2termConn1, restSubId)

	assert.Equal(t, len(e2SubsIds), 2)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}
func TestRESTPolicySubReqAndSubDelOkFullAmount(t *testing.T) {

	subReqCount := 19

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, 19},
		Counter{cSubRespFromE2, 19},
		Counter{cSubRespToXapp, 19},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, 19},
		Counter{cSubDelRespFromE2, 19},
		Counter{cSubDelRespToXapp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount, actionDefinitionPresent, policyParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	e2SubsIds := sendAndReceiveMultipleE2SubReqs(t, subReqCount, xappConn1, e2termConn1, restSubId)

	assert.Equal(t, len(e2SubsIds), 19)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}
func TestRESTTwoReportSubReqAndSubDelOk(t *testing.T) {

	subReqCount := 2
	parameterSet := 1
	actionDefinitionPresent := true
	actionParamCount := 1

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, uint64(subReqCount)},
		Counter{cSubRespFromE2, uint64(subReqCount)},
		Counter{cSubRespToXapp, uint64(subReqCount)},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, uint64(subReqCount)},
		Counter{cSubDelRespFromE2, uint64(subReqCount)},
		Counter{cSubDelRespToXapp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	e2SubsIds := sendAndReceiveMultipleE2SubReqs(t, subReqCount, xappConn1, e2termConn1, restSubId)

	assert.Equal(t, len(e2SubsIds), subReqCount)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

func TestRESTTwoReportSubReqAndSubDelOkNoActParams(t *testing.T) {

	subReqCount := 2
	parameterSet := 1
	actionDefinitionPresent := false
	actionParamCount := 0

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, uint64(subReqCount)},
		Counter{cSubRespFromE2, uint64(subReqCount)},
		Counter{cSubRespToXapp, uint64(subReqCount)},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, uint64(subReqCount)},
		Counter{cSubDelRespFromE2, uint64(subReqCount)},
		Counter{cSubDelRespToXapp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	e2SubsIds := sendAndReceiveMultipleE2SubReqs(t, subReqCount, xappConn1, e2termConn1, restSubId)

	assert.Equal(t, len(e2SubsIds), subReqCount)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

func TestRESTFullAmountReportSubReqAndSubDelOk(t *testing.T) {

	subReqCount := 19
	parameterSet := 1
	actionDefinitionPresent := false
	actionParamCount := 0

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 1},
		Counter{cSubReqToE2, uint64(subReqCount)},
		Counter{cSubRespFromE2, uint64(subReqCount)},
		Counter{cSubRespToXapp, uint64(subReqCount)},
		Counter{cSubDelReqFromXapp, 1},
		Counter{cSubDelReqToE2, uint64(subReqCount)},
		Counter{cSubDelRespFromE2, uint64(subReqCount)},
		Counter{cSubDelRespToXapp, 1},
	})

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	e2SubsIds := sendAndReceiveMultipleE2SubReqs(t, subReqCount, xappConn1, e2termConn1, restSubId)

	assert.Equal(t, len(e2SubsIds), subReqCount)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	sendAndReceiveMultipleE2DelReqs(t, e2SubsIds, e2termConn1)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

	mainCtrl.VerifyCounterValues(t)
}

func TestRESTSubReqReportSameActionDiffEventTriggerDefinitionLen(t *testing.T) {
	CaseBegin("TestRESTSubReqReportSameActionDiffEventTriggerDefinitionLen")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_1")
	eventTriggerDefinition := "1234"
	params.SetSubEventTriggerDefinition(eventTriggerDefinition)

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)

}

func TestRESTSubReqReportSameActionDiffActionListLen(t *testing.T) {
	CaseBegin("TestRESTSubReqReportSameActionDiffActionListLen")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_1")

	actionId := int64(1)
	actionType := "report"
	actionDefinition := "56781"
	subsequestActionType := "continue"
	timeToWait := "w10ms"
	params.AppendActionToActionToBeSetupList(actionId, actionType, actionDefinition, subsequestActionType, timeToWait)

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)

}

func TestRESTSubReqReportSameActionDiffActionID(t *testing.T) {
	CaseBegin("TestRESTSubReqReportSameActionDiffActionID")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_1")
	params.SetSubActionIDs(int64(2))

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)

}

func TestRESTSubReqDiffActionType(t *testing.T) {
	CaseBegin("TestRESTSubReqDiffActionType")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	// Req1
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount, actionDefinitionPresent, policyParamCount)

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_1")

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)

}

func TestRESTSubReqPolicyAndSubDelOkSameAction(t *testing.T) {
	CaseBegin("TestRESTSubReqPolicyAndSubDelOkSameAction")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	// Req1
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount, actionDefinitionPresent, policyParamCount)

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqPolicyParams(subReqCount, actionDefinitionPresent, policyParamCount)
	params.SetMeid("RAN_NAME_1")

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)

}

func TestRESTSubReqReportSameActionDiffActionDefinitionLen(t *testing.T) {
	CaseBegin("TestRESTSubReqReportSameActionDiffActionDefinitionLen")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_1")
	actionDefinition := "5678"
	params.SetSubActionDefinition(actionDefinition)

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)

}

func TestRESTSubReqReportSameActionDiffActionDefinitionContents(t *testing.T) {
	CaseBegin("TestRESTSubReqReportSameActionDiffActionDefinitionContents")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_1")
	actionDefinition := "56782"
	params.SetSubActionDefinition(actionDefinition)

	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)

}

func TestRESTSubReqReportSameActionDiffSubsAction(t *testing.T) {
	CaseBegin("TestRESTSubReqReportSameActionDiffSubsAction")

	mainCtrl.CounterValuesToBeVeriefied(t, CountersToBeAdded{
		Counter{cSubReqFromXapp, 2},
		Counter{cSubReqToE2, 2},
		Counter{cSubRespFromE2, 2},
		Counter{cSubRespToXapp, 2},
		Counter{cSubDelReqFromXapp, 2},
		Counter{cSubDelReqToE2, 2},
		Counter{cSubDelRespFromE2, 2},
		Counter{cSubDelRespToXapp, 2},
	})

	// Req1
	var params *teststube2ap.RESTSubsReqParams = nil

	//Subs Create
	restSubId1, e2SubsId1 := createSubscription(t, xappConn1, e2termConn1, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId1)

	queryXappSubscription(t, int64(e2SubsId1), "RAN_NAME_1", []string{"localhost:13560"})

	// Req2
	params = xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	params.SetMeid("RAN_NAME_1")
	params.SetTimeToWait("w200ms")
	restSubId2 := xappConn2.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId2)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	xappConn2.ExpectRESTNotification(t, restSubId2)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn2.WaitRESTNotification(t, restSubId2)

	deleteXapp1Subscription(t, &restSubId1)
	deleteXapp2Subscription(t, &restSubId2)

	waitSubsCleanup(t, e2SubsId1, 10)
	waitSubsCleanup(t, e2SubsId2, 10)

	mainCtrl.VerifyCounterValues(t)

}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionResponseDecodeFail
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp | ASN.1 decode fails
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------|
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionResponseDecodeFail(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionResponseDecodeFail")
	subReqCount := 1
	parameterSet := 1 // E2SM-gNB-X2
	actionDefinitionPresent := true
	actionParamCount := 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	// Decode of this response fails which will result resending original request
	e2termConn1.SendInvalidE2Asn1Resp(t, cremsg, xapp.RIC_SUB_RESP)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotification(t, restSubId)

	// Subscription already created in E2 Node.
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", instanceId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, crereq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionResponseUnknownInstanceId
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp | Unknown instanceId
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------|
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionResponseUnknownInstanceId(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionResponseUnknownInstanceId")
	subReqCount := 1
	parameterSet := 1 // E2SM-gNB-X2
	actionDefinitionPresent := true
	actionParamCount := 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	// Unknown instanceId in this response which will result resending original request
	orgInstanceId := crereq.RequestId.InstanceId
	crereq.RequestId.InstanceId = 0
	e2termConn1.SendSubsResp(t, crereq, cremsg)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotification(t, restSubId)

	// Subscription already created in E2 Node.
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", instanceId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, orgInstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionResponseNoTransaction
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp | No transaction for the response
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------|
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionResponseNoTransaction(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionResponseNoTransaction")
	subReqCount := 1
	parameterSet := 1 // E2SM-gNB-X2
	actionDefinitionPresent := true
	actionParamCount := 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	mainCtrl.MakeTransactionNil(t, crereq.RequestId.InstanceId)
	// No transaction exist for this response which will result resending original request
	e2termConn1.SendSubsResp(t, crereq, cremsg)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotification(t, restSubId)

	// Subscription already created in E2 Node.
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Resending happens because there no transaction
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", instanceId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, crereq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)

}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionFailureDecodeFail
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | ASN.1 decode fails
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------|
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionFailureDecodeFail(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionFailureDecodeFail")
	subReqCount := 1
	parameterSet := 1 // E2SM-gNB-X2
	actionDefinitionPresent := true
	actionParamCount := 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	// Decode of this response fails which will result resending original request
	e2termConn1.SendInvalidE2Asn1Resp(t, cremsg, xapp.RIC_SUB_FAILURE)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotification(t, restSubId)

	// Subscription already created in E2 Node.
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", instanceId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, crereq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionResponseUnknownInstanceId
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Unknown instanceId
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------|
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionFailureUnknownInstanceId(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionFailureUnknownInstanceId")
	subReqCount := 1
	parameterSet := 1 // E2SM-gNB-X2
	actionDefinitionPresent := true
	actionParamCount := 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	// Unknown instanceId in this response which will result resending original request
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	fparams.Fail.RequestId.InstanceId = 0
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotification(t, restSubId)

	// Subscription already created in E2 Node.
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", instanceId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, crereq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionFailureNoTransaction
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     | RestSubReq      |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | No transaction for the response
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubFail | Duplicated action
//     |                 |<-------------|
//     | RESTNotif (fail)|              |
//     |<----------------|              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionFailureNoTransaction(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionFailureNoTransaction")
	subReqCount := 1
	parameterSet := 1 // E2SM-gNB-X2
	actionDefinitionPresent := true
	actionParamCount := 1

	// Req
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	mainCtrl.MakeTransactionNil(t, crereq.RequestId.InstanceId)

	// No transaction exist for this response which will result resending original request
	fparams := &teststube2ap.E2StubSubsFailParams{}
	fparams.Set(crereq)
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	_, cremsg = e2termConn1.RecvSubsReq(t)

	xappConn1.ExpectRESTNotification(t, restSubId)

	// Subscription already created in E2 Node.
	fparams.SetCauseVal(0, 1, 3) // CauseRIC / duplicate-action
	e2termConn1.SendSubsFail(t, fparams, cremsg)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Resending happens because there no transaction
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	instanceId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", instanceId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, crereq.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteResponseDecodeFail
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp | ASN.1 decode fails
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionDeleteResponseDecodeFail(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionDeleteResponseDecodeFail")

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	// Decode of this response fails which will result resending original request
	e2termConn1.SendInvalidE2Asn1Resp(t, delmsg, xapp.RIC_SUB_DEL_REQ)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteResponseUnknownInstanceId
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp | Unknown instanceId
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionDeleteResponseUnknownInstanceId(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionDeleteResponseUnknownInstanceId")

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	// Unknown instanceId in this response which will result resending original request
	delreq.RequestId.InstanceId = 0
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// E2t: Receive 2nd SubsDelReq
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteResponseNoTransaction
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelResp | No transaction for the response
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionDeleteResponseNoTransaction(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionDeleteResponseNoTransaction")

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	mainCtrl.MakeTransactionNil(t, e2SubsId)

	// No transaction exist for this response which will result resending original request
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// E2t: Receive 2nd SubsDelReq
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteFailureDecodeFail
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | ASN.1 decode fails
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionDeleteFailureDecodeFail(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionDeleteFailureDecodeFail")

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	// Decode of this response fails which will result resending original request
	e2termConn1.SendInvalidE2Asn1Resp(t, delmsg, xapp.RIC_SUB_DEL_FAILURE)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteailureUnknownInstanceId
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Unknown instanceId
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionDeleteailureUnknownInstanceId(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionDeleteailureUnknownInstanceId")

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	// Unknown instanceId in this response which will result resending original request
	delreq.RequestId.InstanceId = 0
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// E2t: Receive 2nd SubsDelReq
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTUnpackSubscriptionDeleteFailureNoTransaction
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq   |              |
//     |---------------->|              |
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | No transaction for the response
//     |                 |<-------------|
//     |                 |              |
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     |                 |   SubDelFail | Subscription does exist any more
//     |                 |<-------------|
//
//-----------------------------------------------------------------------------
func TestRESTUnpackSubscriptionDeleteFailureNoTransaction(t *testing.T) {
	xapp.Logger.Info("TEST: TestRESTUnpackSubscriptionDeleteFailureNoTransaction")

	// Req
	var params *teststube2ap.RESTSubsReqParams = nil
	restSubId, e2SubsId := createSubscription(t, xappConn1, e2termConn1, params)

	// Del
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)

	mainCtrl.MakeTransactionNil(t, e2SubsId)

	// No transaction exist for this response which will result resending original request
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// E2t: Receive 2nd SubsDelReq
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)

	// Subscription does not exist in in E2 Node.
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

////////////////////////////////////////////////////////////////////////////////////
//   Services for UT cases
////////////////////////////////////////////////////////////////////////////////////
const subReqCount int = 1
const parameterSet = 1
const actionDefinitionPresent bool = true
const actionParamCount int = 1
const policyParamCount int = 1
const host string = "localhost"

func createSubscription(t *testing.T, fromXappConn *teststube2ap.E2Stub, toE2termConn *teststube2ap.E2Stub, params *teststube2ap.RESTSubsReqParams) (string, uint32) {
	if params == nil {
		params = fromXappConn.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	}
	restSubId := fromXappConn.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := toE2termConn.RecvSubsReq(t)
	fromXappConn.ExpectRESTNotification(t, restSubId)
	toE2termConn.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := fromXappConn.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("REST notification received e2SubsId=%v", e2SubsId)

	return restSubId, e2SubsId
}

func createXapp2MergedSubscription(t *testing.T, meid string) (string, uint32) {

	params := xappConn2.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	if meid != "" {
		params.SetMeid(meid)
	}
	xapp.Subscription.SetResponseCB(xappConn2.SubscriptionRespHandler)
	restSubId := xappConn2.SendRESTSubsReq(t, params)
	xappConn2.ExpectRESTNotification(t, restSubId)
	xapp.Logger.Info("Send REST subscriber request for subscriberId : %v", restSubId)
	e2SubsId := xappConn2.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("REST notification received e2SubsId=%v", e2SubsId)

	return restSubId, e2SubsId
}

func createXapp1PolicySubscription(t *testing.T) (string, uint32) {
	params := xappConn1.GetRESTSubsReqPolicyParams(subReqCount, actionDefinitionPresent, policyParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)
	xapp.Logger.Info("Send REST Policy subscriber request for subscriberId : %v", restSubId)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("REST notification received e2SubsId=%v", e2SubsId)

	return restSubId, e2SubsId
}

func createXapp1ReportSubscriptionE2Fail(t *testing.T) (string, uint32) {
	params := xappConn1.GetRESTSubsReqReportParams(subReqCount, parameterSet, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTSubsReq(t, params)

	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	xappConn1.ExpectRESTNotification(t, restSubId)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t, restSubId)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	return restSubId, e2SubsId
}

func deleteSubscription(t *testing.T, fromXappConn *teststube2ap.E2Stub, toE2termConn *teststube2ap.E2Stub, restSubId *string) {
	fromXappConn.SendRESTSubsDelReq(t, restSubId)
	delreq, delmsg := toE2termConn.RecvSubsDelReq(t)
	toE2termConn.SendSubsDelResp(t, delreq, delmsg)
}

func deleteXapp1Subscription(t *testing.T, restSubId *string) {
	xappConn1.SendRESTSubsDelReq(t, restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
}

func deleteXapp2Subscription(t *testing.T, restSubId *string) {
	xappConn2.SendRESTSubsDelReq(t, restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)
}

func queryXappSubscription(t *testing.T, e2SubsId int64, meid string, endpoint []string) {
	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, e2SubsId)
	assert.Equal(t, resp[0].Meid, meid)
	assert.Equal(t, resp[0].ClientEndpoint, endpoint)
}

func waitSubsCleanup(t *testing.T, e2SubsId uint32, timeout int) {
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, timeout)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, timeout)
}

func sendAndReceiveMultipleE2SubReqs(t *testing.T, count int, fromXappConn *teststube2ap.E2Stub, toE2termConn *teststube2ap.E2Stub, restSubId string) []uint32 {

	var e2SubsId []uint32

	for i := 0; i < count; i++ {
		xapp.Logger.Info("TEST: %d ===================================== BEGIN CRE ============================================", i+1)
		crereq, cremsg := toE2termConn.RecvSubsReq(t)
		fromXappConn.ExpectRESTNotification(t, restSubId)
		toE2termConn.SendSubsResp(t, crereq, cremsg)
		instanceId := fromXappConn.WaitRESTNotification(t, restSubId)
		e2SubsId = append(e2SubsId, instanceId)
		xapp.Logger.Info("TEST: %v", e2SubsId)
		xapp.Logger.Info("TEST: %d ===================================== END CRE ============================================", i+1)
		<-time.After(100 * time.Millisecond)
	}
	return e2SubsId
}

func sendAndReceiveMultipleE2DelReqs(t *testing.T, e2SubsIds []uint32, toE2termConn *teststube2ap.E2Stub) {

	for i := 0; i < len(e2SubsIds); i++ {
		xapp.Logger.Info("TEST: %d ===================================== BEGIN DEL ============================================", i+1)
		delreq, delmsg := toE2termConn.RecvSubsDelReq(t)
		toE2termConn.SendSubsDelResp(t, delreq, delmsg)
		<-time.After(1 * time.Second)
		xapp.Logger.Info("TEST: %d ===================================== END DEL ============================================", i+1)
		<-time.After(100 * time.Millisecond)
	}

	// Wait that subs is cleaned
	for i := 0; i < len(e2SubsIds); i++ {
		mainCtrl.wait_subs_clean(t, e2SubsIds[i], 10)
	}

}
