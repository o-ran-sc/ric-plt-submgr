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
	"testing"
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
	xapp.Logger.Info("TestSubReqAndRouteNok")

	waiter := rtmgrHttp.AllocNextEvent(false)
	newSubsId := mainCtrl.get_subid(t)
	xappConn1.handle_xapp_subs_req(t, nil, nil)
	waiter.WaitResult(t)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, newSubsId, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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
	xapp.Logger.Info("TestSubReqAndSubDelOk")

	waiter := rtmgrHttp.AllocNextEvent(true)
	cretrans := xappConn1.handle_xapp_subs_req(t, nil, nil)
	waiter.WaitResult(t)

	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq, cremsg)
	e2SubsId := xappConn1.handle_xapp_subs_resp(t, cretrans)
	deltrans := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId)
	delreq, delmsg := e2termConn.handle_e2term_subs_del_req(t)

	waiter = rtmgrHttp.AllocNextEvent(true)
	e2termConn.handle_e2term_subs_del_resp(t, delreq, delmsg)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans)
	waiter.WaitResult(t)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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
	xapp.Logger.Info("TestSubReqRetransmission")

	//Subs Create
	cretrans := xappConn1.handle_xapp_subs_req(t, nil, nil)
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)

	seqBef := mainCtrl.get_msgcounter(t)
	xappConn1.handle_xapp_subs_req(t, nil, cretrans) //Retransmitted SubReq
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	e2termConn.handle_e2term_subs_resp(t, crereq, cremsg)
	e2SubsId := xappConn1.handle_xapp_subs_resp(t, cretrans)

	//Subs Delete
	deltrans := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId)
	delreq, delmsg := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq, delmsg)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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
	xapp.Logger.Info("TestSubDelReqRetransmission")

	//Subs Create
	cretrans := xappConn1.handle_xapp_subs_req(t, nil, nil)
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq, cremsg)
	e2SubsId := xappConn1.handle_xapp_subs_resp(t, cretrans)

	//Subs Delete
	deltrans := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId)
	delreq, delmsg := e2termConn.handle_e2term_subs_del_req(t)

	seqBef := mainCtrl.get_msgcounter(t)
	xappConn1.handle_xapp_subs_del_req(t, deltrans, e2SubsId) //Retransmitted SubDelReq
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	e2termConn.handle_e2term_subs_del_resp(t, delreq, delmsg)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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
	xapp.Logger.Info("TestSubDelReqCollision")

	//Subs Create
	cretrans := xappConn1.handle_xapp_subs_req(t, nil, nil)
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq, cremsg)
	e2SubsId := xappConn1.handle_xapp_subs_resp(t, cretrans)

	//Subs Delete
	xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId)
	delreq1, delmsg1 := e2termConn.handle_e2term_subs_del_req(t)

	// Subs Delete colliding
	seqBef := mainCtrl.get_msgcounter(t)
	deltranscol2 := xappConn1.newXappTransaction("", "RAN_NAME_1")
	xappConn1.handle_xapp_subs_del_req(t, deltranscol2, e2SubsId) //Colliding SubDelReq
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	// Del resp for first and second
	e2termConn.handle_e2term_subs_del_resp(t, delreq1, delmsg1)

	// don't care in which order responses are received
	xappConn1.handle_xapp_subs_del_resp(t, nil)
	xappConn1.handle_xapp_subs_del_resp(t, nil)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkTwoParallel
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |              |              |
//     |              |              |
//     | SubReq1      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq1      |
//     |              |------------->|
//     |              |              |
//     | SubReq2      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq2      |
//     |              |------------->|
//     |              |              |
//     |              |    SubResp1  |
//     |              |<-------------|
//     |              |    SubResp2  |
//     |              |<-------------|
//     |              |              |
//     |    SubResp1  |              |
//     |<-------------|              |
//     |    SubResp2  |              |
//     |<-------------|              |
//     |              |              |
//     |        [SUBS 1 DELETE]      |
//     |              |              |
//     |        [SUBS 2 DELETE]      |
//     |              |              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOkTwoParallel(t *testing.T) {
	xapp.Logger.Info("TestSubReqAndSubDelOkTwoParallel")

	//Req1
	rparams1 := &test_subs_req_params{}
	rparams1.Init()
	rparams1.req.EventTriggerDefinition.ProcedureCode = 5
	cretrans1 := xappConn1.handle_xapp_subs_req(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn.handle_e2term_subs_req(t)

	//Req2
	rparams2 := &test_subs_req_params{}
	rparams2.Init()
	rparams2.req.EventTriggerDefinition.ProcedureCode = 28
	cretrans2 := xappConn2.handle_xapp_subs_req(t, rparams2, nil)
	crereq2, cremsg2 := e2termConn.handle_e2term_subs_req(t)

	//Resp1
	e2termConn.handle_e2term_subs_resp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.handle_xapp_subs_resp(t, cretrans1)

	//Resp2
	e2termConn.handle_e2term_subs_resp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn2.handle_xapp_subs_resp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId1)
	delreq1, delmsg1 := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq1, delmsg1)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans1)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn2.handle_xapp_subs_del_req(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq2, delmsg2)
	xappConn2.handle_xapp_subs_del_resp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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
	xapp.Logger.Info("TestSameSubsDiffRan")

	//Req1
	cretrans1 := xappConn1.newXappTransaction("", "RAN_NAME_1")
	xappConn1.handle_xapp_subs_req(t, nil, cretrans1)
	crereq1, cremsg1 := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.handle_xapp_subs_resp(t, cretrans1)

	//Req2
	cretrans2 := xappConn1.newXappTransaction("", "RAN_NAME_2")
	xappConn1.handle_xapp_subs_req(t, nil, cretrans2)
	crereq2, cremsg2 := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.handle_xapp_subs_resp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.newXappTransaction("", "RAN_NAME_1")
	xappConn1.handle_xapp_subs_del_req(t, deltrans1, e2SubsId1)
	delreq1, delmsg1 := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq1, delmsg1)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans1)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn1.newXappTransaction("", "RAN_NAME_2")
	xappConn1.handle_xapp_subs_del_req(t, deltrans2, e2SubsId2)
	delreq2, delmsg2 := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq2, delmsg2)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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

	xapp.Logger.Info("TestSubReqRetryInSubmgr start")

	// Xapp: Send SubsReq
	cretrans := xappConn1.handle_xapp_subs_req(t, nil, nil)

	// E2t: Receive 1st SubsReq
	e2termConn.handle_e2term_subs_req(t)

	// E2t: Receive 2nd SubsReq and send SubsResp
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq, cremsg)

	// Xapp: Receive SubsResp
	e2SubsId := xappConn1.handle_xapp_subs_resp(t, cretrans)

	deltrans := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId)
	delreq, delmsg := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq, delmsg)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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

	xapp.Logger.Info("TestSubReqTwoRetriesNoRespSubDelRespInSubmgr start")

	// Xapp: Send SubsReq
	xappConn1.handle_xapp_subs_req(t, nil, nil)

	// E2t: Receive 1st SubsReq
	e2termConn.handle_e2term_subs_req(t)

	// E2t: Receive 2nd SubsReq
	e2termConn.handle_e2term_subs_req(t)

	// E2t: Send receive SubsDelReq and send SubsResp
	delreq, delmsg := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq, delmsg)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq.RequestId.Seq, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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

	xapp.Logger.Info("TestSubReqTwoRetriesNoRespAtAllInSubmgr start")

	// Xapp: Send SubsReq
	xappConn1.handle_xapp_subs_req(t, nil, nil)

	// E2t: Receive 1st SubsReq
	e2termConn.handle_e2term_subs_req(t)

	// E2t: Receive 2nd SubsReq
	e2termConn.handle_e2term_subs_req(t)

	// E2t: Receive 1st SubsDelReq
	e2termConn.handle_e2term_subs_del_req(t)

	// E2t: Receive 2nd SubsDelReq
	delreq, _ := e2termConn.handle_e2term_subs_del_req(t)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq.RequestId.Seq, 15)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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
//     |      SubFail |              |
//     |<-------------|              |
//     |              |              |
//
//-----------------------------------------------------------------------------

func TestSubReqSubFailRespInSubmgr(t *testing.T) {

	xapp.Logger.Info("TestSubReqSubFailRespInSubmgr start")

	// Xapp: Send SubsReq
	cretrans := xappConn1.handle_xapp_subs_req(t, nil, nil)

	// E2t: Receive SubsReq and send SubsFail
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)
	fparams := &test_subs_fail_params{}
	fparams.Set(crereq)
	e2termConn.handle_e2term_subs_fail(t, fparams, cremsg)

	// Xapp: Receive SubsFail
	e2SubsId := xappConn1.handle_xapp_subs_fail(t, cretrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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

	xapp.Logger.Info("TestSubDelReqRetryInSubmgr start")

	// Subs Create
	cretrans := xappConn1.handle_xapp_subs_req(t, nil, nil)
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq, cremsg)
	e2SubsId := xappConn1.handle_xapp_subs_resp(t, cretrans)

	// Subs Delete
	// Xapp: Send SubsDelReq
	deltrans := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId)

	// E2t: Receive 1st SubsDelReq
	e2termConn.handle_e2term_subs_del_req(t)

	// E2t: Receive 2nd SubsDelReq and send SubsDelResp
	delreq, delmsg := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq, delmsg)

	// Xapp: Receive SubsDelResp
	xappConn1.handle_xapp_subs_del_resp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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

	xapp.Logger.Info("TestSubDelReTwoRetriesNoRespInSubmgr start")

	// Subs Create
	cretrans := xappConn1.handle_xapp_subs_req(t, nil, nil)
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq, cremsg)
	e2SubsId := xappConn1.handle_xapp_subs_resp(t, cretrans)

	// Subs Delete
	// Xapp: Send SubsDelReq
	deltrans := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId)

	// E2t: Receive 1st SubsDelReq
	e2termConn.handle_e2term_subs_del_req(t)

	// E2t: Receive 2nd SubsDelReq
	e2termConn.handle_e2term_subs_del_req(t)

	// Xapp: Receive SubsDelResp
	xappConn1.handle_xapp_subs_del_resp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
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

	xapp.Logger.Info("TestSubReqSubDelFailRespInSubmgr start")

	// Subs Create
	cretrans := xappConn1.handle_xapp_subs_req(t, nil, nil)
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq, cremsg)
	e2SubsId := xappConn1.handle_xapp_subs_resp(t, cretrans)

	// Xapp: Send SubsDelReq
	deltrans := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId)

	// E2t: Send receive SubsDelReq and send SubsDelFail
	delreq, delmsg := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_fail(t, delreq, delmsg)

	// Xapp: Receive SubsDelResp
	xappConn1.handle_xapp_subs_del_resp(t, deltrans)

	// Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameAction
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |              |              |
//     |              |              |
//     | SubReq1      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq1      |
//     |              |------------->|
//     |              |    SubResp1  |
//     |              |<-------------|
//     |    SubResp1  |              |
//     |<-------------|              |
//     |              |              |
//     | SubReq2      |              |
//     |------------->|              |
//     |              |              |
//     |    SubResp2  |              |
//     |<-------------|              |
//     |              |              |
//     | SubDelReq 1  |              |
//     |------------->|              |
//     |              |              |
//     | SubDelResp 1 |              |
//     |<-------------|              |
//     |              |              |
//     | SubDelReq 2  |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq 2  |
//     |              |------------->|
//     |              |              |
//     |              | SubDelReq 2  |
//     |              |------------->|
//     |              |              |
//     | SubDelResp 2 |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOkSameAction(t *testing.T) {
	xapp.Logger.Info("TestSubReqAndSubDelOkSameAction")

	//Req1
	rparams1 := &test_subs_req_params{}
	rparams1.Init()
	cretrans1 := xappConn1.handle_xapp_subs_req(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.handle_xapp_subs_resp(t, cretrans1)

	//Req2
	rparams2 := &test_subs_req_params{}
	rparams2.Init()
	cretrans2 := xappConn2.handle_xapp_subs_req(t, rparams2, nil)
	//crereq2, cremsg2 := e2termConn.handle_e2term_subs_req(t)
	//e2termConn.handle_e2term_subs_resp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn2.handle_xapp_subs_resp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId1)
	//e2termConn.handle_e2term_subs_del_req(t)
	//e2termConn.handle_e2term_subs_del_resp(t, delreq1, delmsg1)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans1)
	//Wait that subs is cleaned
	//mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn2.handle_xapp_subs_del_req(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq2, delmsg2)
	xappConn2.handle_xapp_subs_del_resp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelOkSameActionParallel
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |              |              |
//     |              |              |
//     | SubReq1      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq1      |
//     |              |------------->|
//     | SubReq2      |              |
//     |------------->|              |
//     |              |    SubResp1  |
//     |              |<-------------|
//     |    SubResp1  |              |
//     |<-------------|              |
//     |              |              |
//     |    SubResp2  |              |
//     |<-------------|              |
//     |              |              |
//     | SubDelReq 1  |              |
//     |------------->|              |
//     |              |              |
//     | SubDelResp 1 |              |
//     |<-------------|              |
//     |              |              |
//     | SubDelReq 2  |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq 2  |
//     |              |------------->|
//     |              |              |
//     |              | SubDelReq 2  |
//     |              |------------->|
//     |              |              |
//     | SubDelResp 2 |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelOkSameActionParallel(t *testing.T) {
	xapp.Logger.Info("TestSubReqAndSubDelOkSameActionParallel")

	//Req1
	rparams1 := &test_subs_req_params{}
	rparams1.Init()
	cretrans1 := xappConn1.handle_xapp_subs_req(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn.handle_e2term_subs_req(t)

	//Req2
	rparams2 := &test_subs_req_params{}
	rparams2.Init()
	cretrans2 := xappConn2.handle_xapp_subs_req(t, rparams2, nil)

	//Resp1
	e2termConn.handle_e2term_subs_resp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.handle_xapp_subs_resp(t, cretrans1)

	//Resp2
	e2SubsId2 := xappConn2.handle_xapp_subs_resp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId1)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans1)

	//Del2
	deltrans2 := xappConn2.handle_xapp_subs_del_req(t, nil, e2SubsId2)
	delreq2, delmsg2 := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq2, delmsg2)
	xappConn2.handle_xapp_subs_del_resp(t, deltrans2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelNokSameActionParallel
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |              |              |
//     |              |              |
//     | SubReq1      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq1      |
//     |              |------------->|
//     | SubReq2      |              |
//     |------------->|              |
//     |              |    SubFail1  |
//     |              |<-------------|
//     |    SubFail1  |              |
//     |<-------------|              |
//     |              |              |
//     |    SubFail2  |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelNokSameActionParallel(t *testing.T) {
	xapp.Logger.Info("TestSubReqAndSubDelNokSameActionParallel")

	//Req1
	rparams1 := &test_subs_req_params{}
	rparams1.Init()
	cretrans1 := xappConn1.handle_xapp_subs_req(t, rparams1, nil)
	crereq1, cremsg1 := e2termConn.handle_e2term_subs_req(t)

	//Req2
	rparams2 := &test_subs_req_params{}
	rparams2.Init()
	seqBef2 := mainCtrl.get_msgcounter(t)
	cretrans2 := xappConn2.handle_xapp_subs_req(t, rparams2, nil)
	mainCtrl.wait_msgcounter_change(t, seqBef2, 10)

	//E2T Fail
	fparams := &test_subs_fail_params{}
	fparams.Set(crereq1)
	e2termConn.handle_e2term_subs_fail(t, fparams, cremsg1)

	//Fail1
	e2SubsId1 := xappConn1.handle_xapp_subs_fail(t, cretrans1)
	//Fail2
	xappConn2.handle_xapp_subs_fail(t, cretrans2)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 15)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
	mainCtrl.wait_registry_empty(t, 10)
}

//-----------------------------------------------------------------------------
// TestSubReqAndSubDelNoAnswerSameActionParallel
//
//   stub                          stub
// +-------+     +---------+    +---------+
// | xapp  |     | submgr  |    | e2term  |
// +-------+     +---------+    +---------+
//     |              |              |
//     |              |              |
//     |              |              |
//     | SubReq1      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq1      |
//     |              |------------->|
//     | SubReq2      |              |
//     |------------->|              |
//     |              |              |
//     |              | SubReq1      |
//     |              |------------->|
//     |              |              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//
//-----------------------------------------------------------------------------
func TestSubReqAndSubDelNoAnswerSameActionParallel(t *testing.T) {
	xapp.Logger.Info("TestSubReqAndSubDelNoAnswerSameActionParallel")

	//Req1
	rparams1 := &test_subs_req_params{}
	rparams1.Init()
	xappConn1.handle_xapp_subs_req(t, rparams1, nil)

	e2termConn.handle_e2term_subs_req(t)

	//Req2
	rparams2 := &test_subs_req_params{}
	rparams2.Init()
	seqBef2 := mainCtrl.get_msgcounter(t)
	xappConn2.handle_xapp_subs_req(t, rparams2, nil)
	mainCtrl.wait_msgcounter_change(t, seqBef2, 10)

	//Req1 (retransmitted)
	e2termConn.handle_e2term_subs_req(t)

	delreq1, delmsg1 := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq1, delmsg1)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, delreq1.RequestId.Seq, 10)

	xappConn1.TestMsgCnt(t)
	xappConn2.TestMsgCnt(t)
	e2termConn.TestMsgCnt(t)
	mainCtrl.wait_registry_empty(t, 15)
}
