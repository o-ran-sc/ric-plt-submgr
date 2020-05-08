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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststube2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

//-----------------------------------------------------------------------------
// TestSubReqAndRouteNok
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | rtmgr   |
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
}

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

	cretrans := xappConn1.SendSubsReq(t, nil, nil)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	resp, _ := xapp.Subscription.QuerySubscriptions()
	assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId))
	assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	assert.Equal(t, resp[0].Endpoint, []string{"localhost:13560"})

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
	rparams1.Req.EventTriggerDefinition.ProcedureCode = 5
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	rparams2.Req.EventTriggerDefinition.ProcedureCode = 28
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
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqRetryNoRespSubDelRespInSubmgr(t *testing.T) {

	CaseBegin("TestSubReqTwoRetriesNoRespSubDelRespInSubmgr start")

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
}

//-----------------------------------------------------------------------------
// TestSubReqSubFailRespInSubmgrWithDuplicate
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

func TestSubReqSubFailRespInSubmgrWithDuplicate(t *testing.T) {

	CaseBegin("TestSubReqSubFailRespInSubmgrWithDuplicate start")

	// Xapp: Send SubsReq
	cretrans := xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive SubsReq and send SubsFail (first)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	fparams1.SetCauseVal(-1, 5, 3)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// E2t: Receive SubsDelReq and send SubsDelResp (internal)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// E2t: Receive SubsReq and send SubsResp (second)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)

	// XAPP: Receive SubsResp
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Delete
	deltrans2 := xappConn1.SendSubsDelReq(t, nil, e2SubsId)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	xappConn1.RecvSubsDelResp(t, deltrans2)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqSubFailRespInSubmgrWithDuplicateFail
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
//     |      SubFail |              |
//     |<-------------|              |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqSubFailRespInSubmgrWithDuplicateFail(t *testing.T) {

	CaseBegin("TestSubReqSubFailRespInSubmgrWithDuplicateFail start")

	// Xapp: Send SubsReq
	cretrans := xappConn1.SendSubsReq(t, nil, nil)

	// E2t: Receive SubsReq and send SubsFail (first)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	fparams1.SetCauseVal(-1, 5, 3)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// E2t: Receive SubsDelReq and send SubsDelResp (internal first)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// E2t: Receive SubsReq and send SubsFail (second)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)
	fparams2 := &teststube2ap.E2StubSubsFailParams{}
	fparams2.Set(crereq2)
	fparams2.SetCauseVal(-1, 5, 3)
	e2termConn1.SendSubsFail(t, fparams2, cremsg2)

	// E2t: Receive SubsDelReq and send SubsDelResp (internal second)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)

	// Xapp: Receive SubsFail
	e2SubsId := xappConn1.RecvSubsFail(t, cretrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
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
	assert.Equal(t, resp[0].Endpoint, []string{"localhost:13560", "localhost:13660"})

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
	CaseBegin("TestSubReqPolicyChangeAndSubDelOk")

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

/******************************************************************************/
//  REST interface test cases
/******************************************************************************/

//-----------------------------------------------------------------------------
// TestRESTSubReqAndRouteNok   It is not possible currently to indicate failure to xapp via REST interface
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

	waiter := rtmgrHttp.AllocNextEvent(false)
	newSubsId := mainCtrl.get_registry_next_subid(t)

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)
	waiter.WaitResult(t)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, newSubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTReportSubReqAndSubDelOk and
// TestRESTPolicySubReqAndSubDelOk
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
	RESTReportSubReqAndSubDelOk(t /*subReqCount=*/, 2 /*actionDefinitionPresent=*/, true /*actionParamCount=*/, 1 /*testIndex=*/, 1)
	RESTReportSubReqAndSubDelOk(t /*subReqCount=*/, 19 /*actionDefinitionPresent=*/, false /*actionParamCount=*/, 0 /*testIndex=*/, 2)
}

func RESTReportSubReqAndSubDelOk(t *testing.T, subReqCount int, actionDefinitionPresent bool, actionParamCount int, testIndex uint32) {
	xapp.Logger.Info("TEST: TestRESTReportSubReqAndSubDelOk with parameter set %v", testIndex)

	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	var e2SubsId []uint32
	for i := 0; i < subReqCount; i++ {
		crereq, cremsg := e2termConn1.RecvSubsReq(t)
		e2termConn1.SendSubsResp(t, crereq, cremsg)
		instanceId := xappConn1.WaitRESTNotification(t)
		xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", instanceId)
		e2SubsId = append(e2SubsId, instanceId)

		// Test REST interface query
		//resp, _ := xapp.Subscription.QuerySubscriptions()
		//assert.Equal(t, resp[i].SubscriptionID, (int64)(instanceId))
		//assert.Equal(t, resp[i].Meid, "RAN_NAME_1")
		//assert.Equal(t, resp[i].Endpoint, []string{"localhost:13560"})
	}

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

func TestRESTPolicySubReqAndSubDelOk(t *testing.T) {
	CaseBegin("TestRESTReportSubReqAndSubDelOk")
	RESTPolicySubReqAndSubDelOk(t /*subReqCount=*/, 2 /*actionDefinitionPresent=*/, true /*policyParamCount=*/, 1 /*testIndex=*/, 1)
	RESTPolicySubReqAndSubDelOk(t /*subReqCount=*/, 19 /*actionDefinitionPresent=*/, false /*policyParamCount=*/, 0 /*testIndex=*/, 2)
}

func RESTPolicySubReqAndSubDelOk(t *testing.T, subReqCount int, actionDefinitionPresent bool, policyParamCount int, testIndex uint32) {
	xapp.Logger.Info("TEST: TestRESTPolicySubReqAndSubDelOk with parameter set %v", testIndex)

	params := xappConn1.GetRESTSubsReqPolicyParams1(subReqCount, actionDefinitionPresent, policyParamCount)
	restSubId := xappConn1.SendRESTPolicySubsReq(t, params)

	var e2SubsId []uint32
	for i := 0; i < subReqCount; i++ {
		crereq, cremsg := e2termConn1.RecvSubsReq(t)
		e2termConn1.SendSubsResp(t, crereq, cremsg)
		instanceId := xappConn1.WaitRESTNotification(t)
		xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", instanceId)
		e2SubsId = append(e2SubsId, instanceId)
	}

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
//-----------------------------------------------------------------------------
// TestRESTSubReqRetransmission
//
// Should Submgr accept only one request per xapp at a time and process that completely before next one accepted?
// Submgr is not able to detect resent REST requests. What about merging equal subscription Requests from same endpoint
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |  RESTSubReq     |              |
//     |---------------->|              |
//     |                 |              |
//     |         SubResp |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |  RESTSubReq     |              |
//     | (retrans)       |              |
//     |---------------->|              |
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     | RESTNotif       |              |
//     | InstanceId=1    |              |
//     |<----------------|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |     Bad request |              |
//     |<----------------|              |
//     |                 |              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqRetransmission(t *testing.T) {
	CaseBegin("TestRESTSubReqRetransmission")

	// Subs Create
	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	// Retry
	xappConn1.SendRESTReportSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Subs Delete
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	for i := 0; i < subReqCount; i++ {
		delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
		e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	}

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}


//-----------------------------------------------------------------------------
// TestRESTSubDelReqRetransmission
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
//     |                 | SubDelReq    |
//     |                 |------------->|
//     |                 |              |
//     | RESTSubDelReq   |              |
//     | (same Id)       |              |
//     |---------------->|              |
//     |                 |              |
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//     |  RESTSubDelResp |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubDelReqRetransmission(t *testing.T) {
	CaseBegin("TestRESTSubDelReqRetransmission")

	// Subs Create
	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Subs Delete
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Retry
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	for i := 0; i < subReqCount; i++ {
		delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
		e2termConn1.SendSubsDelResp(t, delreq, delmsg)
	}

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}
*/
/*
//-----------------------------------------------------------------------------
// TestRESTSubDelReqCollision   In REST we do not have this case exactly
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |            [SUBS CREATE]       |
//     |                 |              |
//     |                 |              |
//     | RESTSubDelReq1  |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelReq1   |
//     |                 |------------->|
//     |                 |              |
//     | RESTSubDelReq2  |              |
//     | (same sub)      |              |
//     | (diff xid)      |              |
//     |---------------->|              |
//     |                 |              |
//     |                 | SubDelResp1  |
//     |                 |<-------------|
//     |                 |              |
//     | RESTSubDelResp1 |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubDelResp2 |              |
//     |<----------------|              |
//
//-----------------------------------------------------------------------------

func TestRESTSubDelReqCollision(t *testing.T) {
	CaseBegin("TestRESTSubDelReqCollision")

	// Subs Create
	cretrans := xappConn1.SendSubsReq(t, nil, nil)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.RecvSubsResp(t, cretrans)

	// Subs Delete
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

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}
*/

//-----------------------------------------------------------------------------
// TestRESTSubReqAndSubDelOkTwoParallel
//
//   stub       stub                                stub
// +-------+  +-------+           +---------+    +---------+
// | xapp2 |  | xapp1 |           | submgr  |    | e2term  |
// +-------+  +-------+           +---------+    +---------+
//     |          |                 |              |
//     |          |                 |              |
//     |          |                 |              |
//     |          | RESTSubReq1     |              |
//     |          |---------------->|              |
//     |          |                 |              |
//     |          |                 | SubReq1      |
//     |          |                 |------------->|
//     |          |                 |              |
//     | RESTSubReq2                |              |
//     |--------------------------->|              |
//     |          |                 |              |
//     |          |                 | SubReq2      |
//     |          |                 |------------->|
//     |          |                 |              |
//     |          |                 |    SubResp1  |
//     |          |                 |<-------------|
//     |          |    RESTSubResp1 |              |
//     |          |<----------------|              |
//     |          |                 |              |
//     |          |                 |    SubResp2  |
//     |          |                 |<-------------|
//     |               RESTSubResp2 |              |
//     |<---------------------------|              |
//     |          |                 |              |
//     |          |           [SUBS 1 DELETE]      |
//     |          |                 |              |
//     |          |           [SUBS 2 DELETE]      |
//     |          |                 |              |
//
//-----------------------------------------------------------------------------
/*
func TestRESTSubReqAndSubDelOkTwoParallel(t *testing.T) {
	CaseBegin("TestRESTSubReqAndSubDelOkTwoParallel")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	// Req1
	params1 := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	params1.SetEventTriggerDefinitionProcedureCode(5)
	restSubId1 := xappConn1.SendRESTReportSubsReq(t, params1)

	// Req2
	params2 := xappConn2.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	params2.SetEventTriggerDefinitionProcedureCode(28)
	restSubId2 := xappConn2.SendRESTReportSubsReq(t, params2)

	// Resp1
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId1=%v", e2SubsId1)

	// Resp2
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId2=%v", e2SubsId2)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	// Del2
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)
*/

/*
	//Req1
	rparams1 := &teststube2ap.E2StubSubsReqParams{}
	rparams1.Init()
	rparams1.Req.EventTriggerDefinition.ProcedureCode = 5
	cretrans1 := xappConn1.SendSubsReq(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)

	//Req2
	rparams2 := &teststube2ap.E2StubSubsReqParams{}
	rparams2.Init()
	rparams2.Req.EventTriggerDefinition.ProcedureCode = 28
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
*/

/*
	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}
*/

//-----------------------------------------------------------------------------
// TestRESTSameSubsDiffRan
// Same subscription to different RANs
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |                 |              |
//     |                 |              |
//     | RESTSubReq      |              |
//     | ran1            |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |     ran1        |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |       ran1      |              |
//     |<----------------|              |
//     |                 |              |
//     | RESTSubReq      |              |
//     | ran2            |              |
//     |---------------->|              |
//     |                 |              |
//     |     RESTSubResp |              |
//     |     ran2        |              |
//     |<----------------|              |
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |       RESTNotif |              |
//     |       ran2      |              |
//     |<----------------|              |
//     |                 |              |
//     |                 |              |
//     |        [SUBS ran1 DELETE]      |
//     |                 |              |
//     |        [SUBS ran2 DELETE]      |
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSameSubsDiffRan(t *testing.T) {
	CaseBegin("TestRESTSameSubsDiffRan")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1

	// Req1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId1 := xappConn1.SendRESTReportSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId1 := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId1=%v", e2SubsId1)

	// Req2
	params = xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	params.SetRANName("RAN_NAME_2")
	restSubId2 := xappConn1.SendRESTReportSubsReq(t, params)

	crereq, cremsg = e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId2 := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId2=%v", e2SubsId2)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Del2
	xappConn1.SendRESTSubsDelReq(t, &restSubId2)
	delreq, delmsg = e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqRetryInSubmgr
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
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqRetryInSubmgr(t *testing.T) {
	CaseBegin("TestRESTSubReqRetryInSubmgr")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	// E2t: Receive 1st SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 2nd SubsReq and send SubsResp
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
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
	CaseBegin("TestRESTSubReqTwoRetriesNoRespSubDelRespInSubmgr")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	// E2t: Receive 1st SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 2nd SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Send receive SubsDelReq and send SubsResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// REST subscription sill there to be deleted
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq.RequestId.InstanceId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqTwoRetriesNoRespAtAllInSubmgr
//
//   stub                             stub
// +-------+        +---------+    +---------+
// | xapp  |        | submgr  |    | e2term  |
// +-------+        +---------+    +---------+
//     |                 |              |
//     |  RESTSubReq     |              |
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
//     |       RESTNotif |              |
//     |       unsuccess |              |
//     |<----------------|              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqTwoRetriesNoRespAtAllInSubmgr(t *testing.T) {
	CaseBegin("TestRESTSubReqTwoRetriesNoRespAtAllInSubmgr")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	// E2t: Receive 1st SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 2nd SubsReq
	e2termConn1.RecvSubsReq(t)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq
	delreq, _ := e2termConn1.RecvSubsDelReq(t)

	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// REST subscription sill there to be deleted
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq.RequestId.InstanceId, 15)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
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

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	// E2t: Receive SubsReq and send SubsFail (first)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// E2t: Receive SubsDelReq and send SubsDelResp (internal first)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// REST subscription sill there to be deleted
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqSubFailRespInSubmgrWithDuplicate
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
//     |                 | SubReq       |
//     |                 |------------->|
//     |                 |              |
//     |                 |      SubResp |
//     |                 |<-------------|
//     |                 |              |
//     |       RESTNotif |              |
//     |<----------------|              |
//     |                 |              |
//     |            [SUBS DELETE]       |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqSubFailRespInSubmgrWithDuplicate(t *testing.T) {
	CaseBegin("TestRESTSubReqSubFailRespInSubmgrWithDuplicate")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	// E2t: Receive SubsReq and send SubsFail (first)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	fparams1.SetCauseVal(-1, 5, 3)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// E2t: Receive SubsDelReq and send SubsDelResp (internal)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// E2t: Receive SubsReq and send SubsResp (second)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Delete
	xappConn1.SendRESTSubsDelReq(t, &restSubId)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestRESTSubReqSubFailRespInSubmgrWithDuplicateFail
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
//     |       RESTNotif |              |
//     |       unsuccess |              |
//     |<----------------|              |
//     |                 |              |
//
//-----------------------------------------------------------------------------

func TestRESTSubReqSubFailRespInSubmgrWithDuplicateFail(t *testing.T) {
	CaseBegin("TestRESTSubReqSubFailRespInSubmgrWithDuplicateFail")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	xappConn1.SendRESTReportSubsReq(t, params)

	// E2t: Receive SubsReq and send SubsFail (first)
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	fparams1.SetCauseVal(-1, 5, 3)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// E2t: Receive SubsDelReq and send SubsDelResp (internal first)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// E2t: Receive SubsReq and send SubsFail (second)
	crereq2, cremsg2 := e2termConn1.RecvSubsReq(t)
	fparams2 := &teststube2ap.E2StubSubsFailParams{}
	fparams2.Set(crereq2)
	fparams2.SetCauseVal(-1, 5, 3)
	e2termConn1.SendSubsFail(t, fparams2, cremsg2)

	// E2t: Receive SubsDelReq and send SubsDelResp (internal second)
	delreq2, delmsg2 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq2, delmsg2)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
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

	// Subs Create
	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	// Subs Create
	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Subs Delete
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
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

func TestRSETSubDelReqTwoRetriesNoRespInSubmgr(t *testing.T) {
	CaseBegin("TestRESTSubDelReTwoRetriesNoRespInSubmgr")

	// Subs Create
	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Subs Delete
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Receive 1st SubsDelReq
	e2termConn1.RecvSubsDelReq(t)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
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
	CaseBegin("TestRESTSubReqSubDelFailRespInSubmgr")

	// Subs Create
	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Subs Delete
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	// E2t: Send receive SubsDelReq and send SubsDelFail
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelFail(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

/*
//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameAction
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
func TestSubReqAndSubDelOkSameAction(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkSameAction")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)

	// Req1
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId1 := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId1)

	// Req2
	restSubId2 := xappConn2.SendRESTReportSubsReq(t, params)

	e2SubsId2 := xappConn2.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId2)

	//resp, _ := xapp.Subscription.QuerySubscriptions()
	//assert.Equal(t, resp[0].SubscriptionID, int64(e2SubsId1))
	//assert.Equal(t, resp[0].Meid, "RAN_NAME_1")
	//assert.Equal(t, resp[0].Endpoint, []string{"localhost:13560", "localhost:13660"})

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)

	//Wait that subs is cleaned
	//mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	// Del2
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)
	delreq, delmsg := e2termConn2.RecvSubsDelReq(t)
	e2termConn2.SendSubsDelResp(t, delreq, delmsg)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameActionParallel
//
//   stub          stub                             stub
// +-------+     +-------+        +---------+    +---------+
// | xapp2 |     | xapp1 |        | submgr  |    | e2term  |
// +-------+     +-------+        +---------+    +---------+
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
//     |                  RESTSubResp2 |              |
//     |<------------------------------|              |
//     |             |                 |    SubResp1  |
//     |             |                 |<-------------|
//     |             |      RESTNotif1 |              |
//     |             |<----------------|              |
//     |             |                 |              |
//     |                    RESTNotif2 |              |
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
//     |             |                 | SubDelReq2   |
//     |             |                 |------------->|
//     |             |                 |              |
//     |             |                 | SubDelReq2   |
//     |             |                 |------------->|
//     |             |                 |              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOkSameActionParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkSameActionParallel")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)

	// Req1
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	// Req2
	restSubId2 := xappConn2.SendRESTReportSubsReq(t, params)

	e2termConn1.SendSubsResp(t, crereq, cremsg)

	e2SubsId1 := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId1)

	e2SubsId2 := xappConn2.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId2)

	// Del1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)

	// Del2
	xappConn2.SendRESTSubsDelReq(t, &restSubId2)
	delreq, delmsg := e2termConn2.RecvSubsDelReq(t)
	e2termConn2.SendSubsDelResp(t, delreq, delmsg)

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
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelNokSameActionParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelNokSameActionParallel")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)

	// Req1
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	E2t: Receive SubsReq (first)
	crereq, cremsg := e2termConn1.RecvSubsReq(t)

	// Req2
	subepcnt2 := mainCtrl.get_subs_entrypoint_cnt(t, crereq1.RequestId.InstanceId)
	restSubId2 := xappConn2.SendRESTReportSubsReq(t, params)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, crereq1.RequestId.InstanceId, subepcnt2, 10)

	// E2t: send SubsFail (first)
	fparams1 := &teststube2ap.E2StubSubsFailParams{}
	fparams1.Set(crereq1)
	e2termConn1.SendSubsFail(t, fparams1, cremsg1)

	// E2t: internal delete
	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	//Fail1
	e2SubsId1 := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId1)

	//Fail2
	e2SubsId2 := xappConn2.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId2)

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
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelNoAnswerSameActionParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelNoAnswerSameActionParallel")

	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)

	// Req1
	restSubId := xappConn1.SendRESTReportSubsReq(t, params)

	crereq1, _ := e2termConn1.RecvSubsReq(t)

	// Req2
	subepcnt2 := mainCtrl.get_subs_entrypoint_cnt(t, crereq1.RequestId.InstanceId)
	restSubId2 := xappConn2.SendRESTReportSubsReq(t, params)
	mainCtrl.wait_subs_entrypoint_cnt_change(t, crereq1.RequestId.InstanceId, subepcnt2, 10)

	//Req1 (retransmitted)
	e2termConn1.RecvSubsReq(t)

	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	e2SubsId1 := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId1)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq1.RequestId.InstanceId, 10)

	xappConn1.TestMsgChanEmpty(t)
	xappConn2.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 15)
}
*/

//-----------------------------  Policy cases ---------------------------------
//-----------------------------------------------------------------------------
// TestRESTSubReqPolicyAndSubDelOk  This duplicate
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
//     |                 |   SubDelResp |
//     |                 |<-------------|
//     |                 |              |
//
//-----------------------------------------------------------------------------
func TestRESTSubReqPolicyAndSubDelOk(t *testing.T) {
	CaseBegin("TestRESTSubReqPolicyAndSubDelOk")

	// Subs Create
	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const policyParamCount int = 1
	params := xappConn1.GetRESTSubsReqPolicyParams1(subReqCount, actionDefinitionPresent, policyParamCount)
	restSubId := xappConn1.SendRESTPolicySubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Subs Delete
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}

/*
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

	// Subs Create
	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const policyParamCount int = 1
	params := xappConn1.GetRESTSubsReqPolicyParams1(subReqCount, actionDefinitionPresent, policyParamCount)
	restSubId, _ := xappConn1.SendRESTPolicySubsReq(t, params)

	crereq, cremsg := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Policy change
	params.SetRANParameterTestCondition("greaterthan")
	restSubId, _ = xappConn1.SendRESTPolicySubsReq(t, params)

	crereq, cremsg = e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq, cremsg)
	e2SubsId := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId=%v", e2SubsId)

	// Subs Delete
	xappConn1.SendRESTSubsDelReq(t, &restSubId)

	delreq, delmsg := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}
*/
/*
//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkTwoE2termParallel
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
func TestSubReqAndSubDelOkTwoE2termParallel(t *testing.T) {
	CaseBegin("TestSubReqAndSubDelOkTwoE2termParallel")

	// Subs Create 1
	const subReqCount int = 1
	const actionDefinitionPresent bool = true
	const actionParamCount int = 1
	params := xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	restSubId1 := xappConn1.SendRESTReportSubsReq(t, params)

	// Subs Create 2
	params = xappConn1.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount)
	params.SetRANName("RAN_NAME_11")
	restSubId2 := xappConn1.SendRESTReportSubsReq(t, params)

	// Resp1
	crereq1, cremsg1 := e2termConn1.RecvSubsReq(t)
	e2termConn1.SendSubsResp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId1=%v", e2SubsId1)

	// Resp2
	crereq2, cremsg2 := e2termConn2.RecvSubsReq(t)
	e2termConn2.SendSubsResp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.WaitRESTNotification(t)
	xapp.Logger.Info("TEST: REST notification received e2SubsId2=%v", e2SubsId2)

	// Subs Delete 1
	xappConn1.SendRESTSubsDelReq(t, &restSubId1)
	delreq1, delmsg1 := e2termConn1.RecvSubsDelReq(t)
	e2termConn1.SendSubsDelResp(t, delreq1, delmsg1)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	// Subs Delete 2
	xappConn1.SendRESTSubsDelReq(t, &restSubId2)
	delreq2, delmsg2 := e2termConn2.RecvSubsDelReq(t)
	e2termConn2.SendSubsDelResp(t, delreq2, delmsg2)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)
	xappConn1.TestMsgChanEmpty(t)
	e2termConn1.TestMsgChanEmpty(t)
	e2termConn2.TestMsgChanEmpty(t)
	mainCtrl.wait_registry_empty(t, 10)
}
*/
