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

package e2ap_tests

import (
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"testing"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionRequestWithData(t *testing.T, areqenc *e2ap.E2APSubscriptionRequest) {

	e2SubsReq := testCtxt.packerif.NewPackerSubscriptionRequest()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")
	seterr := e2SubsReq.Set(areqenc)
	if seterr != nil {
		testCtxt.testError(t, "set err: %s", seterr.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsReq.String())
	testCtxt.testPrint("pack")
	err, packedMsg := e2SubsReq.Pack(nil)
	if err != nil {
		testCtxt.testError(t, "Pack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("unpack")
	err = e2SubsReq.UnPack(packedMsg)
	if err != nil {
		testCtxt.testError(t, "UnPack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsReq.String())
	geterr, areqdec := e2SubsReq.Get()
	if geterr != nil {
		testCtxt.testError(t, "get nil: %s", geterr.Error())
		return
	}
	testCtxt.testValueEquality(t, "msg", areqenc, areqdec)
	testCtxt.testValueEquality(t, "EventTriggerDefinition", &areqenc.EventTriggerDefinition, &areqdec.EventTriggerDefinition)
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionRequest(t *testing.T) {

	areqenc := e2ap.E2APSubscriptionRequest{}
	areqenc.RequestId.Id = 1
	areqenc.RequestId.Seq = 22
	areqenc.FunctionId = 33
	//Bits 20, 28(works), 18, 21 (asn1 problems)
	areqenc.EventTriggerDefinition.InterfaceDirection = e2ap.E2AP_InterfaceDirectionIncoming
	areqenc.EventTriggerDefinition.ProcedureCode = 35
	areqenc.EventTriggerDefinition.TypeOfMessage = e2ap.E2AP_InitiatingMessage
	for index := 0; index < 16; index++ {
		item := e2ap.ActionToBeSetupItem{}
		item.ActionId = uint64(index)
		item.ActionType = e2ap.E2AP_ActionTypeInsert
		// NOT SUPPORTED CURRENTLY
		//item.ActionDefinition.Present = true
		//item.ActionDefinition.StyleId = 255
		//item.ActionDefinition.ParamId = 222
		item.SubsequentAction.Present = true
		item.SubsequentAction.Type = e2ap.E2AP_SubSeqActionTypeContinue
		item.SubsequentAction.TimetoWait = e2ap.E2AP_TimeToWaitW100ms
		areqenc.ActionSetups = append(areqenc.ActionSetups, item)
	}

	areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
	areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.StringPut("310150")
	areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDHomeBits28
	areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = 202251
	testCtxt.SetDesc("SubsReq-28bit")
	testCtxt.E2ApTestMsgSubscriptionRequestWithData(t, &areqenc)

	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.StringPut("310150")
	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDShortMacroits18
	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = 55
	//testCtxt.SetDesc("SubsReq-18bit")
	//testCtxt.E2ApTestMsgSubscriptionRequestWithData(t,&areqenc)

	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.StringPut("310150")
	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDMacroPBits20
	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = 55
	//testCtxt.SetDesc("SubsReq-20bit")
	//testCtxt.E2ApTestMsgSubscriptionRequestWithData(t,&areqenc)

	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.StringPut("310150")
	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDlongMacroBits21
	//areqenc.EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = 55
	//testCtxt.SetDesc("SubsReq-21bit")
	//testCtxt.E2ApTestMsgSubscriptionRequestWithData(t,&areqenc)

}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionResponse(t *testing.T) {

	testCtxt.SetDesc("SubsResp")

	e2SubsResp := testCtxt.packerif.NewPackerSubscriptionResponse()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")

	arespenc := e2ap.E2APSubscriptionResponse{}
	arespenc.RequestId.Id = 1
	arespenc.RequestId.Seq = 22
	arespenc.FunctionId = 33
	for index := uint64(0); index < 16; index++ {
		item := e2ap.ActionAdmittedItem{}
		item.ActionId = index
		arespenc.ActionAdmittedList.Items = append(arespenc.ActionAdmittedList.Items, item)
	}
	for index := uint64(0); index < 16; index++ {
		item := e2ap.ActionNotAdmittedItem{}
		item.ActionId = index
		item.Cause.Content = 1
		item.Cause.CauseVal = 1
		arespenc.ActionNotAdmittedList.Items = append(arespenc.ActionNotAdmittedList.Items, item)
	}

	seterr := e2SubsResp.Set(&arespenc)
	if seterr != nil {
		testCtxt.testError(t, "set err: %s", seterr.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsResp.String())
	testCtxt.testPrint("pack")
	err, packedMsg := e2SubsResp.Pack(nil)
	if err != nil {
		testCtxt.testError(t, "Pack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("unpack")
	err = e2SubsResp.UnPack(packedMsg)
	if err != nil {
		testCtxt.testError(t, "UnPack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsResp.String())
	geterr, arespdec := e2SubsResp.Get()
	if geterr != nil {
		testCtxt.testError(t, "get nil: %s", geterr.Error())
		return
	}
	testCtxt.testValueEquality(t, "msg", &arespenc, arespdec)
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionFailure(t *testing.T) {

	testCtxt.SetDesc("SubsFail")

	e2SubsFail := testCtxt.packerif.NewPackerSubscriptionFailure()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")

	afailenc := e2ap.E2APSubscriptionFailure{}
	afailenc.RequestId.Id = 1
	afailenc.RequestId.Seq = 22
	afailenc.FunctionId = 33
	for index := uint64(0); index < 16; index++ {
		item := e2ap.ActionNotAdmittedItem{}
		item.ActionId = index
		item.Cause.Content = 1
		item.Cause.CauseVal = 1
		afailenc.ActionNotAdmittedList.Items = append(afailenc.ActionNotAdmittedList.Items, item)
	}
	// NOT SUPPORTED CURRENTLY
	afailenc.CriticalityDiagnostics.Present = false
	//	afailenc.CriticalityDiagnostics.ProcCodePresent = true
	//	afailenc.CriticalityDiagnostics.ProcCode = 1
	//	afailenc.CriticalityDiagnostics.TrigMsgPresent = true
	//	afailenc.CriticalityDiagnostics.TrigMsg = 2
	//	afailenc.CriticalityDiagnostics.ProcCritPresent = true
	//	afailenc.CriticalityDiagnostics.ProcCrit = e2ap.E2AP_CriticalityReject
	//	for index := uint32(0); index < 256; index++ {
	//		ieitem := e2ap.CriticalityDiagnosticsIEListItem{}
	//		ieitem.IeCriticality = e2ap.E2AP_CriticalityReject
	//		ieitem.IeID = index
	//		ieitem.TypeOfError = 1
	//		afailenc.CriticalityDiagnostics.CriticalityDiagnosticsIEList.Items = append(afailenc.CriticalityDiagnostics.CriticalityDiagnosticsIEList.Items, ieitem)
	//	}

	seterr := e2SubsFail.Set(&afailenc)
	if seterr != nil {
		testCtxt.testError(t, "set err: %s", seterr.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsFail.String())
	testCtxt.testPrint("pack")
	err, packedMsg := e2SubsFail.Pack(nil)
	if err != nil {
		testCtxt.testError(t, "Pack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("unpack")
	err = e2SubsFail.UnPack(packedMsg)
	if err != nil {
		testCtxt.testError(t, "UnPack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("print:\n%s", e2SubsFail.String())
	geterr, afaildec := e2SubsFail.Get()
	if geterr != nil {
		testCtxt.testError(t, "get nil: %s", geterr.Error())
		return
	}
	testCtxt.testValueEquality(t, "msg", &afailenc, afaildec)
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionRequestBuffers(t *testing.T) {

	testfunc := func(buffer string) {
		packedData := testCtxt.toPackedData(t, buffer)
		if packedData == nil {
			return
		}
		e2SubResp := testCtxt.packerif.NewPackerSubscriptionRequest()
		err := e2SubResp.UnPack(packedData)
		if err != nil {
			testCtxt.testError(t, "UnPack() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		err, _ = e2SubResp.Get()
		if err != nil {
			testCtxt.testError(t, "Get() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		testCtxt.testPrint("OK [%s]", buffer)
	}

	testCtxt.SetDesc("SubReqBuffer")
	testfunc("00c9402c000003ea7e00050000010000ea6300020001ea810016000b00130051407b000000054000ea6b000420000000")
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionResponseBuffers(t *testing.T) {

	testfunc := func(buffer string) {
		packedData := testCtxt.toPackedData(t, buffer)
		if packedData == nil {
			return
		}
		e2SubResp := testCtxt.packerif.NewPackerSubscriptionResponse()
		err := e2SubResp.UnPack(packedData)
		if err != nil {
			testCtxt.testError(t, "UnPack() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		err, _ = e2SubResp.Get()
		if err != nil {
			testCtxt.testError(t, "Get() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		testCtxt.testPrint("OK [%s]", buffer)
	}

	testCtxt.SetDesc("SubRespBuffer")
	testfunc("20c9402a000004ea7e00050000018009ea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106e7ea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106e8ea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106e9ea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106eaea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106ebea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106ecea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106edea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106eeea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106efea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106f0ea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106f4ea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106f5ea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")
	testfunc("20c9402a000004ea7e000500000106f6ea6300020001ea6c000700ea6d00020000ea6e000908ea6f000400000040")

}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionFailureBuffers(t *testing.T) {

	testfunc := func(buffer string) {
		packedData := testCtxt.toPackedData(t, buffer)
		if packedData == nil {
			return
		}
		e2SubResp := testCtxt.packerif.NewPackerSubscriptionFailure()
		err := e2SubResp.UnPack(packedData)
		if err != nil {
			testCtxt.testError(t, "UnPack() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		err, _ = e2SubResp.Get()
		if err != nil {
			testCtxt.testError(t, "Get() Failed: %s [%s]", err.Error(), buffer)
			return
		}
		testCtxt.testPrint("OK [%s]", buffer)
	}

	testCtxt.SetDesc("SubFailBuffer")
	testfunc("40c94017000003ea7e000500000106f3ea6300020001ea6e000100")
}
