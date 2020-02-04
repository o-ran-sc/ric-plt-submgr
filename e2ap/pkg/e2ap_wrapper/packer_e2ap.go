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

package e2ap_wrapper

// #cgo LDFLAGS: -le2ap_wrapper -le2ap -lstdc++
// #include <stdlib.h>
// #include <c_types.h>
// #include <E2AP_if.h>
// #include <strings.h>
//
// void initSubsRequest(RICSubscriptionRequest_t *data){
//   bzero(data,sizeof(RICSubscriptionRequest_t));
// }
// void initSubsResponse(RICSubscriptionResponse_t *data){
//   bzero(data,sizeof(RICSubscriptionResponse_t));
// }
// void initSubsFailure(RICSubscriptionFailure_t *data){
//   bzero(data,sizeof(RICSubscriptionFailure_t));
// }
// void initSubsDeleteRequest(RICSubscriptionDeleteRequest_t *data){
//   bzero(data,sizeof(RICSubscriptionDeleteRequest_t));
// }
// void initSubsDeleteResponse(RICSubscriptionDeleteResponse_t *data){
//   bzero(data,sizeof(RICSubscriptionDeleteResponse_t));
// }
// void initSubsDeleteFailure(RICSubscriptionDeleteFailure_t *data){
//   bzero(data,sizeof(RICSubscriptionDeleteFailure_t));
// }
//
import "C"

import (
	"bytes"
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/conv"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/packer"
	"unsafe"
)

const cMsgBufferMaxSize = 40960
const cMsgBufferExtra = 512

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryRequestID struct {
	entry *C.RICRequestID_t
}

func (e2Item *e2apEntryRequestID) set(id *e2ap.RequestId) error {
	e2Item.entry.ricRequestorID = (C.uint32_t)(id.Id)
	e2Item.entry.ricRequestSequenceNumber = (C.uint32_t)(id.Seq)
	return nil
}

func (e2Item *e2apEntryRequestID) get(id *e2ap.RequestId) error {
	id.Id = (uint32)(e2Item.entry.ricRequestorID)
	id.Seq = (uint32)(e2Item.entry.ricRequestSequenceNumber)
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryActionToBeSetupItem struct {
	entry *C.RICActionToBeSetupItem_t
}

func (e2Item *e2apEntryActionToBeSetupItem) set(id *e2ap.ActionToBeSetupItem) error {

	e2Item.entry.ricActionID = (C.ulong)(id.ActionId)
	e2Item.entry.ricActionType = (C.uint64_t)(id.ActionType)

	if id.ActionDefinition.Present {
		e2Item.entry.ricActionDefinitionPresent = true
		e2Item.entry.ricActionDefinition.styleID = (C.uint64_t)(id.ActionDefinition.StyleId)
		e2Item.entry.ricActionDefinition.sequenceOfActionParameters.parameterID = (C.uint32_t)(id.ActionDefinition.ParamId)
		//e2Item.entry.ricActionDefinition.sequenceOfActionParameters.ParameterValue = id.ActionDefinition.ParamValue
	}

	if id.SubsequentAction.Present {
		e2Item.entry.ricSubsequentActionPresent = true
		e2Item.entry.ricSubsequentAction.ricSubsequentActionType = (C.uint64_t)(id.SubsequentAction.Type)
		e2Item.entry.ricSubsequentAction.ricTimeToWait = (C.uint64_t)(id.SubsequentAction.TimetoWait)
	}
	return nil
}

func (e2Item *e2apEntryActionToBeSetupItem) get(id *e2ap.ActionToBeSetupItem) error {

	id.ActionId = (uint64)(e2Item.entry.ricActionID)
	id.ActionType = (uint64)(e2Item.entry.ricActionType)

	if e2Item.entry.ricActionDefinitionPresent {
		id.ActionDefinition.Present = true
		id.ActionDefinition.StyleId = (uint64)(e2Item.entry.ricActionDefinition.styleID)
		id.ActionDefinition.ParamId = (uint32)(e2Item.entry.ricActionDefinition.sequenceOfActionParameters.parameterID)
		//id.ActionDefinition.ParamValue=e2Item.entry.ricActionDefinition.sequenceOfActionParameters.ParameterValue
	}

	if e2Item.entry.ricSubsequentActionPresent {
		id.SubsequentAction.Present = true
		id.SubsequentAction.Type = (uint64)(e2Item.entry.ricSubsequentAction.ricSubsequentActionType)
		id.SubsequentAction.TimetoWait = (uint64)(e2Item.entry.ricSubsequentAction.ricTimeToWait)
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryPlmnIdentity struct {
	entry *C.PLMNIdentity_t
}

func (plmnId *e2apEntryPlmnIdentity) set(id *conv.PlmnIdentity) error {

	plmnId.entry.contentLength = (C.uint8_t)(len(id.Val))
	for i := 0; i < len(id.Val); i++ {
		plmnId.entry.pLMNIdentityVal[i] = (C.uint8_t)(id.Val[i])
	}
	return nil
}

func (plmnId *e2apEntryPlmnIdentity) get(id *conv.PlmnIdentity) error {
	conlen := (int)(plmnId.entry.contentLength)
	bcdBuf := make([]uint8, conlen)
	for i := 0; i < conlen; i++ {
		bcdBuf[i] = (uint8)(plmnId.entry.pLMNIdentityVal[i])
	}
	id.BcdPut(bcdBuf)
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryGlobalEnbId struct {
	entry *C.GlobalNodeID_t
}

func (enbId *e2apEntryGlobalEnbId) checkbits(bits uint8) error {
	switch bits {
	case e2ap.E2AP_ENBIDMacroPBits20:
		return nil
	case e2ap.E2AP_ENBIDHomeBits28:
		return nil
	case e2ap.E2AP_ENBIDShortMacroits18:
		return nil
	case e2ap.E2AP_ENBIDlongMacroBits21:
		return nil
	}
	return fmt.Errorf("GlobalEnbId: given bits %d not match allowed: 20,28,18,21", bits)
}

func (enbId *e2apEntryGlobalEnbId) set(id *e2ap.GlobalNodeId) error {
	if err := enbId.checkbits(id.NodeId.Bits); err != nil {
		return err
	}
	enbId.entry.nodeID.bits = (C.uchar)(id.NodeId.Bits)
	enbId.entry.nodeID.nodeID = (C.uint32_t)(id.NodeId.Id)
	return (&e2apEntryPlmnIdentity{entry: &enbId.entry.pLMNIdentity}).set(&id.PlmnIdentity)
}

func (enbId *e2apEntryGlobalEnbId) get(id *e2ap.GlobalNodeId) error {
	if err := enbId.checkbits((uint8)(enbId.entry.nodeID.bits)); err != nil {
		return err
	}
	id.NodeId.Bits = (uint8)(enbId.entry.nodeID.bits)
	id.NodeId.Id = (uint32)(enbId.entry.nodeID.nodeID)
	return (&e2apEntryPlmnIdentity{entry: &enbId.entry.pLMNIdentity}).get(&id.PlmnIdentity)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryGlobalGnbId struct {
	entry *C.GlobalNodeID_t
}

func (gnbId *e2apEntryGlobalGnbId) checkbits(bits uint8) error {
	if bits < 22 || bits > 32 {
		return fmt.Errorf("GlobalGnbId: given bits %d not match allowed: 22-32", bits)
	}
	return nil
}

func (gnbId *e2apEntryGlobalGnbId) set(id *e2ap.GlobalNodeId) error {
	if err := gnbId.checkbits(id.NodeId.Bits); err != nil {
		return err
	}
	gnbId.entry.nodeID.bits = (C.uchar)(id.NodeId.Bits)
	gnbId.entry.nodeID.nodeID = (C.uint32_t)(id.NodeId.Id)
	return (&e2apEntryPlmnIdentity{entry: &gnbId.entry.pLMNIdentity}).set(&id.PlmnIdentity)
}

func (gnbId *e2apEntryGlobalGnbId) get(id *e2ap.GlobalNodeId) error {
	if err := gnbId.checkbits((uint8)(gnbId.entry.nodeID.bits)); err != nil {
		return err
	}
	id.NodeId.Bits = (uint8)(gnbId.entry.nodeID.bits)
	id.NodeId.Id = (uint32)(gnbId.entry.nodeID.nodeID)
	return (&e2apEntryPlmnIdentity{entry: &gnbId.entry.pLMNIdentity}).get(&id.PlmnIdentity)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryInterfaceId struct {
	entry *C.InterfaceID_t
}

func (indId *e2apEntryInterfaceId) set(id *e2ap.InterfaceId) error {
	if id.GlobalEnbId.Present {
		indId.entry.globalENBIDPresent = true
		if err := (&e2apEntryGlobalEnbId{entry: &indId.entry.globalENBID}).set(&id.GlobalEnbId); err != nil {
			return err
		}
	}

	if id.GlobalGnbId.Present {
		indId.entry.globalGNBIDPresent = true
		if err := (&e2apEntryGlobalGnbId{entry: &indId.entry.globalGNBID}).set(&id.GlobalGnbId); err != nil {
			return err
		}
	}
	return nil
}

func (indId *e2apEntryInterfaceId) get(id *e2ap.InterfaceId) error {
	if indId.entry.globalENBIDPresent == true {
		id.GlobalEnbId.Present = true
		if err := (&e2apEntryGlobalEnbId{entry: &indId.entry.globalENBID}).get(&id.GlobalEnbId); err != nil {
			return err
		}
	}

	if indId.entry.globalGNBIDPresent == true {
		id.GlobalGnbId.Present = true
		if err := (&e2apEntryGlobalGnbId{entry: &indId.entry.globalGNBID}).get(&id.GlobalGnbId); err != nil {
			return err
		}
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryEventTrigger struct {
	entry *C.RICEventTriggerDefinition_t
}

func (evtTrig *e2apEntryEventTrigger) set(id *e2ap.EventTriggerDefinition) error {
	evtTrig.entry.interfaceDirection = (C.uint8_t)(id.InterfaceDirection)
	evtTrig.entry.interfaceMessageType.procedureCode = (C.uint8_t)(id.ProcedureCode)
	evtTrig.entry.interfaceMessageType.typeOfMessage = (C.uint8_t)(id.TypeOfMessage)
	return (&e2apEntryInterfaceId{entry: &evtTrig.entry.interfaceID}).set(&id.InterfaceId)
}

func (evtTrig *e2apEntryEventTrigger) get(id *e2ap.EventTriggerDefinition) error {
	id.InterfaceDirection = (uint32)(evtTrig.entry.interfaceDirection)
	id.ProcedureCode = (uint32)(evtTrig.entry.interfaceMessageType.procedureCode)
	id.TypeOfMessage = (uint64)(evtTrig.entry.interfaceMessageType.typeOfMessage)
	return (&e2apEntryInterfaceId{entry: &evtTrig.entry.interfaceID}).get(&id.InterfaceId)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryAdmittedList struct {
	entry *C.RICActionAdmittedList_t
}

func (item *e2apEntryAdmittedList) set(data *e2ap.ActionAdmittedList) error {

	if len(data.Items) > 16 {
		return fmt.Errorf("ActionAdmittedList: too long %d while allowed %d", len(data.Items), 16)
	}

	item.entry.contentLength = 0
	for i := 0; i < len(data.Items); i++ {
		item.entry.ricActionID[item.entry.contentLength] = (C.ulong)(data.Items[i].ActionId)
		item.entry.contentLength++
	}
	return nil
}

func (item *e2apEntryAdmittedList) get(data *e2ap.ActionAdmittedList) error {
	conlen := (int)(item.entry.contentLength)
	data.Items = make([]e2ap.ActionAdmittedItem, conlen)
	for i := 0; i < conlen; i++ {
		data.Items[i].ActionId = (uint64)(item.entry.ricActionID[i])
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryNotAdmittedList struct {
	entry *C.RICActionNotAdmittedList_t
}

func (item *e2apEntryNotAdmittedList) set(data *e2ap.ActionNotAdmittedList) error {

	if len(data.Items) > 16 {
		return fmt.Errorf("e2apEntryNotAdmittedList: too long %d while allowed %d", len(data.Items), 16)
	}

	item.entry.contentLength = 0
	for i := 0; i < len(data.Items); i++ {
		item.entry.RICActionNotAdmittedItem[item.entry.contentLength].ricActionID = (C.ulong)(data.Items[i].ActionId)
		item.entry.RICActionNotAdmittedItem[item.entry.contentLength].ricCause.content = (C.uchar)(data.Items[i].Cause.Content)
		item.entry.RICActionNotAdmittedItem[item.entry.contentLength].ricCause.cause = (C.uchar)(data.Items[i].Cause.CauseVal)
		item.entry.contentLength++
	}

	return nil
}

func (item *e2apEntryNotAdmittedList) get(data *e2ap.ActionNotAdmittedList) error {
	conlen := (int)(item.entry.contentLength)
	data.Items = make([]e2ap.ActionNotAdmittedItem, conlen)
	for i := 0; i < conlen; i++ {
		data.Items[i].ActionId = (uint64)(item.entry.RICActionNotAdmittedItem[i].ricActionID)
		data.Items[i].Cause.Content = (uint8)(item.entry.RICActionNotAdmittedItem[i].ricCause.content)
		data.Items[i].Cause.CauseVal = (uint8)(item.entry.RICActionNotAdmittedItem[i].ricCause.cause)
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryCriticalityDiagnostic struct {
	entry *C.CriticalityDiagnostics__t
}

func (item *e2apEntryCriticalityDiagnostic) set(data *e2ap.CriticalityDiagnostics) error {

	item.entry.procedureCodePresent = (C.bool)(data.ProcCodePresent)
	item.entry.procedureCode = (C.uchar)(data.ProcCode)

	item.entry.triggeringMessagePresent = (C.bool)(data.TrigMsgPresent)
	item.entry.triggeringMessage = (C.uchar)(data.TrigMsg)

	item.entry.procedureCriticalityPresent = (C.bool)(data.ProcCritPresent)
	item.entry.procedureCriticality = (C.uchar)(data.ProcCrit)

	item.entry.criticalityDiagnosticsIELength = 0
	item.entry.iEsCriticalityDiagnosticsPresent = false
	for i := 0; i < len(data.CriticalityDiagnosticsIEList.Items); i++ {
		item.entry.criticalityDiagnosticsIEListItem[i].iECriticality = (C.uint8_t)(data.CriticalityDiagnosticsIEList.Items[i].IeCriticality)
		item.entry.criticalityDiagnosticsIEListItem[i].iE_ID = (C.uint32_t)(data.CriticalityDiagnosticsIEList.Items[i].IeID)
		item.entry.criticalityDiagnosticsIEListItem[i].typeOfError = (C.uint8_t)(data.CriticalityDiagnosticsIEList.Items[i].TypeOfError)
		item.entry.criticalityDiagnosticsIELength++
		item.entry.iEsCriticalityDiagnosticsPresent = true
	}
	return nil
}

func (item *e2apEntryCriticalityDiagnostic) get(data *e2ap.CriticalityDiagnostics) error {

	data.ProcCodePresent = (bool)(item.entry.procedureCodePresent)
	data.ProcCode = (uint64)(item.entry.procedureCode)

	data.TrigMsgPresent = (bool)(item.entry.triggeringMessagePresent)
	data.TrigMsg = (uint64)(item.entry.triggeringMessage)

	data.ProcCritPresent = (bool)(item.entry.procedureCriticalityPresent)
	data.ProcCrit = (uint8)(item.entry.procedureCriticality)

	if item.entry.iEsCriticalityDiagnosticsPresent == true {
		conlen := (int)(item.entry.criticalityDiagnosticsIELength)
		data.CriticalityDiagnosticsIEList.Items = make([]e2ap.CriticalityDiagnosticsIEListItem, conlen)
		for i := 0; i < conlen; i++ {
			data.CriticalityDiagnosticsIEList.Items[i].IeCriticality = (uint8)(item.entry.criticalityDiagnosticsIEListItem[i].iECriticality)
			data.CriticalityDiagnosticsIEList.Items[i].IeID = (uint32)(item.entry.criticalityDiagnosticsIEListItem[i].iE_ID)
			data.CriticalityDiagnosticsIEList.Items[i].TypeOfError = (uint8)(item.entry.criticalityDiagnosticsIEListItem[i].typeOfError)
		}
	}
	return nil
}

/*
//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryCallProcessId struct {
	entry *C.RICCallProcessID_t
}

func (callProcId *e2apEntryCallProcessId) set(data *e2ap.CallProcessId) error {
	callProcId.entry.ricCallProcessIDVal = (C.uint64_t)(data.CallProcessIDVal)
	return nil
}

func (callProcId *e2apEntryCallProcessId) get(data *e2ap.CallProcessId) error {
	data.CallProcessIDVal = (uint32)(callProcId.entry.ricCallProcessIDVal)
	return nil
}
*/

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type e2apMessage struct {
	pdu         *C.e2ap_pdu_ptr_t
	messageInfo C.E2MessageInfo_t
}

func (e2apMsg *e2apMessage) PduUnPack(logBuf []byte, data *packer.PackedData) error {
	e2apMsg.pdu = C.unpackE2AP_pdu((C.size_t)(len(data.Buf)), (*C.uchar)(unsafe.Pointer(&data.Buf[0])), (*C.char)(unsafe.Pointer(&logBuf[0])), &e2apMsg.messageInfo)
	return nil
}

func (e2apMsg *e2apMessage) MessageInfo() *packer.MessageInfo {

	msgInfo := &packer.MessageInfo{}

	switch e2apMsg.messageInfo.messageType {
	case C.cE2InitiatingMessage:
		msgInfo.MsgType = e2ap.E2AP_InitiatingMessage
		switch e2apMsg.messageInfo.messageId {
		case C.cRICSubscriptionRequest:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionRequest
			return msgInfo
		case C.cRICSubscriptionDeleteRequest:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionDeleteRequest
			return msgInfo
		}
	case C.cE2SuccessfulOutcome:
		msgInfo.MsgType = e2ap.E2AP_SuccessfulOutcome
		switch e2apMsg.messageInfo.messageId {
		case C.cRICSubscriptionResponse:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionResponse
			return msgInfo
		case C.cRICsubscriptionDeleteResponse:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionDeleteResponse
			return msgInfo
		}
	case C.cE2UnsuccessfulOutcome:
		msgInfo.MsgType = e2ap.E2AP_UnsuccessfulOutcome
		switch e2apMsg.messageInfo.messageId {
		case C.cRICSubscriptionFailure:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionFailure
			return msgInfo
		case C.cRICsubscriptionDeleteFailure:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionDeleteFailure
			return msgInfo
		}

	}
	return nil
}

func (e2apMsg *e2apMessage) String() string {
	msgInfo := e2apMsg.MessageInfo()
	if msgInfo == nil {
		return "N/A"
	}
	return msgInfo.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type e2apMsgSubscriptionRequest struct {
	e2apMessage
	msgC *C.RICSubscriptionRequest_t
}

func (e2apMsg *e2apMsgSubscriptionRequest) PduPack(logBuf []byte, data *packer.PackedData) error {
	p := C.malloc(C.size_t(cMsgBufferMaxSize))
	defer C.free(p)
	plen := C.size_t(cMsgBufferMaxSize) - cMsgBufferExtra
	errorNro := C.packRICSubscriptionRequest(&plen, (*C.uchar)(p), (*C.char)(unsafe.Pointer(&logBuf[0])), e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	data.Buf = C.GoBytes(p, C.int(plen))
	return nil

}

func (e2apMsg *e2apMsgSubscriptionRequest) PduUnPack(logBuf []byte, data *packer.PackedData) error {
	e2apMsg.msgC = &C.RICSubscriptionRequest_t{}
	C.initSubsRequest(e2apMsg.msgC)
	e2apMsg.e2apMessage.PduUnPack(logBuf, data)
	if e2apMsg.e2apMessage.messageInfo.messageType != C.cE2InitiatingMessage || e2apMsg.e2apMessage.messageInfo.messageId != C.cRICSubscriptionRequest {
		return fmt.Errorf("unpackE2AP_pdu failed -> %s", e2apMsg.e2apMessage.String())
	}
	errorNro := C.getRICSubscriptionRequestData(e2apMsg.e2apMessage.pdu, e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	return nil
}

func (e2apMsg *e2apMsgSubscriptionRequest) Pack(data *e2ap.E2APSubscriptionRequest) (error, *packer.PackedData) {

	e2apMsg.msgC = &C.RICSubscriptionRequest_t{}
	C.initSubsRequest(e2apMsg.msgC)
	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(data.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&data.RequestId); err != nil {
		return err, nil
	}
	if err := (&e2apEntryEventTrigger{entry: &e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition}).set(&data.EventTriggerDefinition); err != nil {
		return err, nil
	}
	if len(data.ActionSetups) > 16 {
		return fmt.Errorf("IndicationMessage.InterfaceMessage: too long %d while allowed %d", len(data.ActionSetups), 16), nil
	}
	e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.contentLength = 0
	for i := 0; i < len(data.ActionSetups); i++ {
		item := &e2apEntryActionToBeSetupItem{entry: &e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.contentLength]}
		e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.contentLength += 1
		if err := item.set(&data.ActionSetups[i]); err != nil {
			return err, nil
		}
	}
	return packer.PduPackerPack(e2apMsg)
}

func (e2apMsg *e2apMsgSubscriptionRequest) UnPack(msg *packer.PackedData) (error, *e2ap.E2APSubscriptionRequest) {
	data := &e2ap.E2APSubscriptionRequest{}
	if err := packer.PduPackerUnPack(e2apMsg, msg); err != nil {
		return err, data
	}
	data.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&data.RequestId); err != nil {
		return err, data
	}
	if err := (&e2apEntryEventTrigger{entry: &e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition}).get(&data.EventTriggerDefinition); err != nil {
		return err, data
	}
	conlen := (int)(e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.contentLength)
	data.ActionSetups = make([]e2ap.ActionToBeSetupItem, conlen)
	for i := 0; i < conlen; i++ {
		item := &e2apEntryActionToBeSetupItem{entry: &e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[i]}
		if err := item.get(&data.ActionSetups[i]); err != nil {
			return err, data
		}
	}
	return nil, data

}

func (e2apMsg *e2apMsgSubscriptionRequest) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionRequest.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "     ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "     ricRequestSequenceNumber =", e2apMsg.msgC.ricRequestID.ricRequestSequenceNumber)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	fmt.Fprintln(&b, "  ricSubscription.")
	fmt.Fprintln(&b, "    ricEventTriggerDefinition.")
	fmt.Fprintln(&b, "      contentLength =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.octetString.contentLength)
	fmt.Fprintln(&b, "      interfaceID.globalENBIDPresent =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBIDPresent)
	if e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBIDPresent {
		fmt.Fprintln(&b, "      interfaceID.globalENBID.pLMNIdentity.contentLength =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.contentLength)
		fmt.Fprintln(&b, "      interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[0] =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[0])
		fmt.Fprintln(&b, "      interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[1] =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[1])
		fmt.Fprintln(&b, "      interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[2] =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[2])
		fmt.Fprintln(&b, "      interfaceID.globalENBID.nodeID.bits =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.nodeID.bits)
		fmt.Fprintln(&b, "      interfaceID.globalENBID.nodeID.nodeID =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID)
	}
	fmt.Fprintln(&b, "      interfaceID.globalGNBIDPresent =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalGNBIDPresent)
	if e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalGNBIDPresent {
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.pLMNIdentity.contentLength =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.contentLength)
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[0] =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[0])
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[1] =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[1])
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[2] =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[2])
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.nodeID.bits =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalGNBID.nodeID.bits)
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.nodeID.nodeID =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceID.globalGNBID.nodeID.nodeID)
	}
	fmt.Fprintln(&b, "      interfaceDirection= ", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceDirection)
	fmt.Fprintln(&b, "      interfaceMessageType.procedureCode =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceMessageType.procedureCode)
	fmt.Fprintln(&b, "      interfaceMessageType.typeOfMessage =", e2apMsg.msgC.ricSubscription.ricEventTriggerDefinition.interfaceMessageType.typeOfMessage)
	fmt.Fprintln(&b, "    ricActionToBeSetupItemIEs.")
	fmt.Fprintln(&b, "      contentLength =", e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.contentLength)
	var index uint8
	index = 0
	for (C.uchar)(index) < e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.contentLength {
		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionID =", e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID)
		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionType =", e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType)

		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionDefinitionPresent =", e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent)
		if e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent {
			fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionDefinition.styleID =", e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinition.styleID)
			fmt.Fprintln(&b, "      ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinition.sequenceOfActionParameters.parameterID =", e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinition.sequenceOfActionParameters.parameterID)
		}

		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricSubsequentActionPresent =", e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent)
		if e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent {
			fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType =", e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType)
			fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait =", e2apMsg.msgC.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait)
		}
		index++
	}
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgSubscriptionResponse struct {
	e2apMessage
	msgC *C.RICSubscriptionResponse_t
}

func (e2apMsg *e2apMsgSubscriptionResponse) PduPack(logBuf []byte, data *packer.PackedData) error {
	p := C.malloc(C.size_t(cMsgBufferMaxSize))
	defer C.free(p)
	plen := C.size_t(cMsgBufferMaxSize) - cMsgBufferExtra
	errorNro := C.packRICSubscriptionResponse(&plen, (*C.uchar)(p), (*C.char)(unsafe.Pointer(&logBuf[0])), e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	data.Buf = C.GoBytes(p, C.int(plen))
	return nil
}

func (e2apMsg *e2apMsgSubscriptionResponse) PduUnPack(logBuf []byte, data *packer.PackedData) error {
	e2apMsg.msgC = &C.RICSubscriptionResponse_t{}
	C.initSubsResponse(e2apMsg.msgC)

	e2apMsg.e2apMessage.PduUnPack(logBuf, data)
	if e2apMsg.e2apMessage.messageInfo.messageType != C.cE2SuccessfulOutcome || e2apMsg.e2apMessage.messageInfo.messageId != C.cRICSubscriptionResponse {
		return fmt.Errorf("unpackE2AP_pdu failed -> %s", e2apMsg.e2apMessage.String())
	}
	errorNro := C.getRICSubscriptionResponseData(e2apMsg.e2apMessage.pdu, e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	return nil
}

func (e2apMsg *e2apMsgSubscriptionResponse) Pack(data *e2ap.E2APSubscriptionResponse) (error, *packer.PackedData) {
	e2apMsg.msgC = &C.RICSubscriptionResponse_t{}
	C.initSubsResponse(e2apMsg.msgC)
	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(data.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&data.RequestId); err != nil {
		return err, nil
	}
	if err := (&e2apEntryAdmittedList{entry: &e2apMsg.msgC.ricActionAdmittedList}).set(&data.ActionAdmittedList); err != nil {
		return err, nil
	}
	e2apMsg.msgC.ricActionNotAdmittedListPresent = false
	if len(data.ActionNotAdmittedList.Items) > 0 {
		e2apMsg.msgC.ricActionNotAdmittedListPresent = true
		if err := (&e2apEntryNotAdmittedList{entry: &e2apMsg.msgC.ricActionNotAdmittedList}).set(&data.ActionNotAdmittedList); err != nil {
			return err, nil
		}
	}
	return packer.PduPackerPack(e2apMsg)
}

func (e2apMsg *e2apMsgSubscriptionResponse) UnPack(msg *packer.PackedData) (error, *e2ap.E2APSubscriptionResponse) {
	data := &e2ap.E2APSubscriptionResponse{}

	if err := packer.PduPackerUnPack(e2apMsg, msg); err != nil {
		return err, data
	}

	data.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&data.RequestId); err != nil {
		return err, data
	}
	if err := (&e2apEntryAdmittedList{entry: &e2apMsg.msgC.ricActionAdmittedList}).get(&data.ActionAdmittedList); err != nil {
		return err, data
	}
	if e2apMsg.msgC.ricActionNotAdmittedListPresent == true {
		if err := (&e2apEntryNotAdmittedList{entry: &e2apMsg.msgC.ricActionNotAdmittedList}).get(&data.ActionNotAdmittedList); err != nil {
			return err, data
		}
	}
	return nil, data

}

func (e2apMsg *e2apMsgSubscriptionResponse) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionResponse.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricRequestSequenceNumber =", e2apMsg.msgC.ricRequestID.ricRequestSequenceNumber)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	fmt.Fprintln(&b, "  ricActionAdmittedList.")
	fmt.Fprintln(&b, "    contentLength =", e2apMsg.msgC.ricActionAdmittedList.contentLength)
	var index uint8
	index = 0
	for (C.uchar)(index) < e2apMsg.msgC.ricActionAdmittedList.contentLength {
		fmt.Fprintln(&b, "    ricActionAdmittedList.ricActionID[index] =", e2apMsg.msgC.ricActionAdmittedList.ricActionID[index])
		index++
	}
	if e2apMsg.msgC.ricActionNotAdmittedListPresent {
		fmt.Fprintln(&b, "  ricActionNotAdmittedListPresent =", e2apMsg.msgC.ricActionNotAdmittedListPresent)
		fmt.Fprintln(&b, "    ricActionNotAdmittedList.")
		fmt.Fprintln(&b, "    contentLength =", e2apMsg.msgC.ricActionNotAdmittedList.contentLength)
		index = 0
		for (C.uchar)(index) < e2apMsg.msgC.ricActionNotAdmittedList.contentLength {
			fmt.Fprintln(&b, "      RICActionNotAdmittedItem[index].ricActionID =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID)
			fmt.Fprintln(&b, "      RICActionNotAdmittedItem[index].ricCause.content =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content)
			fmt.Fprintln(&b, "      RICActionNotAdmittedItem[index].ricCause.cause =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause)
			index++
		}
	}
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgSubscriptionFailure struct {
	e2apMessage
	msgC *C.RICSubscriptionFailure_t
}

func (e2apMsg *e2apMsgSubscriptionFailure) PduPack(logBuf []byte, data *packer.PackedData) error {
	p := C.malloc(C.size_t(cMsgBufferMaxSize))
	defer C.free(p)
	plen := C.size_t(cMsgBufferMaxSize) - cMsgBufferExtra
	errorNro := C.packRICSubscriptionFailure(&plen, (*C.uchar)(p), (*C.char)(unsafe.Pointer(&logBuf[0])), e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	data.Buf = C.GoBytes(p, C.int(plen))
	return nil
}

func (e2apMsg *e2apMsgSubscriptionFailure) PduUnPack(logBuf []byte, data *packer.PackedData) error {
	e2apMsg.msgC = &C.RICSubscriptionFailure_t{}
	C.initSubsFailure(e2apMsg.msgC)
	e2apMsg.e2apMessage.PduUnPack(logBuf, data)
	if e2apMsg.e2apMessage.messageInfo.messageType != C.cE2UnsuccessfulOutcome || e2apMsg.e2apMessage.messageInfo.messageId != C.cRICSubscriptionFailure {
		return fmt.Errorf("unpackE2AP_pdu failed -> %s", e2apMsg.e2apMessage.String())
	}
	errorNro := C.getRICSubscriptionFailureData(e2apMsg.e2apMessage.pdu, e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	return nil

}

func (e2apMsg *e2apMsgSubscriptionFailure) Pack(data *e2ap.E2APSubscriptionFailure) (error, *packer.PackedData) {
	e2apMsg.msgC = &C.RICSubscriptionFailure_t{}
	C.initSubsFailure(e2apMsg.msgC)
	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(data.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&data.RequestId); err != nil {
		return err, nil
	}
	if err := (&e2apEntryNotAdmittedList{entry: &e2apMsg.msgC.ricActionNotAdmittedList}).set(&data.ActionNotAdmittedList); err != nil {
		return err, nil
	}
	e2apMsg.msgC.criticalityDiagnosticsPresent = false
	if data.CriticalityDiagnostics.Present {
		e2apMsg.msgC.criticalityDiagnosticsPresent = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).set(&data.CriticalityDiagnostics); err != nil {
			return err, nil
		}
	}
	return packer.PduPackerPack(e2apMsg)
}

func (e2apMsg *e2apMsgSubscriptionFailure) UnPack(msg *packer.PackedData) (error, *e2ap.E2APSubscriptionFailure) {
	data := &e2ap.E2APSubscriptionFailure{}
	if err := packer.PduPackerUnPack(e2apMsg, msg); err != nil {
		return err, data
	}
	data.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&data.RequestId); err != nil {
		return err, data
	}
	if err := (&e2apEntryNotAdmittedList{entry: &e2apMsg.msgC.ricActionNotAdmittedList}).get(&data.ActionNotAdmittedList); err != nil {
		return err, data
	}
	if e2apMsg.msgC.criticalityDiagnosticsPresent == true {
		data.CriticalityDiagnostics.Present = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).get(&data.CriticalityDiagnostics); err != nil {
			return err, data
		}
	}
	return nil, data
}

func (e2apMsg *e2apMsgSubscriptionFailure) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionFailure.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricRequestSequenceNumber =", e2apMsg.msgC.ricRequestID.ricRequestSequenceNumber)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	fmt.Fprintln(&b, "  ricActionNotAdmittedList.")
	fmt.Fprintln(&b, "    contentLength =", e2apMsg.msgC.ricActionNotAdmittedList.contentLength)
	var index uint8
	index = 0
	for (C.uchar)(index) < e2apMsg.msgC.ricActionNotAdmittedList.contentLength {
		fmt.Fprintln(&b, "    RICActionNotAdmittedItem[index].ricActionID =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID)
		fmt.Fprintln(&b, "    RICActionNotAdmittedItem[index].ricCause.content =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content)
		fmt.Fprintln(&b, "    RICActionNotAdmittedItem[index].ricCause.cause =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause)
		index++
	}
	/* NOT SUPPORTED
	if e2apMsg.msgC.criticalityDiagnosticsPresent {
		fmt.Fprintln(&b, "  criticalityDiagnosticsPresent =", e2apMsg.msgC.criticalityDiagnosticsPresent)
		fmt.Fprintln(&b, "    criticalityDiagnostics.")
		fmt.Fprintln(&b, "    procedureCodePresent =", e2apMsg.msgC.criticalityDiagnostics.procedureCodePresent)
		fmt.Fprintln(&b, "      procedureCode =", e2apMsg.msgC.criticalityDiagnostics.procedureCode)
		fmt.Fprintln(&b, "    triggeringMessagePresent =", e2apMsg.msgC.criticalityDiagnostics.triggeringMessagePresent)
		fmt.Fprintln(&b, "      triggeringMessage =", e2apMsg.msgC.criticalityDiagnostics.triggeringMessage)
		fmt.Fprintln(&b, "    procedureCriticalityPresent=", e2apMsg.msgC.criticalityDiagnostics.procedureCriticalityPresent)
		fmt.Fprintln(&b, "      procedureCriticality =", e2apMsg.msgC.criticalityDiagnostics.procedureCriticality)
		fmt.Fprintln(&b, "    iEsCriticalityDiagnosticsPresent =", e2apMsg.msgC.criticalityDiagnostics.iEsCriticalityDiagnosticsPresent)
		fmt.Fprintln(&b, "      criticalityDiagnosticsIELength =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIELength)
		var index2 uint16
		index2 = 0
		for (C.ushort)(index2) < e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIELength {
			fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].iECriticality =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iECriticality)
			fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].iE_ID =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iE_ID)
			fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].typeOfError =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].typeOfError)
			index2++
		}
	}
	*/
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgSubscriptionDeleteRequest struct {
	e2apMessage
	msgC *C.RICSubscriptionDeleteRequest_t
}

func (e2apMsg *e2apMsgSubscriptionDeleteRequest) PduPack(logBuf []byte, data *packer.PackedData) error {
	p := C.malloc(C.size_t(cMsgBufferMaxSize))
	defer C.free(p)
	plen := C.size_t(cMsgBufferMaxSize) - cMsgBufferExtra
	errorNro := C.packRICSubscriptionDeleteRequest(&plen, (*C.uchar)(p), (*C.char)(unsafe.Pointer(&logBuf[0])), e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	data.Buf = C.GoBytes(p, C.int(plen))
	return nil
}

func (e2apMsg *e2apMsgSubscriptionDeleteRequest) PduUnPack(logBuf []byte, data *packer.PackedData) error {
	e2apMsg.msgC = &C.RICSubscriptionDeleteRequest_t{}
	C.initSubsDeleteRequest(e2apMsg.msgC)
	e2apMsg.e2apMessage.PduUnPack(logBuf, data)
	if e2apMsg.e2apMessage.messageInfo.messageType != C.cE2InitiatingMessage || e2apMsg.e2apMessage.messageInfo.messageId != C.cRICSubscriptionDeleteRequest {
		return fmt.Errorf("unpackE2AP_pdu failed -> %s", e2apMsg.e2apMessage.String())
	}
	errorNro := C.getRICSubscriptionDeleteRequestData(e2apMsg.e2apMessage.pdu, e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	return nil
}

func (e2apMsg *e2apMsgSubscriptionDeleteRequest) Pack(data *e2ap.E2APSubscriptionDeleteRequest) (error, *packer.PackedData) {
	e2apMsg.msgC = &C.RICSubscriptionDeleteRequest_t{}
	C.initSubsDeleteRequest(e2apMsg.msgC)
	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(data.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&data.RequestId); err != nil {
		return err, nil
	}
	return packer.PduPackerPack(e2apMsg)
}

func (e2apMsg *e2apMsgSubscriptionDeleteRequest) Pack21(data *e2ap.E2APSubscriptionDeleteRequest) (error, *packer.PackedData) {
	e2apMsg.msgC = &C.RICSubscriptionDeleteRequest_t{}
	C.initSubsDeleteRequest(e2apMsg.msgC)
	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(data.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&data.RequestId); err != nil {
		return err, nil
	}
	return nil, nil
}

func (e2apMsg *e2apMsgSubscriptionDeleteRequest) Pack22(data *e2ap.E2APSubscriptionDeleteRequest) (error, *packer.PackedData) {
	return packer.PduPackerPack(e2apMsg)
}

func (e2apMsg *e2apMsgSubscriptionDeleteRequest) UnPack(msg *packer.PackedData) (error, *e2ap.E2APSubscriptionDeleteRequest) {
	data := &e2ap.E2APSubscriptionDeleteRequest{}
	if err := packer.PduPackerUnPack(e2apMsg, msg); err != nil {
		return err, data
	}
	data.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&data.RequestId); err != nil {
		return err, data
	}
	return nil, data

}

func (e2apMsg *e2apMsgSubscriptionDeleteRequest) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionDeleteRequest.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "     ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "     ricRequestSequenceNumber =", e2apMsg.msgC.ricRequestID.ricRequestSequenceNumber)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgSubscriptionDeleteResponse struct {
	e2apMessage
	msgC *C.RICSubscriptionDeleteResponse_t
}

func (e2apMsg *e2apMsgSubscriptionDeleteResponse) PduPack(logBuf []byte, data *packer.PackedData) error {
	p := C.malloc(C.size_t(cMsgBufferMaxSize))
	defer C.free(p)
	plen := C.size_t(cMsgBufferMaxSize) - cMsgBufferExtra
	errorNro := C.packRICSubscriptionDeleteResponse(&plen, (*C.uchar)(p), (*C.char)(unsafe.Pointer(&logBuf[0])), e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	data.Buf = C.GoBytes(p, C.int(plen))
	return nil
}

func (e2apMsg *e2apMsgSubscriptionDeleteResponse) PduUnPack(logBuf []byte, data *packer.PackedData) error {
	e2apMsg.msgC = &C.RICSubscriptionDeleteResponse_t{}
	C.initSubsDeleteResponse(e2apMsg.msgC)
	e2apMsg.e2apMessage.PduUnPack(logBuf, data)
	if e2apMsg.e2apMessage.messageInfo.messageType != C.cE2SuccessfulOutcome || e2apMsg.e2apMessage.messageInfo.messageId != C.cRICsubscriptionDeleteResponse {
		return fmt.Errorf("unpackE2AP_pdu failed -> %s", e2apMsg.e2apMessage.String())
	}
	errorNro := C.getRICSubscriptionDeleteResponseData(e2apMsg.e2apMessage.pdu, e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	return nil
}

func (e2apMsg *e2apMsgSubscriptionDeleteResponse) Pack(data *e2ap.E2APSubscriptionDeleteResponse) (error, *packer.PackedData) {
	e2apMsg.msgC = &C.RICSubscriptionDeleteResponse_t{}
	C.initSubsDeleteResponse(e2apMsg.msgC)
	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(data.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&data.RequestId); err != nil {
		return err, nil
	}
	return packer.PduPackerPack(e2apMsg)
}

func (e2apMsg *e2apMsgSubscriptionDeleteResponse) UnPack(msg *packer.PackedData) (error, *e2ap.E2APSubscriptionDeleteResponse) {
	data := &e2ap.E2APSubscriptionDeleteResponse{}
	if err := packer.PduPackerUnPack(e2apMsg, msg); err != nil {
		return err, data
	}
	data.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&data.RequestId); err != nil {
		return err, data
	}
	return nil, data
}

func (e2apMsg *e2apMsgSubscriptionDeleteResponse) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionDeleteResponse.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricRequestSequenceNumber =", e2apMsg.msgC.ricRequestID.ricRequestSequenceNumber)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgSubscriptionDeleteFailure struct {
	e2apMessage
	msgC *C.RICSubscriptionDeleteFailure_t
}

func (e2apMsg *e2apMsgSubscriptionDeleteFailure) PduPack(logBuf []byte, data *packer.PackedData) error {
	p := C.malloc(C.size_t(cMsgBufferMaxSize))
	defer C.free(p)
	plen := C.size_t(cMsgBufferMaxSize) - cMsgBufferExtra
	errorNro := C.packRICSubscriptionDeleteFailure(&plen, (*C.uchar)(p), (*C.char)(unsafe.Pointer(&logBuf[0])), e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	data.Buf = C.GoBytes(p, C.int(plen))
	return nil
}

func (e2apMsg *e2apMsgSubscriptionDeleteFailure) PduUnPack(logBuf []byte, data *packer.PackedData) error {
	e2apMsg.msgC = &C.RICSubscriptionDeleteFailure_t{}
	C.initSubsDeleteFailure(e2apMsg.msgC)
	e2apMsg.e2apMessage.PduUnPack(logBuf, data)
	if e2apMsg.e2apMessage.messageInfo.messageType != C.cE2UnsuccessfulOutcome || e2apMsg.e2apMessage.messageInfo.messageId != C.cRICsubscriptionDeleteFailure {
		return fmt.Errorf("unpackE2AP_pdu failed -> %s", e2apMsg.e2apMessage.String())
	}
	errorNro := C.getRICSubscriptionDeleteFailureData(e2apMsg.e2apMessage.pdu, e2apMsg.msgC)
	if errorNro != C.e2err_OK {
		return fmt.Errorf("%s", C.GoString(C.getE2ErrorString(errorNro)))
	}
	return nil

}

func (e2apMsg *e2apMsgSubscriptionDeleteFailure) Pack(data *e2ap.E2APSubscriptionDeleteFailure) (error, *packer.PackedData) {
	e2apMsg.msgC = &C.RICSubscriptionDeleteFailure_t{}
	C.initSubsDeleteFailure(e2apMsg.msgC)
	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(data.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&data.RequestId); err != nil {
		return err, nil
	}
	e2apMsg.msgC.ricCause.content = (C.uchar)(data.Cause.Content)
	e2apMsg.msgC.ricCause.cause = (C.uchar)(data.Cause.CauseVal)
	e2apMsg.msgC.criticalityDiagnosticsPresent = false
	if data.CriticalityDiagnostics.Present {
		e2apMsg.msgC.criticalityDiagnosticsPresent = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).set(&data.CriticalityDiagnostics); err != nil {
			return err, nil
		}
	}

	return packer.PduPackerPack(e2apMsg)
}

func (e2apMsg *e2apMsgSubscriptionDeleteFailure) UnPack(msg *packer.PackedData) (error, *e2ap.E2APSubscriptionDeleteFailure) {
	data := &e2ap.E2APSubscriptionDeleteFailure{}
	if err := packer.PduPackerUnPack(e2apMsg, msg); err != nil {
		return err, data
	}
	data.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&data.RequestId); err != nil {
		return err, data
	}
	data.Cause.Content = (uint8)(e2apMsg.msgC.ricCause.content)
	data.Cause.CauseVal = (uint8)(e2apMsg.msgC.ricCause.cause)
	if e2apMsg.msgC.criticalityDiagnosticsPresent == true {
		data.CriticalityDiagnostics.Present = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).get(&data.CriticalityDiagnostics); err != nil {
			return err, data
		}
	}
	return nil, data
}

func (e2apMsg *e2apMsgSubscriptionDeleteFailure) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionDeleteFailure.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricRequestSequenceNumber =", e2apMsg.msgC.ricRequestID.ricRequestSequenceNumber)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	/*	NOT SUPPORTED
		if e2apMsg.msgC.criticalityDiagnosticsPresent {
			fmt.Fprintln(&b, "  criticalityDiagnosticsPresent =", e2apMsg.msgC.criticalityDiagnosticsPresent)
			fmt.Fprintln(&b, "    criticalityDiagnostics.")
			fmt.Fprintln(&b, "    procedureCodePresent =", e2apMsg.msgC.criticalityDiagnostics.procedureCodePresent)
			fmt.Fprintln(&b, "      procedureCode =", e2apMsg.msgC.criticalityDiagnostics.procedureCode)
			fmt.Fprintln(&b, "    triggeringMessagePresent =", e2apMsg.msgC.criticalityDiagnostics.triggeringMessagePresent)
			fmt.Fprintln(&b, "      triggeringMessage =", e2apMsg.msgC.criticalityDiagnostics.triggeringMessage)
			fmt.Fprintln(&b, "    procedureCriticalityPresent=", e2apMsg.msgC.criticalityDiagnostics.procedureCriticalityPresent)
			fmt.Fprintln(&b, "      procedureCriticality =", e2apMsg.msgC.criticalityDiagnostics.procedureCriticality)
			fmt.Fprintln(&b, "    iEsCriticalityDiagnosticsPresent =", e2apMsg.msgC.criticalityDiagnostics.iEsCriticalityDiagnosticsPresent)
			fmt.Fprintln(&b, "      criticalityDiagnosticsIELength =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIELength)
			var index2 uint16
			index2 = 0
			for (C.ushort)(index2) < e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIELength {
				fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].iECriticality =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iECriticality)
				fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].iE_ID =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iE_ID)
				fmt.Fprintln(&b, "      criticalityDiagnosticsIEListItem[index2].typeOfError =", e2apMsg.msgC.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].typeOfError)
				index2++
			}
		}
	*/
	return b.String()
}

//-----------------------------------------------------------------------------
// Public E2AP packer creators
//-----------------------------------------------------------------------------

type cppasn1E2APPacker struct{}

func (*cppasn1E2APPacker) NewPackerSubscriptionRequest() e2ap.E2APMsgPackerSubscriptionRequestIf {
	return &e2apMsgSubscriptionRequest{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionResponse() e2ap.E2APMsgPackerSubscriptionResponseIf {
	return &e2apMsgSubscriptionResponse{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionFailure() e2ap.E2APMsgPackerSubscriptionFailureIf {
	return &e2apMsgSubscriptionFailure{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionDeleteRequest() e2ap.E2APMsgPackerSubscriptionDeleteRequestIf {
	return &e2apMsgSubscriptionDeleteRequest{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionDeleteResponse() e2ap.E2APMsgPackerSubscriptionDeleteResponseIf {
	return &e2apMsgSubscriptionDeleteResponse{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionDeleteFailure() e2ap.E2APMsgPackerSubscriptionDeleteFailureIf {
	return &e2apMsgSubscriptionDeleteFailure{}
}

func NewAsn1E2Packer() e2ap.E2APPackerIf {
	return &cppasn1E2APPacker{}
}
