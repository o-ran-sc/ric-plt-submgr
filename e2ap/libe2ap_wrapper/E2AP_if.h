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

#ifndef E2AP_IF_H
#define E2AP_IF_H

#ifndef ASN_DISABLE_OER_SUPPORT
#define ASN_DISABLE_OER_SUPPORT
#endif

#include <stdbool.h>
#include <stdint.h>
#include <ProcedureCode.h>
#include <PrintableString.h>
#include "memtrack.h"

#ifdef	__cplusplus
extern "C" {
#endif

extern const int cCauseRICRequest;
extern const int cCauseRICService;
extern const int cCauseTransport;
extern const int cCauseProtocol;
extern const int cCauseMisc;

typedef unsigned char byte;

//extern const int64_t cMaxNrOfErrors;

extern const uint64_t cMaxSizeOfOctetString;

typedef struct { // Octet string in ASN.1 does not have maximum length!
    size_t contentLength;
    uint8_t data[1024]; // Table size is const cMaxSizeOfOctetString
} OctetString_t;

typedef struct { // Octet string in ASN.1 does not have maximum length!
    size_t length;
    uint8_t* data;
} DynOctetString_t;

typedef struct {
    uint8_t unusedBits; // Trailing unused bits 0 - 7
    size_t byteLength;  // Length in bytes
    uint8_t* data;
} DynBitString_t;

typedef struct  {
	uint32_t ricRequestorID; // 0..65535
	uint32_t ricInstanceID;  // 0..65535
} RICRequestID_t;

typedef uint16_t RANFunctionID_t; // 0..4095

typedef uint64_t RICActionID_t; // 0..255

enum RICActionType_t {
     RICActionType_report,
     RICActionType_insert,
     RICActionType_policy
};

typedef uint64_t StyleID_t;

typedef uint32_t ParameterID_t;  // 0..255 (maxofActionParameters)

typedef struct { // CHOICE. Only one value can be present
    bool valueIntPresent;
	int64_t valueInt;
	bool valueEnumPresent;
	int64_t valueEnum;
    bool valueBoolPresent;
	bool valueBool;
    bool valueBitSPresent;
	DynBitString_t valueBitS;
    bool valueOctSPresent;
	DynOctetString_t valueOctS;
	bool valuePrtSPresent;
	DynOctetString_t valuePrtS;
} ActionParameterValue_t;

typedef struct {
    ParameterID_t parameterID;
    ActionParameterValue_t actionParameterValue;
} ActionParameterItem_t;

typedef struct {
    StyleID_t styleID;
    uint8_t actionParameterCount;
    ActionParameterItem_t actionParameterItem[255]; // OPTIONAL. 1..255 (maxofRANParameters)
} E2SMgNBX2actionDefinition_t;

enum RANParameterTest_t {
	RANParameterTest_equal,
	RANParameterTest_greaterthan,
	RANParameterTest_lessthan,
	RANParameterTest_contains,
	RANParameterTest_present
};

typedef struct {
    bool valueIntPresent;
	int64_t valueInt;
	bool valueEnumPresent;
	int64_t valueEnum;
    bool valueBoolPresent;
	bool valueBool;
    bool valueBitSPresent;
	DynBitString_t valueBitS;
    bool valueOctSPresent;
	DynOctetString_t valueOctS;
	bool valuePrtSPresent;
	DynOctetString_t valuePrtS;
} RANParameterValue_t;

typedef int64_t RANueGroupID_t; // INTEGER
typedef uint32_t RANParameterID_t; // 0..255 (maxofRANParameters)

typedef struct {
	RANParameterID_t ranParameterID;
	RANParameterValue_t ranParameterValue;
} RANParameterItem_t;

typedef struct {
	RANParameterID_t ranParameterID;
	uint8_t ranParameterTest;               // This is type of enum RANParameterTest_t
	RANParameterValue_t ranParameterValue;
} RANueGroupDefItem_t;

typedef struct {
    uint8_t ranUeGroupDefCount;
	RANueGroupDefItem_t ranUeGroupDefItem[255]; //OPTIONAL. 1..255 (maxofRANParameters)
} RANueGroupDefinition_t;

typedef struct {
    uint8_t ranParameterCount;
	RANParameterItem_t ranParameterItem[255]; //OPTIONAL. 1..255 (maxofRANParameters)
} RANimperativePolicy_t;

typedef struct {
    RANueGroupID_t ranUEgroupID;
	RANueGroupDefinition_t ranUEgroupDefinition;
	RANimperativePolicy_t ranPolicy;
} RANueGroupItem_t;

typedef struct {
    uint8_t ranUeGroupCount;
    RANueGroupItem_t ranUeGroupItem[15]; // OPTIONAL. 1..15 (maxofUEgroup)
} E2SMgNBX2ActionDefinitionFormat2_t;

enum RICSubsequentActionType_t {
	RICSubsequentActionType_Continue,
	RICSubsequentActionType_wait
};

typedef struct {
    uint8_t ranParameterCount;
	RANParameterItem_t ranParameterList[255];	// OPTIONAL. 1..255 (maxofRANParameters)
} E2SMgNBNRTActionDefinitionFormat1_t;

typedef struct {
    OctetString_t octetString;   // This element is E2AP spec format
    // CHOICE. Only one value can be present
    bool actionDefinitionX2Format1Present;
	E2SMgNBX2actionDefinition_t* actionDefinitionX2Format1; // This element is E2SM-gNB-X2 format
	bool actionDefinitionX2Format2Present;
	E2SMgNBX2ActionDefinitionFormat2_t* actionDefinitionX2Format2; // This element is E2SM-gNB-X2 format
	bool actionDefinitionNRTFormat1Present;
    E2SMgNBNRTActionDefinitionFormat1_t* actionDefinitionNRTFormat1; // This element is E2SM-gNB-NRT format
} RICActionDefinitionChoice_t;

enum RICTimeToWait_t {
	RICTimeToWait_zero,
	RICTimeToWait_w1ms,
	RICTimeToWait_w2ms,
	RICTimeToWait_w5ms,
	RICTimeToWait_w10ms,
	RICTimeToWait_w20ms,
	RICTimeToWait_w30ms,
	RICTimeToWait_w40ms,
	RICTimeToWait_w50ms,
	RICTimeToWait_w100ms,
	RICTimeToWait_w200ms,
    RICTimeToWait_w500ms,
	RICTimeToWait_w1s,
	RICTimeToWait_w2s,
	RICTimeToWait_w5s,
	RICTimeToWait_w10s,
	RICTimeToWait_w20s,
	RICTimeToWait_w60s
};

typedef struct {
	uint64_t ricSubsequentActionType;  // This is type of enum RICSubsequentActionType_t
	uint64_t ricTimeToWait;  // This is type of enum RICTimeToWait_t
} RICSubsequentAction_t;

typedef struct  {
	RICActionID_t ricActionID;
	uint64_t ricActionType;  // This is type of enum RICActionType_t
	bool ricActionDefinitionPresent;
	RICActionDefinitionChoice_t ricActionDefinitionChoice;
	bool ricSubsequentActionPresent;
	RICSubsequentAction_t ricSubsequentAction;
} RICActionToBeSetupItem_t;

static const uint64_t cMaxofRICactionID = 16;

typedef struct  {
    uint8_t contentLength;
    RICActionToBeSetupItem_t ricActionToBeSetupItem[16];  // 1..16 // Table size is const cMaxofRICactionID
} RICActionToBeSetupList_t;

typedef struct {
    uint8_t contentLength;
    uint8_t pLMNIdentityVal[3];
} PLMNIdentity_t;

// size of eNB-id
extern const size_t cMacroENBIDP_20Bits;
extern const size_t cHomeENBID_28Bits;
extern const size_t cShortMacroENBID_18Bits;
extern const size_t clongMacroENBIDP_21Bits;

typedef struct {   // gNB-ID (SIZE 22..32 bits) or eNB-ID (SIZE 18, 20,21 or 28 bits)
    uint8_t bits;
    uint32_t nodeID;
} NodeID_t;

typedef struct {
	PLMNIdentity_t  pLMNIdentity;
	NodeID_t        nodeID;
}  GlobalNodeID_t;

typedef struct {   // CHOICE. Only either value can be present
	bool globalENBIDPresent;
	GlobalNodeID_t globalENBID;
	bool globalGNBIDPresent;
	GlobalNodeID_t globalGNBID;
} InterfaceID_t;

enum InterfaceDirection__t {
	InterfaceDirection__incoming,
	InterfaceDirection__outgoing
};

typedef uint8_t ProcedureCode__t;

enum TypeOfMessage_t {
    TypeOfMessage_nothing,
    TypeOfMessage_InitiatingMessage,
    TypeOfMessage_SuccessfulOutcome,
    TypeOfMessage_UnsuccessfulOutcome
};

typedef struct  {
	ProcedureCode__t procedureCode;
	uint8_t typeOfMessage;  // This is type of enum TypeOfMessage_t
} InterfaceMessageType_t;

typedef uint32_t InterfaceProtocolIEID_t;

enum InterfaceProtocolIETest_t {
	ProtocolIEtestCondition_equal,
	ProtocolIEtestCondition_greaterthan,
	ProtocolIEtestCondition_lessthan,
	ProtocolIEtestCondition_contains,
	ProtocolIEtestCondition_present
};

typedef struct {   // CHOICE. Only one value can be present
    bool valueIntPresent;
	int64_t valueInt;
	bool valueEnumPresent;
	int64_t valueEnum;
    bool valueBoolPresent;
	bool valueBool;
    bool valueBitStringPresent;
	DynBitString_t valueBitString;
    bool octetstringPresent;
	DynOctetString_t octetString;
} InterfaceProtocolIEValue_t;

typedef struct {
    InterfaceProtocolIEID_t interfaceProtocolIEID;
    uint8_t interfaceProtocolIETest;                        // This is type of enum InterfaceProtocolIETest_t
    InterfaceProtocolIEValue_t  interfaceProtocolIEValue;
} InterfacProtocolIE_t;

static const uint64_t cMaxofProtocolIE = 15;

typedef struct {
    InterfacProtocolIE_t InterfacProtocolIE[15]; // Table size is const cMaxofProtocolIE
} InterfaceProtocolIEList_t;

typedef struct {
    InterfaceID_t interfaceID;
    uint8_t interfaceDirection;  // This is type of enum InterfaceDirection_t
    InterfaceMessageType_t interfaceMessageType;
    bool interfaceProtocolIEListPresent;
    InterfaceProtocolIEList_t interfaceProtocolIEList;  // OPTIONAL. Not used in RIC currently
} E2SMgNBX2eventTriggerDefinition_t;

enum NRTTriggerNature_t {
    NRTTriggerNature_t_now,
    NRTTriggerNature_t_onchange
};

typedef struct {
	uint8_t triggerNature;  // This is type of enum NRTTriggerNature_t
} E2SMgNBNRTEventTriggerDefinitionFormat1_t;

typedef struct {
    E2SMgNBNRTEventTriggerDefinitionFormat1_t eventDefinitionFormat1;
} E2SMgNBNRTEventTriggerDefinition_t;

typedef struct {
    OctetString_t octetString;   // This element is E2AP spec format
    // CHOICE. Only one value can be present.
    bool E2SMgNBX2EventTriggerDefinitionPresent;
    E2SMgNBX2eventTriggerDefinition_t e2SMgNBX2eventTriggerDefinition;  // This element is E2SM-gNB-X2 spec format
    bool E2SMgNBNRTEventTriggerDefinitionPresent;
    E2SMgNBNRTEventTriggerDefinition_t e2SMgNBNRTEventTriggerDefinition; // This element is E2SM-gNB-NRT spec format
} RICEventTriggerDefinition_t;

typedef struct {
    RICEventTriggerDefinition_t ricEventTriggerDefinition;
    RICActionToBeSetupList_t ricActionToBeSetupItemIEs;
} RICSubscriptionDetails_t;

typedef struct {
    uint8_t contentLength;
	RICActionID_t ricActionID[16]; // Table size is const cMaxofRICactionID
} RICActionAdmittedList_t;

extern const int cCauseRIC; // This is content of type CauseRIC_t
extern const int cCauseRICService; // This is content of type CauseRICservice_t
extern const int cRICCauseTransport; // This is content of type CauseTransport_t
extern const int cRICCauseProtocol; // This is content of type CauseProtocol_t
extern const int cRICCauseMisc; // This is content of type CauseMisc_t

typedef struct {
    uint8_t content; // See above constants
    uint8_t causeVal; // This is type of enum CauseRIC_t
} RICCause_t;

typedef struct {
	RICActionID_t ricActionID;
    RICCause_t cause;
} RICActionNotAdmittedItem_t;

typedef struct {
    uint8_t contentLength;
    RICActionNotAdmittedItem_t RICActionNotAdmittedItem[16];  // Table size is const cMaxofRICactionID
} RICActionNotAdmittedList_t;

enum Criticality_t {
    Criticality__reject,
    Criticality__ignore,
    Criticality__notify
};

typedef uint32_t ProtocolIE_ID__t;

enum TriggeringMessage__t {
    TriggeringMessage__initiating_message,
    TriggeringMessage__successful_outcome,
    TriggeringMessage__unsuccessful_outcome
};

enum TypeOfError_t {
	TypeOfError_not_understood,
	TypeOfError_missing
};

typedef struct {
	uint8_t iECriticality; // This is type of enum Criticality_t
	ProtocolIE_ID__t iE_ID;
	uint8_t typeOfError; // This is type of enum TypeOfError_t
	//iE-Extensions  // This has no content in E2AP ASN.1 specification
} CriticalityDiagnosticsIEListItem_t;

typedef struct {
    bool procedureCodePresent;
	ProcedureCode__t procedureCode;  // OPTIONAL
	bool triggeringMessagePresent;
	uint8_t triggeringMessage;       // OPTIONAL. This is type of enum TriggeringMessage_t
	bool procedureCriticalityPresent;
	uint8_t procedureCriticality;    // OPTIONAL. This is type of enum Criticality_t
	bool ricRequestorIDPresent;
    RICRequestID_t ricRequestorID;   //OPTIONAL
	bool iEsCriticalityDiagnosticsPresent;
    uint16_t criticalityDiagnosticsIELength; // 1..256
	CriticalityDiagnosticsIEListItem_t criticalityDiagnosticsIEListItem[256];  // OPTIONAL. Table size is const cMaxNrOfErrors
} CriticalityDiagnostics__t;

typedef struct {
    OctetString_t octetString;    // E2AP spec format, the other elements for E2SM-X2 format
    uint64_t ricCallProcessIDVal;
} RICCallProcessID_t;

//////////////////////////////////////////////////////////////////////
// E2 Error codes
enum e2err {
    e2err_OK,
    e2err_RICSubscriptionRequestAllocRICrequestIDFail,
    e2err_RICSubscriptionRequestAllocRANfunctionIDFail,
    e2err_RICSubscriptionRequestAllocRICeventTriggerDefinitionBufFail,
    e2err_RICSubscriptionRequestAllocRICaction_ToBeSetup_ItemIEsFail,
    e2err_RICSubscriptionRequestAllocactionParameterValueValueBitSFail,
    e2err_RICSubscriptionRequestAllocactionParameterValueValueOctSFail,
    e2err_RICSubscriptionRequestAllocactionParameterValueValuePrtsSFail,
    e2err_RICSubscriptionRequestAllocactionRanParameterValueValueBitSFail,
    e2err_RICSubscriptionRequestAllocactionRanParameterValueValueOctSFail,
    e2err_RICSubscriptionRequestAllocactionRanParameterValueValuePrtsSFail,
    e2err_RICSubscriptionRequestAllocactionRanParameterValue2ValueBitSFail,
    e2err_RICSubscriptionRequestAllocactionRanParameterValue2ValueOctSFail,
    e2err_RICSubscriptionRequestAllocactionRanParameterValue2ValuePrtsSFail,
    e2err_RICSubscriptionRequestAllocactionDefinitionX2Format1Fail,
    e2err_RICSubscriptionRequestAllocactionDefinitionX2Format2Fail,
    e2err_RICSubscriptionRequestAllocactionDefinitionNRTFormat1Fail,
    e2err_RICSubscriptionRequestAllocRICactionDefinitionBufFail,
    e2err_RICSubscriptionRequestAllocRICactionDefinitionFail,
    e2err_RICSubscriptionRequestRICActionDefinitionEmpty,
    e2err_RICSubscriptionRequestRICActionDefinitionEmptyE2_E2SM_gNB_X2_actionDefinition,
    e2err_RICSubscriptionRequestRICActionDefinitionEmptyE2_E2SM_gNB_NRT_actionDefinition,
    e2err_RICSubscriptionRequestActionParameterItemFail,
    e2err_RICActionDefinitionChoicePackFail_1,
    e2err_RICActionDefinitionChoicePackFail_2,
    e2err_RICSubscriptionRequestAllocE2_RANueGroupDef_ItemFail,
    e2err_RICSubscriptionRequestAllocRANParameter_ItemFail,
    e2err_RICSubscriptionRequestRanranUeGroupDefItemParameterValueEmptyFail,
    e2err_RICSubscriptionRequestRanParameterItemRanParameterValueEmptyFail,
    e2err_RICSubscriptionRequestAllocActionDefinitionFail,
    e2err_RICSubscriptionRequestAllocNRTRANParameter_ItemFail,
    e2err_RICSubscriptionRequestAllocactionNRTRanParameterValue2ValueBitSFail,
    e2err_RICSubscriptionRequestAllocactionNRTRanParameterValue2ValueOctSFail,
    e2err_RICSubscriptionRequestAllocactionNRTRanParameterValue2ValuePrtsSFail,
    e2err_RICSubscriptionRequestRanParameterItemNRTRanParameterValueEmptyFail,
    e2err_RICSubscriptionRequestAsn_set_addE2_ActionParameter_ItemFail,
    e2err_RICSubscriptionRequestAsn_set_addRANueGroupDef_ItemFail,
    e2err_RICSubscriptionRequestAsn_set_addE2_RANParameter_ItemFail,
    e2err_RICSubscriptionRequestAsn_set_addE2_NRTRANParameter_ItemFail,
    e2err_RICActionDefinitionChoiceWMOREFail,
    e2err_RICActionDefinitionChoiceDecodeFAIL,
    e2err_RICActionDefinitionChoiceDecodeDefaultFail,
    e2err_RICNRTActionDefinitionChoiceWMOREFail,
    e2err_RICNRTActionDefinitionChoiceDecodeFAIL,
    e2err_RICNRTActionDefinitionChoiceDecodeDefaultFail,
    e2err_RICActionDefinitionChoiceEmptyFAIL,
    e2err_RICNRTEventTriggerDefinitionDecodeWMOREFail,
    e2err_RICNRTEventTriggerDefinitionDecodeFAIL,
    e2err_RICNRTEventTriggerDefinitionDecodeDefaultFail,
    e2err_RICEventTriggerDefinitionEmptyDecodeDefaultFail,
    e2err_RICSubscriptionRequestAllocE2_E2SM_gNB_X2_ActionDefinitionChoiceFail,
    e2err_RICSubscriptionRequestAllocE2_E2SM_gNB_NRT_ActionDefinitionFormat1Fail,
    e2err_RICSubscriptionRequestNRTRanParameterItemRanParameterValueEmptyFail,
    e2err_RICSubscriptionRequestNRTAllocActionDefinitionFail,
    e2err_RICSubscriptionRequestAllocE2_E2SM_gNB_NRT_ActionDefinitionFail,
    e2err_RICSubscriptionRequestAllocRICsubsequentActionFail,
    e2err_RICSubscriptionRequestAllocRICsubscriptionRequest_IEsFail,
    e2err_RICSubscriptionRequestEncodeFail,
    e2err_RICSubscriptionRequestAllocE2AP_PDUFail,
    e2err_RICEventTriggerDefinitionIEValueFail_1,
    e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDpLMN_IdentityBufFail,
    e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDmacro_eNB_IDBufFail,
    e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDhome_eNB_IDBufFail,
    e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDshort_Macro_eNB_IDBufFail,
    e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDlong_Macro_eNB_IDBufFail,
    e2err_RICEventTriggerDefinitionIEValueFail_2,
    e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_gNB_IDpLMN_IdentityBufFail,
    e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_gNB_IDgNB_IDBufFail,
    e2err_RICEventTriggerDefinitionIEValueFail_3,
    e2err_RICEventTriggerDefinitionIEValueFail_4,
    e2err_RICEventTriggerDefinitionPackFail_1,
    e2err_RICEventTriggerDefinitionPackFail_2,
    e2err_RICENRTventTriggerDefinitionPackFail_1,
    e2err_RICNRTEventTriggerDefinitionPackFail_2,
    e2err_RICEventTriggerDefinitionAllocE2SM_gNB_X2_eventTriggerDefinitionFail,
    e2err_RICEventTriggerDefinitionAllocE2SM_gNB_NRT_eventTriggerDefinitionFail,
    e2err_RICEventTriggerDefinitionAllocEventTriggerDefinitionEmptyFail,
    e2err_RICSubscriptionResponseAllocRICrequestIDFail,
    e2err_RICSubscriptionResponseAllocRANfunctionIDFail,
    e2err_RICSubscriptionResponseAllocRICaction_Admitted_ItemIEsFail,
    e2err_RICSubscriptionResponseAllocRICActionAdmittedListFail,
    e2err_RICSubscriptionResponseAllocRICaction_NotAdmitted_ItemIEsFail,
    e2err_RICSubscriptionResponseEncodeFail,
    e2err_RICSubscriptionResponseAllocE2AP_PDUFail,
    e2err_RICSubscriptionFailureAllocRICrequestIDFail,
    e2err_RICSubscriptionFailureAllocRANfunctionIDFail,
    e2err_RICSubscriptionFailureAllocRICaction_NotAdmitted_ItemIEsFail,
    e2err_RICSubscriptionFailureAllocRICActionAdmittedListFail,
    e2err_RICSubscriptionFailureEncodeFail,
    e2err_RICSubscriptionFailureAllocE2AP_PDUFail,
    e2err_E2SM_gNB_X2_indicationMessageAllocE2AP_PDUFail,
    e2err_RICSubscriptionDeleteRequestAllocRICrequestIDFail,
    e2err_RICSubscriptionDeleteRequestAllocRANfunctionIDFail,
    e2err_RICSubscriptionDeleteRequestEncodeFail,
    e2err_RICSubscriptionDeleteRequestAllocE2AP_PDUFail,
    e2err_RICSubscriptionDeleteResponseAllocRICrequestIDFail,
    e2err_RICSubscriptionDeleteResponseAllocRANfunctionIDFail,
    e2err_RICSubscriptionDeleteResponseEncodeFail,
    e2err_RICSubscriptionDeleteResponseAllocE2AP_PDUFail,
    e2err_RICSubscriptionDeleteFailureAllocRICrequestIDFail,
    e2err_RICSubscriptionDeleteFailureAllocRANfunctionIDFail,
    e2err_RICSubscriptionDeleteFailureAllocRICcauseFail,
    e2err_RICSubscriptionDeleteFailureEncodeFail,
    e2err_RICSubscriptionDeleteFailureAllocE2AP_PDUFail,
    e2err_RICsubscriptionRequestRICrequestIDMissing,
    e2err_RICsubscriptionRequestRANfunctionIDMissing,
    e2err_RICsubscriptionRequestICsubscriptionMissing,
    e2err_RICEventTriggerDefinitionIEValueFail_5,
    e2err_RICEventTriggerDefinitionIEValueFail_6,
    e2err_RICEventTriggerDefinitionIEValueFail_7,
    e2err_RICEventTriggerDefinitionIEValueFail_8,
    e2err_RICEventTriggerDefinitionDecodeWMOREFail,
    e2err_RICEventTriggerDefinitionDecodeFAIL,
    e2err_RICEventTriggerDefinitionDecodeDefaultFail,
    e2err_RICsubscriptionResponseRICrequestIDMissing,
    e2err_RICsubscriptionResponseRANfunctionIDMissing,
    e2err_RICsubscriptionResponseRICaction_Admitted_ListMissing,
    e2err_RICsubscriptionFailureRICrequestIDMissing,
    e2err_RICsubscriptionFailureRANfunctionIDMissing,
    e2err_RICsubscriptionFailureRICaction_NotAdmitted_ListMissing,
    e2err_RICEventTriggerDefinitionIEValueFail_9,
    e2err_RICEventTriggerDefinitionIEValueFail_10,
    e2err_RICEventTriggerDefinitionIEValueFail_11,
    e2err_RICsubscriptionDeleteRequestRICrequestIDMissing,
    e2err_RICsubscriptionDeleteRequestRANfunctionIDMissing,
    e2err_RICsubscriptionDeleteResponseRICrequestIDMissing,
    e2err_RICsubscriptionDeleteResponseRANfunctionIDMissing,
    e2err_RICsubscriptionDeleteFailureRICrequestIDMissing,
    e2err_RICsubscriptionDeleteFailureRANfunctionIDMissing,
    e2err_RICsubscriptionDeleteFailureRICcauseMissing
};

static const char* const E2ErrorStrings[] = {
    "e2err_OK",
    "e2err_RICSubscriptionRequestAllocRICrequestIDFail",
    "e2err_RICSubscriptionRequestAllocRANfunctionIDFail",
    "e2err_RICSubscriptionRequestAllocRICeventTriggerDefinitionBufFail",
    "e2err_RICSubscriptionRequestAllocRICaction_ToBeSetup_ItemIEsFail",
    "e2err_RICSubscriptionRequestAllocactionParameterValueValueBitSFail",
    "e2err_RICSubscriptionRequestAllocactionParameterValueValueOctSFail",
    "e2err_RICSubscriptionRequestAllocactionParameterValueValuePrtsSFail",
    "e2err_RICSubscriptionRequestAllocactionRanParameterValueValueBitSFail",
    "e2err_RICSubscriptionRequestAllocactionRanParameterValueValueOctSFail",
    "e2err_RICSubscriptionRequestAllocactionRanParameterValueValuePrtsSFail",
    "e2err_RICSubscriptionRequestAllocactionRanParameterValue2ValueBitSFail",
    "e2err_RICSubscriptionRequestAllocactionRanParameterValue2ValueOctSFail",
    "e2err_RICSubscriptionRequestAllocactionRanParameterValue2ValuePrtsSFail",
    "e2err_RICSubscriptionRequestAllocactionDefinitionX2Format1Fail",
    "e2err_RICSubscriptionRequestAllocactionDefinitionX2Format2Fail",
    "e2err_RICSubscriptionRequestAllocactionDefinitionNRTFormat1Fail",
    "e2err_RICSubscriptionRequestAllocRICactionDefinitionBufFail",
    "e2err_RICSubscriptionRequestAllocRICactionDefinitionFail",
    "e2err_RICSubscriptionRequestRICActionDefinitionEmpty",
    "e2err_RICSubscriptionRequestRICActionDefinitionEmptyE2_E2SM_gNB_X2_actionDefinition",
    "e2err_RICSubscriptionRequestRICActionDefinitionEmptyE2_E2SM_gNB_NRT_actionDefinition",
    "e2err_RICSubscriptionRequestActionParameterItemFail",
    "e2err_RICActionDefinitionChoicePackFail_1",
    "e2err_RICActionDefinitionChoicePackFail_2",
    "e2err_RICSubscriptionRequestAllocE2_RANueGroupDef_ItemFail",
    "e2err_RICSubscriptionRequestAllocRANParameter_ItemFail",
    "e2err_RICSubscriptionRequestRanranUeGroupDefItemParameterValueEmptyFail",
    "e2err_RICSubscriptionRequestRanParameterItemRanParameterValueEmptyFail",
    "e2err_RICSubscriptionRequestAllocActionDefinitionFail",
    "e2err_RICSubscriptionRequestAllocNRTRANParameter_ItemFail",
    "e2err_RICSubscriptionRequestAllocactionNRTRanParameterValue2ValueBitSFail",
    "e2err_RICSubscriptionRequestAllocactionNRTRanParameterValue2ValueOctSFail",
    "e2err_RICSubscriptionRequestAllocactionNRTRanParameterValue2ValuePrtsSFail",
    "e2err_RICSubscriptionRequestRanParameterItemNRTRanParameterValueEmptyFail",
    "e2err_RICSubscriptionRequestAsn_set_addE2_ActionParameter_ItemFail",
    "e2err_RICSubscriptionRequestAsn_set_addRANueGroupDef_ItemFail",
    "e2err_RICSubscriptionRequestAsn_set_addE2_RANParameter_ItemFail",
    "e2err_RICSubscriptionRequestAsn_set_addE2_NRTRANParameter_ItemFail",
    "e2err_RICActionDefinitionChoiceWMOREFail",
    "e2err_RICActionDefinitionChoiceDecodeFAIL",
    "e2err_RICActionDefinitionChoiceDecodeDefaultFail",
    "e2err_RICNRTActionDefinitionChoiceWMOREFail",
    "e2err_RICNRTActionDefinitionChoiceDecodeFAIL",
    "e2err_RICNRTActionDefinitionChoiceDecodeDefaultFail",
    "e2err_RICActionDefinitionChoiceEmptyFAIL",
    "e2err_RICNRTEventTriggerDefinitionDecodeWMOREFail",
    "e2err_RICNRTEventTriggerDefinitionDecodeFAIL",
    "e2err_RICNRTEventTriggerDefinitionDecodeDefaultFail",
    "e2err_RICEventTriggerDefinitionEmptyDecodeDefaultFail",
    "e2err_RICSubscriptionRequestAllocE2_E2SM_gNB_X2_ActionDefinitionChoiceFail",
    "e2err_RICSubscriptionRequestAllocE2_E2SM_gNB_NRT_ActionDefinitionFormat1Fail",
    "e2err_RICSubscriptionRequestNRTRanParameterItemRanParameterValueEmptyFail",
    "e2err_RICSubscriptionRequestNRTAllocActionDefinitionFail",
    "e2err_RICSubscriptionRequestAllocE2_E2SM_gNB_NRT_ActionDefinitionFail",
    "e2err_RICSubscriptionRequestAllocRICsubsequentActionFail",
    "e2err_RICSubscriptionRequestAllocRICsubscriptionRequest_IEsFail",
    "e2err_RICSubscriptionRequestEncodeFail",
    "e2err_RICSubscriptionRequestAllocE2AP_PDUFail",
    "e2err_RICEventTriggerDefinitionIEValueFail_1",
    "e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDpLMN_IdentityBufFail",
    "e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDmacro_eNB_IDBufFail",
    "e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDhome_eNB_IDBufFail",
    "e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDshort_Macro_eNB_IDBufFail",
    "e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDlong_Macro_eNB_IDBufFail",
    "e2err_RICEventTriggerDefinitionIEValueFail_2",
    "e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_gNB_IDpLMN_IdentityBufFail",
    "e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_gNB_IDgNB_IDBufFail",
    "e2err_RICEventTriggerDefinitionIEValueFail_3",
    "e2err_RICEventTriggerDefinitionIEValueFail_4",
    "e2err_RICEventTriggerDefinitionPackFail_1",
    "e2err_RICEventTriggerDefinitionPackFail_2",
    "e2err_RICENRTventTriggerDefinitionPackFail_1",
    "e2err_RICNRTEventTriggerDefinitionPackFail_2",
    "e2err_RICEventTriggerDefinitionAllocE2SM_gNB_X2_eventTriggerDefinitionFail",
    "e2err_RICEventTriggerDefinitionAllocE2SM_gNB_NRT_eventTriggerDefinitionFail",
    "e2err_RICEventTriggerDefinitionAllocEventTriggerDefinitionEmptyFail",
    "e2err_RICSubscriptionResponseAllocRICrequestIDFail",
    "e2err_RICSubscriptionResponseAllocRANfunctionIDFail",
    "e2err_RICSubscriptionResponseAllocRICaction_Admitted_ItemIEsFail",
    "e2err_RICSubscriptionResponseAllocRICActionAdmittedListFail",
    "e2err_RICSubscriptionResponseAllocRICaction_NotAdmitted_ItemIEsFail",
    "e2err_RICSubscriptionResponseEncodeFail",
    "e2err_RICSubscriptionResponseAllocE2AP_PDUFail",
    "e2err_RICSubscriptionFailureAllocRICrequestIDFail",
    "e2err_RICSubscriptionFailureAllocRANfunctionIDFail",
    "e2err_RICSubscriptionFailureAllocRICaction_NotAdmitted_ItemIEsFail",
    "e2err_RICSubscriptionFailureAllocRICActionAdmittedListFail",
    "e2err_RICSubscriptionFailureEncodeFail",
    "e2err_RICSubscriptionFailureAllocE2AP_PDUFail",
    "e2err_E2SM_gNB_X2_indicationMessageAllocE2AP_PDUFail",
    "e2err_RICSubscriptionDeleteRequestAllocRICrequestIDFail",
    "e2err_RICSubscriptionDeleteRequestAllocRANfunctionIDFail",
    "e2err_RICSubscriptionDeleteRequestEncodeFail",
    "e2err_RICSubscriptionDeleteRequestAllocE2AP_PDUFail",
    "e2err_RICSubscriptionDeleteResponseAllocRICrequestIDFail",
    "e2err_RICSubscriptionDeleteResponseAllocRANfunctionIDFail",
    "e2err_RICSubscriptionDeleteResponseEncodeFail",
    "e2err_RICSubscriptionDeleteResponseAllocE2AP_PDUFail",
    "e2err_RICSubscriptionDeleteFailureAllocRICrequestIDFail",
    "e2err_RICSubscriptionDeleteFailureAllocRANfunctionIDFail",
    "e2err_RICSubscriptionDeleteFailureAllocRICcauseFail",
    "e2err_RICSubscriptionDeleteFailureEncodeFail",
    "e2err_RICSubscriptionDeleteFailureAllocE2AP_PDUFail",
    "e2err_RICsubscriptionRequestRICrequestIDMissing",
    "e2err_RICsubscriptionRequestRANfunctionIDMissing",
    "e2err_RICsubscriptionRequestICsubscriptionMissing",
    "e2err_RICEventTriggerDefinitionIEValueFail_5",
    "e2err_RICEventTriggerDefinitionIEValueFail_6",
    "e2err_RICEventTriggerDefinitionIEValueFail_7",
    "e2err_RICEventTriggerDefinitionIEValueFail_8",
    "e2err_RICEventTriggerDefinitionDecodeWMOREFail",
    "e2err_RICEventTriggerDefinitionDecodeFAIL",
    "e2err_RICEventTriggerDefinitionDecodeDefaultFail",
    "e2err_RICsubscriptionResponseRICrequestIDMissing",
    "e2err_RICsubscriptionResponseRANfunctionIDMissing",
    "e2err_RICsubscriptionResponseRICaction_Admitted_ListMissing",
    "e2err_RICsubscriptionFailureRICrequestIDMissing",
    "e2err_RICsubscriptionFailureRANfunctionIDMissing",
    "e2err_RICsubscriptionFailureRICaction_NotAdmitted_ListMissing",
    "e2err_RICEventTriggerDefinitionIEValueFail_9",
    "e2err_RICEventTriggerDefinitionIEValueFail_10",
    "e2err_RICEventTriggerDefinitionIEValueFail_11",
    "e2err_RICsubscriptionDeleteRequestRICrequestIDMissing",
    "e2err_RICsubscriptionDeleteRequestRANfunctionIDMissing",
    "e2err_RICsubscriptionDeleteResponseRICrequestIDMissing",
    "e2err_RICsubscriptionDeleteResponseRANfunctionIDMissing",
    "e2err_RICsubscriptionDeleteFailureRICrequestIDMissing",
    "e2err_RICsubscriptionDeleteFailureRANfunctionIDMissing",
    "e2err_RICsubscriptionDeleteFailureRICcauseMissing"
};

typedef struct {
    uint64_t messageType; // Initiating message or Successful outcome or Unsuccessful outcome
    uint64_t messageId;
} E2MessageInfo_t;

//////////////////////////////////////////////////////////////////////
// Message definitons

// Below constant values are same as in E2AP, E2SM and X2AP specs
extern const uint64_t cE2InitiatingMessage;
extern const uint64_t cE2SuccessfulOutcome;
extern const uint64_t cE2UnsuccessfulOutcome;

// E2AP messages. Below message id constant values are the same as in ASN.1 specification

// Initiating message
extern const uint64_t cRICSubscriptionRequest;
extern const uint64_t cRICSubscriptionDeleteRequest;

// Successful outcome
extern const uint64_t cRICSubscriptionResponse;
extern const uint64_t cRICsubscriptionDeleteResponse;

// Unsuccessful outcome
extern const uint64_t cRICSubscriptionFailure;
extern const uint64_t cRICsubscriptionDeleteFailure;

typedef struct {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
    RICSubscriptionDetails_t ricSubscriptionDetails;
} RICSubscriptionRequest_t;

typedef struct {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
    RICActionAdmittedList_t ricActionAdmittedList;
    bool ricActionNotAdmittedListPresent;
    RICActionNotAdmittedList_t ricActionNotAdmittedList;
} RICSubscriptionResponse_t;

typedef struct {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
    RICActionNotAdmittedList_t ricActionNotAdmittedList;
    bool criticalityDiagnosticsPresent;
    CriticalityDiagnostics__t criticalityDiagnostics;
} RICSubscriptionFailure_t;

typedef struct {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
} RICSubscriptionDeleteRequest_t;

typedef struct  {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
} RICSubscriptionDeleteResponse_t;

typedef struct  {
    RICRequestID_t ricRequestID;
    RANFunctionID_t ranFunctionID;
    RICCause_t cause;
    bool criticalityDiagnosticsPresent;
    CriticalityDiagnostics__t criticalityDiagnostics; // OPTIONAL. Not used in RIC currently
} RICSubscriptionDeleteFailure_t;

//////////////////////////////////////////////////////////////////////
// Function declarations

const char* getE2ErrorString(uint64_t);

typedef void* e2ap_pdu_ptr_t;

uint64_t packRICSubscriptionRequest(size_t*, byte*, char*,RICSubscriptionRequest_t*);
uint64_t packRICEventTriggerDefinition(char*,RICEventTriggerDefinition_t*);
uint64_t packRICActionDefinition(char*, RICActionDefinitionChoice_t*);
uint64_t packRICEventTriggerDefinitionX2Format(char* pLogBuffer, RICEventTriggerDefinition_t*);
uint64_t packRICEventTriggerDefinitionNRTFormat(char* pLogBuffer, RICEventTriggerDefinition_t*);
uint64_t packActionDefinitionX2Format(char*, RICActionDefinitionChoice_t*);
uint64_t packActionDefinitionNRTFormat(char*, RICActionDefinitionChoice_t*);
uint64_t packRICSubscriptionResponse(size_t*, byte*, char*,RICSubscriptionResponse_t*);
uint64_t packRICSubscriptionFailure(size_t*, byte*, char*,RICSubscriptionFailure_t*);
uint64_t packRICSubscriptionDeleteRequest(size_t*, byte*, char*,RICSubscriptionDeleteRequest_t*);
uint64_t packRICSubscriptionDeleteResponse(size_t*, byte*, char*,RICSubscriptionDeleteResponse_t*);
uint64_t packRICSubscriptionDeleteFailure(size_t*, byte*, char*,RICSubscriptionDeleteFailure_t*);

e2ap_pdu_ptr_t* unpackE2AP_pdu(const size_t, const byte*, char*, E2MessageInfo_t*);
uint64_t getRICSubscriptionRequestData(mem_track_hdr_t *, e2ap_pdu_ptr_t*, RICSubscriptionRequest_t*);
uint64_t getRICEventTriggerDefinitionData(RICEventTriggerDefinition_t*);
uint64_t getRICEventTriggerDefinitionDataX2Format(RICEventTriggerDefinition_t*);
uint64_t getRICEventTriggerDefinitionDataNRTFormat(RICEventTriggerDefinition_t*);
uint64_t getRICActionDefinitionData(mem_track_hdr_t *, RICActionDefinitionChoice_t*);
uint64_t getRICActionDefinitionDataX2Format(mem_track_hdr_t*, RICActionDefinitionChoice_t*);
uint64_t getRICActionDefinitionDataNRTFormat(mem_track_hdr_t*, RICActionDefinitionChoice_t*);
uint64_t getRICSubscriptionResponseData(e2ap_pdu_ptr_t*, RICSubscriptionResponse_t*);
uint64_t getRICSubscriptionFailureData(e2ap_pdu_ptr_t*, RICSubscriptionFailure_t*);
uint64_t getRICSubscriptionDeleteRequestData(e2ap_pdu_ptr_t*, RICSubscriptionDeleteRequest_t*);
uint64_t getRICSubscriptionDeleteResponseData(e2ap_pdu_ptr_t*, RICSubscriptionDeleteResponse_t*);
uint64_t getRICSubscriptionDeleteFailureData(e2ap_pdu_ptr_t*, RICSubscriptionDeleteFailure_t*);

void* allocDynMem(mem_track_hdr_t*, size_t);
bool addOctetString(mem_track_hdr_t *, DynOctetString_t*, uint64_t, void*);
bool addBitString(mem_track_hdr_t *, DynBitString_t*, uint64_t, void*, uint8_t);

uint64_t allocActionDefinitionX2Format1(mem_track_hdr_t*, E2SMgNBX2actionDefinition_t**);
uint64_t allocActionDefinitionX2Format2(mem_track_hdr_t*, E2SMgNBX2ActionDefinitionFormat2_t**);
uint64_t allocActionDefinitionNRTFormat1(mem_track_hdr_t*, E2SMgNBNRTActionDefinitionFormat1_t**);

uint64_t allocateOctetStringBuffer(DynOctetString_t*, uint64_t);
uint64_t allocateBitStringBuffer(mem_track_hdr_t *, DynBitString_t*, uint64_t);

#if DEBUG
bool TestRICSubscriptionRequest();
bool TestRICSubscriptionResponse();
bool TestRICSubscriptionFailure();
bool TestRICSubscriptionDeleteRequest();
bool TestRICSubscriptionDeleteResponse();
bool TestRICSubscriptionDeleteFailure();

void printRICSubscriptionRequest(const RICSubscriptionRequest_t*);
void printRICSubscriptionResponse(const RICSubscriptionResponse_t*);
void printRICSubscriptionFailure(const RICSubscriptionFailure_t*);
void printRICSubscriptionDeleteRequest(const RICSubscriptionDeleteRequest_t*);
void printRICSubscriptionDeleteResponse(const RICSubscriptionDeleteResponse_t*);
void printRICSubscriptionDeleteFailure(const RICSubscriptionDeleteFailure_t*);
#endif

#ifdef	__cplusplus
}
#endif

#endif
