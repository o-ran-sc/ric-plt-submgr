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
	"strconv"
	"strings"
	"testing"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
var xapp_e2asnpacker e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type xappTransaction struct {
	tc   *testingXappStub
	xid  string
	meid *xapp.RMRMeid
}

type testingXappStub struct {
	testingRmrStubControl
	xid_seq uint64
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func createNewXappStub(desc string, rtfile string, port string, stat string) *testingXappStub {
	xappCtrl := &testingXappStub{}
	xappCtrl.testingRmrStubControl.init(desc, rtfile, port, stat, xappCtrl)
	xappCtrl.xid_seq = 1
	return xappCtrl
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *testingXappStub) newXid() string {
	var xid string
	xid = tc.desc + "_XID_" + strconv.FormatUint(uint64(tc.xid_seq), 10)
	tc.xid_seq++
	return xid
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *testingXappStub) newXappTransaction(xid *string, ranname string) *xappTransaction {
	trans := &xappTransaction{}
	trans.tc = tc
	if xid == nil {
		trans.xid = tc.newXid()
	} else {
		trans.xid = *xid
	}
	trans.meid = &xapp.RMRMeid{RanName: ranname}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *testingXappStub) Consume(params *xapp.RMRParams) (err error) {
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil
	msg := &RMRParams{params}

	if params.Mtype == 55555 {
		xapp.Logger.Info("(%s) Testing message ignore %s", tc.desc, msg.String())
		tc.active = true
		return
	}

	if strings.Contains(msg.Xid, tc.desc) {
		xapp.Logger.Info("(%s) Consume %s", tc.desc, msg.String())
		tc.IncMsgCnt()
		tc.rmrConChan <- msg
	} else {
		xapp.Logger.Info("(%s) Ignore %s", tc.desc, msg.String())
	}
	return
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type test_subs_req_params struct {
	req *e2ap.E2APSubscriptionRequest
}

func (p *test_subs_req_params) Init() {
	p.req = &e2ap.E2APSubscriptionRequest{}

	p.req.RequestId.Id = 1
	p.req.RequestId.Seq = 0
	p.req.FunctionId = 1

	p.req.EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
	p.req.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.StringPut("310150")
	p.req.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = 123
	p.req.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDHomeBits28

	// gnb -> enb outgoing
	// enb -> gnb incoming
	// X2 36423-f40.doc
	p.req.EventTriggerDefinition.InterfaceDirection = e2ap.E2AP_InterfaceDirectionIncoming
	p.req.EventTriggerDefinition.ProcedureCode = 5 //28 35
	p.req.EventTriggerDefinition.TypeOfMessage = e2ap.E2AP_InitiatingMessage

	p.req.ActionSetups = make([]e2ap.ActionToBeSetupItem, 1)
	p.req.ActionSetups[0].ActionId = 0
	p.req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypeReport
	p.req.ActionSetups[0].ActionDefinition.Present = false
	//p.req.ActionSetups[index].ActionDefinition.StyleId = 255
	//p.req.ActionSetups[index].ActionDefinition.ParamId = 222
	p.req.ActionSetups[0].SubsequentAction.Present = true
	p.req.ActionSetups[0].SubsequentAction.Type = e2ap.E2AP_SubSeqActionTypeContinue
	p.req.ActionSetups[0].SubsequentAction.TimetoWait = e2ap.E2AP_TimeToWaitZero

}

func (xappConn *testingXappStub) handle_xapp_subs_req(t *testing.T, rparams *test_subs_req_params, oldTrans *xappTransaction) *xappTransaction {
	xapp.Logger.Info("(%s) handle_xapp_subs_req", xappConn.desc)
	e2SubsReq := xapp_e2asnpacker.NewPackerSubscriptionRequest()

	//---------------------------------
	// xapp activity: Send Subs Req
	//---------------------------------
	xapp.Logger.Info("(%s) Send Subs Req", xappConn.desc)

	myparams := rparams

	if myparams == nil {
		myparams = &test_subs_req_params{}
		myparams.Init()
	}

	e2SubsReq.Set(myparams.req)
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

	params := &RMRParams{&xapp.RMRParams{}}
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
func (xappConn *testingXappStub) handle_xapp_subs_resp(t *testing.T, trans *xappTransaction) int {
	xapp.Logger.Info("(%s) handle_xapp_subs_resp", xappConn.desc)
	e2SubsResp := xapp_e2asnpacker.NewPackerSubscriptionResponse()
	var e2SubsId int

	//---------------------------------
	// xapp activity: Recv Subs Resp
	//---------------------------------
	select {
	case msg := <-xappConn.rmrConChan:
		xappConn.DecMsgCnt()
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
func (xappConn *testingXappStub) handle_xapp_subs_fail(t *testing.T, trans *xappTransaction) int {
	xapp.Logger.Info("(%s) handle_xapp_subs_fail", xappConn.desc)
	e2SubsFail := xapp_e2asnpacker.NewPackerSubscriptionFailure()
	var e2SubsId int

	//-------------------------------
	// xapp activity: Recv Subs Fail
	//-------------------------------
	select {
	case msg := <-xappConn.rmrConChan:
		xappConn.DecMsgCnt()
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_FAILURE"] {
			testError(t, "(%s) Received RIC_SUB_FAILURE wrong mtype expected %s got %s, error", xappConn.desc, "RIC_SUB_FAILURE", xapp.RicMessageTypeToName[msg.Mtype])
			return -1
		} else if msg.Xid != trans.xid {
			testError(t, "(%s) Received RIC_SUB_FAILURE wrong xid expected %s got %s, error", xappConn.desc, trans.xid, msg.Xid)
			return -1
		} else {
			packedData := &packer.PackedData{}
			packedData.Buf = msg.Payload
			e2SubsId = msg.SubId
			unpackerr := e2SubsFail.UnPack(packedData)

			if unpackerr != nil {
				testError(t, "(%s) RIC_SUB_FAILURE unpack failed err: %s", xappConn.desc, unpackerr.Error())
			}
			geterr, resp := e2SubsFail.Get()
			if geterr != nil {
				testError(t, "(%s) RIC_SUB_FAILURE get failed err: %s", xappConn.desc, geterr.Error())
			}

			xapp.Logger.Info("(%s) Recv Subs Fail rmr: xid=%s subid=%d, asn: seqnro=%d", xappConn.desc, msg.Xid, msg.SubId, resp.RequestId.Seq)
			return e2SubsId
		}
	case <-time.After(15 * time.Second):
		testError(t, "(%s) Not Received RIC_SUB_FAILURE within 15 secs", xappConn.desc)
		return -1
	}
	return -1
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (xappConn *testingXappStub) handle_xapp_subs_del_req(t *testing.T, oldTrans *xappTransaction, e2SubsId int) *xappTransaction {
	xapp.Logger.Info("(%s) handle_xapp_subs_del_req", xappConn.desc)
	e2SubsDelReq := xapp_e2asnpacker.NewPackerSubscriptionDeleteRequest()

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

	params := &RMRParams{&xapp.RMRParams{}}
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
func (xappConn *testingXappStub) handle_xapp_subs_del_resp(t *testing.T, trans *xappTransaction) {
	xapp.Logger.Info("(%s) handle_xapp_subs_del_resp", xappConn.desc)
	e2SubsDelResp := xapp_e2asnpacker.NewPackerSubscriptionDeleteResponse()

	//---------------------------------
	// xapp activity: Recv Subs Del Resp
	//---------------------------------
	select {
	case msg := <-xappConn.rmrConChan:
		xappConn.DecMsgCnt()
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
