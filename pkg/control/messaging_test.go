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

func createSubsReq() *e2ap.E2APSubscriptionRequest {
	req := &e2ap.E2APSubscriptionRequest{}

	req.RequestId.Id = 1
	req.RequestId.Seq = 22
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

	return req
}

func createSubsResp(req *e2ap.E2APSubscriptionRequest) *e2ap.E2APSubscriptionResponse {

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

	return resp
}

//-----------------------------------------------------------------------------
// TestSubRequestSubResponseOk
//
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
//
//-----------------------------------------------------------------------------
func TestSubRequestSubResponseOk(t *testing.T) {

	xapp.Logger.Info("TestSimple start")
	e2SubsReq := e2asnpacker.NewPackerSubscriptionRequest()
	e2SubsResp := e2asnpacker.NewPackerSubscriptionResponse()

	//---------------------------------
	// xapp activity
	//---------------------------------
	select {
	case <-time.After(5 * time.Second):
		req := createSubsReq()
		e2SubsReq.Set(req)
		xapp.Logger.Debug("%s", e2SubsReq.String())
		err, packedMsg := e2SubsReq.Pack(nil)
		if err != nil {
			testError(t, "(xappConn) pack NOK %s", err.Error())
		} else {
			xapp.Logger.Info("(xappConn) pack OK")
		}

		params := &xapp.RMRParams{}
		params.Mtype = xapp.RIC_SUB_REQ
		params.SubId = -1
		params.Payload = packedMsg.Buf
		params.Meid = &xapp.RMRMeid{RanName: "RAN_NAME_1"}
		params.Xid = "XID_1"
		params.Mbuf = nil

		snderr := xappConn.RmrSend(params)
		if snderr != nil {
			testError(t, "(xappConn) RMR SEND FAILED: %s", snderr.Error())
		}
	}

	//---------------------------------
	// e2term activity
	//---------------------------------
	select {
	case msg := <-e2termConn.rmrConChan:
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_REQ"] {
			testError(t, "(e2termConn) Received non RIC_SUB_REQ message")
		} else {

			packedData := &packer.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr := e2SubsReq.UnPack(packedData)
			if unpackerr != nil {
				testError(t, "(e2termConn) RIC_SUB_REQ unpack failed err: %s", unpackerr.Error())
			}
			geterr, req := e2SubsReq.Get()
			if geterr != nil {
				testError(t, "(e2termConn) RIC_SUB_REQ get failed err: %s", geterr.Error())
			}

			resp := createSubsResp(req)
			e2SubsResp.Set(resp)
			xapp.Logger.Debug("%s", e2SubsResp.String())
			packerr, packedMsg := e2SubsResp.Pack(nil)
			if packerr != nil {
				testError(t, "(e2termConn) pack NOK %s", packerr.Error())
			} else {
				xapp.Logger.Info("(e2termConn) pack OK")
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
				testError(t, "(e2termConn) RMR SEND FAILED: %s", snderr.Error())
			}

		}
	case <-time.After(15 * time.Second):
		testError(t, "(e2termConn) Not Received RIC_SUB_REQ within 15 secs")
	}

	//---------------------------------
	// xapp activity
	//---------------------------------
	select {
	case msg := <-xappConn.rmrConChan:
		if msg.Mtype != xapp.RICMessageTypes["RIC_SUB_RESP"] {
			testError(t, "(xappConn) Received non RIC_SUB_RESP message")
		} else {

			packedData := &packer.PackedData{}
			packedData.Buf = msg.Payload
			unpackerr := e2SubsResp.UnPack(packedData)
			if unpackerr != nil {
				testError(t, "(xappConn) RIC_SUB_RESP unpack failed err: %s", unpackerr.Error())
			}
			geterr, _ := e2SubsResp.Get()
			if geterr != nil {
				testError(t, "(xappConn) RIC_SUB_RESP get failed err: %s", geterr.Error())
			}

		}
	case <-time.After(15 * time.Second):
		testError(t, "(xappConn) Not Received RIC_SUB_RESP within 15 secs")
	}

}
