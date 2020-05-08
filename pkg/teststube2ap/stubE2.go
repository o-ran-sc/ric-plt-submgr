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
	"fmt"
	"strconv"
	"testing"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/teststub"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/xapptweaks"
	clientmodel "gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/clientmodel"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
var e2asnpacker e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RmrTransactionId struct {
	xid  string
	meid *xapp.RMRMeid
}

func (trans *RmrTransactionId) String() string {
	return "trans(" + trans.xid + "/" + (&xapptweaks.RMRMeid{trans.meid}).String() + ")"
}

type E2Stub struct {
	teststub.RmrStubControl
	xid_seq           uint64
	subscriptionId    string
	requestCount      int
	RESTNotifications chan int64
	clientEndpoint    string
	meid              string
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CreateNewE2Stub(desc string, srcId teststub.RmrSrcId, rtgSvc teststub.RmrRtgSvc, stat string, mtypeseed int, ranName string, clientEndPoint string) *E2Stub {
	tc := &E2Stub{}
	tc.RmrStubControl.Init(desc, srcId, rtgSvc, stat, mtypeseed)
	tc.xid_seq = 1
	tc.SetCheckXid(true)
	tc.RESTNotifications = make(chan int64)
	tc.clientEndpoint = clientEndPoint // Real endpoint example: service-ricxapp-ueec-http.ricxapp:8080
	tc.meid = ranName
	return tc
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func CreateNewE2termStub(desc string, srcId teststub.RmrSrcId, rtgSvc teststub.RmrRtgSvc, stat string, mtypeseed int) *E2Stub {
	tc := &E2Stub{}
	tc.RmrStubControl.Init(desc, srcId, rtgSvc, stat, mtypeseed)
	tc.xid_seq = 1
	tc.SetCheckXid(false)
	return tc
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
	p.Req.RequestId.InstanceId = 0
	p.Req.FunctionId = 1

	// gnb -> enb outgoing
	// enb -> gnb incoming
	// X2 36423-f40.doc
	p.Req.EventTriggerDefinition.NBX2EventTriggerDefinitionPresent = true
	p.Req.EventTriggerDefinition.NBNRTEventTriggerDefinitionPresent = false
	if p.Req.EventTriggerDefinition.NBX2EventTriggerDefinitionPresent == true {
		p.Req.EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
		p.Req.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Mcc = "310"
		p.Req.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Mnc = "150"
		p.Req.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = 123
		p.Req.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDHomeBits28
		p.Req.EventTriggerDefinition.InterfaceDirection = e2ap.E2AP_InterfaceDirectionIncoming
		p.Req.EventTriggerDefinition.ProcedureCode = 5 //28 35
		p.Req.EventTriggerDefinition.TypeOfMessage = e2ap.E2AP_InitiatingMessage
	} else if p.Req.EventTriggerDefinition.NBNRTEventTriggerDefinitionPresent == true {
		p.Req.EventTriggerDefinition.NBNRTEventTriggerDefinition.TriggerNature = e2ap.NRTTriggerNature_now
	}

	p.Req.ActionSetups = make([]e2ap.ActionToBeSetupItem, 1)

	p.Req.ActionSetups[0].ActionId = 0
	p.Req.ActionSetups[0].ActionType = e2ap.E2AP_ActionTypeReport
	p.Req.ActionSetups[0].RicActionDefinitionPresent = true
	p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionX2Format1Present = false
	p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionX2Format2Present = true
	p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionNRTFormat1Present = false

	if p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionX2Format1Present {
		p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionX2Format1.StyleID = 99
		// 1..255
		for index := 0; index < 1; index++ {
			actionParameterItem := e2ap.ActionParameterItem{}
			actionParameterItem.ParameterID = 11
			actionParameterItem.ActionParameterValue.ValueIntPresent = true
			actionParameterItem.ActionParameterValue.ValueInt = 100
			p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionX2Format1.ActionParameterItems =
				append(p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionX2Format1.ActionParameterItems, actionParameterItem)
		}
	} else if p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionX2Format2Present {
		// 1..15
		for index := 0; index < 2; index++ {
			ranUEgroupItem := e2ap.RANueGroupItem{}
			// 1..255
			for index2 := 0; index2 < 2; index2++ {
				ranUEGroupDefItem := e2ap.RANueGroupDefItem{}
				ranUEGroupDefItem.RanParameterID = 22
				ranUEGroupDefItem.RanParameterTest = e2ap.RANParameterTest_equal
				ranUEGroupDefItem.RanParameterValue.ValueIntPresent = true
				ranUEGroupDefItem.RanParameterValue.ValueInt = 100
				ranUEgroupItem.RanUEgroupDefinition.RanUEGroupDefItems =
					append(ranUEgroupItem.RanUEgroupDefinition.RanUEGroupDefItems, ranUEGroupDefItem)
			}
			// 1..255
			for index3 := 0; index3 < 1; index3++ {
				ranParameterItem := e2ap.RANParameterItem{}
				ranParameterItem.RanParameterID = 33
				ranParameterItem.RanParameterValue.ValueIntPresent = true
				ranParameterItem.RanParameterValue.ValueInt = 100
				ranUEgroupItem.RanPolicy.RanParameterItems =
					append(ranUEgroupItem.RanPolicy.RanParameterItems, ranParameterItem)
			}
			ranUEgroupItem.RanUEgroupID = int64(index)
			p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionX2Format2.RanUEgroupItems =
				append(p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionX2Format2.RanUEgroupItems, ranUEgroupItem)
		}
	} else if p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionNRTFormat1Present {
		// 1..255
		for index := 0; index < 1; index++ {
			ranParameterItem := e2ap.RANParameterItem{}
			ranParameterItem.RanParameterID = 33
			ranParameterItem.RanParameterValue.ValueIntPresent = true
			ranParameterItem.RanParameterValue.ValueInt = 100
			p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionNRTFormat1.RanParameterList =
				append(p.Req.ActionSetups[0].ActionDefinitionChoice.ActionDefinitionNRTFormat1.RanParameterList, ranParameterItem)
		}
	}
	p.Req.ActionSetups[0].SubsequentAction.Present = true
	p.Req.ActionSetups[0].SubsequentAction.Type = e2ap.E2AP_SubSeqActionTypeContinue
	p.Req.ActionSetups[0].SubsequentAction.TimetoWait = e2ap.E2AP_TimeToWaitZero
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
	p.Fail.RequestId.InstanceId = p.Req.RequestId.InstanceId
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
			p.Fail.ActionNotAdmittedList.Items[index].Cause.Value = causeval
		}
		return
	}
	p.Fail.ActionNotAdmittedList.Items[ind].Cause.Content = content
	p.Fail.ActionNotAdmittedList.Items[ind].Cause.Value = causeval
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func (tc *E2Stub) SendSubsReq(t *testing.T, rparams *E2StubSubsReqParams, oldTrans *RmrTransactionId) *RmrTransactionId {

	trans := oldTrans
	if oldTrans == nil {
		trans = tc.NewRmrTransactionId("", "RAN_NAME_1")
	}

	tc.Logger.Info("SendSubsReq %s", trans.String())
	e2SubsReq := e2asnpacker.NewPackerSubscriptionRequest()

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
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = trans.meid
	params.Xid = trans.xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB REQ: %s", params.String())
	snderr := tc.RmrSend(params, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s %s", trans.String(), snderr.Error())
		return nil
	}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) RecvSubsReq(t *testing.T) (*e2ap.E2APSubscriptionRequest, *xapptweaks.RMRParams) {
	tc.Logger.Info("RecvSubsReq")
	e2SubsReq := e2asnpacker.NewPackerSubscriptionRequest()

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

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendSubsResp(t *testing.T, req *e2ap.E2APSubscriptionRequest, msg *xapptweaks.RMRParams) {
	tc.Logger.Info("SendSubsResp")
	e2SubsResp := e2asnpacker.NewPackerSubscriptionResponse()

	//---------------------------------
	// e2term activity: Send Subs Resp
	//---------------------------------
	if req == nil {
		tc.TestError(t, "SendSubsResp: Req == nil")
		return
	}

	resp := &e2ap.E2APSubscriptionResponse{}
	resp.RequestId.Id = req.RequestId.Id
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
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Logger.Debug("%s", e2SubsResp.String())

	params := xapptweaks.NewParams(nil)
	params.Mtype = xapp.RIC_SUB_RESP
	//params.SubId = msg.SubId
	params.SubId = -1
	params.Payload = packedMsg.Buf
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = msg.Meid
	//params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB RESP: %s", params.String())
	snderr := tc.RmrSend(params, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) RecvSubsResp(t *testing.T, trans *RmrTransactionId) uint32 {
	tc.Logger.Info("RecvSubsResp")
	e2SubsResp := e2asnpacker.NewPackerSubscriptionResponse()
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
			tc.Logger.Info("Recv Subs Resp rmr: xid=%s subid=%d, asn: instanceid=%d", msg.Xid, msg.SubId, resp.RequestId.InstanceId)
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

func (tc *E2Stub) SendSubsFail(t *testing.T, fparams *E2StubSubsFailParams, msg *xapptweaks.RMRParams) {
	tc.Logger.Info("SendSubsFail")
	e2SubsFail := e2asnpacker.NewPackerSubscriptionFailure()

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
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB FAIL: %s", params.String())
	snderr := tc.RmrSend(params, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) RecvSubsFail(t *testing.T, trans *RmrTransactionId) uint32 {
	tc.Logger.Info("RecvSubsFail")
	e2SubsFail := e2asnpacker.NewPackerSubscriptionFailure()
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
			tc.Logger.Info("Recv Subs Fail rmr: xid=%s subid=%d, asn: instanceid=%d", msg.Xid, msg.SubId, resp.RequestId.InstanceId)
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
func (tc *E2Stub) SendSubsDelReq(t *testing.T, oldTrans *RmrTransactionId, e2SubsId uint32) *RmrTransactionId {

	trans := oldTrans
	if oldTrans == nil {
		trans = tc.NewRmrTransactionId("", "RAN_NAME_1")
	}

	tc.Logger.Info("SendSubsDelReq %s", trans.String())
	e2SubsDelReq := e2asnpacker.NewPackerSubscriptionDeleteRequest()
	//---------------------------------
	// xapp activity: Send Subs Del Req
	//---------------------------------
	req := &e2ap.E2APSubscriptionDeleteRequest{}
	req.RequestId.Id = 1
	req.RequestId.InstanceId = e2SubsId
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
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = trans.meid
	params.Xid = trans.xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB DEL REQ: %s", params.String())
	snderr := tc.RmrSend(params, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s %s", trans.String(), snderr.Error())
		return nil
	}
	return trans
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) RecvSubsDelReq(t *testing.T) (*e2ap.E2APSubscriptionDeleteRequest, *xapptweaks.RMRParams) {
	tc.Logger.Info("RecvSubsDelReq")
	e2SubsDelReq := e2asnpacker.NewPackerSubscriptionDeleteRequest()

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

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendSubsDelResp(t *testing.T, req *e2ap.E2APSubscriptionDeleteRequest, msg *xapptweaks.RMRParams) {
	tc.Logger.Info("SendSubsDelResp")
	e2SubsDelResp := e2asnpacker.NewPackerSubscriptionDeleteResponse()

	//---------------------------------
	// e2term activity: Send Subs Del Resp
	//---------------------------------
	resp := &e2ap.E2APSubscriptionDeleteResponse{}
	resp.RequestId.Id = req.RequestId.Id
	resp.RequestId.InstanceId = req.RequestId.InstanceId
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
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB DEL RESP: %s", params.String())
	snderr := tc.RmrSend(params, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) RecvSubsDelResp(t *testing.T, trans *RmrTransactionId) {
	tc.Logger.Info("RecvSubsDelResp")
	e2SubsDelResp := e2asnpacker.NewPackerSubscriptionDeleteResponse()

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
			tc.Logger.Info("Recv Subs Del Resp rmr: xid=%s subid=%d, asn: instanceid=%d", msg.Xid, msg.SubId, resp.RequestId.InstanceId)
			return
		}
	} else {
		tc.TestError(t, "Not Received msg within %d secs", 15)
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendSubsDelFail(t *testing.T, req *e2ap.E2APSubscriptionDeleteRequest, msg *xapptweaks.RMRParams) {
	tc.Logger.Info("SendSubsDelFail")
	e2SubsDelFail := e2asnpacker.NewPackerSubscriptionDeleteFailure()

	//---------------------------------
	// e2term activity: Send Subs Del Fail
	//---------------------------------
	resp := &e2ap.E2APSubscriptionDeleteFailure{}
	resp.RequestId.Id = req.RequestId.Id
	resp.RequestId.InstanceId = req.RequestId.InstanceId
	resp.FunctionId = req.FunctionId
	resp.Cause.Content = 4 // CauseMisc
	resp.Cause.Value = 3   // unspecified

	packerr, packedMsg := e2SubsDelFail.Pack(resp)
	if packerr != nil {
		tc.TestError(t, "pack NOK %s", packerr.Error())
	}
	tc.Logger.Debug("%s", e2SubsDelFail.String())

	params := xapptweaks.NewParams(nil)
	params.Mtype = xapp.RIC_SUB_DEL_FAILURE
	params.SubId = msg.SubId
	params.Payload = packedMsg.Buf
	params.PayloadLen = len(packedMsg.Buf)
	params.Meid = msg.Meid
	params.Xid = msg.Xid
	params.Mbuf = nil

	tc.Logger.Info("SEND SUB DEL FAIL: %s", params.String())
	snderr := tc.RmrSend(params, 5)
	if snderr != nil {
		tc.TestError(t, "RMR SEND FAILED: %s", snderr.Error())
	}
}

/*****************************************************************************/
// REST interface specific functions are below

//-----------------------------------------------------------------------------
// Callback handler for subscription response notifications
//-----------------------------------------------------------------------------
func (tc *E2Stub) SubscriptionRespHandler(resp *clientmodel.SubscriptionResponse) {
	if tc.subscriptionId == *resp.SubscriptionID {
		tc.Logger.Info("REST notification received SubscriptionID=%s, InstanceID=%v, RequestorID=%v",
			*resp.SubscriptionID, *resp.SubscriptionInstances[0].InstanceID, *resp.SubscriptionInstances[0].RequestorID)
		tc.RESTNotifications <- *resp.SubscriptionInstances[0].InstanceID
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) WaitRESTNotification(t *testing.T) uint32 {
	tc.Logger.Info("Started waiting REST notification")
	select {
	case instanceId := <-tc.RESTNotifications:
		tc.requestCount--
		if tc.requestCount == 0 {
			tc.Logger.Info("All expected REST notifications received for endpoint=%s", tc.clientEndpoint)
		}
		time.Sleep(time.Millisecond * 200)
		return (uint32)(instanceId) // This delay solves random problems in test cases but not all of them
	case <-time.After(15 * time.Second):
		err := fmt.Errorf("Timeout 15s expired while waiting REST notification")
		tc.TestError(t, "%s", err.Error())
		return 0
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendRESTReportSubsReq(t *testing.T, params *RESTSubsReqReportParams) string {
	tc.Logger.Info("Posting REST Report subscriptions to Submgr")

	xapp.Subscription.SetResponseCB(tc.SubscriptionRespHandler)

	if params == nil {
		tc.Logger.Info("SendRESTReportSubsReq: params == nil")
		return ""
	}

	resp, err := xapp.Subscription.SubscribeReport(&params.ReportParams)
	if err != nil {
		// Swagger generated code makes checks for the values that are inserted the subscription function
		// If error cause is unknown and POST is not done, the problem is in the inserted values
		tc.Logger.Error("REST report subscriptions failed %s", err.Error())
	}
	tc.subscriptionId = *resp.SubscriptionID
	tc.Logger.Info("REST report subscriptions pushed successfully. SubscriptionID = %s, RequestCount = %v", *resp.SubscriptionID, tc.requestCount)
	return *resp.SubscriptionID
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) GetRESTSubsReqReportParams1(subReqCount int, actionDefinitionPresent bool, actionParamCount int) *RESTSubsReqReportParams {

	reportParams := RESTSubsReqReportParams{}
	reportParams.GetRESTSubsReqReportParams1(subReqCount, actionDefinitionPresent, actionParamCount, &tc.clientEndpoint, &tc.meid)
	tc.requestCount = subReqCount
	return &reportParams
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RESTSubsReqReportParams struct {
	ReportParams clientmodel.ReportParams
}

func (p *RESTSubsReqReportParams) GetRESTSubsReqReportParams1(subReqCount int, actionDefinitionPresent bool, actionParamCount int, clientEndpoint *string, meid *string) {

	p.ReportParams.ClientEndpoint = clientEndpoint
	p.ReportParams.Meid = *meid
	var rANFunctionID int64 = 33
	p.ReportParams.RANFunctionID = &rANFunctionID

	for requestCount := 0; requestCount < subReqCount; requestCount++ {
		var procedureCode int64
		procedureCode = int64(requestCount)
		eventTrigger := clientmodel.EventTrigger{}
		eventTrigger.ENBID = "202251"
		eventTrigger.PlmnID = "31050"
		eventTrigger.InterfaceDirection = (int64)(e2ap.E2AP_InterfaceDirectionIncoming)
		eventTrigger.ProcedureCode = procedureCode
		eventTrigger.TypeOfMessage = (int64)(e2ap.E2AP_InitiatingMessage)
		p.ReportParams.EventTriggers = append(p.ReportParams.EventTriggers, &eventTrigger)
	}

	reportActionDefinitions := clientmodel.ReportActionDefinition{}
	p.ReportParams.ReportActionDefinitions = &reportActionDefinitions
	if actionDefinitionPresent == true {
		format1ActionDefinition := clientmodel.Format1ActionDefinition{}
		p.ReportParams.ReportActionDefinitions.ActionDefinitionFormat1 = &format1ActionDefinition
		var styleID int64 = 1
		p.ReportParams.ReportActionDefinitions.ActionDefinitionFormat1.StyleID = &styleID

		for i := 0; i < actionParamCount; i++ {
			var actionParameterID int64
			actionParameterID = (int64)(actionParamCount)
			var actionParameterValue bool = true
			actionParameterItem := clientmodel.ActionParameters{}
			actionParameterItem.ActionParameterID = &actionParameterID
			actionParameterItem.ActionParameterValue = &actionParameterValue
			p.ReportParams.ReportActionDefinitions.ActionDefinitionFormat1.ActionParameters =
				append(p.ReportParams.ReportActionDefinitions.ActionDefinitionFormat1.ActionParameters, &actionParameterItem)
		}
	}
}

func (p *RESTSubsReqReportParams) SetEventTriggerDefinitionProcedureCode(procedureCode int64) {
	for i := 0; i < len(p.ReportParams.EventTriggers); i++ {
		p.ReportParams.EventTriggers[i].ProcedureCode = procedureCode
	}
}

func (p *RESTSubsReqReportParams) SetRANName(ranName string) {
	p.ReportParams.Meid = ranName
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendRESTPolicySubsReq(t *testing.T, params *RESTSubsReqPolicyParams) string {
	tc.Logger.Info("Posting REST policy subscriptions to Submgr")

	xapp.Subscription.SetResponseCB(tc.SubscriptionRespHandler)

	resp, err := xapp.Subscription.SubscribePolicy(&params.PolicyParams)
	if err != nil {
		// Swagger generated code makes checks for the values that are inserted the subscription function
		// If error cause is unknown and POST is not done, the problem is in the inserted values
		tc.Logger.Error("REST policy subscriptions failed %s", err.Error())
	}
	tc.subscriptionId = *resp.SubscriptionID
	tc.Logger.Info("REST policy subscriptions pushed successfully. SubscriptionID = %s, RequestCount = v", *resp.SubscriptionID, tc.requestCount)
	return *resp.SubscriptionID
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) GetRESTSubsReqPolicyParams1(subReqCount int, actionDefinitionPresent bool, policyParamCount int) *RESTSubsReqPolicyParams {

	policyParams := RESTSubsReqPolicyParams{}
	policyParams.GetRESTSubsReqPolicyParams1(subReqCount, actionDefinitionPresent, policyParamCount, &tc.clientEndpoint, &tc.meid)
	tc.requestCount = subReqCount
	return &policyParams
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RESTSubsReqPolicyParams struct {
	PolicyParams clientmodel.PolicyParams
}

func (p *RESTSubsReqPolicyParams) GetRESTSubsReqPolicyParams1(subReqCount int, actionDefinitionPresent bool, policyParamCount int, clientEndpoint *string, meid *string) {

	p.PolicyParams.ClientEndpoint = clientEndpoint
	p.PolicyParams.Meid = meid
	var rANFunctionID int64 = 33
	p.PolicyParams.RANFunctionID = &rANFunctionID

	for requestCount := 0; requestCount < subReqCount; requestCount++ {
		var procedureCode int64
		procedureCode = int64(requestCount)
		eventTrigger := clientmodel.EventTrigger{}
		eventTrigger.ENBID = "202251"
		eventTrigger.PlmnID = "31050"
		eventTrigger.InterfaceDirection = (int64)(e2ap.E2AP_InterfaceDirectionIncoming)
		eventTrigger.ProcedureCode = procedureCode
		eventTrigger.TypeOfMessage = (int64)(e2ap.E2AP_InitiatingMessage)
		p.PolicyParams.EventTriggers = append(p.PolicyParams.EventTriggers, &eventTrigger)
	}

	policyActionDefinitions := clientmodel.PolicyActionDefinition{}
	p.PolicyParams.PolicyActionDefinitions = &policyActionDefinitions
	if actionDefinitionPresent == true {
		format2ActionDefinition := clientmodel.Format2ActionDefinition{}
		p.PolicyParams.PolicyActionDefinitions.ActionDefinitionFormat2 = &format2ActionDefinition

		for i := 0; i < policyParamCount; i++ {
			var policyParameterID int64
			policyParameterID = (int64)(policyParamCount)
			var policyParameterValue int64 = 2
			rANImperativePolicy := clientmodel.ImperativePolicyDefinition{}
			rANImperativePolicy.PolicyParameterID = &policyParameterID
			rANImperativePolicy.PolicyParameterValue = &policyParameterValue

			var rANParameterID int64
			rANParameterID = (int64)(policyParamCount)
			var rANParameterTestCondition = "equal" // equal or greaterthan or lessthan or contains or present
			var rANParameterValue int64 = 3
			rANUeGroupDefinition := clientmodel.RANUeGroupParams{}
			rANUeGroupDefinition.RANParameterID = &rANParameterID
			rANUeGroupDefinition.RANParameterTestCondition = rANParameterTestCondition
			rANUeGroupDefinition.RANParameterValue = &rANParameterValue

			var rANUeGroupID int64
			rANUeGroupID = (int64)(policyParamCount)
			rANUeGroupParametersItem := clientmodel.RANUeGroupList{}
			rANUeGroupParametersItem.RANImperativePolicy = &rANImperativePolicy
			rANUeGroupParametersItem.RANUeGroupDefinition = &rANUeGroupDefinition
			rANUeGroupParametersItem.RANUeGroupID = &rANUeGroupID
			p.PolicyParams.PolicyActionDefinitions.ActionDefinitionFormat2.RANUeGroupParameters =
				append(p.PolicyParams.PolicyActionDefinitions.ActionDefinitionFormat2.RANUeGroupParameters, &rANUeGroupParametersItem)
		}
	}

}

func (p *RESTSubsReqPolicyParams) SetRANParameterTestCondition(testCondition string) {

	for i := 0; i < len(p.PolicyParams.PolicyActionDefinitions.ActionDefinitionFormat2.RANUeGroupParameters); i++ {
		p.PolicyParams.PolicyActionDefinitions.ActionDefinitionFormat2.RANUeGroupParameters[i].RANUeGroupDefinition.RANParameterTestCondition = testCondition
	}
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (tc *E2Stub) SendRESTSubsDelReq(t *testing.T, subscriptionID *string) {

	if *subscriptionID == "" {
		tc.Logger.Error("REST error in deleting subscriptions. Empty SubscriptionID = %s", *subscriptionID)
	}
	tc.Logger.Info("REST deleting E2 subscriptions. SubscriptionID = %s", *subscriptionID)

	err := xapp.Subscription.UnSubscribe(*subscriptionID)
	if err != nil {
		tc.Logger.Error("REST Delete subscription failed %s", err.Error())
	}
	tc.Logger.Info("REST delete subscription pushed to submgr successfully. SubscriptionID = %s", *subscriptionID)
}
