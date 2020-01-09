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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/packer"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"testing"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
var e2asnpacker e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (xappConn *testingXappControl) handle_xapp_subs_req(t *testing.T, oldTrans *xappTransaction) *xappTransaction {
	xapp.Logger.Info("handle_xapp_subs_req")
	e2SubsReq := e2asnpacker.NewPackerSubscriptionRequest()

	//---------------------------------
	// xapp activity: Send Subs Req
	//---------------------------------
	xapp.Logger.Info("(%s) Send Subs Req", xappConn.desc)

	req := &e2ap.E2APSubscriptionRequest{}

	req.RequestId.Id = 1
	req.RequestId.Seq = 0
	req.FunctionId = 1

	req.EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
	req.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.StringPut("310150")
	req.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = 123
	req.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDHomeBits28

	// gnb -> enb outgoing
	// enb -> gnb incoming
	// X2 36423-f40.doc
	req.EventTriggerDefinition.InterfaceDirection = e2ap.E2AP_InterfaceDirectionIncoming
	req.EventTriggerDefinition.ProcedureCode = 5 //28 35
	req.EventTriggerDefinition.TypeOfMessage = e2ap.E2AP_InitiatingMessage

	req.ActionSetups = make([]e2ap.ActionToBeSetupItem, 1)
	req.ActionSetups[0].ActionId = 0
	req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypeReport
	req.ActionSetups[0].ActionDefinition.Present = false
	//req.ActionSetups[index].ActionDefinition.StyleId = 255
	//req.ActionSetups[index].ActionDefinition.ParamId = 222
	req.ActionSetups[0].SubsequentAction.Present = true
	req.ActionSetups[0].SubsequentAction.Type = e2ap.E2AP_SubSeqActionTypeContinue
	req.ActionSetups[0].SubsequentAction.TimetoWait = e2ap.E2AP_TimeToWaitZero

	e2SubsReq.Set(req)
	xapp.Logger.Debug("%s", e2SubsReq.String())
	err, packedMsg := e2SubsReq.Pack(nil)
	if err != nil {
		testError(t, "(%s) pack NOK %s", xappConn.desc, err.Error())
		return nil
	}

	var trans *xappTransaction = oldTrans
	if trans == nil {
		trans = xappConn.newXappTransaction(nil, "RAN_NAME_1")
	}

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_REQ
	params.SubId = -1
	params.Payload = packedMsg.Buf
	params.Meid = trans.meid
	params.Xid = trans.xid
	params.Mbuf = nil

	snderr := xappConn.RmrSend(params)
	if snderr != nil {
		testError(t, "(%s) RMR SEND FAILED: %s", xappConn.desc, snderr.Error())
		return nil
	}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (xappConn *testingXappControl) handle_xapp_subs_resp(t *testing.T, trans *xappTransaction) int {
	xapp.Logger.Info("handle_xapp_subs_resp")
	e2SubsResp := e2asnpacker.NewPackerSubscriptionResponse()
	var e2SubsId int

	//---------------------------------
	// xapp activity: Recv Subs Resp
	//---------------------------------
	select {
	case msg := <-xappConn.rmrConChan:
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_RESP"] {
			testError(t, "(%s) Received RIC_SUB_RESP wrong mtype expected %s got %s, error", xappConn.desc, "RIC_SUB_RESP", xapp.RicMessageTypeToName[msg.Mtype])
			return -1
		} else if msg.Xid != trans.xid {
			testError(t, "(%s) Received RIC_SUB_RESP wrong xid expected %s got %s, error", xappConn.desc, trans.xid, msg.Xid)
			return -1
		} else {
			packedData := &packer.PackedData{}
			packedData.Buf = msg.Payload
			e2SubsId = msg.SubId
			unpackerr := e2SubsResp.UnPack(packedData)

			if unpackerr != nil {
				testError(t, "(%s) RIC_SUB_RESP unpack failed err: %s", xappConn.desc, unpackerr.Error())
			}
			geterr, resp := e2SubsResp.Get()
			if geterr != nil {
				testError(t, "(%s) RIC_SUB_RESP get failed err: %s", xappConn.desc, geterr.Error())
			}

			xapp.Logger.Info("(%s) Recv Subs Resp rmr: xid=%s subid=%d, asn: seqnro=%d", xappConn.desc, msg.Xid, msg.SubId, resp.RequestId.Seq)
			return e2SubsId
		}
	case <-time.After(15 * time.Second):
		testError(t, "(%s) Not Received RIC_SUB_RESP within 15 secs", xappConn.desc)
		return -1
	}
	return -1
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (xappConn *testingXappControl) handle_xapp_subs_del_req(t *testing.T, oldTrans *xappTransaction, e2SubsId int) *xappTransaction {
	xapp.Logger.Info("handle_xapp_subs_del_req")
	e2SubsDelReq := e2asnpacker.NewPackerSubscriptionDeleteRequest()

	//---------------------------------
	// xapp activity: Send Subs Del Req
	//---------------------------------
	xapp.Logger.Info("(%s) Send Subs Del Req", xappConn.desc)

	req := &e2ap.E2APSubscriptionDeleteRequest{}
	req.RequestId.Id = 1
	req.RequestId.Seq = uint32(e2SubsId)
	req.FunctionId = 1

	e2SubsDelReq.Set(req)
	xapp.Logger.Debug("%s", e2SubsDelReq.String())
	err, packedMsg := e2SubsDelReq.Pack(nil)
	if err != nil {
		testError(t, "(%s) pack NOK %s", xappConn.desc, err.Error())
		return nil
	}

	var trans *xappTransaction = oldTrans
	if trans == nil {
		trans = xappConn.newXappTransaction(nil, "RAN_NAME_1")
	}

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_DEL_REQ
	params.SubId = e2SubsId
	params.Payload = packedMsg.Buf
	params.Meid = trans.meid
	params.Xid = trans.xid
	params.Mbuf = nil

	snderr := xappConn.RmrSend(params)
	if snderr != nil {
		testError(t, "(%s) RMR SEND FAILED: %s", xappConn.desc, snderr.Error())
		return nil
	}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (xappConn *testingXappControl) handle_xapp_subs_del_resp(t *testing.T, trans *xappTransaction) {
	xapp.Logger.Info("handle_xapp_subs_del_resp")
	e2SubsDelResp := e2asnpacker.NewPackerSubscriptionDeleteResponse()

	//---------------------------------
	// xapp activity: Recv Subs Del Resp
	//---------------------------------
	select {
	case msg := <-xappConn.rmrConChan:
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_DEL_RESP"] {
			testError(t, "(%s) Received RIC_SUB_DEL_RESP wrong mtype expected %s got %s, error", xappConn.desc, "RIC_SUB_DEL_RESP", xapp.RicMessageTypeToName[msg.Mtype])
			return
		} else if msg.Xid != trans.xid {
			testError(t, "(%s) Received RIC_SUB_DEL_RESP wrong xid expected %s got %s, error", xappConn.desc, trans.xid, msg.Xid)
			return
		} else {
			packedData := &packer.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr := e2SubsDelResp.UnPack(packedData)
			if unpackerr != nil {
				testError(t, "(%s) RIC_SUB_DEL_RESP unpack failed err: %s", xappConn.desc, unpackerr.Error())
			}
			geterr, resp := e2SubsDelResp.Get()
			if geterr != nil {
				testError(t, "(%s) RIC_SUB_DEL_RESP get failed err: %s", xappConn.desc, geterr.Error())
			}
			xapp.Logger.Info("(%s) Recv Subs Del Resp rmr: xid=%s subid=%d, asn: seqnro=%d", xappConn.desc, msg.Xid, msg.SubId, resp.RequestId.Seq)
			return
		}
	case <-time.After(15 * time.Second):
		testError(t, "(%s) Not Received RIC_SUB_DEL_RESP within 15 secs", xappConn.desc)
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e2termConn *testingE2termControl) handle_e2term_subs_req(t *testing.T) (*e2ap.E2APSubscriptionRequest, *xapp.RMRParams) {
	xapp.Logger.Info("handle_e2term_subs_req")
	e2SubsReq := e2asnpacker.NewPackerSubscriptionRequest()

	//---------------------------------
	// e2term activity: Recv Subs Req
	//---------------------------------
	select {
	case msg := <-e2termConn.rmrConChan:
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_REQ"] {
			testError(t, "(%s) Received wrong mtype expected %s got %s, error", e2termConn.desc, "RIC_SUB_REQ", xapp.RicMessageTypeToName[msg.Mtype])
		} else {
			xapp.Logger.Info("(%s) Recv Subs Req", e2termConn.desc)
			packedData := &packer.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr := e2SubsReq.UnPack(packedData)
			if unpackerr != nil {
				testError(t, "(%s) RIC_SUB_REQ unpack failed err: %s", e2termConn.desc, unpackerr.Error())
			}
			geterr, req := e2SubsReq.Get()
			if geterr != nil {
				testError(t, "(%s) RIC_SUB_REQ get failed err: %s", e2termConn.desc, geterr.Error())
			}
			return req, msg
		}
	case <-time.After(15 * time.Second):
		testError(t, "(%s) Not Received RIC_SUB_REQ within 15 secs", e2termConn.desc)
	}
	return nil, nil
}

func (e2termConn *testingE2termControl) handle_e2term_subs_resp(t *testing.T, req *e2ap.E2APSubscriptionRequest, msg *xapp.RMRParams) {
	xapp.Logger.Info("handle_e2term_subs_resp")
	e2SubsResp := e2asnpacker.NewPackerSubscriptionResponse()

	//---------------------------------
	// e2term activity: Send Subs Resp
	//---------------------------------
	xapp.Logger.Info("(%s) Send Subs Resp", e2termConn.desc)

	resp := &e2ap.E2APSubscriptionResponse{}

	resp.RequestId.Id = req.RequestId.Id
	resp.RequestId.Seq = req.RequestId.Seq
	resp.FunctionId = req.FunctionId

	resp.ActionAdmittedList.Items = make([]e2ap.ActionAdmittedItem, len(req.ActionSetups))
	for index := int(0); index < len(req.ActionSetups); index++ {
		resp.ActionAdmittedList.Items[index].ActionId = req.ActionSetups[index].ActionId
	}

	for index := uint64(0); index < 1; index++ {
		item := e2ap.ActionNotAdmittedItem{}
		item.ActionId = index
		item.Cause.Content = 1
		item.Cause.CauseVal = 1
		resp.ActionNotAdmittedList.Items = append(resp.ActionNotAdmittedList.Items, item)
	}

	e2SubsResp.Set(resp)
	xapp.Logger.Debug("%s", e2SubsResp.String())
	packerr, packedMsg := e2SubsResp.Pack(nil)
	if packerr != nil {
		testError(t, "(%s) pack NOK %s", e2termConn.desc, packerr.Error())
	}

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_RESP
	params.SubId = msg.SubId
	params.Payload = packedMsg.Buf
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	snderr := e2termConn.RmrSend(params)
	if snderr != nil {
		testError(t, "(%s) RMR SEND FAILED: %s", e2termConn.desc, snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (e2termConn *testingE2termControl) handle_e2term_subs_del_req(t *testing.T) (*e2ap.E2APSubscriptionDeleteRequest, *xapp.RMRParams) {
	xapp.Logger.Info("handle_e2term_subs_del_req")
	e2SubsDelReq := e2asnpacker.NewPackerSubscriptionDeleteRequest()

	//---------------------------------
	// e2term activity: Recv Subs Del Req
	//---------------------------------
	select {
	case msg := <-e2termConn.rmrConChan:
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_DEL_REQ"] {
			testError(t, "(%s) Received wrong mtype expected %s got %s, error", e2termConn.desc, "RIC_SUB_DEL_REQ", xapp.RicMessageTypeToName[msg.Mtype])
		} else {
			xapp.Logger.Info("(%s) Recv Subs Del Req", e2termConn.desc)

			packedData := &packer.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr := e2SubsDelReq.UnPack(packedData)
			if unpackerr != nil {
				testError(t, "(%s) RIC_SUB_DEL_REQ unpack failed err: %s", e2termConn.desc, unpackerr.Error())
			}
			geterr, req := e2SubsDelReq.Get()
			if geterr != nil {
				testError(t, "(%s) RIC_SUB_DEL_REQ get failed err: %s", e2termConn.desc, geterr.Error())
			}
			return req, msg
		}
	case <-time.After(15 * time.Second):
		testError(t, "(%s) Not Received RIC_SUB_DEL_REQ within 15 secs", e2termConn.desc)
	}
	return nil, nil
}

func handle_e2term_recv_empty() bool {
	if len(e2termConn.rmrConChan) > 0 {
		return false
	}
	return true
}

func (e2termConn *testingE2termControl) handle_e2term_subs_del_resp(t *testing.T, req *e2ap.E2APSubscriptionDeleteRequest, msg *xapp.RMRParams) {
	xapp.Logger.Info("handle_e2term_subs_del_resp")
	e2SubsDelResp := e2asnpacker.NewPackerSubscriptionDeleteResponse()

	//---------------------------------
	// e2term activity: Send Subs Del Resp
	//---------------------------------
	xapp.Logger.Info("(%s) Send Subs Del Resp", e2termConn.desc)

	resp := &e2ap.E2APSubscriptionDeleteResponse{}
	resp.RequestId.Id = req.RequestId.Id
	resp.RequestId.Seq = req.RequestId.Seq
	resp.FunctionId = req.FunctionId

	e2SubsDelResp.Set(resp)
	xapp.Logger.Debug("%s", e2SubsDelResp.String())
	packerr, packedMsg := e2SubsDelResp.Pack(nil)
	if packerr != nil {
		testError(t, "(%s) pack NOK %s", e2termConn.desc, packerr.Error())
	}

	params := &xapp.RMRParams{}
	params.Mtype = xapp.RIC_SUB_DEL_RESP
	params.SubId = msg.SubId
	params.Payload = packedMsg.Buf
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	snderr := e2termConn.RmrSend(params)
	if snderr != nil {
		testError(t, "(%s) RMR SEND FAILED: %s", e2termConn.desc, snderr.Error())
	}

}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (mc *testingMainControl) wait_subs_clean(t *testing.T, e2SubsId int, secs int) bool {
	i := 1
	for ; i <= secs*2; i++ {
		if mc.c.registry.GetSubscription(uint16(e2SubsId)) == nil {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	testError(t, "(general) no clean within %d secs", secs)
	return false
}

func (mc *testingMainControl) wait_subs_trans_clean(t *testing.T, e2SubsId int, secs int) bool {
	i := 1
	for ; i <= secs*2; i++ {
		subs := mc.c.registry.GetSubscription(uint16(e2SubsId))
		if subs == nil {
			return true
		}
		trans := subs.GetTransaction()
		if trans == nil {
			return true
		}
		time.Sleep(500 * time.Millisecond)
	}
	testError(t, "(general) no clean within %d secs", secs)
	return false
}

func (mc *testingMainControl) get_seqcnt(t *testing.T) uint16 {
	mc.c.registry.mutex.Lock()
	defer mc.c.registry.mutex.Unlock()
	return mc.c.registry.counter
}

func (mc *testingMainControl) wait_seqcnt_change(t *testing.T, orig uint16, secs int) (uint16, bool) {
	i := 1
	for ; i <= secs*2; i++ {
		mc.c.registry.mutex.Lock()
		curr := mc.c.registry.counter
		mc.c.registry.mutex.Unlock()
		if curr != orig {
			return curr, true
		}
		time.Sleep(500 * time.Millisecond)
	}
	testError(t, "(general) no seq change within %d secs", secs)
	return 0, false
}

func (mc *testingMainControl) get_msgcounter(t *testing.T) uint64 {
	return mc.c.msgCounter
}

func (mc *testingMainControl) wait_msgcounter_change(t *testing.T, orig uint64, secs int) (uint64, bool) {
	i := 1
	for ; i <= secs*2; i++ {
		curr := mc.c.msgCounter
		if curr != orig {
			return curr, true
		}
		time.Sleep(500 * time.Millisecond)
	}
	testError(t, "(general) no msg counter change within %d secs", secs)
	return 0, false
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

	cretrans := xappConn1.handle_xapp_subs_req(t, nil)
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq, cremsg)
	e2SubsId := xappConn1.handle_xapp_subs_resp(t, cretrans)

	deltrans := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId)
	delreq, delmsg := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq, delmsg)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
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
	cretrans := xappConn1.handle_xapp_subs_req(t, nil)
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)

	seqBef := mainCtrl.get_msgcounter(t)
	xappConn1.handle_xapp_subs_req(t, cretrans) //Retransmitted SubReq
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
	cretrans := xappConn1.handle_xapp_subs_req(t, nil)
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
//     | SubDelReq    |              |
//     |------------->|              |
//     |              |              |
//     |              | SubDelReq    |
//     |              |------------->|
//     |              |              |
//     | SubDelReq    |              |
//     | (same sub)   |              |
//     | (diff xid)   |              |
//     |------------->|              |
//     |              |              |
//     |              |   SubDelResp |
//     |              |<-------------|
//     |              |              |
//     |   SubDelResp |              |
//     |<-------------|              |
//
//-----------------------------------------------------------------------------
func TestSubDelReqCollision(t *testing.T) {
	xapp.Logger.Info("TestSubDelReqCollision")

	//Subs Create
	cretrans := xappConn1.handle_xapp_subs_req(t, nil)
	crereq, cremsg := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq, cremsg)
	e2SubsId := xappConn1.handle_xapp_subs_resp(t, cretrans)

	//Subs Delete
	deltrans := xappConn1.handle_xapp_subs_del_req(t, nil, e2SubsId)
	delreq, delmsg := e2termConn.handle_e2term_subs_del_req(t)

	seqBef := mainCtrl.get_msgcounter(t)
	deltranscol := xappConn1.newXappTransaction(nil, "RAN_NAME_1")
	xappConn1.handle_xapp_subs_del_req(t, deltranscol, e2SubsId) //Colliding SubDelReq
	mainCtrl.wait_msgcounter_change(t, seqBef, 10)

	e2termConn.handle_e2term_subs_del_resp(t, delreq, delmsg)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans)

	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId, 10)
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
	cretrans1 := xappConn1.handle_xapp_subs_req(t, nil)
	crereq1, cremsg1 := e2termConn.handle_e2term_subs_req(t)

	//Req2
	cretrans2 := xappConn2.handle_xapp_subs_req(t, nil)
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
	cretrans1 := xappConn1.newXappTransaction(nil, "RAN_NAME_1")
	xappConn1.handle_xapp_subs_req(t, cretrans1)
	crereq1, cremsg1 := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq1, cremsg1)
	e2SubsId1 := xappConn1.handle_xapp_subs_resp(t, cretrans1)

	//Req2
	cretrans2 := xappConn1.newXappTransaction(nil, "RAN_NAME_2")
	xappConn1.handle_xapp_subs_req(t, cretrans2)
	crereq2, cremsg2 := e2termConn.handle_e2term_subs_req(t)
	e2termConn.handle_e2term_subs_resp(t, crereq2, cremsg2)
	e2SubsId2 := xappConn1.handle_xapp_subs_resp(t, cretrans2)

	//Del1
	deltrans1 := xappConn1.newXappTransaction(nil, "RAN_NAME_1")
	xappConn1.handle_xapp_subs_del_req(t, deltrans1, e2SubsId1)
	delreq1, delmsg1 := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq1, delmsg1)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans1)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId1, 10)

	//Del2
	deltrans2 := xappConn1.newXappTransaction(nil, "RAN_NAME_2")
	xappConn1.handle_xapp_subs_del_req(t, deltrans2, e2SubsId2)
	delreq2, delmsg2 := e2termConn.handle_e2term_subs_del_req(t)
	e2termConn.handle_e2term_subs_del_resp(t, delreq2, delmsg2)
	xappConn1.handle_xapp_subs_del_resp(t, deltrans2)
	//Wait that subs is cleaned
	mainCtrl.wait_subs_clean(t, e2SubsId2, 10)
}
