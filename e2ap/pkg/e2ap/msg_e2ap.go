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

package e2ap

import (
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/conv"
	"strconv"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_InitiatingMessage   uint64 = 1
	E2AP_SuccessfulOutcome   uint64 = 2
	E2AP_UnsuccessfulOutcome uint64 = 3
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
// E2AP messages
// Initiating message
const (
	E2AP_RICSubscriptionRequest       uint64 = 1
	E2AP_RICSubscriptionDeleteRequest uint64 = 2

	// E2AP_RICServiceUpdate uint64 = 3
	// E2AP_RICControlRequest uint64 = 4
	//
	// //E2AP_X2SetupRequest uint64 = 5;
	// E2AP_ENDCX2SetupRequest uint64 = 6
	// E2AP_ResourceStatusRequest uint64 = 7
	// E2AP_ENBConfigurationUpdate uint64 = 8
	// E2AP_ENDCConfigurationUpdate uint64 = 9
	// E2AP_ResetRequest uint64 = 10
	// E2AP_RICIndication uint64 = 11

	// E2AP_RICServiceQuery uint64 = 12
	// E2AP_LoadInformation uint64 = 13
	// E2AP_GNBStatusIndication uint64 = 14
	// E2AP_ResourceStatusUpdate uint64 = 15
	// E2AP_ErrorIndication uint64 = 16
	//
)

// E2AP messages
// Successful outcome
const (
	E2AP_RICSubscriptionResponse       uint64 = 1
	E2AP_RICSubscriptionDeleteResponse uint64 = 2

	// E2AP_RICserviceUpdateAcknowledge uint64 = 3
	// E2AP_RICcontrolAcknowledge uint64 = 4
	//
	// //E2AP_X2SetupResponse uint64 = 5;
	// E2AP_ENDCX2SetupResponse uint64 = 6
	// E2AP_ResourceStatusResponse uint64 = 7
	// E2AP_ENBConfigurationUpdateAcknowledge uint64 = 8
	// E2AP_ENDCConfigurationUpdateAcknowledge uint64 = 9
	// E2AP_ResetResponse uint64 = 10
	//
)

// E2AP messages
// Unsuccessful outcome
const (
	E2AP_RICSubscriptionFailure       uint64 = 1
	E2AP_RICSubscriptionDeleteFailure uint64 = 2

	// E2AP_RICserviceUpdateFailure uint64 = 3
	// E2AP_RICcontrolFailure uint64 = 4
	//
	// //E2AP_X2SetupFailure uint64 = 5;
	// E2AP_ENDCX2SetupFailure uint64 = 6
	// E2AP_ResourceStatusFailure uint64 = 7
	// E2AP_ENBConfigurationUpdateFailure uint64 = 8
	// E2AP_ENDCConfigurationUpdateFailure uint64 = 9
	//
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_IndicationTypeReport uint64 = 0
	E2AP_IndicationTypeInsert uint64 = 1
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_ActionTypeReport  uint64 = 0
	E2AP_ActionTypeInsert  uint64 = 1
	E2AP_ActionTypePolicy  uint64 = 2
	E2AP_ActionTypeInvalid uint64 = 99 // For RIC internal usage only
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_SubSeqActionTypeContinue uint64 = 0
	E2AP_SubSeqActionTypeWait     uint64 = 1
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_TimeToWaitZero   uint64 = 0
	E2AP_TimeToWaitW1ms   uint64 = 1
	E2AP_TimeToWaitW2ms   uint64 = 2
	E2AP_TimeToWaitW5ms   uint64 = 3
	E2AP_TimeToWaitW10ms  uint64 = 4
	E2AP_TimeToWaitW20ms  uint64 = 4
	E2AP_TimeToWaitW30ms  uint64 = 5
	E2AP_TimeToWaitW40ms  uint64 = 6
	E2AP_TimeToWaitW50ms  uint64 = 7
	E2AP_TimeToWaitW100ms uint64 = 8
	E2AP_TimeToWaitW200ms uint64 = 9
	E2AP_TimeToWaitW500ms uint64 = 10
	E2AP_TimeToWaitW1s    uint64 = 11
	E2AP_TimeToWaitW2s    uint64 = 12
	E2AP_TimeToWaitW5s    uint64 = 13
	E2AP_TimeToWaitW10s   uint64 = 14
	E2AP_TimeToWaitW20s   uint64 = 15
	E2AP_TimeToWaitW60    uint64 = 16
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_InterfaceDirectionIncoming uint32 = 0
	E2AP_InterfaceDirectionOutgoing uint32 = 1
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_CriticalityReject uint8 = 0
	E2AP_CriticalityIgnore uint8 = 1
	E2AP_CriticalityNotify uint8 = 2
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const (
	E2AP_ENBIDMacroPBits20    uint8 = 20
	E2AP_ENBIDHomeBits28      uint8 = 28
	E2AP_ENBIDShortMacroits18 uint8 = 18
	E2AP_ENBIDlongMacroBits21 uint8 = 21
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type PackedData struct {
	Buf []byte
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type MessageInfo struct {
	MsgType uint64
	MsgId   uint64
}

func (msgInfo *MessageInfo) String() string {
	return "msginfo(" + strconv.FormatUint((uint64)(msgInfo.MsgType), 10) + string(":") + strconv.FormatUint((uint64)(msgInfo.MsgId), 10) + ")"
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RequestId struct {
	Id         uint32
	InstanceId uint32
}

func (rid *RequestId) String() string {
	return strconv.FormatUint((uint64)(rid.Id), 10) + string(":") + strconv.FormatUint((uint64)(rid.InstanceId), 10)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type FunctionId uint16

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type NodeId struct {
	Bits uint8
	Id   uint32
}

func (nid *NodeId) String() string {
	return strconv.FormatUint((uint64)(nid.Id), 10)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type GlobalNodeId struct {
	Present      bool
	PlmnIdentity conv.PlmnIdentity
	NodeId       NodeId
}

func (gnid *GlobalNodeId) String() string {
	return gnid.PlmnIdentity.String() + string(":") + gnid.NodeId.String()
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type InterfaceId struct {
	GlobalEnbId GlobalNodeId
	GlobalGnbId GlobalNodeId
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type EventTriggerDefinition struct {
	NBX2EventTriggerDefinitionPresent bool
	X2EventTriggerDefinition
	NBNRTEventTriggerDefinitionPresent bool
	NBNRTEventTriggerDefinition
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type X2EventTriggerDefinition struct {
	InterfaceId
	InterfaceDirection uint32
	ProcedureCode      uint32
	TypeOfMessage      uint64
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type NBNRTEventTriggerDefinition struct {
	TriggerNature uint8
}

const ( // enum NRTTriggerNature
	NRTTriggerNature_now = iota
	NRTTriggerNature_onchange
)

/*
//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type CallProcessId struct {
  CallProcessIDVal uint32
}
*/

type ActionDefinitionChoice struct {
	ActionDefinitionX2Format1Present  bool
	ActionDefinitionX2Format1         E2SMgNBX2actionDefinition
	ActionDefinitionX2Format2Present  bool
	ActionDefinitionX2Format2         ActionDefinitionFormat2
	ActionDefinitionNRTFormat1Present bool
	ActionDefinitionNRTFormat1        E2SMgNBNRTActionDefinitionFormat1
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2SMgNBNRTActionDefinitionFormat1 struct {
	RanParameterList []RANParameterItem // 1..255
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2SMgNBX2actionDefinition struct {
	StyleID              uint64
	ActionParameterItems []ActionParameterItem // 1..255
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionParameterItem struct {
	ParameterID          uint32 // 1..255
	ActionParameterValue ActionParameterValue
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionParameterValue struct {
	ValueIntPresent  bool
	ValueInt         int64
	ValueEnumPresent bool
	ValueEnum        int64
	ValueBoolPresent bool
	ValueBool        bool
	ValueBitSPresent bool
	ValueBitS        BitString
	ValueOctSPresent bool
	ValueOctS        OctetString
	ValuePrtSPresent bool
	ValuePrtS        OctetString
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionDefinitionFormat2 struct {
	RanUEgroupItems []RANueGroupItem // 1..15
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RANueGroupItem struct {
	RanUEgroupID         int64
	RanUEgroupDefinition RANueGroupDefinition
	RanPolicy            RANimperativePolicy
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RANueGroupDefinition struct {
	RanUEGroupDefItems []RANueGroupDefItem // 1..255
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RANimperativePolicy struct {
	RanParameterItems []RANParameterItem // 1..255
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RANueGroupDefItem struct {
	RanParameterID    uint32 // 1..255
	RanParameterTest  uint8
	RanParameterValue RANParameterValue
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RANParameterItem struct {
	RanParameterID    uint8 // 1..255
	RanParameterValue RANParameterValue
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const ( // enum RANParameterTest
	RANParameterTest_equal = iota
	RANParameterTest_greaterthan
	RANParameterTest_lessthan
	RANParameterTest_contains
	RANParameterTest_present
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RANParameterValue struct {
	ValueIntPresent  bool
	ValueInt         int64
	ValueEnumPresent bool
	ValueEnum        int64
	ValueBoolPresent bool
	ValueBool        bool
	ValueBitSPresent bool
	ValueBitS        BitString
	ValueOctSPresent bool
	ValueOctS        OctetString
	ValuePrtSPresent bool
	ValuePrtS        OctetString
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type BitString struct {
	UnusedBits uint8
	Length     uint64
	Data       []uint8
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type OctetString struct {
	Length uint64
	Data   []uint8
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type SubsequentAction struct {
	Present    bool
	Type       uint64
	TimetoWait uint64
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionToBeSetupItem struct {
	ActionId                   uint64
	ActionType                 uint64
	RicActionDefinitionPresent bool
	ActionDefinitionChoice
	SubsequentAction
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type Cause struct {
	Content  uint8
	CauseVal uint8
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionAdmittedItem struct {
	ActionId uint64
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionAdmittedList struct {
	Items []ActionAdmittedItem //16
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionNotAdmittedItem struct {
	ActionId uint64
	Cause    Cause
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionNotAdmittedList struct {
	Items []ActionNotAdmittedItem //16
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type CriticalityDiagnosticsIEListItem struct {
	IeCriticality uint8 //Crit
	IeID          uint32
	TypeOfError   uint8
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type CriticalityDiagnosticsIEList struct {
	Items []CriticalityDiagnosticsIEListItem //256
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type CriticalityDiagnostics struct {
	Present         bool
	ProcCodePresent bool
	ProcCode        uint64
	TrigMsgPresent  bool
	TrigMsg         uint64
	ProcCritPresent bool
	ProcCrit        uint8 //Crit
	CriticalityDiagnosticsIEList
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type SubscriptionTestMsgContent struct {
	NBX2EventTriggerDefinitionPresent  bool
	NBNRTEventTriggerDefinitionPresent bool
	ActionDefinitionX2Format1Present   bool
	ActionDefinitionX2Format2Present   bool
	ActionDefinitionNRTFormat1Present  bool
	ActionParameterValueIntPresent     bool
	ActionParameterValueEnumPresent    bool
	ActionParameterValueBoolPresent    bool
	ActionParameterValueBitSPresent    bool
	ActionParameterValueOctSPresent    bool
	ActionParameterValuePrtSPresent    bool
	RANParameterValueIntPresent        bool
	RANParameterValueEnumPresent       bool
	RANParameterValueBoolPresent       bool
	RANParameterValueBitSPresent       bool
	RANParameterValueOctSPresent       bool
	RANParameterValuePrtSPresent       bool
}
