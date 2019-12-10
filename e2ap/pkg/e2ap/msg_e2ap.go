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

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

// E2AP messages
// Initiating message
const E2AP_RICSubscriptionRequest uint64 = 1
const E2AP_RICSubscriptionDeleteRequest uint64 = 2

// const E2AP_RICServiceUpdate uint64 = 3
// const E2AP_RICControlRequest uint64 = 4
//
// //const E2AP_X2SetupRequest uint64 = 5;
// const E2AP_ENDCX2SetupRequest uint64 = 6
// const E2AP_ResourceStatusRequest uint64 = 7
// const E2AP_ENBConfigurationUpdate uint64 = 8
// const E2AP_ENDCConfigurationUpdate uint64 = 9
// const E2AP_ResetRequest uint64 = 10
const E2AP_RICIndication uint64 = 11

// const E2AP_RICServiceQuery uint64 = 12
// const E2AP_LoadInformation uint64 = 13
// const E2AP_GNBStatusIndication uint64 = 14
// const E2AP_ResourceStatusUpdate uint64 = 15
// const E2AP_ErrorIndication uint64 = 16
//
// // Successful outcome
const E2AP_RICSubscriptionResponse uint64 = 1
const E2AP_RICSubscriptionDeleteResponse uint64 = 2

// const E2AP_RICserviceUpdateAcknowledge uint64 = 3
// const E2AP_RICcontrolAcknowledge uint64 = 4
//
// //const E2AP_X2SetupResponse uint64 = 5;
// const E2AP_ENDCX2SetupResponse uint64 = 6
// const E2AP_ResourceStatusResponse uint64 = 7
// const E2AP_ENBConfigurationUpdateAcknowledge uint64 = 8
// const E2AP_ENDCConfigurationUpdateAcknowledge uint64 = 9
// const E2AP_ResetResponse uint64 = 10
//
// // Unsuccessful outcome
const E2AP_RICSubscriptionFailure uint64 = 1
const E2AP_RICSubscriptionDeleteFailure uint64 = 2

// const E2AP_RICserviceUpdateFailure uint64 = 3
// const E2AP_RICcontrolFailure uint64 = 4
//
// //const E2AP_X2SetupFailure uint64 = 5;
// const E2AP_ENDCX2SetupFailure uint64 = 6
// const E2AP_ResourceStatusFailure uint64 = 7
// const E2AP_ENBConfigurationUpdateFailure uint64 = 8
// const E2AP_ENDCConfigurationUpdateFailure uint64 = 9
//

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const E2AP_IndicationTypeReport uint64 = 0
const E2AP_IndicationTypeInsert uint64 = 1

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const E2AP_ActionTypeReport uint64 = 0
const E2AP_ActionTypeInsert uint64 = 1
const E2AP_ActionTypePolicy uint64 = 2

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const E2AP_SubSeqActionTypeContinue uint64 = 0
const E2AP_SubSeqActionTypeWait uint64 = 1

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const E2AP_TimeToWaitZero uint64 = 0
const E2AP_TimeToWaitW1ms uint64 = 1
const E2AP_TimeToWaitW2ms uint64 = 2
const E2AP_TimeToWaitW5ms uint64 = 3
const E2AP_TimeToWaitW10ms uint64 = 4
const E2AP_TimeToWaitW20ms uint64 = 4
const E2AP_TimeToWaitW30ms uint64 = 5
const E2AP_TimeToWaitW40ms uint64 = 6
const E2AP_TimeToWaitW50ms uint64 = 7
const E2AP_TimeToWaitW100ms uint64 = 8
const E2AP_TimeToWaitW200ms uint64 = 9
const E2AP_TimeToWaitW500ms uint64 = 10
const E2AP_TimeToWaitW1s uint64 = 11
const E2AP_TimeToWaitW2s uint64 = 12
const E2AP_TimeToWaitW5s uint64 = 13
const E2AP_TimeToWaitW10s uint64 = 14
const E2AP_TimeToWaitW20s uint64 = 15
const E2AP_TimeToWaitW60 uint64 = 16

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const E2AP_InterfaceDirectionIncoming uint32 = 0
const E2AP_InterfaceDirectionOutgoing uint32 = 1

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const E2AP_CriticalityReject uint8 = 0
const E2AP_CriticalityIgnore uint8 = 1
const E2AP_CriticalityNotify uint8 = 2

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
const E2AP_ENBIDMacroPBits20 uint8 = 20
const E2AP_ENBIDHomeBits28 uint8 = 28
const E2AP_ENBIDShortMacroits18 uint8 = 18
const E2AP_ENBIDlongMacroBits21 uint8 = 21

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
