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
func (testCtxt *E2ApTests) E2ApTestMsgIndication(t *testing.T) {

	testCtxt.SetDesc("MsgIndication")
	e2Ind := testCtxt.packerif.NewPackerIndication()

	testCtxt.testPrint("########## Indication ##########")
	testCtxt.testPrint("Indication: init")

	aindenc := e2ap.E2APIndication{}
	aindenc.RequestId.Id = 1
	aindenc.RequestId.Seq = 22
	aindenc.FunctionId = 33
	aindenc.IndicationSn = 1
	aindenc.IndicationType = e2ap.E2AP_IndicationTypeReport
	aindenc.IndicationHeader.InterfaceId.GlobalEnbId.Present = true
	aindenc.IndicationHeader.InterfaceId.GlobalEnbId.PlmnIdentity.StringPut("310150")
	//Bits 20, 28(works), 18, 21 (asn1 problems)
	aindenc.IndicationHeader.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDHomeBits28
	aindenc.IndicationHeader.InterfaceId.GlobalEnbId.NodeId.Id = 202251
	aindenc.IndicationHeader.InterfaceDirection = 0
	aindenc.IndicationMessage.InterfaceMessage.Buf = []uint8{1, 2, 3, 4, 5}
	//aindenc.CallProcessId.CallProcessIDVal=100

	seterr := e2Ind.Set(&aindenc)
	if seterr != nil {
		testCtxt.testError(t, "set err: %s", seterr.Error())
		return
	}

	testCtxt.testPrint("Indication: print:\n%s", e2Ind.String())
	testCtxt.testPrint("Indication: pack")
	err, packedMsg := e2Ind.Pack(nil)
	if err != nil {
		testCtxt.testError(t, "Indication Pack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("Indication: unpack")
	err = e2Ind.UnPack(packedMsg)
	if err != nil {
		testCtxt.testError(t, "Indication UnPack failed: %s", err.Error())
		return
	}
	testCtxt.testPrint("Indication: print:\n%s", e2Ind.String())
	geterr, ainddec := e2Ind.Get()
	if geterr != nil {
		testCtxt.testError(t, "Indication get nil: %s", geterr.Error())
		return
	}
	testCtxt.testValueEquality(t, "msg", &aindenc, ainddec)
}
