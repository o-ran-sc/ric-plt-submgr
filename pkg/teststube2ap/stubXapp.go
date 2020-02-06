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
type XappTransaction struct {
	xid  string
	meid *xapp.RMRMeid
}

func (trans *XappTransaction) String() string {
	return "trans(" + trans.xid + "/" + trans.meid.RanName + ")"
}

type XappStub struct {
	teststub.RmrStubControl
	xid_seq uint64
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CreateNewXappStub(desc string, rtfile string, port string, stat string, mtypeseed int) *XappStub {
	xappCtrl := &XappStub{}
	xappCtrl.RmrStubControl.Init(desc, rtfile, port, stat, xappCtrl, mtypeseed)
	xappCtrl.xid_seq = 1
	xappCtrl.SetCheckXid(true)
	return xappCtrl
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *XappStub) NewXappTransaction(xid string, ranname string) *XappTransaction {
	trans := &XappTransaction{}
	if len(xid) == 0 {
		trans.xid = tc.GetDesc() + "_XID_" + strconv.FormatUint(uint64(tc.xid_seq), 10)
		tc.xid_seq++
	} else {
		trans.xid = xid
	}
	trans.meid = &xapp.RMRMeid{RanName: ranname}
	xapp.Logger.Info("(%s) New test %s", tc.GetDesc(), trans.String())
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Test_subs_req_params struct {
	Req *e2ap.E2APSubscriptionRequest
}

func (p *Test_subs_req_params) Init() {
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

func (tc *XappStub) Handle_xapp_subs_req(t *testing.T, rparams *Test_subs_req_params, oldTrans *XappTransaction) *XappTransaction {

	trans := oldTrans
	if oldTrans == nil {
		trans = tc.NewXappTransaction("", "RAN_NAME_1")
	}

	xapp.Logger.Info("(%s) Handle_xapp_subs_req %s", tc.GetDesc(), trans.String())
	e2SubsReq := xapp_e2asnpacker.NewPackerSubscriptionRequest()

	//---------------------------------
	// xapp activity: Send Subs Req
	//---------------------------------
	xapp.Logger.Info("(%s) Send Subs Req %s", tc.GetDesc(), trans.String())

	myparams := rparams

	if myparams == nil {
		myparams = &Test_subs_req_params{}
		myparams.Init()
	}

	err, packedMsg := e2SubsReq.Pack(myparams.Req)
	if err != nil {
		teststub.TestError(t, "(%s) pack NOK %s %s", tc.GetDesc(), trans.String(), err.Error())
		return nil
	}
	xapp.Logger.Debug("(%s) %s %s", tc.GetDesc(), trans.String(), e2SubsReq.String())

	params := xapptweaks.NewParams(nil)
	params.Mtype = xapp.RIC_SUB_REQ
	params.SubId = -1
	params.Payload = packedMsg.Buf
	params.Meid = trans.meid
	params.Xid = trans.xid
	params.Mbuf = nil

	snderr := tc.RmrSend("subs_req", params)
	if snderr != nil {
		teststub.TestError(t, "(%s) RMR SEND FAILED: %s %s", tc.GetDesc(), trans.String(), snderr.Error())
		return nil
	}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *XappStub) Handle_xapp_subs_resp(t *testing.T, trans *XappTransaction) uint32 {
	xapp.Logger.Info("(%s) Handle_xapp_subs_resp", tc.GetDesc())
	e2SubsResp := xapp_e2asnpacker.NewPackerSubscriptionResponse()
	var e2SubsId uint32

	//---------------------------------
	// xapp activity: Recv Subs Resp
	//---------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_RESP"] {
			teststub.TestError(t, "(%s) Received RIC_SUB_RESP wrong mtype expected %s got %s, error", tc.GetDesc(), "RIC_SUB_RESP", xapp.RicMessageTypeToName[msg.Mtype])
			return 0
		} else if msg.Xid != trans.xid {
			teststub.TestError(t, "(%s) Received RIC_SUB_RESP wrong xid expected %s got %s, error", tc.GetDesc(), trans.xid, msg.Xid)
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
				teststub.TestError(t, "(%s) RIC_SUB_RESP unpack failed err: %s", tc.GetDesc(), unpackerr.Error())
			}
			xapp.Logger.Info("(%s) Recv Subs Resp rmr: xid=%s subid=%d, asn: seqnro=%d", tc.GetDesc(), msg.Xid, msg.SubId, resp.RequestId.Seq)
			return e2SubsId
		}
	} else {
		teststub.TestError(t, "(%s) Not Received msg within %d secs", tc.GetDesc(), 15)
	}
	return 0
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *XappStub) Handle_xapp_subs_fail(t *testing.T, trans *XappTransaction) uint32 {
	xapp.Logger.Info("(%s) Handle_xapp_subs_fail", tc.GetDesc())
	e2SubsFail := xapp_e2asnpacker.NewPackerSubscriptionFailure()
	var e2SubsId uint32

	//-------------------------------
	// xapp activity: Recv Subs Fail
	//-------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_FAILURE"] {
			teststub.TestError(t, "(%s) Received RIC_SUB_FAILURE wrong mtype expected %s got %s, error", tc.GetDesc(), "RIC_SUB_FAILURE", xapp.RicMessageTypeToName[msg.Mtype])
			return 0
		} else if msg.Xid != trans.xid {
			teststub.TestError(t, "(%s) Received RIC_SUB_FAILURE wrong xid expected %s got %s, error", tc.GetDesc(), trans.xid, msg.Xid)
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
				teststub.TestError(t, "(%s) RIC_SUB_FAILURE unpack failed err: %s", tc.GetDesc(), unpackerr.Error())
			}
			xapp.Logger.Info("(%s) Recv Subs Fail rmr: xid=%s subid=%d, asn: seqnro=%d", tc.GetDesc(), msg.Xid, msg.SubId, resp.RequestId.Seq)
			return e2SubsId
		}
	} else {
		teststub.TestError(t, "(%s) Not Received msg within %d secs", tc.GetDesc(), 15)
	}
	return 0
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *XappStub) Handle_xapp_subs_del_req(t *testing.T, oldTrans *XappTransaction, e2SubsId uint32) *XappTransaction {

	trans := oldTrans
	if oldTrans == nil {
		trans = tc.NewXappTransaction("", "RAN_NAME_1")
	}

	xapp.Logger.Info("(%s) Handle_xapp_subs_del_req %s", tc.GetDesc(), trans.String())
	e2SubsDelReq := xapp_e2asnpacker.NewPackerSubscriptionDeleteRequest()
	//---------------------------------
	// xapp activity: Send Subs Del Req
	//---------------------------------
	xapp.Logger.Info("(%s) Send Subs Del Req  %s", tc.GetDesc(), trans.String())

	req := &e2ap.E2APSubscriptionDeleteRequest{}
	req.RequestId.Id = 1
	req.RequestId.Seq = e2SubsId
	req.FunctionId = 1

	err, packedMsg := e2SubsDelReq.Pack(req)
	if err != nil {
		teststub.TestError(t, "(%s) pack NOK %s %s", tc.GetDesc(), trans.String(), err.Error())
		return nil
	}
	xapp.Logger.Debug("(%s) %s %s", tc.GetDesc(), trans.String(), e2SubsDelReq.String())

	params := xapptweaks.NewParams(nil)
	params.Mtype = xapp.RIC_SUB_DEL_REQ
	params.SubId = int(e2SubsId)
	params.Payload = packedMsg.Buf
	params.Meid = trans.meid
	params.Xid = trans.xid
	params.Mbuf = nil

	snderr := tc.RmrSend("subs_del_req", params)

	if snderr != nil {
		teststub.TestError(t, "(%s) RMR SEND FAILED: %s %s", tc.GetDesc(), trans.String(), snderr.Error())
		return nil
	}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *XappStub) Handle_xapp_subs_del_resp(t *testing.T, trans *XappTransaction) {
	xapp.Logger.Info("(%s) Handle_xapp_subs_del_resp", tc.GetDesc())
	e2SubsDelResp := xapp_e2asnpacker.NewPackerSubscriptionDeleteResponse()

	//---------------------------------
	// xapp activity: Recv Subs Del Resp
	//---------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_DEL_RESP"] {
			teststub.TestError(t, "(%s) Received RIC_SUB_DEL_RESP wrong mtype expected %s got %s, error", tc.GetDesc(), "RIC_SUB_DEL_RESP", xapp.RicMessageTypeToName[msg.Mtype])
			return
		} else if trans != nil && msg.Xid != trans.xid {
			teststub.TestError(t, "(%s) Received RIC_SUB_DEL_RESP wrong xid expected %s got %s, error", tc.GetDesc(), trans.xid, msg.Xid)
			return
		} else {
			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr, resp := e2SubsDelResp.UnPack(packedData)
			if unpackerr != nil {
				teststub.TestError(t, "(%s) RIC_SUB_DEL_RESP unpack failed err: %s", tc.GetDesc(), unpackerr.Error())
			}
			xapp.Logger.Info("(%s) Recv Subs Del Resp rmr: xid=%s subid=%d, asn: seqnro=%d", tc.GetDesc(), msg.Xid, msg.SubId, resp.RequestId.Seq)
			return
		}
	} else {
		teststub.TestError(t, "(%s) Not Received msg within %d secs", tc.GetDesc(), 15)
	}
}
