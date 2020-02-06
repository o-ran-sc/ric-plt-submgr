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

package teststube2ap

import (
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/xapptweaks"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strconv"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
var xapp_e2asnpacker e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrTransactionId struct {
	xid  string
	meid *xapp.RMRMeid
}

func (trans *RmrTransactionId) String() string {
	return "trans(" + trans.xid + "/" + trans.meid.RanName + ")"
}

type E2Stub struct {
	teststub.RmrStubControl
	xid_seq uint64
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CreateNewE2Stub(desc string, rtfile string, port string, stat string, mtypeseed int) *E2Stub {
	xappCtrl := &E2Stub{}
	xappCtrl.RmrStubControl.Init(desc, rtfile, port, stat, mtypeseed)
	xappCtrl.xid_seq = 1
	xappCtrl.SetCheckXid(true)
	return xappCtrl
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) NewRmrTransactionId(xid string, ranname string) *RmrTransactionId {
	trans := &RmrTransactionId{}
	if len(xid) == 0 {
		trans.xid = tc.GetDesc() + "_XID_" + strconv.FormatUint(uint64(tc.xid_seq), 10)
		tc.xid_seq++
	} else {
		trans.xid = xid
	}
	trans.meid = &xapp.RMRMeid{RanName: ranname}
	tc.Logger.Info("New test %s", trans.String())
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2StubSubsReqParams struct {
	Req *e2ap.E2APSubscriptionRequest
}

func (p *E2StubSubsReqParams) Init() {
	p.Req = &e2ap.E2APSubscriptionRequest{}

	p.Req.RequestId.Id = 1
	p.Req.RequestId.Seq = 0
	p.Req.FunctionId = 1

	p.Req.EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
	p.Req.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.StringPut("310150")
	p.Req.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = 123
	p.Req.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDHomeBits28

	// gnb -> enb outgoing
	// enb -> gnb incoming
	// X2 36423-f40.doc
	p.Req.EventTriggerDefinition.InterfaceDirection = e2ap.E2AP_InterfaceDirectionIncoming
	p.Req.EventTriggerDefinition.ProcedureCode = 5 //28 35
	p.Req.EventTriggerDefinition.TypeOfMessage = e2ap.E2AP_InitiatingMessage

	p.Req.ActionSetups = make([]e2ap.ActionToBeSetupItem, 1)
	p.Req.ActionSetups[0].ActionId = 0
	p.Req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypeReport
	p.Req.ActionSetups[0].ActionDefinition.Present = false
	//p.Req.ActionSetups[index].ActionDefinition.StyleId = 255
	//p.Req.ActionSetups[index].ActionDefinition.ParamId = 222
	p.Req.ActionSetups[0].SubsequentAction.Present = true
	p.Req.ActionSetups[0].SubsequentAction.Type = e2ap.E2AP_SubSeqActionTypeContinue
	p.Req.ActionSetups[0].SubsequentAction.TimetoWait = e2ap.E2AP_TimeToWaitZero

}

func (tc *E2Stub) Handle_xapp_subs_req(t *testing.T, rparams *E2StubSubsReqParams, oldTrans *RmrTransactionId) *RmrTransactionId {

	trans := oldTrans
	if oldTrans == nil {
		trans = tc.NewRmrTransactionId("", "RAN_NAME_1")
	}

	tc.Logger.Info("Handle_xapp_subs_req %s", trans.String())
	e2SubsReq := xapp_e2asnpacker.NewPackerSubscriptionRequest()

	//---------------------------------
	// xapp activity: Send Subs Req
	//---------------------------------
	myparams := rparams

	if myparams == nil {
		myparams = &E2StubSubsReqParams{}
		myparams.Init()
	}

	err, packedMsg := e2SubsReq.Pack(myparams.Req)
	if err != nil {
		tc.TestError(t, "pack NOK %s %s", trans.String(), err.Error())
		return nil
	}
	tc.Logger.Debug("%s %s", trans.String(), e2SubsReq.String())

	params := xapptweaks.NewParams(nil)
	params.Mtype = xapp.RIC_SUB_REQ
	params.SubId = -1
	params.Payload = packedMsg.Buf
	params.Meid = trans.meid
	params.Xid = trans.xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB REQ: %s", params.String())
	snderr := tc.RmrSend(params)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s %s", trans.String(), snderr.Error())
		return nil
	}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) Handle_xapp_subs_resp(t *testing.T, trans *RmrTransactionId) uint32 {
	tc.Logger.Info("Handle_xapp_subs_resp")
	e2SubsResp := xapp_e2asnpacker.NewPackerSubscriptionResponse()
	var e2SubsId uint32

	//---------------------------------
	// xapp activity: Recv Subs Resp
	//---------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_RESP"] {
			tc.TestError(t, "Received RIC_SUB_RESP wrong mtype expected %s got %s, error", "RIC_SUB_RESP", xapp.RicMessageTypeToName[msg.Mtype])
			return 0
		} else if msg.Xid != trans.xid {
			tc.TestError(t, "Received RIC_SUB_RESP wrong xid expected %s got %s, error", trans.xid, msg.Xid)
			return 0
		} else {
			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			if msg.SubId > 0 {
				e2SubsId = uint32(msg.SubId)
			} else {
				e2SubsId = 0
			}
			unpackerr, resp := e2SubsResp.UnPack(packedData)
			if unpackerr != nil {
				tc.TestError(t, "RIC_SUB_RESP unpack failed err: %s", unpackerr.Error())
			}
			tc.Logger.Info("Recv Subs Resp rmr: xid=%s subid=%d, asn: seqnro=%d", msg.Xid, msg.SubId, resp.RequestId.Seq)
			return e2SubsId
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
	return 0
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) Handle_xapp_subs_fail(t *testing.T, trans *RmrTransactionId) uint32 {
	tc.Logger.Info("Handle_xapp_subs_fail")
	e2SubsFail := xapp_e2asnpacker.NewPackerSubscriptionFailure()
	var e2SubsId uint32

	//-------------------------------
	// xapp activity: Recv Subs Fail
	//-------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_FAILURE"] {
			tc.TestError(t, "Received RIC_SUB_FAILURE wrong mtype expected %s got %s, error", "RIC_SUB_FAILURE", xapp.RicMessageTypeToName[msg.Mtype])
			return 0
		} else if msg.Xid != trans.xid {
			tc.TestError(t, "Received RIC_SUB_FAILURE wrong xid expected %s got %s, error", trans.xid, msg.Xid)
			return 0
		} else {
			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			if msg.SubId > 0 {
				e2SubsId = uint32(msg.SubId)
			} else {
				e2SubsId = 0
			}
			unpackerr, resp := e2SubsFail.UnPack(packedData)
			if unpackerr != nil {
				tc.TestError(t, "RIC_SUB_FAILURE unpack failed err: %s", unpackerr.Error())
			}
			tc.Logger.Info("Recv Subs Fail rmr: xid=%s subid=%d, asn: seqnro=%d", msg.Xid, msg.SubId, resp.RequestId.Seq)
			return e2SubsId
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
	return 0
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) Handle_xapp_subs_del_req(t *testing.T, oldTrans *RmrTransactionId, e2SubsId uint32) *RmrTransactionId {

	trans := oldTrans
	if oldTrans == nil {
		trans = tc.NewRmrTransactionId("", "RAN_NAME_1")
	}

	tc.Logger.Info("Handle_xapp_subs_del_req %s", trans.String())
	e2SubsDelReq := xapp_e2asnpacker.NewPackerSubscriptionDeleteRequest()
	//---------------------------------
	// xapp activity: Send Subs Del Req
	//---------------------------------
	req := &e2ap.E2APSubscriptionDeleteRequest{}
	req.RequestId.Id = 1
	req.RequestId.Seq = e2SubsId
	req.FunctionId = 1

	err, packedMsg := e2SubsDelReq.Pack(req)
	if err != nil {
		tc.TestError(t, "pack NOK %s %s", trans.String(), err.Error())
		return nil
	}
	tc.Logger.Debug("%s %s", trans.String(), e2SubsDelReq.String())

	params := xapptweaks.NewParams(nil)
	params.Mtype = xapp.RIC_SUB_DEL_REQ
	params.SubId = int(e2SubsId)
	params.Payload = packedMsg.Buf
	params.Meid = trans.meid
	params.Xid = trans.xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB DEL REQ: %s", params.String())
	snderr := tc.RmrSend(params)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s %s", trans.String(), snderr.Error())
		return nil
	}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) Handle_xapp_subs_del_resp(t *testing.T, trans *RmrTransactionId) {
	tc.Logger.Info("Handle_xapp_subs_del_resp")
	e2SubsDelResp := xapp_e2asnpacker.NewPackerSubscriptionDeleteResponse()

	//---------------------------------
	// xapp activity: Recv Subs Del Resp
	//---------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_DEL_RESP"] {
			tc.TestError(t, "Received RIC_SUB_DEL_RESP wrong mtype expected %s got %s, error", "RIC_SUB_DEL_RESP", xapp.RicMessageTypeToName[msg.Mtype])
			return
		} else if trans != nil && msg.Xid != trans.xid {
			tc.TestError(t, "Received RIC_SUB_DEL_RESP wrong xid expected %s got %s, error", trans.xid, msg.Xid)
			return
		} else {
			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr, resp := e2SubsDelResp.UnPack(packedData)
			if unpackerr != nil {
				tc.TestError(t, "RIC_SUB_DEL_RESP unpack failed err: %s", unpackerr.Error())
			}
			tc.Logger.Info("Recv Subs Del Resp rmr: xid=%s subid=%d, asn: seqnro=%d", msg.Xid, msg.SubId, resp.RequestId.Seq)
			return
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
}
