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

// #cgo LDFLAGS: -le2ap_wrapper -le2ap -lgnbx2 -lgnbnrt -lstdc++
// #include <stdlib.h>
// #include <c_types.h>
// #include <E2AP_if.h>
// #include <memtrack.h>
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
	"strings"
	"unsafe"
)

const cLogBufferMaxSize = 40960
const cMsgBufferMaxSize = 40960
const cMsgBufferExtra = 512

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func cMessageInfoToMessageInfo(minfo *C.E2MessageInfo_t) *e2ap.MessageInfo {

	msgInfo := &e2ap.MessageInfo{}

	switch minfo.messageType {
	case C.cE2InitiatingMessage:
		msgInfo.MsgType = e2ap.E2AP_InitiatingMessage
		switch minfo.messageId {
		case C.cRICSubscriptionRequest:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionRequest
			return msgInfo
		case C.cRICSubscriptionDeleteRequest:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionDeleteRequest
			return msgInfo
		}
	case C.cE2SuccessfulOutcome:
		msgInfo.MsgType = e2ap.E2AP_SuccessfulOutcome
		switch minfo.messageId {
		case C.cRICSubscriptionResponse:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionResponse
			return msgInfo
		case C.cRICsubscriptionDeleteResponse:
			msgInfo.MsgId = e2ap.E2AP_RICSubscriptionDeleteResponse
			return msgInfo
		}
	case C.cE2UnsuccessfulOutcome:
		msgInfo.MsgType = e2ap.E2AP_UnsuccessfulOutcome
		switch minfo.messageId {
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

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryRequestID struct {
	entry *C.RICRequestID_t
}

func (e2Item *e2apEntryRequestID) set(id *e2ap.RequestId) error {
	e2Item.entry.ricRequestorID = (C.uint32_t)(id.Id)
	e2Item.entry.ricInstanceID = (C.uint32_t)(id.InstanceId)
	return nil
}

func (e2Item *e2apEntryRequestID) get(id *e2ap.RequestId) error {
	id.Id = (uint32)(e2Item.entry.ricRequestorID)
	id.InstanceId = (uint32)(e2Item.entry.ricInstanceID)
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryActionToBeSetupItem struct {
	entry *C.RICActionToBeSetupItem_t
}

func (e2Item *e2apEntryActionToBeSetupItem) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.ActionToBeSetupItem) error {

	e2Item.entry.ricActionID = (C.ulong)(id.ActionId)
	e2Item.entry.ricActionType = (C.uint64_t)(id.ActionType)
	if id.RicActionDefinitionPresent {
		e2Item.entry.ricActionDefinitionPresent = true
		if err := (&e2apEntryActionDefinitionChoice{entry: &e2Item.entry.ricActionDefinitionChoice}).set(dynMemHead, &id.ActionDefinitionChoice); err != nil {
			return err
		}
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
		id.RicActionDefinitionPresent = true
		if err := (&e2apEntryActionDefinitionChoice{entry: &e2Item.entry.ricActionDefinitionChoice}).get(&id.ActionDefinitionChoice); err != nil {
			return err
		}
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
type e2apEntryActionDefinitionChoice struct {
	entry *C.RICActionDefinitionChoice_t
}

func (e2Item *e2apEntryActionDefinitionChoice) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.ActionDefinitionChoice) error {

	if id.ActionDefinitionX2Format1Present {
		e2Item.entry.actionDefinitionX2Format1Present = true
		errorNro := C.allocActionDefinitionX2Format1(dynMemHead, &e2Item.entry.actionDefinitionX2Format1)
		if errorNro != C.e2err_OK {
			return fmt.Errorf("e2err(%s)", C.GoString(C.getE2ErrorString(errorNro)))
		}
		if err := (&e2apEntryE2SMgNBX2actionDefinition{entry: e2Item.entry.actionDefinitionX2Format1}).set(dynMemHead, &id.ActionDefinitionX2Format1); err != nil {
			return err
		}

	} else if id.ActionDefinitionX2Format2Present {
		e2Item.entry.actionDefinitionX2Format2Present = true
		errorNro := C.allocActionDefinitionX2Format2(dynMemHead, &e2Item.entry.actionDefinitionX2Format2)
		if errorNro != C.e2err_OK {
			return fmt.Errorf("e2err(%s)", C.GoString(C.getE2ErrorString(errorNro)))
		}
		if err := (&e2apEntryActionDefinitionFormat2{entry: e2Item.entry.actionDefinitionX2Format2}).set(dynMemHead, &id.ActionDefinitionX2Format2); err != nil {
			return err
		}
	} else if id.ActionDefinitionNRTFormat1Present {
		e2Item.entry.actionDefinitionNRTFormat1Present = true
		errorNro := C.allocActionDefinitionNRTFormat1(dynMemHead, &e2Item.entry.actionDefinitionNRTFormat1)
		if errorNro != C.e2err_OK {
			return fmt.Errorf("e2err(%s)", C.GoString(C.getE2ErrorString(errorNro)))
		}
		if err := (&e2apEntryE2SMgNBNRTactionDefinitionFormat1{entry: e2Item.entry.actionDefinitionNRTFormat1}).set(dynMemHead, &id.ActionDefinitionNRTFormat1); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Set() Missing mandatory ActionDefinition element")
	}
	return nil
}

func (e2Item *e2apEntryActionDefinitionChoice) get(id *e2ap.ActionDefinitionChoice) error {
	if e2Item.entry.actionDefinitionX2Format1Present {
		id.ActionDefinitionX2Format1Present = true
		if err := (&e2apEntryE2SMgNBX2actionDefinition{entry: e2Item.entry.actionDefinitionX2Format1}).get(&id.ActionDefinitionX2Format1); err != nil {
			return err
		}
	} else if e2Item.entry.actionDefinitionX2Format2Present {
		id.ActionDefinitionX2Format2Present = true
		if err := (&e2apEntryActionDefinitionFormat2{entry: e2Item.entry.actionDefinitionX2Format2}).get(&id.ActionDefinitionX2Format2); err != nil {
			return err
		}
	} else if e2Item.entry.actionDefinitionNRTFormat1Present {
		id.ActionDefinitionNRTFormat1Present = true
		if err := (&e2apEntryE2SMgNBNRTactionDefinitionFormat1{entry: e2Item.entry.actionDefinitionNRTFormat1}).get(&id.ActionDefinitionNRTFormat1); err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Get() Missing mandatory ActionDefinition element")
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryE2SMgNBX2actionDefinition struct {
	entry *C.E2SMgNBX2actionDefinition_t
}

func (e2Item *e2apEntryE2SMgNBX2actionDefinition) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.E2SMgNBX2actionDefinition) error {

	e2Item.entry.styleID = (C.uint64_t)(id.StyleID)
	// 1..255
	e2Item.entry.actionParameterCount = 0
	for i := 0; i < len(id.ActionParameterItems); i++ {
		if err := (&e2apEntryActionParameterItem{entry: &e2Item.entry.actionParameterItem[i]}).set(dynMemHead, &id.ActionParameterItems[i]); err != nil {
			return err
		}
		e2Item.entry.actionParameterCount++
	}
	return nil
}

func (e2Item *e2apEntryE2SMgNBX2actionDefinition) get(id *e2ap.E2SMgNBX2actionDefinition) error {

	id.StyleID = (uint64)(e2Item.entry.styleID)

	// 1..255
	length := (int)(e2Item.entry.actionParameterCount)
	id.ActionParameterItems = make([]e2ap.ActionParameterItem, length)
	for i := 0; i < length; i++ {
		if err := (&e2apEntryActionParameterItem{entry: &e2Item.entry.actionParameterItem[i]}).get(&id.ActionParameterItems[i]); err != nil {
			return err
		}
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryE2SMgNBNRTactionDefinitionFormat1 struct {
	entry *C.E2SMgNBNRTActionDefinitionFormat1_t
}

func (e2Item *e2apEntryE2SMgNBNRTactionDefinitionFormat1) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.E2SMgNBNRTActionDefinitionFormat1) error {
	// 1..255
	e2Item.entry.ranParameterCount = 0
	for i := 0; i < len(id.RanParameterList); i++ {
		if err := (&e2apEntryRANParameterItem{entry: &e2Item.entry.ranParameterList[i]}).set(dynMemHead, &id.RanParameterList[i]); err != nil {
			return err
		}
		e2Item.entry.ranParameterCount++
	}
	return nil
}

func (e2Item *e2apEntryE2SMgNBNRTactionDefinitionFormat1) get(id *e2ap.E2SMgNBNRTActionDefinitionFormat1) error {
	// 1..255
	length := (int)(e2Item.entry.ranParameterCount)
	id.RanParameterList = make([]e2ap.RANParameterItem, length)
	for i := 0; i < length; i++ {
		if err := (&e2apEntryRANParameterItem{entry: &e2Item.entry.ranParameterList[i]}).get(&id.RanParameterList[i]); err != nil {
			return err
		}
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryActionParameterItem struct {
	entry *C.ActionParameterItem_t
}

func (e2Item *e2apEntryActionParameterItem) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.ActionParameterItem) error {

	e2Item.entry.parameterID = (C.uint32_t)(id.ParameterID)
	if err := (&e2apEntryActionParameterValue{entry: &e2Item.entry.actionParameterValue}).set(dynMemHead, &id.ActionParameterValue); err != nil {
		return err
	}
	return nil
}

func (e2Item *e2apEntryActionParameterItem) get(id *e2ap.ActionParameterItem) error {

	id.ParameterID = (uint32)(e2Item.entry.parameterID)
	if err := (&e2apEntryActionParameterValue{entry: &e2Item.entry.actionParameterValue}).get(&id.ActionParameterValue); err != nil {
		return err
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryActionParameterValue struct {
	entry *C.ActionParameterValue_t
}

func (e2Item *e2apEntryActionParameterValue) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.ActionParameterValue) error {

	if id.ValueIntPresent {
		e2Item.entry.valueInt = (C.int64_t)(id.ValueInt)
		e2Item.entry.valueIntPresent = true
	} else if id.ValueEnumPresent {
		e2Item.entry.valueEnum = (C.int64_t)(id.ValueEnum)
		e2Item.entry.valueEnumPresent = true
	} else if id.ValueBoolPresent {
		e2Item.entry.valueBool = (C.bool)(id.ValueBool)
		e2Item.entry.valueBoolPresent = true
	} else if id.ValueBitSPresent {
		if C.addBitString(dynMemHead, &e2Item.entry.valueBitS, (C.uint64_t)(id.ValueBitS.Length), unsafe.Pointer(&id.ValueBitS.Data[0]), (C.uint8_t)(id.ValueBitS.UnusedBits)) == nil {
			return fmt.Errorf("Alloc valueBitS fail")
		}
		e2Item.entry.valueBitSPresent = true
	} else if id.ValueOctSPresent {
		if C.addOctetString(dynMemHead, &e2Item.entry.valueOctS, (C.uint64_t)(id.ValueOctS.Length), unsafe.Pointer(&id.ValueOctS.Data[0])) == nil {
			return fmt.Errorf("Alloc valueOctS fail")
		}
		e2Item.entry.valueOctSPresent = true
	} else if id.ValuePrtSPresent {
		if C.addOctetString(dynMemHead, &e2Item.entry.valuePrtS, (C.uint64_t)(id.ValuePrtS.Length), unsafe.Pointer(&id.ValuePrtS.Data[0])) == nil {
			return fmt.Errorf("Alloc valuePrtS fail")
		}
		e2Item.entry.valuePrtSPresent = true
	}
	return nil
}

func (e2Item *e2apEntryActionParameterValue) get(id *e2ap.ActionParameterValue) error {

	if e2Item.entry.valueIntPresent {
		id.ValueInt = (int64)(e2Item.entry.valueInt)
		id.ValueIntPresent = true
	} else if e2Item.entry.valueEnumPresent {
		id.ValueEnum = (int64)(e2Item.entry.valueEnum)
		id.ValueEnumPresent = true
	} else if e2Item.entry.valueBoolPresent {
		id.ValueBool = (bool)(e2Item.entry.valueBool)
		id.ValueBoolPresent = true
	} else if e2Item.entry.valueBitSPresent {
		id.ValueBitSPresent = true
		id.ValueBitS.Length = (uint64)(e2Item.entry.valueBitS.byteLength)
		id.ValueBitS.UnusedBits = (uint8)(e2Item.entry.valueBitS.unusedBits)
		C.memcpy(unsafe.Pointer(&id.ValueBitS.Data), unsafe.Pointer(&e2Item.entry.valueBitS.data), C.size_t(e2Item.entry.valueBitS.byteLength))
	} else if e2Item.entry.valueOctSPresent {
		id.ValueOctSPresent = true
		id.ValueOctS.Length = (uint64)(e2Item.entry.valueOctS.length)
		C.memcpy(unsafe.Pointer(&id.ValueBitS.Data), unsafe.Pointer(&e2Item.entry.valueOctS.data), C.size_t(e2Item.entry.valueOctS.length))
	} else if e2Item.entry.valuePrtSPresent {
		id.ValuePrtSPresent = true
		id.ValuePrtS.Length = (uint64)(e2Item.entry.valuePrtS.length)
		C.memcpy(unsafe.Pointer(&id.ValueBitS.Data), unsafe.Pointer(&e2Item.entry.valuePrtS.data), C.size_t(e2Item.entry.valuePrtS.length))
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryActionDefinitionFormat2 struct {
	entry *C.E2SMgNBX2ActionDefinitionFormat2_t
}

func (e2Item *e2apEntryActionDefinitionFormat2) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.ActionDefinitionFormat2) error {
	// 1..15
	e2Item.entry.ranUeGroupCount = 0
	for i := 0; i < len(id.RanUEgroupItems); i++ {
		if err := (&e2apEntryRANueGroupItem{entry: &e2Item.entry.ranUeGroupItem[i]}).set(dynMemHead, &id.RanUEgroupItems[i]); err != nil {
			return err
		}
		e2Item.entry.ranUeGroupCount++
	}
	return nil
}

func (e2Item *e2apEntryActionDefinitionFormat2) get(id *e2ap.ActionDefinitionFormat2) error {
	// 1..15
	length := (int)(e2Item.entry.ranUeGroupCount)
	id.RanUEgroupItems = make([]e2ap.RANueGroupItem, length)
	for i := 0; i < length; i++ {
		if err := (&e2apEntryRANueGroupItem{entry: &e2Item.entry.ranUeGroupItem[i]}).get(&id.RanUEgroupItems[i]); err != nil {
			return err
		}
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryRANueGroupItem struct {
	entry *C.RANueGroupItem_t
}

func (e2Item *e2apEntryRANueGroupItem) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.RANueGroupItem) error {
	e2Item.entry.ranUEgroupID = (C.int64_t)(id.RanUEgroupID)
	if err := (&e2apEntryRANueGroupDefinition{entry: &e2Item.entry.ranUEgroupDefinition}).set(dynMemHead, &id.RanUEgroupDefinition); err != nil {
		return err
	}
	if err := (&e2apEntryRANimperativePolicy{entry: &e2Item.entry.ranPolicy}).set(dynMemHead, &id.RanPolicy); err != nil {
		return err
	}
	return nil
}

func (e2Item *e2apEntryRANueGroupItem) get(id *e2ap.RANueGroupItem) error {
	id.RanUEgroupID = (int64)(e2Item.entry.ranUEgroupID)
	if err := (&e2apEntryRANueGroupDefinition{entry: &e2Item.entry.ranUEgroupDefinition}).get(&id.RanUEgroupDefinition); err != nil {
		return err
	}
	if err := (&e2apEntryRANimperativePolicy{entry: &e2Item.entry.ranPolicy}).get(&id.RanPolicy); err != nil {
		return err
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryRANueGroupDefinition struct {
	entry *C.RANueGroupDefinition_t
}

func (e2Item *e2apEntryRANueGroupDefinition) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.RANueGroupDefinition) error {
	// 1..255
	e2Item.entry.ranUeGroupDefCount = 0
	for i := 0; i < len(id.RanUEGroupDefItems); i++ {
		if err := (&e2apEntryRANueGroupDefItem{entry: &e2Item.entry.ranUeGroupDefItem[i]}).set(dynMemHead, &id.RanUEGroupDefItems[i]); err != nil {
			return err
		}
		e2Item.entry.ranUeGroupDefCount++
	}
	return nil
}

func (e2Item *e2apEntryRANueGroupDefinition) get(id *e2ap.RANueGroupDefinition) error {
	// 1..255
	length := (int)(e2Item.entry.ranUeGroupDefCount)
	id.RanUEGroupDefItems = make([]e2ap.RANueGroupDefItem, length)
	for i := 0; i < length; i++ {
		if err := (&e2apEntryRANueGroupDefItem{entry: &e2Item.entry.ranUeGroupDefItem[i]}).get(&id.RanUEGroupDefItems[i]); err != nil {
			return err
		}
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryRANimperativePolicy struct {
	entry *C.RANimperativePolicy_t
}

func (e2Item *e2apEntryRANimperativePolicy) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.RANimperativePolicy) error {
	// 1..255
	e2Item.entry.ranParameterCount = 0
	for i := 0; i < len(id.RanParameterItems); i++ {
		if err := (&e2apEntryRANParameterItem{entry: &e2Item.entry.ranParameterItem[i]}).set(dynMemHead, &id.RanParameterItems[i]); err != nil {
			return err
		}
		e2Item.entry.ranParameterCount++
	}
	return nil
}

func (e2Item *e2apEntryRANimperativePolicy) get(id *e2ap.RANimperativePolicy) error {
	// 1..255
	length := (int)(e2Item.entry.ranParameterCount)
	id.RanParameterItems = make([]e2ap.RANParameterItem, length)
	for i := 0; i < length; i++ {
		if err := (&e2apEntryRANParameterItem{entry: &e2Item.entry.ranParameterItem[i]}).get(&id.RanParameterItems[i]); err != nil {
			return err
		}
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryRANueGroupDefItem struct {
	entry *C.RANueGroupDefItem_t
}

func (e2Item *e2apEntryRANueGroupDefItem) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.RANueGroupDefItem) error {

	e2Item.entry.ranParameterID = (C.uint32_t)(id.RanParameterID)
	e2Item.entry.ranParameterTest = (C.uint8_t)(id.RanParameterTest)
	if err := (&e2apEntryRANParameterValue{entry: &e2Item.entry.ranParameterValue}).set(dynMemHead, &id.RanParameterValue); err != nil {
		return err
	}
	return nil
}

func (e2Item *e2apEntryRANueGroupDefItem) get(id *e2ap.RANueGroupDefItem) error {

	id.RanParameterID = (uint32)(e2Item.entry.ranParameterID)
	id.RanParameterTest = (uint8)(e2Item.entry.ranParameterTest)
	if err := (&e2apEntryRANParameterValue{entry: &e2Item.entry.ranParameterValue}).get(&id.RanParameterValue); err != nil {
		return err
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryRANParameterItem struct {
	entry *C.RANParameterItem_t
}

func (e2Item *e2apEntryRANParameterItem) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.RANParameterItem) error {

	e2Item.entry.ranParameterID = (C.uint32_t)(id.RanParameterID)
	if err := (&e2apEntryRANParameterValue{entry: &e2Item.entry.ranParameterValue}).set(dynMemHead, &id.RanParameterValue); err != nil {
		return err
	}
	return nil
}

func (e2Item *e2apEntryRANParameterItem) get(id *e2ap.RANParameterItem) error {

	id.RanParameterID = (uint8)(e2Item.entry.ranParameterID)
	if err := (&e2apEntryRANParameterValue{entry: &e2Item.entry.ranParameterValue}).get(&id.RanParameterValue); err != nil {
		return err
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryRANParameterValue struct {
	entry *C.RANParameterValue_t
}

func (e2Item *e2apEntryRANParameterValue) set(dynMemHead *C.mem_track_hdr_t, id *e2ap.RANParameterValue) error {

	if id.ValueIntPresent {
		e2Item.entry.valueInt = (C.int64_t)(id.ValueInt)
		e2Item.entry.valueIntPresent = true
	} else if id.ValueEnumPresent {
		e2Item.entry.valueEnum = (C.int64_t)(id.ValueEnum)
		e2Item.entry.valueEnumPresent = true
	} else if id.ValueBoolPresent {
		e2Item.entry.valueBool = (C.bool)(id.ValueBool)
		e2Item.entry.valueBoolPresent = true
	} else if id.ValueBitSPresent {
		if C.addBitString(dynMemHead, &e2Item.entry.valueBitS, (C.uint64_t)(id.ValueBitS.Length), unsafe.Pointer(&id.ValueBitS.Data[0]), (C.uint8_t)(id.ValueBitS.UnusedBits)) == nil {
			return fmt.Errorf("Alloc valueBitS fail")
		}
		e2Item.entry.valueBitSPresent = true
	} else if id.ValueOctSPresent {
		if C.addOctetString(dynMemHead, &e2Item.entry.valueOctS, (C.uint64_t)(id.ValueOctS.Length), unsafe.Pointer(&id.ValueOctS.Data[0])) == nil {
			return fmt.Errorf("Alloc valueOctS fail")
		}
		e2Item.entry.valueOctSPresent = true
	} else if id.ValuePrtSPresent {
		if C.addOctetString(dynMemHead, &e2Item.entry.valuePrtS, (C.uint64_t)(id.ValuePrtS.Length), unsafe.Pointer(&id.ValuePrtS.Data[0])) == nil {
			return fmt.Errorf("Alloc valuePrtS fail")
		}
		e2Item.entry.valuePrtSPresent = true
	}
	return nil
}

func (e2Item *e2apEntryRANParameterValue) get(id *e2ap.RANParameterValue) error {

	if e2Item.entry.valueIntPresent {
		id.ValueInt = (int64)(e2Item.entry.valueInt)
		id.ValueIntPresent = true
	} else if e2Item.entry.valueEnumPresent {
		id.ValueEnum = (int64)(e2Item.entry.valueEnum)
		id.ValueEnumPresent = true
	} else if e2Item.entry.valueBoolPresent {
		id.ValueBool = (bool)(e2Item.entry.valueBool)
		id.ValueBoolPresent = true
	} else if e2Item.entry.valueBitSPresent {
		id.ValueBitSPresent = true
		id.ValueBitS.Length = (uint64)(e2Item.entry.valueBitS.byteLength)
		id.ValueBitS.UnusedBits = (uint8)(e2Item.entry.valueBitS.unusedBits)
		C.memcpy(unsafe.Pointer(&id.ValueBitS.Data), unsafe.Pointer(&e2Item.entry.valueBitS.data), C.size_t(e2Item.entry.valueBitS.byteLength))
	} else if e2Item.entry.valueOctSPresent {
		id.ValueOctSPresent = true
		id.ValueOctS.Length = (uint64)(e2Item.entry.valueOctS.length)
		C.memcpy(unsafe.Pointer(&id.ValueBitS.Data), unsafe.Pointer(&e2Item.entry.valueOctS.data), C.size_t(e2Item.entry.valueOctS.length))
	} else if e2Item.entry.valuePrtSPresent {
		id.ValuePrtSPresent = true
		id.ValuePrtS.Length = (uint64)(e2Item.entry.valuePrtS.length)
		C.memcpy(unsafe.Pointer(&id.ValueBitS.Data), unsafe.Pointer(&e2Item.entry.valuePrtS.data), C.size_t(e2Item.entry.valuePrtS.length))
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
	if id.NBX2EventTriggerDefinitionPresent {
		evtTrig.entry.E2SMgNBX2EventTriggerDefinitionPresent = true
		return (&e2apEntryX2EventTrigger{entry: &evtTrig.entry.e2SMgNBX2eventTriggerDefinition}).set(&id.X2EventTriggerDefinition)

	} else if id.NBNRTEventTriggerDefinitionPresent {
		evtTrig.entry.E2SMgNBNRTEventTriggerDefinitionPresent = true
		return (&e2apEntryNRTEventTrigger{entry: &evtTrig.entry.e2SMgNBNRTEventTriggerDefinition}).set(&id.NBNRTEventTriggerDefinition)
	}
	return fmt.Errorf("Set() empty EventTriggerDefinition")
}

func (evtTrig *e2apEntryEventTrigger) get(id *e2ap.EventTriggerDefinition) error {
	if evtTrig.entry.E2SMgNBX2EventTriggerDefinitionPresent {
		id.NBX2EventTriggerDefinitionPresent = true
		return (&e2apEntryX2EventTrigger{entry: &evtTrig.entry.e2SMgNBX2eventTriggerDefinition}).get(&id.X2EventTriggerDefinition)

	} else if evtTrig.entry.E2SMgNBNRTEventTriggerDefinitionPresent {
		id.NBNRTEventTriggerDefinitionPresent = true
		return (&e2apEntryNRTEventTrigger{entry: &evtTrig.entry.e2SMgNBNRTEventTriggerDefinition}).get(&id.NBNRTEventTriggerDefinition)
	}
	return fmt.Errorf("Get() empty EventTriggerDefinition")
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryX2EventTrigger struct {
	entry *C.E2SMgNBX2eventTriggerDefinition_t
}

func (evtTrig *e2apEntryX2EventTrigger) set(id *e2ap.X2EventTriggerDefinition) error {
	evtTrig.entry.interfaceDirection = (C.uint8_t)(id.InterfaceDirection)
	evtTrig.entry.interfaceMessageType.procedureCode = (C.uint8_t)(id.ProcedureCode)
	evtTrig.entry.interfaceMessageType.typeOfMessage = (C.uint8_t)(id.TypeOfMessage)
	return (&e2apEntryInterfaceId{entry: &evtTrig.entry.interfaceID}).set(&id.InterfaceId)
}

func (evtTrig *e2apEntryX2EventTrigger) get(id *e2ap.X2EventTriggerDefinition) error {
	id.InterfaceDirection = (uint32)(evtTrig.entry.interfaceDirection)
	id.ProcedureCode = (uint32)(evtTrig.entry.interfaceMessageType.procedureCode)
	id.TypeOfMessage = (uint64)(evtTrig.entry.interfaceMessageType.typeOfMessage)
	return (&e2apEntryInterfaceId{entry: &evtTrig.entry.interfaceID}).get(&id.InterfaceId)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apEntryNRTEventTrigger struct {
	entry *C.E2SMgNBNRTEventTriggerDefinition_t
}

func (evtTrig *e2apEntryNRTEventTrigger) set(id *e2ap.NBNRTEventTriggerDefinition) error {
	evtTrig.entry.eventDefinitionFormat1.triggerNature = (C.uint8_t)(id.TriggerNature)
	return nil
}

func (evtTrig *e2apEntryNRTEventTrigger) get(id *e2ap.NBNRTEventTriggerDefinition) error {
	id.TriggerNature = (uint8)(evtTrig.entry.eventDefinitionFormat1.triggerNature)
	return nil
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
		item.entry.RICActionNotAdmittedItem[item.entry.contentLength].cause.content = (C.uchar)(data.Items[i].Cause.Content)
		item.entry.RICActionNotAdmittedItem[item.entry.contentLength].cause.causeVal = (C.uchar)(data.Items[i].Cause.CauseVal)
		item.entry.contentLength++
	}

	return nil
}

func (item *e2apEntryNotAdmittedList) get(data *e2ap.ActionNotAdmittedList) error {
	conlen := (int)(item.entry.contentLength)
	data.Items = make([]e2ap.ActionNotAdmittedItem, conlen)
	for i := 0; i < conlen; i++ {
		data.Items[i].ActionId = (uint64)(item.entry.RICActionNotAdmittedItem[i].ricActionID)
		data.Items[i].Cause.Content = (uint8)(item.entry.RICActionNotAdmittedItem[i].cause.content)
		data.Items[i].Cause.CauseVal = (uint8)(item.entry.RICActionNotAdmittedItem[i].cause.causeVal)
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

type e2apMessagePacker struct {
	expectedInfo C.E2MessageInfo_t
	pduMsgInfo   C.E2MessageInfo_t
	pdu          *C.e2ap_pdu_ptr_t
	lb           []byte
	p            unsafe.Pointer
	plen         C.size_t
}

func (e2apMsg *e2apMessagePacker) init(minfo C.E2MessageInfo_t) {
	e2apMsg.expectedInfo = minfo
	e2apMsg.lb = make([]byte, cLogBufferMaxSize)
	e2apMsg.lb[0] = 0
	e2apMsg.p = C.malloc(C.size_t(cMsgBufferMaxSize))
	e2apMsg.plen = C.size_t(cMsgBufferMaxSize) - cMsgBufferExtra
}

func (e2apMsg *e2apMessagePacker) fini() {
	C.free(e2apMsg.p)
	e2apMsg.plen = 0
	e2apMsg.p = nil
}

func (e2apMsg *e2apMessagePacker) lbString() string {
	return "logbuffer(" + string(e2apMsg.lb[:strings.Index(string(e2apMsg.lb[:]), "\000")]) + ")"
}

func (e2apMsg *e2apMessagePacker) packeddata() *e2ap.PackedData {
	return &e2ap.PackedData{C.GoBytes(e2apMsg.p, C.int(e2apMsg.plen))}
}

func (e2apMsg *e2apMessagePacker) checkerr(errorNro C.uint64_t) error {
	if errorNro != C.e2err_OK {
		return fmt.Errorf("e2err(%s) %s", C.GoString(C.getE2ErrorString(errorNro)), e2apMsg.lbString())
	}
	return nil
}

func (e2apMsg *e2apMessagePacker) unpacktopdu(data *e2ap.PackedData) error {
	e2apMsg.pdu = C.unpackE2AP_pdu((C.size_t)(len(data.Buf)), (*C.uchar)(unsafe.Pointer(&data.Buf[0])), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), &e2apMsg.pduMsgInfo)
	if e2apMsg.pduMsgInfo.messageType != e2apMsg.expectedInfo.messageType || e2apMsg.pduMsgInfo.messageId != e2apMsg.expectedInfo.messageId {
		return fmt.Errorf("unpack e2ap %s %s", e2apMsg.lbString(), e2apMsg.String())
	}
	return nil
}

func (e2apMsg *e2apMessagePacker) messageInfoPdu() *e2ap.MessageInfo {
	return cMessageInfoToMessageInfo(&e2apMsg.pduMsgInfo)
}

func (e2apMsg *e2apMessagePacker) messageInfoExpected() *e2ap.MessageInfo {
	return cMessageInfoToMessageInfo(&e2apMsg.expectedInfo)
}

func (e2apMsg *e2apMessagePacker) String() string {
	var ret string
	pduInfo := e2apMsg.messageInfoPdu()
	if pduInfo != nil {
		ret += "pduinfo(" + pduInfo.String() + ")"
	} else {
		ret += "pduinfo(N/A)"
	}
	expInfo := e2apMsg.messageInfoExpected()
	if expInfo != nil {
		ret += " expinfo(" + expInfo.String() + ")"
	} else {
		ret += " expinfo(N/A)"
	}
	return ret
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

type e2apMsgPackerSubscriptionRequest struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionRequest_t
	msgG *e2ap.E2APSubscriptionRequest
}

func (e2apMsg *e2apMsgPackerSubscriptionRequest) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2InitiatingMessage, C.cRICSubscriptionRequest})
	e2apMsg.msgC = &C.RICSubscriptionRequest_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionRequest{}
	C.initSubsRequest(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionRequest) Pack(data *e2ap.E2APSubscriptionRequest) (error, *e2ap.PackedData) {

	e2apMsg.init()

	defer e2apMsg.fini()
	e2apMsg.msgG = data

	var dynMemHead C.mem_track_hdr_t
	C.mem_track_init(&dynMemHead)
	defer C.mem_track_free(&dynMemHead)

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}
	if err := (&e2apEntryEventTrigger{entry: &e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition}).set(&e2apMsg.msgG.EventTriggerDefinition); err != nil {
		return err, nil
	}
	if len(e2apMsg.msgG.ActionSetups) > 16 {
		return fmt.Errorf("IndicationMessage.InterfaceMessage: too long %d while allowed %d", len(e2apMsg.msgG.ActionSetups), 16), nil
	}
	e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength = 0
	for i := 0; i < len(e2apMsg.msgG.ActionSetups); i++ {
		item := &e2apEntryActionToBeSetupItem{entry: &e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength]}
		e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength += 1
		if err := item.set(&dynMemHead, &e2apMsg.msgG.ActionSetups[i]); err != nil {
			return err, nil
		}
	}
	errorNro := C.packRICSubscriptionRequest(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionRequest) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionRequest) {

	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}

	var dynMemHead C.mem_track_hdr_t
	C.mem_track_init(&dynMemHead)
	defer C.mem_track_free(&dynMemHead)

	errorNro := C.getRICSubscriptionRequestData(&dynMemHead, e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	if err := (&e2apEntryEventTrigger{entry: &e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition}).get(&e2apMsg.msgG.EventTriggerDefinition); err != nil {
		return err, e2apMsg.msgG
	}
	conlen := (int)(e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength)
	e2apMsg.msgG.ActionSetups = make([]e2ap.ActionToBeSetupItem, conlen)
	for i := 0; i < conlen; i++ {
		item := &e2apEntryActionToBeSetupItem{entry: &e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[i]}
		if err := item.get(&e2apMsg.msgG.ActionSetups[i]); err != nil {
			return err, e2apMsg.msgG
		}
	}
	return nil, e2apMsg.msgG

}

func (e2apMsg *e2apMsgPackerSubscriptionRequest) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionRequest.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "     ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "     ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	fmt.Fprintln(&b, "  ricSubscriptionDetails.")
	fmt.Fprintln(&b, "    ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.")
	fmt.Fprintln(&b, "      contentLength =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.octetString.contentLength)
	fmt.Fprintln(&b, "      interfaceID.globalENBIDPresent =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBIDPresent)
	if e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBIDPresent {
		fmt.Fprintln(&b, "      interfaceID.globalENBID.pLMNIdentity.contentLength =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.contentLength)
		fmt.Fprintln(&b, "      interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[0] =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[0])
		fmt.Fprintln(&b, "      interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[1] =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[1])
		fmt.Fprintln(&b, "      interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[2] =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[2])
		fmt.Fprintln(&b, "      interfaceID.globalENBID.nodeID.bits =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits)
		fmt.Fprintln(&b, "      interfaceID.globalENBID.nodeID.nodeID =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID)
	}
	fmt.Fprintln(&b, "      interfaceID.globalGNBIDPresent =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBIDPresent)
	if e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBIDPresent {
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.pLMNIdentity.contentLength =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.contentLength)
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[0] =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[0])
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[1] =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[1])
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[2] =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal[2])
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.nodeID.bits =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.nodeID.bits)
		fmt.Fprintln(&b, "      interfaceID.globalGNBID.nodeID.nodeID =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.nodeID.nodeID)
	}
	fmt.Fprintln(&b, "      interfaceDirection = ", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceDirection)
	fmt.Fprintln(&b, "      interfaceMessageType.procedureCode =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceMessageType.procedureCode)
	fmt.Fprintln(&b, "      interfaceMessageType.typeOfMessage =", e2apMsg.msgC.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceMessageType.typeOfMessage)
	fmt.Fprintln(&b, "    ricActionToBeSetupItemIEs.")
	fmt.Fprintln(&b, "      contentLength =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength)
	var index uint8
	index = 0
	for (C.uchar)(index) < e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength {
		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionID =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID)
		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionType =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType)

		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionDefinitionPresent =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent)
		if e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent {
			fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionX2Format1Present =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionX2Format1Present)
			fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionX2Format2Present =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionX2Format2Present)
			fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionNRTFormat1Present =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionNRTFormat1Present)
			// Dynamically allocated C-structs are already freed. Can't print those.
		}

		fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricSubsequentActionPresent =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent)
		if e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent {
			fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType)
			fmt.Fprintln(&b, "      ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait =", e2apMsg.msgC.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait)
		}
		index++
	}
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgPackerSubscriptionResponse struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionResponse_t
	msgG *e2ap.E2APSubscriptionResponse
}

func (e2apMsg *e2apMsgPackerSubscriptionResponse) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2SuccessfulOutcome, C.cRICSubscriptionResponse})
	e2apMsg.msgC = &C.RICSubscriptionResponse_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionResponse{}
	C.initSubsResponse(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionResponse) Pack(data *e2ap.E2APSubscriptionResponse) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}
	if err := (&e2apEntryAdmittedList{entry: &e2apMsg.msgC.ricActionAdmittedList}).set(&e2apMsg.msgG.ActionAdmittedList); err != nil {
		return err, nil
	}
	e2apMsg.msgC.ricActionNotAdmittedListPresent = false
	if len(e2apMsg.msgG.ActionNotAdmittedList.Items) > 0 {
		e2apMsg.msgC.ricActionNotAdmittedListPresent = true
		if err := (&e2apEntryNotAdmittedList{entry: &e2apMsg.msgC.ricActionNotAdmittedList}).set(&e2apMsg.msgG.ActionNotAdmittedList); err != nil {
			return err, nil
		}
	}

	errorNro := C.packRICSubscriptionResponse(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionResponse) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionResponse) {

	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}
	errorNro := C.getRICSubscriptionResponseData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	if err := (&e2apEntryAdmittedList{entry: &e2apMsg.msgC.ricActionAdmittedList}).get(&e2apMsg.msgG.ActionAdmittedList); err != nil {
		return err, e2apMsg.msgG
	}
	if e2apMsg.msgC.ricActionNotAdmittedListPresent == true {
		if err := (&e2apEntryNotAdmittedList{entry: &e2apMsg.msgC.ricActionNotAdmittedList}).get(&e2apMsg.msgG.ActionNotAdmittedList); err != nil {
			return err, e2apMsg.msgG
		}
	}
	return nil, e2apMsg.msgG
}

func (e2apMsg *e2apMsgPackerSubscriptionResponse) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionResponse.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
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
			fmt.Fprintln(&b, "      RICActionNotAdmittedItem[index].cause.content =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content)
			fmt.Fprintln(&b, "      RICActionNotAdmittedItem[index].cause.causeVal =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal)
			index++
		}
	}
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgPackerSubscriptionFailure struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionFailure_t
	msgG *e2ap.E2APSubscriptionFailure
}

func (e2apMsg *e2apMsgPackerSubscriptionFailure) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2UnsuccessfulOutcome, C.cRICSubscriptionFailure})
	e2apMsg.msgC = &C.RICSubscriptionFailure_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionFailure{}
	C.initSubsFailure(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionFailure) Pack(data *e2ap.E2APSubscriptionFailure) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}
	if err := (&e2apEntryNotAdmittedList{entry: &e2apMsg.msgC.ricActionNotAdmittedList}).set(&e2apMsg.msgG.ActionNotAdmittedList); err != nil {
		return err, nil
	}
	e2apMsg.msgC.criticalityDiagnosticsPresent = false
	if e2apMsg.msgG.CriticalityDiagnostics.Present {
		e2apMsg.msgC.criticalityDiagnosticsPresent = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).set(&e2apMsg.msgG.CriticalityDiagnostics); err != nil {
			return err, nil
		}
	}

	errorNro := C.packRICSubscriptionFailure(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionFailure) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionFailure) {
	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}
	errorNro := C.getRICSubscriptionFailureData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	if err := (&e2apEntryNotAdmittedList{entry: &e2apMsg.msgC.ricActionNotAdmittedList}).get(&e2apMsg.msgG.ActionNotAdmittedList); err != nil {
		return err, e2apMsg.msgG
	}
	if e2apMsg.msgC.criticalityDiagnosticsPresent == true {
		e2apMsg.msgG.CriticalityDiagnostics.Present = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).get(&e2apMsg.msgG.CriticalityDiagnostics); err != nil {
			return err, e2apMsg.msgG
		}
	}
	return nil, e2apMsg.msgG
}

func (e2apMsg *e2apMsgPackerSubscriptionFailure) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionFailure.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	fmt.Fprintln(&b, "  ricActionNotAdmittedList.")
	fmt.Fprintln(&b, "    contentLength =", e2apMsg.msgC.ricActionNotAdmittedList.contentLength)
	var index uint8
	index = 0
	for (C.uchar)(index) < e2apMsg.msgC.ricActionNotAdmittedList.contentLength {
		fmt.Fprintln(&b, "    RICActionNotAdmittedItem[index].ricActionID =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID)
		fmt.Fprintln(&b, "    RICActionNotAdmittedItem[index].cause.content =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content)
		fmt.Fprintln(&b, "    RICActionNotAdmittedItem[index].cause.causeVal =", e2apMsg.msgC.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal)
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
type e2apMsgPackerSubscriptionDeleteRequest struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionDeleteRequest_t
	msgG *e2ap.E2APSubscriptionDeleteRequest
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequest) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2InitiatingMessage, C.cRICSubscriptionDeleteRequest})
	e2apMsg.msgC = &C.RICSubscriptionDeleteRequest_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionDeleteRequest{}
	C.initSubsDeleteRequest(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequest) Pack(data *e2ap.E2APSubscriptionDeleteRequest) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}

	errorNro := C.packRICSubscriptionDeleteRequest(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequest) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionDeleteRequest) {
	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}
	errorNro := C.getRICSubscriptionDeleteRequestData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	return nil, e2apMsg.msgG

}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteRequest) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionDeleteRequest.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "     ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "     ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgPackerSubscriptionDeleteResponse struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionDeleteResponse_t
	msgG *e2ap.E2APSubscriptionDeleteResponse
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteResponse) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2SuccessfulOutcome, C.cRICsubscriptionDeleteResponse})
	e2apMsg.msgC = &C.RICSubscriptionDeleteResponse_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionDeleteResponse{}
	C.initSubsDeleteResponse(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteResponse) Pack(data *e2ap.E2APSubscriptionDeleteResponse) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}

	errorNro := C.packRICSubscriptionDeleteResponse(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteResponse) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionDeleteResponse) {
	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}
	errorNro := C.getRICSubscriptionDeleteResponseData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	return nil, e2apMsg.msgG
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteResponse) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionDeleteResponse.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
	fmt.Fprintln(&b, "  ranFunctionID =", e2apMsg.msgC.ranFunctionID)
	return b.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type e2apMsgPackerSubscriptionDeleteFailure struct {
	e2apMessagePacker
	msgC *C.RICSubscriptionDeleteFailure_t
	msgG *e2ap.E2APSubscriptionDeleteFailure
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteFailure) init() {
	e2apMsg.e2apMessagePacker.init(C.E2MessageInfo_t{C.cE2UnsuccessfulOutcome, C.cRICsubscriptionDeleteFailure})
	e2apMsg.msgC = &C.RICSubscriptionDeleteFailure_t{}
	e2apMsg.msgG = &e2ap.E2APSubscriptionDeleteFailure{}
	C.initSubsDeleteFailure(e2apMsg.msgC)
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteFailure) Pack(data *e2ap.E2APSubscriptionDeleteFailure) (error, *e2ap.PackedData) {
	e2apMsg.init()
	defer e2apMsg.fini()
	e2apMsg.msgG = data

	e2apMsg.msgC.ranFunctionID = (C.uint16_t)(e2apMsg.msgG.FunctionId)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).set(&e2apMsg.msgG.RequestId); err != nil {
		return err, nil
	}
	e2apMsg.msgC.cause.content = (C.uchar)(e2apMsg.msgG.Cause.Content)
	e2apMsg.msgC.cause.causeVal = (C.uchar)(e2apMsg.msgG.Cause.CauseVal)
	e2apMsg.msgC.criticalityDiagnosticsPresent = false
	if e2apMsg.msgG.CriticalityDiagnostics.Present {
		e2apMsg.msgC.criticalityDiagnosticsPresent = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).set(&e2apMsg.msgG.CriticalityDiagnostics); err != nil {
			return err, nil
		}
	}

	errorNro := C.packRICSubscriptionDeleteFailure(&e2apMsg.plen, (*C.uchar)(e2apMsg.p), (*C.char)(unsafe.Pointer(&e2apMsg.lb[0])), e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, nil
	}
	return nil, e2apMsg.packeddata()
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteFailure) UnPack(msg *e2ap.PackedData) (error, *e2ap.E2APSubscriptionDeleteFailure) {
	e2apMsg.init()
	defer e2apMsg.fini()

	if err := e2apMsg.e2apMessagePacker.unpacktopdu(msg); err != nil {
		return err, e2apMsg.msgG
	}
	errorNro := C.getRICSubscriptionDeleteFailureData(e2apMsg.e2apMessagePacker.pdu, e2apMsg.msgC)
	if err := e2apMsg.checkerr(errorNro); err != nil {
		return err, e2apMsg.msgG
	}

	e2apMsg.msgG.FunctionId = (e2ap.FunctionId)(e2apMsg.msgC.ranFunctionID)
	if err := (&e2apEntryRequestID{entry: &e2apMsg.msgC.ricRequestID}).get(&e2apMsg.msgG.RequestId); err != nil {
		return err, e2apMsg.msgG
	}
	e2apMsg.msgG.Cause.Content = (uint8)(e2apMsg.msgC.cause.content)
	e2apMsg.msgG.Cause.CauseVal = (uint8)(e2apMsg.msgC.cause.causeVal)
	if e2apMsg.msgC.criticalityDiagnosticsPresent == true {
		e2apMsg.msgG.CriticalityDiagnostics.Present = true
		if err := (&e2apEntryCriticalityDiagnostic{entry: &e2apMsg.msgC.criticalityDiagnostics}).get(&e2apMsg.msgG.CriticalityDiagnostics); err != nil {
			return err, e2apMsg.msgG
		}
	}
	return nil, e2apMsg.msgG
}

func (e2apMsg *e2apMsgPackerSubscriptionDeleteFailure) String() string {
	var b bytes.Buffer
	fmt.Fprintln(&b, "ricSubscriptionDeleteFailure.")
	fmt.Fprintln(&b, "  ricRequestID.")
	fmt.Fprintln(&b, "    ricRequestorID =", e2apMsg.msgC.ricRequestID.ricRequestorID)
	fmt.Fprintln(&b, "    ricInstanceID =", e2apMsg.msgC.ricRequestID.ricInstanceID)
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
	return &e2apMsgPackerSubscriptionRequest{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionResponse() e2ap.E2APMsgPackerSubscriptionResponseIf {
	return &e2apMsgPackerSubscriptionResponse{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionFailure() e2ap.E2APMsgPackerSubscriptionFailureIf {
	return &e2apMsgPackerSubscriptionFailure{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionDeleteRequest() e2ap.E2APMsgPackerSubscriptionDeleteRequestIf {
	return &e2apMsgPackerSubscriptionDeleteRequest{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionDeleteResponse() e2ap.E2APMsgPackerSubscriptionDeleteResponseIf {
	return &e2apMsgPackerSubscriptionDeleteResponse{}
}

func (*cppasn1E2APPacker) NewPackerSubscriptionDeleteFailure() e2ap.E2APMsgPackerSubscriptionDeleteFailureIf {
	return &e2apMsgPackerSubscriptionDeleteFailure{}
}

func NewAsn1E2Packer() e2ap.E2APPackerIf {
	return &cppasn1E2APPacker{}
}
