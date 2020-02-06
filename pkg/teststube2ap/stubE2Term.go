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
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
var e2t_e2asnpacker e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2termStub struct {
	teststub.RmrStubControl
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CreateNewE2termStub(desc string, rtfile string, port string, stat string, mtypeseed int) *E2termStub {
	e2termCtrl := &E2termStub{}
	e2termCtrl.RmrStubControl.Init(desc, rtfile, port, stat, mtypeseed)
	e2termCtrl.SetCheckXid(false)
	return e2termCtrl
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2termStub) Handle_e2term_subs_req(t *testing.T) (*e2ap.E2APSubscriptionRequest, *xapptweaks.RMRParams) {
	tc.Logger.Info("Handle_e2term_subs_req")
	e2SubsReq := e2t_e2asnpacker.NewPackerSubscriptionRequest()

	//---------------------------------
	// e2term activity: Recv Subs Req
	//---------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_REQ"] {
			tc.TestError(t, "Received wrong mtype expected %s got %s, error", "RIC_SUB_REQ", xapp.RicMessageTypeToName[msg.Mtype])
		} else {
			tc.Logger.Info("Recv Subs Req")
			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr, req := e2SubsReq.UnPack(packedData)
			if unpackerr != nil {
				tc.TestError(t, "RIC_SUB_REQ unpack failed err: %s", unpackerr.Error())
			}
			return req, msg
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}

	return nil, nil
}

func (tc *E2termStub) Handle_e2term_subs_resp(t *testing.T, req *e2ap.E2APSubscriptionRequest, msg *xapptweaks.RMRParams) {
	tc.Logger.Info("Handle_e2term_subs_resp")
	e2SubsResp := e2t_e2asnpacker.NewPackerSubscriptionResponse()

	//---------------------------------
	// e2term activity: Send Subs Resp
	//---------------------------------
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

	packerr, packedMsg := e2SubsResp.Pack(resp)
	if packerr != nil {
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Logger.Debug("%s", e2SubsResp.String())

	params := xapptweaks.NewParams(nil)
	params.Mtype = xapp.RIC_SUB_RESP
	//params.SubId = msg.SubId
	params.SubId = -1
	params.Payload = packedMsg.Buf
	params.Meid = msg.Meid
	//params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB RESP: %s", params.String())
	snderr := tc.RmrSend(params)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2StubSubsFailParams struct {
	Req  *e2ap.E2APSubscriptionRequest
	Fail *e2ap.E2APSubscriptionFailure
}

func (p *E2StubSubsFailParams) Set(req *e2ap.E2APSubscriptionRequest) {
	p.Req = req

	p.Fail = &e2ap.E2APSubscriptionFailure{}
	p.Fail.RequestId.Id = p.Req.RequestId.Id
	p.Fail.RequestId.Seq = p.Req.RequestId.Seq
	p.Fail.FunctionId = p.Req.FunctionId
	p.Fail.ActionNotAdmittedList.Items = make([]e2ap.ActionNotAdmittedItem, len(p.Req.ActionSetups))
	for index := int(0); index < len(p.Fail.ActionNotAdmittedList.Items); index++ {
		p.Fail.ActionNotAdmittedList.Items[index].ActionId = p.Req.ActionSetups[index].ActionId
		p.SetCauseVal(index, 5, 1)
	}
}

func (p *E2StubSubsFailParams) SetCauseVal(ind int, content uint8, causeval uint8) {

	if ind < 0 {
		for index := int(0); index < len(p.Fail.ActionNotAdmittedList.Items); index++ {
			p.Fail.ActionNotAdmittedList.Items[index].Cause.Content = content
			p.Fail.ActionNotAdmittedList.Items[index].Cause.CauseVal = causeval
		}
		return
	}
	p.Fail.ActionNotAdmittedList.Items[ind].Cause.Content = content
	p.Fail.ActionNotAdmittedList.Items[ind].Cause.CauseVal = causeval
}

func (tc *E2termStub) Handle_e2term_subs_fail(t *testing.T, fparams *E2StubSubsFailParams, msg *xapptweaks.RMRParams) {
	tc.Logger.Info("Handle_e2term_subs_fail")
	e2SubsFail := e2t_e2asnpacker.NewPackerSubscriptionFailure()

	//---------------------------------
	// e2term activity: Send Subs Fail
	//---------------------------------
	packerr, packedMsg := e2SubsFail.Pack(fparams.Fail)
	if packerr != nil {
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Logger.Debug("%s", e2SubsFail.String())

	params := xapptweaks.NewParams(nil)
	params.Mtype = xapp.RIC_SUB_FAILURE
	params.SubId = msg.SubId
	params.Payload = packedMsg.Buf
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB FAIL: %s", params.String())
	snderr := tc.RmrSend(params)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2termStub) Handle_e2term_subs_del_req(t *testing.T) (*e2ap.E2APSubscriptionDeleteRequest, *xapptweaks.RMRParams) {
	tc.Logger.Info("Handle_e2term_subs_del_req")
	e2SubsDelReq := e2t_e2asnpacker.NewPackerSubscriptionDeleteRequest()

	//---------------------------------
	// e2term activity: Recv Subs Del Req
	//---------------------------------
	msg := tc.WaitMsg(15)
	if msg != nil {
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_DEL_REQ"] {
			tc.TestError(t, "Received wrong mtype expected %s got %s, error", "RIC_SUB_DEL_REQ", xapp.RicMessageTypeToName[msg.Mtype])
		} else {
			tc.Logger.Info("Recv Subs Del Req")

			packedData := &e2ap.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr, req := e2SubsDelReq.UnPack(packedData)
			if unpackerr != nil {
				tc.TestError(t, "RIC_SUB_DEL_REQ unpack failed err: %s", unpackerr.Error())
			}
			return req, msg
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
	return nil, nil
}

func (tc *E2termStub) Handle_e2term_subs_del_resp(t *testing.T, req *e2ap.E2APSubscriptionDeleteRequest, msg *xapptweaks.RMRParams) {
	tc.Logger.Info("Handle_e2term_subs_del_resp")
	e2SubsDelResp := e2t_e2asnpacker.NewPackerSubscriptionDeleteResponse()

	//---------------------------------
	// e2term activity: Send Subs Del Resp
	//---------------------------------
	resp := &e2ap.E2APSubscriptionDeleteResponse{}
	resp.RequestId.Id = req.RequestId.Id
	resp.RequestId.Seq = req.RequestId.Seq
	resp.FunctionId = req.FunctionId

	packerr, packedMsg := e2SubsDelResp.Pack(resp)
	if packerr != nil {
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Logger.Debug("%s", e2SubsDelResp.String())

	params := xapptweaks.NewParams(nil)
	params.Mtype = xapp.RIC_SUB_DEL_RESP
	params.SubId = msg.SubId
	params.Payload = packedMsg.Buf
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB DEL RESP: %s", params.String())
	snderr := tc.RmrSend(params)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2termStub) Handle_e2term_subs_del_fail(t *testing.T, req *e2ap.E2APSubscriptionDeleteRequest, msg *xapptweaks.RMRParams) {
	tc.Logger.Info("Handle_e2term_del_subs_fail")
	e2SubsDelFail := e2t_e2asnpacker.NewPackerSubscriptionDeleteFailure()

	//---------------------------------
	// e2term activity: Send Subs Del Fail
	//---------------------------------
	resp := &e2ap.E2APSubscriptionDeleteFailure{}
	resp.RequestId.Id = req.RequestId.Id
	resp.RequestId.Seq = req.RequestId.Seq
	resp.FunctionId = req.FunctionId
	resp.Cause.Content = 3  // CauseMisc
	resp.Cause.CauseVal = 4 // unspecified

	packerr, packedMsg := e2SubsDelFail.Pack(resp)
	if packerr != nil {
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Logger.Debug("%s", e2SubsDelFail.String())

	params := xapptweaks.NewParams(nil)
	params.Mtype = xapp.RIC_SUB_DEL_FAILURE
	params.SubId = msg.SubId
	params.Payload = packedMsg.Buf
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB DEL FAIL: %s", params.String())
	snderr := tc.RmrSend(params)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}
