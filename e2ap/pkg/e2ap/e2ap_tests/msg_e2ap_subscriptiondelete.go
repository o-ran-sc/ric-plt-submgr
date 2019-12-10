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

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionDeleteRequest(t *testing.T) {

	testCtxt.SetDesc("SubsDeleteReq")

	e2SubsReq := testCtxt.packerif.NewPackerSubscriptionDeleteRequest()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")

	areqenc := e2ap.E2APSubscriptionDeleteRequest{}
	areqenc.RequestId.Id = 1
	areqenc.RequestId.Seq = 22
	areqenc.FunctionId = 33

	seterr := e2SubsReq.Set(&areqenc)
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
	testCtxt.testValueEquality(t, "msg", &areqenc, areqdec)
}

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionDeleteResponse(t *testing.T) {

	testCtxt.SetDesc("SubsDeleteResp")

	e2SubsResp := testCtxt.packerif.NewPackerSubscriptionDeleteResponse()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")

	arespenc := e2ap.E2APSubscriptionDeleteResponse{}
	arespenc.RequestId.Id = 1
	arespenc.RequestId.Seq = 22
	arespenc.FunctionId = 33

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

func (testCtxt *E2ApTests) E2ApTestMsgSubscriptionDeleteFailure(t *testing.T) {

	testCtxt.SetDesc("SubsDeleteFail")

	e2SubsFail := testCtxt.packerif.NewPackerSubscriptionDeleteFailure()

	testCtxt.testPrint("########## ##########")
	testCtxt.testPrint("init")

	afailenc := e2ap.E2APSubscriptionDeleteFailure{}
	afailenc.RequestId.Id = 1
	afailenc.RequestId.Seq = 22
	afailenc.FunctionId = 33
	afailenc.Cause.Content = 1
	afailenc.Cause.CauseVal = 1
	// NOT SUPPORTED CURRENTLY
	//	afailenc.CriticalityDiagnostics.Present = false
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
