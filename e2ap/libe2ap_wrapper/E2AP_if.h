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

#ifdef	__cplusplus
extern "C" {
#endif

typedef unsigned char byte;

extern const int64_t cMaxNrOfErrors;

extern const uint64_t cMaxSizeOfOctetString;

typedef struct { // Octet string in ASN.1 does not have maximum length!
    size_t contentLength;
    uint8_t data[1024]; // table size is const cMaxSizeOfOctetString
} OctetString_t;

typedef struct {
    uint8_t unusedbits; // trailing unused bits 0 - 7
    size_t byteLength;  // length in bytes
    uint8_t data[1024];
} Bitstring_t;

typedef struct  {
	uint32_t ricRequestorID;
	uint32_t ricRequestSequenceNumber;
} RICRequestID_t;

typedef uint16_t RANFunctionID_t;

typedef uint64_t RICActionID_t;

enum RICActionType_t {
     RICActionType_report
    ,RICActionType_insert
    ,RICActionType_policy
};

typedef uint64_t StyleID_t;

typedef uint32_t ParameterID_t;

typedef struct {
    uint32_t dummy; // This data type has no content. This dummy is added here to solve problem with Golang. Golang do not like empty types.
} ParameterValue_t;

typedef struct {
    ParameterID_t parameterID;
    ParameterValue_t ParameterValue;
} SequenceOfActionParameters_t;

typedef struct {
    StyleID_t styleID;
    SequenceOfActionParameters_t sequenceOfActionParameters;
} RICActionDefinition_t;

enum RICSubsequentActionType_t {
	RICSubsequentActionType_Continue,
	RICSubsequentActionType_wait
};

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
	uint64_t ricSubsequentActionType;  // this is type of enum RICSubsequentActionType_t
	uint64_t ricTimeToWait;  // this is type of enum RICTimeToWait_t
} RICSubsequentAction_t;

typedef struct  {
	RICActionID_t ricActionID;
	uint64_t ricActionType;  // this is type of enum RICActionType_t
	bool ricActionDefinitionPresent;
	RICActionDefinition_t ricActionDefinition;
	bool ricSubsequentActionPresent;
	RICSubsequentAction_t ricSubsequentAction;
} RICActionToBeSetupItem_t;

static const uint64_t cMaxofRICactionID = 16;

typedef struct  {
    uint8_t contentLength;
    RICActionToBeSetupItem_t ricActionToBeSetupItem[16];  // table size is const cMaxofRICactionID
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

typedef struct  {
	ProcedureCode__t procedureCode;
	uint8_t typeOfMessage;  // This is X2AP-PDU, CHOICE of InitiatingMessage or SuccessfulOutcome or UnsuccessfulOutcome
} RICInterfaceMessageType_t;

typedef uint32_t InterfaceProtocolIEID_t;

enum ProtocolIEtestCondition_t {
	ProtocolIEtestCondition_equal,
	ProtocolIEtestCondition_greaterthan,
	ProtocolIEtestCondition_lessthan,
	ProtocolIEtestCondition_contains,
	ProtocolIEtestCondition_present
};

typedef struct {   // CHOICE. Only one value can be present
    bool valueIntPresent;
	int64_t integer;           // INTEGER;
	bool valueEnumPresent;
	int64_t valueEnum;         // INTEGER
    bool valueBoolPresent;
	bool valueBool;	           // BOOLEAN
    bool valueBitSPresent;
	Bitstring_t octetstring;   // OCTET STRING,
    bool octetstringPresent;
	OctetString_t octetString; // OCTET STRING,
} InterfaceProtocolIEValue_t;

typedef struct {
    InterfaceProtocolIEID_t interfaceProtocolIEID;
    //ProtocolIEtestCondition_t protocolIEtestCondition;  Golang do not like this line. We do not need this right now.
    InterfaceProtocolIEValue_t  interfaceProtocolIEValue;
} SequenceOfProtocolIE_t;

static const uint64_t cMaxofProtocolIE = 16;

typedef struct {
    SequenceOfProtocolIE_t sequenceOfProtocolIE[16]; // table size is const cMaxofProtocolIE
} SequenceOfProtocolIEList_t;

typedef struct {
    OctetString_t octetString;   // E2AP spec format, the other elements for E2SM-X2 format
    InterfaceID_t interfaceID;
    uint8_t interfaceDirection;  // this is type of enum InterfaceDirection_t
    RICInterfaceMessageType_t interfaceMessageType ;
    bool sequenceOfProtocolIEListPresent;
    SequenceOfProtocolIEList_t SequenceOfProtocolIEList;
} RICEventTriggerDefinition_t;

typedef struct {
    RICEventTriggerDefinition_t ricEventTriggerDefinition;
    RICActionToBeSetupList_t ricActionToBeSetupItemIEs;
} RICSubscription_t;

typedef struct {
    uint8_t contentLength;
	RICActionID_t ricActionID[16]; // table size is const cMaxofRICactionID
} RICActionAdmittedList_t;

enum CauseRIC_t {
	CauseRIC__function_id_Invalid,
	CauseRIC__action_not_supported,
	CauseRIC__excessive_actions,
	CauseRIC__duplicate_action,
	CauseRIC__duplicate_event,
	CauseRIC__function_resource_limit,
	CauseRIC__request_id_unknown,
	CauseRIC__inconsistent_action_subsequent_action_sequence,
	CauseRIC__control_message_invalid,
	CauseRIC__call_process_id_invalid,
	CauseRIC__function_not_required,
	CauseRIC__excessive_functions,
	CauseRIC__ric_resource_limit
};

extern const int cRICCauseRadioNetwork; // this is content of type RICCause_t
extern const int cRICCauseTransport; // this is content of type RICCause_t
extern const int cRICCauseProtocol; // this is content of type RICCause_t
extern const int cRICCauseMisc; // this is content of type RICCause_t
extern const int cRICCauseRic; // this is content of type RICCause_t

typedef struct {
    uint8_t content; // See above constants
    uint8_t cause; // this is type of enum CauseRIC_t
} RICCause_t;

typedef struct {
	RICActionID_t ricActionID;
    RICCause_t ricCause;
} RICActionNotAdmittedItem_t;

typedef struct {
    uint8_t contentLength;
    RICActionNotAdmittedItem_t RICActionNotAdmittedItem[16];  // table size is const cMaxofRICactionID
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
	uint8_t iECriticality; // this is type of enum Criticality_t
	ProtocolIE_ID__t iE_ID;
	uint8_t typeOfError; // this is type of enum TypeOfError_t
	//iE-Extensions  // This has no content in E2AP ASN.1 specification
} CriticalityDiagnosticsIEListItem_t;

typedef struct {
    bool procedureCodePresent;
	ProcedureCode__t procedureCode;
	bool triggeringMessagePresent;
	uint8_t triggeringMessage; // this is type of enum TriggeringMessage_t
	bool procedureCriticalityPresent;
	uint8_t procedureCriticality; // this is type of enum Criticality_t
	bool iEsCriticalityDiagnosticsPresent;
    uint16_t criticalityDiagnosticsIELength;
	CriticalityDiagnosticsIEListItem_t criticalityDiagnosticsIEListItem[256];  // table size is const cMaxNrOfErrors
	//iE-Extensions	  // This has no content in E2AP ASN.1 specification

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
    e2err_RICEventTriggerDefinitionAllocE2SM_gNB_X2_eventTriggerDefinitionFail,
    e2err_RICSubscriptionResponseAllocRICrequestIDFail,
    e2err_RICSubscriptionResponseAllocRANfunctionIDFail,
    e2err_RICSubscriptionResponseAllocRICaction_Admitted_ItemIEsFail,
    e2err_RICSubscriptionResponseAllocRICActionAdmittedListFail,
    e2err_RICSubscriptionResponseAllocRICaction_NotAdmitted_ItemIEsFail,
    e2err_RICSubscriptionResponseAllocRICActionNotAdmittedListFail,
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
    "e2err_RICEventTriggerDefinitionAllocE2SM_gNB_X2_eventTriggerDefinitionFail",
    "e2err_RICSubscriptionResponseAllocRICrequestIDFail",
    "e2err_RICSubscriptionResponseAllocRANfunctionIDFail",
    "e2err_RICSubscriptionResponseAllocRICaction_Admitted_ItemIEsFail",
    "e2err_RICSubscriptionResponseAllocRICActionAdmittedListFail",
    "e2err_RICSubscriptionResponseAllocRICaction_NotAdmitted_ItemIEsFail",
    "e2err_RICSubscriptionResponseAllocRICActionNotAdmittedListFail",
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
    RICSubscription_t ricSubscription;
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
    RICCause_t ricCause;
    bool criticalityDiagnosticsPresent;
    CriticalityDiagnostics__t criticalityDiagnostics; // Not used in RIC currently
} RICSubscriptionDeleteFailure_t;

//////////////////////////////////////////////////////////////////////
// Function declarations

const char* getE2ErrorString(uint64_t);

typedef void* e2ap_pdu_ptr_t;

uint64_t packRICSubscriptionRequest(size_t*, byte*, char*,RICSubscriptionRequest_t*);
uint64_t packRICEventTriggerDefinition(char*,RICEventTriggerDefinition_t*);
uint64_t packRICSubscriptionResponse(size_t*, byte*, char*,RICSubscriptionResponse_t*);
uint64_t packRICSubscriptionFailure(size_t*, byte*, char*,RICSubscriptionFailure_t*);
uint64_t packRICSubscriptionDeleteRequest(size_t*, byte*, char*,RICSubscriptionDeleteRequest_t*);
uint64_t packRICSubscriptionDeleteResponse(size_t*, byte*, char*,RICSubscriptionDeleteResponse_t*);
uint64_t packRICSubscriptionDeleteFailure(size_t*, byte*, char*,RICSubscriptionDeleteFailure_t*);

e2ap_pdu_ptr_t* unpackE2AP_pdu(const size_t, const byte*, char*, E2MessageInfo_t*);
uint64_t getRICSubscriptionRequestData(e2ap_pdu_ptr_t*, RICSubscriptionRequest_t*);
uint64_t getRICEventTriggerDefinitionData(RICEventTriggerDefinition_t*);
uint64_t getRICSubscriptionResponseData(e2ap_pdu_ptr_t*, RICSubscriptionResponse_t*);
uint64_t getRICSubscriptionFailureData(e2ap_pdu_ptr_t*, RICSubscriptionFailure_t*);
uint64_t getRICSubscriptionDeleteRequestData(e2ap_pdu_ptr_t*, RICSubscriptionDeleteRequest_t*);
uint64_t getRICSubscriptionDeleteResponseData(e2ap_pdu_ptr_t*, RICSubscriptionDeleteResponse_t*);
uint64_t getRICSubscriptionDeleteFailureData(e2ap_pdu_ptr_t*, RICSubscriptionDeleteFailure_t*);

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
