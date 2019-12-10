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
const E2AP_InitiatingMessage uint64 = 1
const E2AP_SuccessfulOutcome uint64 = 2
const E2AP_UnsuccessfulOutcome uint64 = 3

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type RequestId struct {
	Id  uint32
	Seq uint32
}

func (rid *RequestId) String() string {
	return strconv.FormatUint((uint64)(rid.Id), 10) + string(":") + strconv.FormatUint((uint64)(rid.Seq), 10)
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
	InterfaceId
	InterfaceDirection uint32
	ProcedureCode      uint32
	TypeOfMessage      uint64
}

/*
//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type CallProcessId struct {
	CallProcessIDVal uint32
}
*/

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type ActionDefinition struct {
	Present bool
	StyleId uint64
	ParamId uint32
	//ParamValue
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
	ActionId   uint64
	ActionType uint64
	ActionDefinition
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
