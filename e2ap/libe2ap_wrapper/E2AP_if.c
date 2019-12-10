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

#include <stdio.h>
#include <stdlib.h>

#include "E2AP-PDU.h"
#include "ProtocolIE-Field.h"
#include "RICsubsequentAction.h"
#include "E2SM-gNB-X2-eventTriggerDefinition.h"
#include "E2SM-gNB-X2-indicationHeader.h"
#include "E2SM-gNB-X2-indicationMessage.h"
#include "asn_constant.h"
#include "E2AP_if.h"

#ifdef DEBUG
    static const bool debug = true;
#else
    static const bool debug = false;
#endif


const int64_t cMaxNrOfErrors = 256;

const uint64_t cMaxSizeOfOctetString = 1024;

const size_t cMacroENBIDP_20Bits = 20;
const size_t cHomeENBID_28Bits = 28;
const size_t cShortMacroENBID_18Bits = 18;
const size_t clongMacroENBIDP_21Bits = 21;

const int cRICCauseRadioNetwork = 1; // this is content of type RICCause_t
const int cRICCauseTransport = 2; // this is content of type RICCause_t
const int cRICCauseProtocol = 3; // this is content of type RICCause_t
const int cRICCauseMisc = 4; // this is content of type RICCause_t
const int cRICCauseRic = 5; // this is content of type RICCause_t

//////////////////////////////////////////////////////////////////////
// Message definitons

// Below constant values are same as in E2AP, E2SM and X2AP specs
const uint64_t cE2InitiatingMessage = 1;
const uint64_t cE2SuccessfulOutcome = 2;
const uint64_t cE2UnsuccessfulOutcome = 3;

// E2AP messages
// Initiating message
const uint64_t cRICSubscriptionRequest = 1;
const uint64_t cRICSubscriptionDeleteRequest = 2;
const uint64_t cRICIndication = 11;

// Successful outcome
const uint64_t cRICSubscriptionResponse = 1;
const uint64_t cRICsubscriptionDeleteResponse = 2;

// Unsuccessful outcome
const uint64_t cRICSubscriptionFailure = 1;
const uint64_t cRICsubscriptionDeleteFailure = 2;

typedef union {
    uint32_t  nodeID;
    uint8_t   octets[4];
} IdOctects_t;

//////////////////////////////////////////////////////////////////////
const char* getE2ErrorString(uint64_t errorCode) {

    return E2ErrorStrings[errorCode];
}

/////////////////////////////////////////////////////////////////////
bool E2encode(E2AP_PDU_t* pE2AP_PDU, size_t* dataBufferSize, byte* dataBuffer, char* pLogBuffer) {

    // Debug print
    if (debug)
        asn_fprint(stdout, &asn_DEF_E2AP_PDU, pE2AP_PDU);

    asn_enc_rval_t rval;
    rval = asn_encode_to_buffer(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2AP_PDU, pE2AP_PDU, dataBuffer, *dataBufferSize);
    if (rval.encoded == -1) {
        sprintf(pLogBuffer,"\nSerialization of %s failed.\n", asn_DEF_E2AP_PDU.name);
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return false;
    }
    else if (rval.encoded > *dataBufferSize) {
        sprintf(pLogBuffer,"\nBuffer of size %zu is too small for %s, need %zu\n",*dataBufferSize, asn_DEF_E2AP_PDU.name, rval.encoded);
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return false;
    }
    else {
        if (debug)
            sprintf(pLogBuffer,"\nSuccessfully encoded %s. Buffer size %zu, encoded size %zu\n\n",asn_DEF_E2AP_PDU.name, *dataBufferSize, rval.encoded);

        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        *dataBufferSize = rval.encoded;
        return true;
    }
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICSubscriptionRequest(size_t* pdataBufferSize, byte* pDataBuffer, char* pLogBuffer, RICSubscriptionRequest_t* pRICSubscriptionRequest) {

    E2AP_PDU_t* pE2AP_PDU = calloc(1, sizeof(E2AP_PDU_t));
    if(pE2AP_PDU)
	{
        pE2AP_PDU->present = E2AP_PDU_PR_initiatingMessage;
        pE2AP_PDU->choice.initiatingMessage.procedureCode = ProcedureCode_id_ricSubscription;
        pE2AP_PDU->choice.initiatingMessage.criticality = Criticality_ignore;
        pE2AP_PDU->choice.initiatingMessage.value.present = RICInitiatingMessage__value_PR_RICsubscriptionRequest;

        // RICrequestID
        RICsubscriptionRequest_IEs_t* pRICsubscriptionRequest_IEs = calloc(1, sizeof(RICsubscriptionRequest_IEs_t));
        if (pRICsubscriptionRequest_IEs) {
            pRICsubscriptionRequest_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionRequest_IEs->criticality = Criticality_reject;
            pRICsubscriptionRequest_IEs->value.present = RICsubscriptionRequest_IEs__value_PR_RICrequestID;
            pRICsubscriptionRequest_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionRequest->ricRequestID.ricRequestorID;
            pRICsubscriptionRequest_IEs->value.choice.RICrequestID.ricRequestSequenceNumber = pRICSubscriptionRequest->ricRequestID.ricRequestSequenceNumber;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionRequest.protocolIEs.list, pRICsubscriptionRequest_IEs);
        }
        else
            return e2err_RICSubscriptionRequestAllocRICrequestIDFail;

        // RANfunctionID
        pRICsubscriptionRequest_IEs = calloc(1, sizeof(RICsubscriptionRequest_IEs_t));
        if (pRICsubscriptionRequest_IEs) {
            pRICsubscriptionRequest_IEs->id = ProtocolIE_ID_id_RANfunctionID;
            pRICsubscriptionRequest_IEs->criticality = Criticality_reject;
            pRICsubscriptionRequest_IEs->value.present = RICsubscriptionRequest_IEs__value_PR_RANfunctionID;
            pRICsubscriptionRequest_IEs->value.choice.RANfunctionID = pRICSubscriptionRequest->ranFunctionID;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionRequest.protocolIEs.list, pRICsubscriptionRequest_IEs);
        }
        else
            return e2err_RICSubscriptionRequestAllocRANfunctionIDFail;

        // RICsubscription
        pRICsubscriptionRequest_IEs = calloc(1, sizeof(RICsubscriptionRequest_IEs_t));
        if (pRICsubscriptionRequest_IEs) {
            pRICsubscriptionRequest_IEs->id = ProtocolIE_ID_id_RICsubscription;
            pRICsubscriptionRequest_IEs->criticality = Criticality_reject;
            pRICsubscriptionRequest_IEs->value.present = RICsubscriptionRequest_IEs__value_PR_RICsubscription;

            // RICeventTriggerDefinition
            uint64_t returnCode;
            if ((returnCode = packRICEventTriggerDefinition(pLogBuffer, &pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition) != e2err_OK))
                return returnCode;

            pRICsubscriptionRequest_IEs->value.choice.RICsubscription.ricEventTriggerDefinition.buf =
              calloc(1, pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.octetString.contentLength);
            if (pRICsubscriptionRequest_IEs->value.choice.RICsubscription.ricEventTriggerDefinition.buf) {
                pRICsubscriptionRequest_IEs->value.choice.RICsubscription.ricEventTriggerDefinition.size =
                  pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.octetString.contentLength;
                memcpy(pRICsubscriptionRequest_IEs->value.choice.RICsubscription.ricEventTriggerDefinition.buf,
                       pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.octetString.data,
                       pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.octetString.contentLength);
            }
            else
                return e2err_RICSubscriptionRequestAllocRICeventTriggerDefinitionBufFail;

            // RICactions-ToBeSetup-List
            uint64_t index = 0;
            while (index < pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.contentLength && index < maxofRICactionID) {

                RICaction_ToBeSetup_ItemIEs_t* pRICaction_ToBeSetup_ItemIEs = calloc(1, sizeof(RICaction_ToBeSetup_ItemIEs_t));
                if (pRICaction_ToBeSetup_ItemIEs) {
                    pRICaction_ToBeSetup_ItemIEs->id = ProtocolIE_ID_id_RICaction_ToBeSetup_Item;
                    pRICaction_ToBeSetup_ItemIEs->criticality = Criticality_reject;
                    pRICaction_ToBeSetup_ItemIEs->value.present = RICaction_ToBeSetup_ItemIEs__value_PR_RICaction_ToBeSetup_Item;
                    // RICActionID
                    pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionID =
                      pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID;
                    // RICActionType
                    pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionType =
                      pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType;
                }
                else
                    return e2err_RICSubscriptionRequestAllocRICaction_ToBeSetup_ItemIEsFail;

                // RICactionDefinition, OPTIONAL
                  // This is not used in RIC

                // RICsubsequentAction, OPTIONAL
                RICsubsequentAction_t* pRICsubsequentAction = calloc(1, sizeof(RICsubsequentAction_t));
                if (pRICsubsequentAction) {
                    pRICsubsequentAction->ricSubsequentActionType =
                      pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType;
                    pRICsubsequentAction->ricTimeToWait =
                      pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait;
                    pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricSubsequentAction = pRICsubsequentAction;
                }
                else
                    return e2err_RICSubscriptionRequestAllocRICsubsequentActionFail;

                ASN_SEQUENCE_ADD(&pRICsubscriptionRequest_IEs->value.choice.RICsubscription.ricAction_ToBeSetup_List.list, pRICaction_ToBeSetup_ItemIEs);
                index++;
            }
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionRequest.protocolIEs.list, pRICsubscriptionRequest_IEs);
        }
        else
            return e2err_RICSubscriptionRequestAllocRICsubscriptionRequest_IEsFail;

        if (E2encode(pE2AP_PDU, pdataBufferSize, pDataBuffer, pLogBuffer))
            return e2err_OK;
        else
            return e2err_RICSubscriptionRequestEncodeFail;
    }
    return e2err_RICSubscriptionRequestAllocE2AP_PDUFail;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICEventTriggerDefinition(char* pLogBuffer, RICEventTriggerDefinition_t* pRICEventTriggerDefinition) {

    E2SM_gNB_X2_eventTriggerDefinition_t* pE2SM_gNB_X2_eventTriggerDefinition = calloc(1, sizeof(E2SM_gNB_X2_eventTriggerDefinition_t));
    if(pE2SM_gNB_X2_eventTriggerDefinition)
	{
        // RICeventTriggerDefinition
        // InterfaceID
        if ((pRICEventTriggerDefinition->interfaceID.globalENBIDPresent == true && pRICEventTriggerDefinition->interfaceID.globalGNBIDPresent == true) ||
            (pRICEventTriggerDefinition->interfaceID.globalENBIDPresent == false && pRICEventTriggerDefinition->interfaceID.globalGNBIDPresent == false))
            return e2err_RICEventTriggerDefinitionIEValueFail_1;

        // GlobalENB-ID or GlobalGNB-ID
        if (pRICEventTriggerDefinition->interfaceID.globalENBIDPresent)
        {
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.present = Interface_ID_PR_global_eNB_ID;

            // GlobalENB-ID
            // PLMN-Identity
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.size =
            pRICEventTriggerDefinition->interfaceID.globalENBID.pLMNIdentity.contentLength;
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf = calloc(1,3);
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf) {
                memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf,
                       pRICEventTriggerDefinition->interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal,
                       pRICEventTriggerDefinition->interfaceID.globalENBID.pLMNIdentity.contentLength);
            }
            else
                return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDpLMN_IdentityBufFail;

            // Add ENB-ID
            if (pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.bits == cMacroENBIDP_20Bits){
                // BIT STRING (SIZE (20)
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_macro_eNB_ID;
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf = calloc(1,3);
                if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf) {
                    pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.size = 3; // bytes
                    pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.bits_unused = 4; // trailing unused bits
                    memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf,
                           (void*)&pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.nodeID,3);
                }
                else
                    return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDmacro_eNB_IDBufFail;
            }
            else if (pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.bits == cHomeENBID_28Bits) {
                // BIT STRING (SIZE (28)
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_home_eNB_ID;
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf = calloc(1,4);
                if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf) {
                    pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.size = 4; // bytes
                    pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.bits_unused = 4; // trailing unused bits
                    memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf,
                           (void*)&pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.nodeID,4);
                }
                else
                    return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDhome_eNB_IDBufFail;
            }
            else if (pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.bits == cShortMacroENBID_18Bits) {
                // BIT STRING (SIZE(18)
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_short_Macro_eNB_ID;
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf = calloc(1,3);
                if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf) {
                    pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.size = 3;
                    pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.bits_unused = 6; // trailing unused bits
                    memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf,
                           (void*)&pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.nodeID,3);
                }
                else
                    return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDshort_Macro_eNB_IDBufFail;
            }
            else if (pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.bits == clongMacroENBIDP_21Bits) {
                // BIT STRING (SIZE(21)
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_long_Macro_eNB_ID;
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf = calloc(1,3);
                if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf) {
                    pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.size = 3; // bytes
                    pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.bits_unused = 3; // trailing unused bits
                    memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf,
                           (void*)&pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.nodeID,3);
                }
                else
                    return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDlong_Macro_eNB_IDBufFail;
            }
            else
                return e2err_RICEventTriggerDefinitionIEValueFail_2;

        }
        else if (pRICEventTriggerDefinition->interfaceID.globalGNBIDPresent) {
            // GlobalGNB-ID
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.present = Interface_ID_PR_global_gNB_ID;

            // PLMN-Identity
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.pLMN_Identity.size =
              pRICEventTriggerDefinition->interfaceID.globalGNBID.pLMNIdentity.contentLength;
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.pLMN_Identity.buf =
              calloc(1,pRICEventTriggerDefinition->interfaceID.globalGNBID.pLMNIdentity.contentLength);
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.pLMN_Identity.buf) {
                memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.pLMN_Identity.buf,
                       (void*)&pRICEventTriggerDefinition->interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal,
                        pRICEventTriggerDefinition->interfaceID.globalGNBID.pLMNIdentity.contentLength);
            }
            else
                return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_gNB_IDpLMN_IdentityBufFail;

            // GNB-ID, BIT STRING (SIZE (22..32)
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.size = 4;  //32bits
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf = calloc(1, 4);
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf) {
                memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf,
                       (void*)&pRICEventTriggerDefinition->interfaceID.globalGNBID,4); //32bits
            }
            else
                return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_gNB_IDgNB_IDBufFail;
        }
        else
            return e2err_RICEventTriggerDefinitionIEValueFail_3;

        // InterfaceDirection
        pE2SM_gNB_X2_eventTriggerDefinition->interfaceDirection = pRICEventTriggerDefinition->interfaceDirection;

        // InterfaceMessageType
        // ProcedureCode
        pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.procedureCode = pRICEventTriggerDefinition->interfaceMessageType.procedureCode;

        // TypeOfMessage
        if(pRICEventTriggerDefinition->interfaceMessageType.typeOfMessage == cE2InitiatingMessage)
            pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage = TypeOfMessage_initiating_message;
        else if(pRICEventTriggerDefinition->interfaceMessageType.typeOfMessage == cE2SuccessfulOutcome)
            pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage = TypeOfMessage_successful_outcome;
        else if(pRICEventTriggerDefinition->interfaceMessageType.typeOfMessage == cE2UnsuccessfulOutcome)
            pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage = TypeOfMessage_unsuccessful_outcome;
        else
            return e2err_RICEventTriggerDefinitionIEValueFail_4;

        // InterfaceProtocolIE-List, OPTIONAL

        // Debug print
        if (debug)
            asn_fprint(stdout, &asn_DEF_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);

        // Encode
        size_t bufferSize = sizeof(pRICEventTriggerDefinition->octetString.data);
        asn_enc_rval_t rval;
        rval = asn_encode_to_buffer(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition,
                                    pRICEventTriggerDefinition->octetString.data, bufferSize);
        if(rval.encoded == -1)
        {
            sprintf(pLogBuffer,"\nSerialization of %s failed.\n", asn_DEF_E2SM_gNB_X2_eventTriggerDefinition.name);
            return e2err_RICEventTriggerDefinitionPackFail_1;
        }
        else if(rval.encoded > bufferSize)
        {
           sprintf(pLogBuffer,"\nBuffer of size %zu is too small for %s, need %zu\n",bufferSize, asn_DEF_E2SM_gNB_X2_eventTriggerDefinition.name, rval.encoded);
            return e2err_RICEventTriggerDefinitionPackFail_2;
        }
        else
        if (debug)
               sprintf(pLogBuffer,"\nSuccessfully encoded %s. Buffer size %zu, encoded size %zu\n\n",asn_DEF_E2SM_gNB_X2_eventTriggerDefinition.name, bufferSize, rval.encoded);

        ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);

        pRICEventTriggerDefinition->octetString.contentLength = rval.encoded;
        return e2err_OK;
    }
    return e2err_RICEventTriggerDefinitionAllocE2SM_gNB_X2_eventTriggerDefinitionFail;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICSubscriptionResponse(size_t* pDataBufferSize, byte* pDataBuffer, char* pLogBuffer, RICSubscriptionResponse_t* pRICSubscriptionResponse) {

    E2AP_PDU_t* pE2AP_PDU = calloc(1, sizeof(E2AP_PDU_t));
    if(pE2AP_PDU)
	{
        pE2AP_PDU->present = E2AP_PDU_PR_successfulOutcome;
        pE2AP_PDU->choice.initiatingMessage.procedureCode = ProcedureCode_id_ricSubscription;
        pE2AP_PDU->choice.initiatingMessage.criticality = Criticality_ignore;
        pE2AP_PDU->choice.initiatingMessage.value.present = RICSuccessfulOutcome__value_PR_RICsubscriptionResponse;

        // RICrequestID
        RICsubscriptionResponse_IEs_t* pRICsubscriptionResponse_IEs = calloc(1, sizeof(RICsubscriptionResponse_IEs_t));
        if (pRICsubscriptionResponse_IEs) {
            pRICsubscriptionResponse_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionResponse_IEs->criticality = Criticality_reject;
            pRICsubscriptionResponse_IEs->value.present = RICsubscriptionResponse_IEs__value_PR_RICrequestID;
            pRICsubscriptionResponse_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionResponse->ricRequestID.ricRequestorID;
            pRICsubscriptionResponse_IEs->value.choice.RICrequestID.ricRequestSequenceNumber = pRICSubscriptionResponse->ricRequestID.ricRequestSequenceNumber;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list, pRICsubscriptionResponse_IEs);
        }
        else
            return e2err_RICSubscriptionResponseAllocRICrequestIDFail;

        // RANfunctionID
        pRICsubscriptionResponse_IEs = calloc(1, sizeof(RICsubscriptionResponse_IEs_t));
        if (pRICsubscriptionResponse_IEs) {
            pRICsubscriptionResponse_IEs->id = ProtocolIE_ID_id_RANfunctionID;
            pRICsubscriptionResponse_IEs->criticality = Criticality_reject;
            pRICsubscriptionResponse_IEs->value.present = RICsubscriptionResponse_IEs__value_PR_RANfunctionID;
            pRICsubscriptionResponse_IEs->value.choice.RANfunctionID = pRICSubscriptionResponse->ranFunctionID;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list, pRICsubscriptionResponse_IEs);
        }
        else
            return e2err_RICSubscriptionResponseAllocRANfunctionIDFail;

        // RICaction-Admitted list
        pRICsubscriptionResponse_IEs = calloc(1, sizeof(RICsubscriptionResponse_IEs_t));
        if (pRICsubscriptionResponse_IEs) {
            pRICsubscriptionResponse_IEs->id = ProtocolIE_ID_id_RICactions_Admitted;
            pRICsubscriptionResponse_IEs->criticality = Criticality_reject;
            pRICsubscriptionResponse_IEs->value.present = RICsubscriptionResponse_IEs__value_PR_RICaction_Admitted_List;

            uint64_t index = 0;
            while (index < pRICSubscriptionResponse->ricActionAdmittedList.contentLength && index < maxofRICactionID) {

                RICaction_Admitted_ItemIEs_t* pRICaction_Admitted_ItemIEs = calloc(1, sizeof (RICaction_Admitted_ItemIEs_t));
                if (pRICaction_Admitted_ItemIEs)
                {
                    pRICaction_Admitted_ItemIEs->id = ProtocolIE_ID_id_RICaction_Admitted_Item;
                    pRICaction_Admitted_ItemIEs->criticality = Criticality_reject;
                    pRICaction_Admitted_ItemIEs->value.present = RICaction_Admitted_ItemIEs__value_PR_RICaction_Admitted_Item;

                    // RICActionID
                    pRICaction_Admitted_ItemIEs->value.choice.RICaction_Admitted_Item.ricActionID = pRICSubscriptionResponse->ricActionAdmittedList.ricActionID[index];
                    ASN_SEQUENCE_ADD(&pRICsubscriptionResponse_IEs->value.choice.RICaction_Admitted_List.list, pRICaction_Admitted_ItemIEs);
                }
                else
                    return e2err_RICSubscriptionResponseAllocRICaction_Admitted_ItemIEsFail;
                index++;
            }
        }
        else
            return e2err_RICSubscriptionResponseAllocRICActionAdmittedListFail;

        ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list, pRICsubscriptionResponse_IEs);

        // RICaction-NotAdmitted list
        if (pRICSubscriptionResponse->ricActionNotAdmittedListPresent) {
            pRICsubscriptionResponse_IEs = calloc(1, sizeof(RICsubscriptionResponse_IEs_t));
            if (pRICsubscriptionResponse_IEs) {
                pRICsubscriptionResponse_IEs->id = ProtocolIE_ID_id_RICactions_NotAdmitted;
                pRICsubscriptionResponse_IEs->criticality = Criticality_reject;
                pRICsubscriptionResponse_IEs->value.present = RICsubscriptionResponse_IEs__value_PR_RICaction_NotAdmitted_List;

                uint64_t index = 0;
                while (index < pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength && index < maxofRICactionID) {

                    RICaction_NotAdmitted_ItemIEs_t* pRICaction_NotAdmitted_ItemIEs = calloc(1, sizeof (RICaction_NotAdmitted_ItemIEs_t));
                    if (pRICaction_NotAdmitted_ItemIEs)
                    {
                        pRICaction_NotAdmitted_ItemIEs->id = ProtocolIE_ID_id_RICaction_NotAdmitted_Item;
                        pRICaction_NotAdmitted_ItemIEs->criticality = Criticality_reject;
                        pRICaction_NotAdmitted_ItemIEs->value.present = RICaction_NotAdmitted_ItemIEs__value_PR_RICaction_NotAdmitted_Item;

                        // RICActionID
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricActionID =
                          pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID;

                        // RICCause
                        if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content == RICcause_PR_radioNetwork) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present = RICcause_PR_radioNetwork;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.radioNetwork =
                              pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause;
                        }
                        else if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content == RICcause_PR_transport) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present = RICcause_PR_transport;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.transport =
                              pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause;
                        }
                        else if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content == RICcause_PR_protocol) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present = RICcause_PR_protocol;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.protocol =
                              pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause;
                        }
                        else if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content == RICcause_PR_misc) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present = RICcause_PR_misc;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.misc =
                              pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause;
                        }
                        else if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content == RICcause_PR_ric) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present = RICcause_PR_ric;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.ric =
                              pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause;
                        }
                        ASN_SEQUENCE_ADD(&pRICsubscriptionResponse_IEs->value.choice.RICaction_NotAdmitted_List.list, pRICaction_NotAdmitted_ItemIEs);
                    }
                    else
                        return e2err_RICSubscriptionResponseAllocRICaction_NotAdmitted_ItemIEsFail;
                    index++;
                }
            }
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list, pRICsubscriptionResponse_IEs);
        }
        else
            return e2err_RICSubscriptionResponseAllocRICActionNotAdmittedListFail;

        if (E2encode(pE2AP_PDU, pDataBufferSize, pDataBuffer, pLogBuffer))
            return e2err_OK;
        else
            return e2err_RICSubscriptionResponseEncodeFail;
    }
    return e2err_RICSubscriptionResponseAllocE2AP_PDUFail;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICSubscriptionFailure(size_t* pDataBufferSize, byte* pDataBuffer, char* pLogBuffer, RICSubscriptionFailure_t* pRICSubscriptionFailure) {

    E2AP_PDU_t* pE2AP_PDU = calloc(1, sizeof(E2AP_PDU_t));
    if(pE2AP_PDU)
	{
        pE2AP_PDU->present = E2AP_PDU_PR_unsuccessfulOutcome;
        pE2AP_PDU->choice.unsuccessfulOutcome.procedureCode = ProcedureCode_id_ricSubscription;
        pE2AP_PDU->choice.unsuccessfulOutcome.criticality = Criticality_ignore;
        pE2AP_PDU->choice.unsuccessfulOutcome.value.present = RICUnsuccessfulOutcome__value_PR_RICsubscriptionFailure;

        // RICrequestID
        RICsubscriptionFailure_IEs_t* pRICsubscriptionFailure_IEs = calloc(1, sizeof(RICsubscriptionFailure_IEs_t));
        if (pRICsubscriptionFailure_IEs) {
            pRICsubscriptionFailure_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionFailure_IEs->criticality = Criticality_reject;
            pRICsubscriptionFailure_IEs->value.present = RICsubscriptionFailure_IEs__value_PR_RICrequestID;
            pRICsubscriptionFailure_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionFailure->ricRequestID.ricRequestorID;
            pRICsubscriptionFailure_IEs->value.choice.RICrequestID.ricRequestSequenceNumber = pRICSubscriptionFailure->ricRequestID.ricRequestSequenceNumber;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure.protocolIEs.list, pRICsubscriptionFailure_IEs);
        }
        else
            return e2err_RICSubscriptionFailureAllocRICrequestIDFail;

        // RANfunctionID
        pRICsubscriptionFailure_IEs = calloc(1, sizeof(RICsubscriptionFailure_IEs_t));
        if (pRICsubscriptionFailure_IEs) {
            pRICsubscriptionFailure_IEs->id = ProtocolIE_ID_id_RANfunctionID;
            pRICsubscriptionFailure_IEs->criticality = Criticality_reject;
            pRICsubscriptionFailure_IEs->value.present = RICsubscriptionFailure_IEs__value_PR_RANfunctionID;
            pRICsubscriptionFailure_IEs->value.choice.RANfunctionID = pRICSubscriptionFailure->ranFunctionID;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure.protocolIEs.list, pRICsubscriptionFailure_IEs);
        }
        else
            return e2err_RICSubscriptionFailureAllocRANfunctionIDFail;

        // RICaction-NotAdmitted list
        pRICsubscriptionFailure_IEs = calloc(1, sizeof(RICsubscriptionFailure_IEs_t));
        if (pRICsubscriptionFailure_IEs) {
            pRICsubscriptionFailure_IEs->id = ProtocolIE_ID_id_RICactions_NotAdmitted;
            pRICsubscriptionFailure_IEs->criticality = Criticality_reject;
            pRICsubscriptionFailure_IEs->value.present = RICsubscriptionFailure_IEs__value_PR_RICaction_NotAdmitted_List;

            uint64_t index = 0;
            while (index < pRICSubscriptionFailure->ricActionNotAdmittedList.contentLength && index < maxofRICactionID) {

                RICaction_NotAdmitted_ItemIEs_t* pRICaction_NotAdmitted_ItemIEs = calloc(1, sizeof (RICaction_NotAdmitted_ItemIEs_t));
                if (pRICaction_NotAdmitted_ItemIEs)
                {
                    pRICaction_NotAdmitted_ItemIEs->id = ProtocolIE_ID_id_RICaction_NotAdmitted_Item;
                    pRICaction_NotAdmitted_ItemIEs->criticality = Criticality_reject;
                    pRICaction_NotAdmitted_ItemIEs->value.present = RICaction_NotAdmitted_ItemIEs__value_PR_RICaction_NotAdmitted_Item;

                    // RICActionID
                    pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricActionID =
                      pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID;

                    // RICCause
                    if (pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content == RICcause_PR_radioNetwork) {
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present = RICcause_PR_radioNetwork;
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.radioNetwork =
                          pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause;
                    }
                    else if (pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content == RICcause_PR_transport) {
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present = RICcause_PR_transport;
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.transport =
                          pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause;
                    }
                    else if (pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content == RICcause_PR_protocol) {
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present = RICcause_PR_protocol;
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.protocol =
                          pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause;
                    }
                    else if (pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content == RICcause_PR_misc) {
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present = RICcause_PR_misc;
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.misc =
                          pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause;
                    }
                    else if (pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content == RICcause_PR_ric) {
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present = RICcause_PR_ric;
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.ric =
                          pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause;
                    }
                    ASN_SEQUENCE_ADD(&pRICsubscriptionFailure_IEs->value.choice.RICaction_NotAdmitted_List.list, pRICaction_NotAdmitted_ItemIEs);
                }
                else
                    return e2err_RICSubscriptionFailureAllocRICaction_NotAdmitted_ItemIEsFail;
                index++;
            }
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure.protocolIEs.list, pRICsubscriptionFailure_IEs);
        }
        else
            return e2err_RICSubscriptionFailureAllocRICActionAdmittedListFail;

        // CriticalityDiagnostics, OPTIONAL. Not used in RIC

        if (E2encode(pE2AP_PDU, pDataBufferSize, pDataBuffer, pLogBuffer))
            return e2err_OK;
        else
            return e2err_RICSubscriptionFailureEncodeFail;
    }
    else
        return e2err_RICSubscriptionFailureAllocE2AP_PDUFail;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICIndication(size_t* pDataBufferSize, byte* pDataBuffer, char* pLogBuffer, RICIndication_t* pRICIndication) {

    E2AP_PDU_t* pE2AP_PDU = calloc(1, sizeof(E2AP_PDU_t));
    if(pE2AP_PDU)
	{
        pE2AP_PDU->present = E2AP_PDU_PR_initiatingMessage;
        pE2AP_PDU->choice.initiatingMessage.procedureCode = ProcedureCode_id_ricIndication;
        pE2AP_PDU->choice.initiatingMessage.criticality = Criticality_ignore;
        pE2AP_PDU->choice.initiatingMessage.value.present = RICInitiatingMessage__value_PR_RICindication;

        // RICrequestID
        RICindication_IEs_t* pRICindication_IEs = calloc(1, sizeof(RICindication_IEs_t));
        if (pRICindication_IEs) {
            pRICindication_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICindication_IEs->criticality = Criticality_reject;
            pRICindication_IEs->value.present = RICindication_IEs__value_PR_RICrequestID;
            pRICindication_IEs->value.choice.RICrequestID.ricRequestorID = pRICIndication->ricRequestID.ricRequestorID;
            pRICindication_IEs->value.choice.RICrequestID.ricRequestSequenceNumber = pRICIndication->ricRequestID.ricRequestSequenceNumber;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list, pRICindication_IEs);
        }
        else
            return e2err_RICIndicationRICrequestIDFail;

        // RANfunctionID
        pRICindication_IEs = calloc(1, sizeof(RICindication_IEs_t));
        if (pRICindication_IEs) {
            pRICindication_IEs->id = ProtocolIE_ID_id_RANfunctionID;
            pRICindication_IEs->criticality = Criticality_reject;
            pRICindication_IEs->value.present = RICindication_IEs__value_PR_RANfunctionID;
            pRICindication_IEs->value.choice.RANfunctionID = pRICIndication->ranFunctionID;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list, pRICindication_IEs);
        }
        else
            return e2err_RICIndicationAllocRANfunctionIDFail;

        // RICactionID
        pRICindication_IEs = calloc(1, sizeof(RICindication_IEs_t));
        if (pRICindication_IEs) {
            pRICindication_IEs->id = ProtocolIE_ID_id_RICactionID;
            pRICindication_IEs->criticality = Criticality_reject;
            pRICindication_IEs->value.present = RICindication_IEs__value_PR_RICactionID;
            pRICindication_IEs->value.choice.RICactionID = pRICIndication->ricActionID;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list, pRICindication_IEs);
        }
        else
            return e2err_RICIndicationAllocRICactionIDFail;

        // RICindicationSN
        pRICindication_IEs = calloc(1, sizeof(RICindication_IEs_t));
        if (pRICindication_IEs) {
            pRICindication_IEs->id = ProtocolIE_ID_id_RICindicationSN;
            pRICindication_IEs->criticality = Criticality_reject;
            pRICindication_IEs->value.present = RICindication_IEs__value_PR_RICindicationSN;
            pRICindication_IEs->value.choice.RICindicationSN = pRICIndication->ricIndicationSN;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list, pRICindication_IEs);
        }
        else
            return e2err_RICIndicationAllocRICindicationSNFail;

        // RICindicationType
        pRICindication_IEs = calloc(1, sizeof(RICindication_IEs_t));
        if (pRICindication_IEs) {
            pRICindication_IEs->id = ProtocolIE_ID_id_RICindicationType;
            pRICindication_IEs->criticality = Criticality_reject;
            pRICindication_IEs->value.present = RICindication_IEs__value_PR_RICindicationType;
            pRICindication_IEs->value.choice.RICindicationType = pRICIndication->ricIndicationType;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list, pRICindication_IEs);
        }
        else
            return e2err_RICIndicationAllocRICindicationTypeFail;

        // RICindicationHeader
        uint64_t returnCode;
        uint64_t logBufferSize = 512;
        char logBuffer[logBufferSize];
        if ((returnCode = packRICIndicationHeader(logBuffer, &pRICIndication->ricIndicationHeader)) != e2err_OK) {
            return returnCode;
        }

        pRICindication_IEs = calloc(1, sizeof(RICindication_IEs_t));
        if (pRICindication_IEs) {
            pRICindication_IEs->id = ProtocolIE_ID_id_RICindicationHeader;
            pRICindication_IEs->criticality = Criticality_reject;
            pRICindication_IEs->value.present = RICindication_IEs__value_PR_RICindicationHeader;
            pRICindication_IEs->value.choice.RICindicationHeader.buf = calloc(1,pRICIndication->ricIndicationHeader.octetString.contentLength);
            if (pRICindication_IEs->value.choice.RICindicationHeader.buf) {
                pRICindication_IEs->value.choice.RICindicationHeader.size = pRICIndication->ricIndicationHeader.octetString.contentLength;
                memcpy(pRICindication_IEs->value.choice.RICindicationHeader.buf,pRICIndication->ricIndicationHeader.octetString.data,
                    pRICIndication->ricIndicationHeader.octetString.contentLength);
                ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list, pRICindication_IEs);
            }
            else
                return e2err_RICIndicationAllocRRICindicationHeaderBufFail;
        }
        else
            return e2err_RICIndicationAllocRICindicationHeaderFail;

        // RICindicationMessage
        if ((returnCode = packRICIndicationMessage(logBuffer, &pRICIndication->ricIndicationMessage)) != e2err_OK) {
            return returnCode;
        }

        pRICindication_IEs = calloc(1, sizeof(RICindication_IEs_t));
        if (pRICindication_IEs) {
            pRICindication_IEs->id = ProtocolIE_ID_id_RICindicationMessage;
            pRICindication_IEs->criticality = Criticality_reject;
            pRICindication_IEs->value.present = RICindication_IEs__value_PR_RICindicationMessage;
            pRICindication_IEs->value.choice.RICindicationMessage.buf = calloc(1,pRICIndication->ricIndicationMessage.octetString.contentLength);
            if (pRICindication_IEs->value.choice.RICindicationMessage.buf) {
                pRICindication_IEs->value.choice.RICindicationMessage.size = pRICIndication->ricIndicationMessage.octetString.contentLength;
                memcpy(pRICindication_IEs->value.choice.RICindicationHeader.buf,pRICIndication->ricIndicationMessage.octetString.data,
                       pRICIndication->ricIndicationMessage.octetString.contentLength);
                ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list, pRICindication_IEs);
            }
            else
                return e2err_RICIndicationAllocRICindicationMessageBufFail;
        }
        else
            return e2err_RICIndicationAllocRICindicationMessageFail;

        // RICcallProcessID, OPTIONAL. Not used in RIC.

        if (E2encode(pE2AP_PDU, pDataBufferSize, pDataBuffer, pLogBuffer))
            return e2err_OK;
        else
            return e2err_RICIndicationEncodeFail;
    }
    else
        return e2err_RICIndicationAllocE2AP_PDUFail;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICIndicationHeader(char* pLogBuffer, RICIndicationHeader_t* pRICIndicationHeader) {

    E2SM_gNB_X2_indicationHeader_t* pE2SM_gNB_X2_indicationHeader = calloc(1, sizeof(E2SM_gNB_X2_indicationHeader_t));
    if(pE2SM_gNB_X2_indicationHeader)
	{
        // InterfaceID
        if ((pRICIndicationHeader->interfaceID.globalENBIDPresent == true && pRICIndicationHeader->interfaceID.globalGNBIDPresent == true) ||
            (pRICIndicationHeader->interfaceID.globalENBIDPresent == false && pRICIndicationHeader->interfaceID.globalGNBIDPresent == false))
            return e2err_RICindicationHeaderIEValueFail_1;

        // GlobalENB-ID or GlobalGNB-ID
        if (pRICIndicationHeader->interfaceID.globalENBIDPresent)
        {
            pE2SM_gNB_X2_indicationHeader->interface_ID.present = Interface_ID_PR_global_eNB_ID;

            // GlobalENB-ID
            // PLMN-Identity
            pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.pLMN_Identity.size =
            pRICIndicationHeader->interfaceID.globalENBID.pLMNIdentity.contentLength;
            pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf = calloc(1,3);
            if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf) {
                memcpy(pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf,
                       pRICIndicationHeader->interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal,
                       pRICIndicationHeader->interfaceID.globalENBID.pLMNIdentity.contentLength);
            }
            else
                return e2err_RICIndicationAllocRICIndicationHeaderglobal_eNB_IDpLMN_IdentityBufFail;

            // Add ENB-ID
            if (pRICIndicationHeader->interfaceID.globalENBID.nodeID.bits == cMacroENBIDP_20Bits){
                // BIT STRING (SIZE (20)
                pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_macro_eNB_ID;

                pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf = calloc(1,3);
                if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf) {
                    pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.size = 3; // bytes
                    pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.bits_unused = 4; // trailing unused bits
                    memcpy(pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf,
                           (void*)&pRICIndicationHeader->interfaceID.globalENBID.nodeID.nodeID,3);
                }
                else
                    return e2err_RICIndicationAllocRICIndicationHeaderglobal_eNB_IDeNB_IDmacro_eNB_IDBufFail;
            }
            else if (pRICIndicationHeader->interfaceID.globalENBID.nodeID.bits == cHomeENBID_28Bits) {
                // BIT STRING (SIZE (28)
                pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_home_eNB_ID;

                pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf = calloc(1,4);
                if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf) {
                    pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.size = 4; // bytes
                    pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.bits_unused = 4; // trailing unused bits
                    memcpy(pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf,
                           (void*)&pRICIndicationHeader->interfaceID.globalENBID.nodeID.nodeID,4);
                }
                else
                    return e2err_RICIndicationAllocRICIndicationHeaderglobal_eNB_IDeNB_IDhome_eNB_IDBufFail;
            }
            else if (pRICIndicationHeader->interfaceID.globalENBID.nodeID.bits == cShortMacroENBID_18Bits) {
                // BIT STRING (SIZE(18)
                pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_short_Macro_eNB_ID;

                pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf = calloc(1,3);
                if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf) {
                    pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.size = 3;
                    pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.bits_unused = 6; // trailing unused bits
                     memcpy(pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf,
                            (void*)&pRICIndicationHeader->interfaceID.globalENBID.nodeID.nodeID,3);
                }
                else
                    return e2err_RICIndicationAllocRICIndicationHeaderglobal_eNB_IDeNB_IDshort_Macro_eNB_IDBufFail;
            }
            else if (pRICIndicationHeader->interfaceID.globalENBID.nodeID.bits == clongMacroENBIDP_21Bits) {
                // BIT STRING (SIZE(21)
                pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_long_Macro_eNB_ID;

                pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf = calloc(1,3);
                if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf) {
                    pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.size = 3; // bytes
                    pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.bits_unused = 3; // trailing unused bits
                    memcpy(pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf,
                           (void*)&pRICIndicationHeader->interfaceID.globalENBID.nodeID.nodeID,3);
                }
                else
                    return e2err_RICIndicationAllocRICIndicationHeaderglobal_eNB_IDeNB_IDlong_Macro_eNB_IDBufFail;
            }
            else
                return e2err_RICindicationHeaderIEValueFail_2;

        }
        else if (pRICIndicationHeader->interfaceID.globalGNBIDPresent) {
            // GlobalGNB-ID
            pE2SM_gNB_X2_indicationHeader->interface_ID.present = Interface_ID_PR_global_gNB_ID;

            // PLMN-Identity
            pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.pLMN_Identity.size =
              pRICIndicationHeader->interfaceID.globalGNBID.pLMNIdentity.contentLength;
            pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.pLMN_Identity.buf =
              calloc(1,pRICIndicationHeader->interfaceID.globalGNBID.pLMNIdentity.contentLength);
            if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.pLMN_Identity.buf) {
                memcpy(pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.pLMN_Identity.buf,
                       (void*)&pRICIndicationHeader->interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal,
                       pRICIndicationHeader->interfaceID.globalGNBID.pLMNIdentity.contentLength);
            }
            else
                return e2err_RICIndicationAllocRICIndicationHeaderglobal_gNB_IDpLMN_IdentityBufFail;

            // GNB-ID, BIT STRING (SIZE (22..32)
            pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.size = 4;  //32bits
            pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf = calloc(1,4);
            if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf) {
                memcpy(pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf,
                       (void*)&pRICIndicationHeader->interfaceID.globalGNBID,4); //32bits
                }
                else
                    return e2err_RICIndicationAllocRICIndicationHeaderglobal_gNB_IDgNB_IDgNB_IDBufFail;
        }
        else
            return e2err_RICindicationHeaderIEValueFail_3;

        // InterfaceDirection
        pE2SM_gNB_X2_indicationHeader->interfaceDirection = pRICIndicationHeader->interfaceDirection;

        // TimeStamp OPTIONAL. Not used in RIC.

        // Debug print
        if (debug)
            asn_fprint(stdout, &asn_DEF_E2SM_gNB_X2_indicationHeader, pE2SM_gNB_X2_indicationHeader);

        // Encode
        size_t bufferSize = sizeof(pRICIndicationHeader->octetString.data);
        asn_enc_rval_t rval;
        rval = asn_encode_to_buffer(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2SM_gNB_X2_indicationHeader, pE2SM_gNB_X2_indicationHeader,
                                    pRICIndicationHeader->octetString.data, bufferSize);
        if(rval.encoded == -1)
        {
            sprintf(pLogBuffer,"\nSerialization of %s failed.\n", asn_DEF_E2SM_gNB_X2_indicationHeader.name);
            return e2err_RICindicationHeaderPackFail_1;
        }
        else if(rval.encoded > bufferSize)
        {
            sprintf(pLogBuffer,"\nBuffer of size %zu is too small for %s, need %zu\n",bufferSize, asn_DEF_E2SM_gNB_X2_indicationHeader.name, rval.encoded);
            return e2err_RICindicationHeaderPackFail_2;
        }
        else
            if (debug)
                sprintf(pLogBuffer,"\nSuccessfully encoded %s. Buffer size %zu, encoded size %zu\n\n",asn_DEF_E2SM_gNB_X2_indicationHeader.name, bufferSize, rval.encoded);

        ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_indicationHeader, pE2SM_gNB_X2_indicationHeader);

        pRICIndicationHeader->octetString.contentLength = rval.encoded;
        return e2err_OK;
    }
    else
        return e2err_RICIndicationHeaderAllocE2AP_PDUFail;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICIndicationMessage(char* pLogBuffer, RICIndicationMessage_t* pRICIndicationMessage) {

    E2SM_gNB_X2_indicationMessage_t* pE2SM_gNB_X2_indicationMessage = calloc(1, sizeof(E2SM_gNB_X2_indicationMessage_t));
    if(pE2SM_gNB_X2_indicationMessage)
    {
        pE2SM_gNB_X2_indicationMessage->interfaceMessage.buf = calloc(1, pRICIndicationMessage->interfaceMessage.contentLength);
        if(pE2SM_gNB_X2_indicationMessage->interfaceMessage.buf)
        {
            pE2SM_gNB_X2_indicationMessage->interfaceMessage.size = pRICIndicationMessage->interfaceMessage.contentLength;
            memcpy(pE2SM_gNB_X2_indicationMessage->interfaceMessage.buf,pRICIndicationMessage->interfaceMessage.data,pRICIndicationMessage->interfaceMessage.contentLength);
        }
        else
            return e2err_RICIndicationMessageAllocinterfaceMessageFail;

        // Debug print
        if (debug)
            asn_fprint(stdout, &asn_DEF_E2SM_gNB_X2_indicationMessage, pE2SM_gNB_X2_indicationMessage);

        // Encode
        size_t bufferSize = sizeof(pRICIndicationMessage->octetString.data);
        asn_enc_rval_t rval;
        rval = asn_encode_to_buffer(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2SM_gNB_X2_indicationMessage, pE2SM_gNB_X2_indicationMessage,
                                    pRICIndicationMessage->octetString.data, bufferSize);
        if(rval.encoded == -1)
        {
            sprintf(pLogBuffer,"\nSerialization of %s failed.\n", asn_DEF_E2SM_gNB_X2_indicationMessage.name);
            return e2err_RICindicationMessagePackFail_1;
        }
        else if(rval.encoded > bufferSize)
        {
            sprintf(pLogBuffer,"\nBuffer of size %zu is too small for %s, need %zu\n",bufferSize, asn_DEF_E2SM_gNB_X2_indicationMessage.name, rval.encoded);
            return e2err_RICindicationMessagePackFail_2;
        }
        else
            if (debug)
                sprintf(pLogBuffer,"\nSuccessfully encoded %s. Buffer size %zu, encoded size %zu\n\n",asn_DEF_E2SM_gNB_X2_indicationMessage.name, bufferSize, rval.encoded);

        ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_indicationMessage, pE2SM_gNB_X2_indicationMessage);

        pRICIndicationMessage->octetString.contentLength = rval.encoded;
        return e2err_OK;
    }
    else
        return e2err_E2SM_gNB_X2_indicationMessageAllocE2AP_PDUFail;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICSubscriptionDeleteRequest(size_t* pDataBufferSize, byte* pDataBuffer, char* pLogBuffer, RICSubscriptionDeleteRequest_t* pRICSubscriptionDeleteRequest) {

    E2AP_PDU_t* pE2AP_PDU = calloc(1, sizeof(E2AP_PDU_t));
    if(pE2AP_PDU)
	{
        pE2AP_PDU->present = E2AP_PDU_PR_initiatingMessage;
        pE2AP_PDU->choice.initiatingMessage.procedureCode = ProcedureCode_id_ricSubscriptionDelete;
        pE2AP_PDU->choice.initiatingMessage.criticality = Criticality_ignore;
        pE2AP_PDU->choice.initiatingMessage.value.present = RICInitiatingMessage__value_PR_RICsubscriptionDeleteRequest;

        // RICrequestID
        RICsubscriptionDeleteRequest_IEs_t* pRICsubscriptionDeleteRequest_IEs = calloc(1, sizeof(RICsubscriptionDeleteRequest_IEs_t));
        if (pRICsubscriptionDeleteRequest_IEs) {
            pRICsubscriptionDeleteRequest_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionDeleteRequest_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteRequest_IEs->value.present = RICsubscriptionDeleteRequest_IEs__value_PR_RICrequestID;
            pRICsubscriptionDeleteRequest_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionDeleteRequest->ricRequestID.ricRequestorID;
            pRICsubscriptionDeleteRequest_IEs->value.choice.RICrequestID.ricRequestSequenceNumber = pRICSubscriptionDeleteRequest->ricRequestID.ricRequestSequenceNumber;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionDeleteRequest.protocolIEs.list, pRICsubscriptionDeleteRequest_IEs);
        }
        else
            return e2err_RICSubscriptionDeleteRequestAllocRICrequestIDFail;

        // RANfunctionID
        pRICsubscriptionDeleteRequest_IEs = calloc(1, sizeof(RICsubscriptionDeleteRequest_IEs_t));
        if (pRICsubscriptionDeleteRequest_IEs) {
            pRICsubscriptionDeleteRequest_IEs->id = ProtocolIE_ID_id_RANfunctionID;
            pRICsubscriptionDeleteRequest_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteRequest_IEs->value.present = RICsubscriptionDeleteRequest_IEs__value_PR_RANfunctionID;
            pRICsubscriptionDeleteRequest_IEs->value.choice.RANfunctionID = pRICSubscriptionDeleteRequest->ranFunctionID;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionDeleteRequest.protocolIEs.list, pRICsubscriptionDeleteRequest_IEs);
        }
        else
            return e2err_RICSubscriptionDeleteRequestAllocRANfunctionIDFail;

        if (E2encode(pE2AP_PDU, pDataBufferSize, pDataBuffer, pLogBuffer))
            return e2err_OK;
        else
            return e2err_RICSubscriptionDeleteRequestEncodeFail;
    }
    else
        return e2err_RICSubscriptionDeleteRequestAllocE2AP_PDUFail;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICSubscriptionDeleteResponse(size_t* pDataBufferSize, byte* pDataBuffer, char* pLogBuffer, RICSubscriptionDeleteResponse_t* pRICSubscriptionDeleteResponse) {

    E2AP_PDU_t* pE2AP_PDU = calloc(1, sizeof(E2AP_PDU_t));
    if(pE2AP_PDU)
	{
        pE2AP_PDU->present = E2AP_PDU_PR_successfulOutcome;
        pE2AP_PDU->choice.successfulOutcome.procedureCode = ProcedureCode_id_ricSubscriptionDelete;
        pE2AP_PDU->choice.successfulOutcome.criticality = Criticality_ignore;
        pE2AP_PDU->choice.successfulOutcome.value.present = RICSuccessfulOutcome__value_PR_RICsubscriptionDeleteResponse;

        // RICrequestID
        RICsubscriptionDeleteResponse_IEs_t* pRICsubscriptionDeleteResponse_IEs = calloc(1, sizeof(RICsubscriptionDeleteResponse_IEs_t));
        if (pRICsubscriptionDeleteResponse_IEs) {
            pRICsubscriptionDeleteResponse_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionDeleteResponse_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteResponse_IEs->value.present = RICsubscriptionDeleteResponse_IEs__value_PR_RICrequestID;
            pRICsubscriptionDeleteResponse_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionDeleteResponse->ricRequestID.ricRequestorID;
            pRICsubscriptionDeleteResponse_IEs->value.choice.RICrequestID.ricRequestSequenceNumber = pRICSubscriptionDeleteResponse->ricRequestID.ricRequestSequenceNumber;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionDeleteResponse.protocolIEs.list, pRICsubscriptionDeleteResponse_IEs);
        }
        else
            return e2err_RICSubscriptionDeleteResponseAllocRICrequestIDFail;

        // RANfunctionID
        pRICsubscriptionDeleteResponse_IEs = calloc(1, sizeof(RICsubscriptionDeleteResponse_IEs_t));
        if (pRICsubscriptionDeleteResponse_IEs) {
            pRICsubscriptionDeleteResponse_IEs->id = ProtocolIE_ID_id_RANfunctionID;
            pRICsubscriptionDeleteResponse_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteResponse_IEs->value.present = RICsubscriptionDeleteResponse_IEs__value_PR_RANfunctionID;
            pRICsubscriptionDeleteResponse_IEs->value.choice.RANfunctionID = pRICSubscriptionDeleteResponse->ranFunctionID;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionDeleteResponse.protocolIEs.list, pRICsubscriptionDeleteResponse_IEs);
        }
        else
            return e2err_RICSubscriptionDeleteResponseAllocRANfunctionIDFail;

        if (E2encode(pE2AP_PDU, pDataBufferSize, pDataBuffer, pLogBuffer))
            return e2err_OK;
        else
            return e2err_RICSubscriptionDeleteResponseEncodeFail;
    }
    else
        return e2err_RICSubscriptionDeleteResponseAllocE2AP_PDUFail;
}

uint64_t packRICSubscriptionDeleteFailure(size_t* pDataBufferSize, byte* pDataBuffer, char* pLogBuffer, RICSubscriptionDeleteFailure_t* pRICSubscriptionDeleteFailure) {

    E2AP_PDU_t* pE2AP_PDU = calloc(1, sizeof(E2AP_PDU_t));
    if(pE2AP_PDU)
	{
        pE2AP_PDU->present = E2AP_PDU_PR_unsuccessfulOutcome;
        pE2AP_PDU->choice.unsuccessfulOutcome.procedureCode = ProcedureCode_id_ricSubscriptionDelete;
        pE2AP_PDU->choice.unsuccessfulOutcome.criticality = Criticality_ignore;
        pE2AP_PDU->choice.unsuccessfulOutcome.value.present = RICUnsuccessfulOutcome__value_PR_RICsubscriptionDeleteFailure;

        // RICrequestID
        RICsubscriptionDeleteFailure_IEs_t* pRICsubscriptionDeleteFailure_IEs = calloc(1, sizeof(RICsubscriptionDeleteFailure_IEs_t));
        if (pRICsubscriptionDeleteFailure_IEs) {
            pRICsubscriptionDeleteFailure_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionDeleteFailure_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteFailure_IEs->value.present = RICsubscriptionDeleteFailure_IEs__value_PR_RICrequestID;
            pRICsubscriptionDeleteFailure_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionDeleteFailure->ricRequestID.ricRequestorID;
            pRICsubscriptionDeleteFailure_IEs->value.choice.RICrequestID.ricRequestSequenceNumber = pRICSubscriptionDeleteFailure->ricRequestID.ricRequestSequenceNumber;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionDeleteFailure.protocolIEs.list, pRICsubscriptionDeleteFailure_IEs);
        }
        else
            return e2err_RICSubscriptionDeleteFailureAllocRICrequestIDFail;

        // RANfunctionID
        pRICsubscriptionDeleteFailure_IEs = calloc(1, sizeof(RICsubscriptionDeleteFailure_IEs_t));
        if (pRICsubscriptionDeleteFailure_IEs) {
            pRICsubscriptionDeleteFailure_IEs->id = ProtocolIE_ID_id_RANfunctionID;
            pRICsubscriptionDeleteFailure_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteFailure_IEs->value.present = RICsubscriptionDeleteFailure_IEs__value_PR_RANfunctionID;
            pRICsubscriptionDeleteFailure_IEs->value.choice.RANfunctionID = pRICSubscriptionDeleteFailure->ranFunctionID;
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionDeleteFailure.protocolIEs.list, pRICsubscriptionDeleteFailure_IEs);
        }
        else
            return e2err_RICSubscriptionDeleteFailureAllocRANfunctionIDFail;

        // RICcause
        pRICsubscriptionDeleteFailure_IEs = calloc(1, sizeof(RICsubscriptionDeleteFailure_IEs_t));
        if (pRICsubscriptionDeleteFailure_IEs) {
            pRICsubscriptionDeleteFailure_IEs->id = ProtocolIE_ID_id_RICcause;
            pRICsubscriptionDeleteFailure_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteFailure_IEs->value.present = RICsubscriptionDeleteFailure_IEs__value_PR_RICcause;
            if (pRICSubscriptionDeleteFailure->ricCause.content == RICcause_PR_radioNetwork) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.present = RICcause_PR_radioNetwork;
                pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.choice.radioNetwork =
                  pRICSubscriptionDeleteFailure->ricCause.cause;
            }
            else if (pRICSubscriptionDeleteFailure->ricCause.content == RICcause_PR_transport) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.present = RICcause_PR_transport;
                pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.choice.transport =
                  pRICSubscriptionDeleteFailure->ricCause.cause;
            }
            else if (pRICSubscriptionDeleteFailure->ricCause.content == RICcause_PR_protocol) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.present = RICcause_PR_protocol;
                pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.choice.protocol =
                  pRICSubscriptionDeleteFailure->ricCause.cause;
            }
            else if (pRICSubscriptionDeleteFailure->ricCause.content == RICcause_PR_misc) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.present = RICcause_PR_misc;
                pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.choice.misc =
                  pRICSubscriptionDeleteFailure->ricCause.cause;
            }
            else if (pRICSubscriptionDeleteFailure->ricCause.content == RICcause_PR_ric) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.present = RICcause_PR_ric;
                pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.choice.ric =
                  pRICSubscriptionDeleteFailure->ricCause.cause;
            }
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionDeleteFailure.protocolIEs.list, pRICsubscriptionDeleteFailure_IEs);
        }
        else
            return e2err_RICSubscriptionDeleteFailureAllocRICcauseFail;

        // CriticalityDiagnostics, OPTIONAL

        if (E2encode(pE2AP_PDU, pDataBufferSize, pDataBuffer, pLogBuffer))
            return e2err_OK;
        else
            return e2err_RICSubscriptionDeleteFailureEncodeFail;
    }
    else
        return e2err_RICSubscriptionDeleteFailureAllocE2AP_PDUFail;
}

//////////////////////////////////////////////////////////////////////
e2ap_pdu_ptr_t* unpackE2AP_pdu(const size_t dataBufferSize, const byte* dataBuffer, char* pLogBuffer, E2MessageInfo_t* pMessageInfo) {

    E2AP_PDU_t* pE2AP_PDU = 0;
    asn_dec_rval_t rval;
    rval = asn_decode(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2AP_PDU, (void **)&pE2AP_PDU, dataBuffer, dataBufferSize);
    switch (rval.code) {
    case RC_OK:
        // Debug print
        if (debug) {
            sprintf(pLogBuffer,"\nSuccessfully decoded E2AP-PDU\n\n");
            asn_fprint(stdout, &asn_DEF_E2AP_PDU, pE2AP_PDU);
        }

        if (pE2AP_PDU->present == E2AP_PDU_PR_initiatingMessage) {
            if (pE2AP_PDU->choice.initiatingMessage.procedureCode == ProcedureCode_id_ricSubscription) {
                if (pE2AP_PDU->choice.initiatingMessage.value.present == RICInitiatingMessage__value_PR_RICsubscriptionRequest) {
                    pMessageInfo->messageType = cE2InitiatingMessage;
                    pMessageInfo->messageId = cRICSubscriptionRequest;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"\nError. Not supported initiatingMessage MessageId = %u\n\n",pE2AP_PDU->choice.initiatingMessage.value.present);
                    return 0;
                }
            }
            else if (pE2AP_PDU->choice.initiatingMessage.procedureCode == ProcedureCode_id_ricIndication) {
                if (pE2AP_PDU->choice.initiatingMessage.value.present == RICInitiatingMessage__value_PR_RICindication) {
                    pMessageInfo->messageType = cE2InitiatingMessage;
                    pMessageInfo->messageId = cRICIndication;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"\nError. Not supported initiatingMessage MessageId = %u\n\n",pE2AP_PDU->choice.initiatingMessage.value.present);
                    return 0;
                }
            }
            else if (pE2AP_PDU->choice.initiatingMessage.procedureCode == ProcedureCode_id_ricSubscriptionDelete) {
                if (pE2AP_PDU->choice.initiatingMessage.value.present == RICInitiatingMessage__value_PR_RICsubscriptionDeleteRequest) {
                    pMessageInfo->messageType = cE2InitiatingMessage;
                    pMessageInfo->messageId = cRICSubscriptionDeleteRequest;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"\nError. Not supported initiatingMessage MessageId = %u\n\n",pE2AP_PDU->choice.initiatingMessage.value.present);
                    return 0;
                }
            }
            else {
                sprintf(pLogBuffer,"\nError. Procedure not supported. ProcedureCode = %li\n\n",pE2AP_PDU->choice.initiatingMessage.procedureCode);
                return 0;
            }
        }
        else if (pE2AP_PDU->present == E2AP_PDU_PR_successfulOutcome) {
            if (pE2AP_PDU->choice.successfulOutcome.procedureCode == ProcedureCode_id_ricSubscription) {
                if (pE2AP_PDU->choice.successfulOutcome.value.present == RICSuccessfulOutcome__value_PR_RICsubscriptionResponse) {
                    pMessageInfo->messageType = cE2SuccessfulOutcome;
                    pMessageInfo->messageId = cRICSubscriptionResponse;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"\nError. Not supported successfulOutcome MessageId = %u\n\n",pE2AP_PDU->choice.successfulOutcome.value.present);
                    return 0;
                }
            }
            else if (pE2AP_PDU->choice.successfulOutcome.procedureCode == ProcedureCode_id_ricSubscriptionDelete) {
                if (pE2AP_PDU->choice.successfulOutcome.value.present == RICSuccessfulOutcome__value_PR_RICsubscriptionDeleteResponse) {
                    pMessageInfo->messageType = cE2SuccessfulOutcome;
                    pMessageInfo->messageId = cRICsubscriptionDeleteResponse;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"\nError. Not supported successfulOutcome MessageId = %u\n\n",pE2AP_PDU->choice.successfulOutcome.value.present);
                    return 0;
                }
            }
            else {
                sprintf(pLogBuffer,"\nError. Procedure not supported. ProcedureCode = %li\n\n",pE2AP_PDU->choice.successfulOutcome.procedureCode);
                return 0;
            }
        }
        else if (pE2AP_PDU->present == E2AP_PDU_PR_unsuccessfulOutcome) {
            if (pE2AP_PDU->choice.unsuccessfulOutcome.procedureCode == ProcedureCode_id_ricSubscription) {
                if (pE2AP_PDU->choice.unsuccessfulOutcome.value.present == RICUnsuccessfulOutcome__value_PR_RICsubscriptionFailure) {
                    pMessageInfo->messageType = cE2UnsuccessfulOutcome;
                    pMessageInfo->messageId = cRICSubscriptionFailure;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"\nError. Not supported unsuccessfulOutcome MessageId = %u\n\n",pE2AP_PDU->choice.unsuccessfulOutcome.value.present);
                    return 0;
                }
            }
            else if (pE2AP_PDU->choice.unsuccessfulOutcome.procedureCode == ProcedureCode_id_ricSubscriptionDelete) {
                if (pE2AP_PDU->choice.unsuccessfulOutcome.value.present == RICUnsuccessfulOutcome__value_PR_RICsubscriptionDeleteFailure) {
                    pMessageInfo->messageType = cE2UnsuccessfulOutcome;
                    pMessageInfo->messageId = cRICsubscriptionDeleteFailure;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"\nError. Not supported unsuccessfulOutcome MessageId = %u\n\n",pE2AP_PDU->choice.unsuccessfulOutcome.value.present);
                    return 0;
                }
            }
        }
        else
            sprintf(pLogBuffer,"\nDecode failed. Invalid message type %u\n",pE2AP_PDU->present);
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return 0;
    case RC_WMORE:
        sprintf(pLogBuffer,"\nDecode failed. More data needed. Buffer size %zu, %s, consumed %zu\n",dataBufferSize, asn_DEF_E2AP_PDU.name, rval.consumed);
        return 0;
    case RC_FAIL:
        sprintf(pLogBuffer,"\nDecode failed. Buffer size %zu, %s, consumed %zu\n",dataBufferSize, asn_DEF_E2AP_PDU.name, rval.consumed);
        return 0;
    default:
        return 0;
    }
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICSubscriptionRequestData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionRequest_t* pRICSubscriptionRequest) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;
    RICsubscriptionRequest_IEs_t* pRICsubscriptionRequest_IEs;
    // RICrequestID
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionRequest.protocolIEs.list.count > 0) {
        pRICsubscriptionRequest_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionRequest.protocolIEs.list.array[0];
        pRICSubscriptionRequest->ricRequestID.ricRequestorID = pRICsubscriptionRequest_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionRequest->ricRequestID.ricRequestSequenceNumber = pRICsubscriptionRequest_IEs->value.choice.RICrequestID.ricRequestSequenceNumber;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionRequestRICrequestIDMissing;
    }

    // RANfunctionID
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionRequest.protocolIEs.list.count > 1) {
        pRICsubscriptionRequest_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionRequest.protocolIEs.list.array[1];
        pRICSubscriptionRequest->ranFunctionID = pRICsubscriptionRequest_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionRequestRANfunctionIDMissing;
    }

    // RICsubscription
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionRequest.protocolIEs.list.count > 2) {
        pRICsubscriptionRequest_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionRequest.protocolIEs.list.array[2];

        // Unpack EventTriggerDefinition
        RICeventTriggerDefinition_t* pRICeventTriggerDefinition =
          (RICeventTriggerDefinition_t*)&pRICsubscriptionRequest_IEs->value.choice.RICsubscription.ricEventTriggerDefinition;
        pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.octetString.contentLength = pRICeventTriggerDefinition->size;
        memcpy(pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.octetString.data, pRICeventTriggerDefinition->buf, pRICeventTriggerDefinition->size); //octetstring

        uint64_t returnCode;
        if ((returnCode = getRICEventTriggerDefinitionData(&pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition) != e2err_OK)) {
            ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
            return returnCode;
        }

        // RICactions-ToBeSetup-List
        RICaction_ToBeSetup_ItemIEs_t* pRICaction_ToBeSetup_ItemIEs;
        uint64_t index = 0;
        while (index < pRICsubscriptionRequest_IEs->value.choice.RICsubscription.ricAction_ToBeSetup_List.list.count)
        {
            pRICaction_ToBeSetup_ItemIEs = (RICaction_ToBeSetup_ItemIEs_t*)pRICsubscriptionRequest_IEs->value.choice.RICsubscription.ricAction_ToBeSetup_List.list.array[index];

            // RICActionID
            pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID =
              pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionID;

            // RICActionType
            pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType =
              pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionType;

            // RICactionDefinition, OPTIONAL
            if (pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition)
            {
                pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = false;
                // not used in RIC
            }
            else
                pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = false;

            // RICsubsequentAction, OPTIONAL
            RICsubsequentAction_t* pRICsubsequentAction;
            if (pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricSubsequentAction)
            {
                pRICsubsequentAction = pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricSubsequentAction;
                pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent = true;
                pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType =
                  pRICsubsequentAction->ricSubsequentActionType;
                pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait =
                  pRICsubsequentAction->ricTimeToWait;
            }
            else
                pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent = false;
            index++;
        }
        pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.contentLength = index;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionRequestICsubscriptionMissing;
    }

    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICEventTriggerDefinitionData(RICEventTriggerDefinition_t* pRICEventTriggerDefinition) {

    E2SM_gNB_X2_eventTriggerDefinition_t* pE2SM_gNB_X2_eventTriggerDefinition = 0;
    asn_dec_rval_t rval;
    rval = asn_decode(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2SM_gNB_X2_eventTriggerDefinition, (void **)&pE2SM_gNB_X2_eventTriggerDefinition,
                      pRICEventTriggerDefinition->octetString.data, pRICEventTriggerDefinition->octetString.contentLength);
    switch(rval.code) {
    case RC_OK:
        // Debug print
        if (debug) {
            printf("\nSuccessfully decoded E2SM_gNB_X2_eventTriggerDefinition\n\n");
            asn_fprint(stdout, &asn_DEF_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
        }

        // InterfaceID, GlobalENB-ID or GlobalGNB-ID
        if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.present == Interface_ID_PR_global_eNB_ID) {

            // GlobalENB-ID
            pRICEventTriggerDefinition->interfaceID.globalENBIDPresent = true;

            // PLMN-Identity
            pRICEventTriggerDefinition->interfaceID.globalENBID.pLMNIdentity.contentLength =
              pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.size;
            memcpy(pRICEventTriggerDefinition->interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal,
              pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf,
              pRICEventTriggerDefinition->interfaceID.globalENBID.pLMNIdentity.contentLength);

            //  ENB-ID
            IdOctects_t eNBOctects;
            memset(eNBOctects.octets, 0, sizeof(eNBOctects));
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present == ENB_ID_PR_macro_eNB_ID) {
                // BIT STRING (SIZE (20)
                pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.bits = cMacroENBIDP_20Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf,
                  pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.size);
                pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present == ENB_ID_PR_home_eNB_ID) {
                // BIT STRING (SIZE (28)
                pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.bits = cHomeENBID_28Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf,
                  pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.size);
                pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present == ENB_ID_PR_short_Macro_eNB_ID) {
                // BIT STRING (SIZE(18)
                pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.bits = cShortMacroENBID_18Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf,
                  pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.size);
                pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present == ENB_ID_PR_long_Macro_eNB_ID) {
                // BIT STRING (SIZE(21)
                pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.bits =  clongMacroENBIDP_21Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf,
                  pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.size);
                pRICEventTriggerDefinition->interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else {
                pRICEventTriggerDefinition->interfaceID.globalENBIDPresent = false;
                pRICEventTriggerDefinition->interfaceID.globalGNBIDPresent = false;
                ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
                return e2err_RICEventTriggerDefinitionIEValueFail_5;
            }
        }
        else if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.present == Interface_ID_PR_global_gNB_ID) {
            // GlobalGNB-ID
            pRICEventTriggerDefinition->interfaceID.globalGNBIDPresent = true;

            // PLMN-Identity
            pRICEventTriggerDefinition->interfaceID.globalGNBID.pLMNIdentity.contentLength =
              pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.size;
            memcpy(pRICEventTriggerDefinition->interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal,
              pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf,
              pRICEventTriggerDefinition->interfaceID.globalGNBID.pLMNIdentity.contentLength);

            // GNB-ID
            IdOctects_t gNBOctects;
            memset(gNBOctects.octets, 0, sizeof(gNBOctects));
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.present == GNB_ID_PR_gNB_ID) {
                pRICEventTriggerDefinition->interfaceID.globalGNBID.nodeID.bits = pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.size;
                memcpy(gNBOctects.octets, pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf,
                   pRICEventTriggerDefinition->interfaceID.globalGNBID.nodeID.bits);
                pRICEventTriggerDefinition->interfaceID.globalGNBID.nodeID.nodeID = gNBOctects.nodeID;
            }
            else {
                pRICEventTriggerDefinition->interfaceID.globalENBIDPresent = false;
                pRICEventTriggerDefinition->interfaceID.globalGNBIDPresent = false;
                ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
                return e2err_RICEventTriggerDefinitionIEValueFail_6;
            }
        }
        else {
            pRICEventTriggerDefinition->interfaceID.globalENBIDPresent = false;
            pRICEventTriggerDefinition->interfaceID.globalGNBIDPresent = false;
            ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
            return e2err_RICEventTriggerDefinitionIEValueFail_7;
        }

        // InterfaceDirection
        pRICEventTriggerDefinition->interfaceDirection = pE2SM_gNB_X2_eventTriggerDefinition->interfaceDirection;

        // InterfaceMessageType
        pRICEventTriggerDefinition->interfaceMessageType.procedureCode = pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.procedureCode;

        if (pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage == TypeOfMessage_initiating_message)
            pRICEventTriggerDefinition->interfaceMessageType.typeOfMessage = cE2InitiatingMessage;
        else if (pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage == TypeOfMessage_successful_outcome)
            pRICEventTriggerDefinition->interfaceMessageType.typeOfMessage = cE2SuccessfulOutcome;
        else if (pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage == TypeOfMessage_unsuccessful_outcome)
            pRICEventTriggerDefinition->interfaceMessageType.typeOfMessage = cE2UnsuccessfulOutcome;
        else {
            ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
            return e2err_RICEventTriggerDefinitionIEValueFail_8;
        }

        ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
        return e2err_OK;
    case RC_WMORE:
        if (debug)
            printf("\nDecode failed. More data needed. Buffer size %zu, %s, consumed %zu\n",pRICEventTriggerDefinition->octetString.contentLength,
                   asn_DEF_E2SM_gNB_X2_eventTriggerDefinition.name, rval.consumed);

        return e2err_RICEventTriggerDefinitionDecodeWMOREFail;
    case RC_FAIL:
        if (debug)
            printf("\nDecode failed. Buffer size %zu, %s, consumed %zu\n",pRICEventTriggerDefinition->octetString.contentLength,
                   asn_DEF_E2SM_gNB_X2_eventTriggerDefinition.name, rval.consumed);

        return e2err_RICEventTriggerDefinitionDecodeFAIL;
    default:
        return e2err_RICEventTriggerDefinitionDecodeDefaultFail;
    }
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICSubscriptionResponseData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionResponse_t* pRICSubscriptionResponse) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;

    // RICrequestID
    RICsubscriptionResponse_IEs_t* pRICsubscriptionResponse_IEs;
    if (pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list.count > 0) {
        pRICsubscriptionResponse_IEs = pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list.array[0];
        pRICSubscriptionResponse->ricRequestID.ricRequestorID = pRICsubscriptionResponse_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionResponse->ricRequestID.ricRequestSequenceNumber = pRICsubscriptionResponse_IEs->value.choice.RICrequestID.ricRequestSequenceNumber;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionResponseRICrequestIDMissing;
    }

    // RANfunctionID
    if (pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list.count > 1) {
        pRICsubscriptionResponse_IEs = pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list.array[1];
        pRICSubscriptionResponse->ranFunctionID = pRICsubscriptionResponse_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionResponseRANfunctionIDMissing;
    }

    // RICaction-Admitted-List
    if (pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list.count > 2) {
        pRICsubscriptionResponse_IEs = pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list.array[2];
        pRICSubscriptionResponse->ricActionAdmittedList.contentLength = 0;
        uint64_t index = 0;
        while ((index < maxofRICactionID) && (index < pRICsubscriptionResponse_IEs->value.choice.RICaction_Admitted_List.list.count)) {
            RICaction_Admitted_ItemIEs_t* pRICaction_Admitted_ItemIEs =
              (RICaction_Admitted_ItemIEs_t*)pRICsubscriptionResponse_IEs->value.choice.RICaction_Admitted_List.list.array[index];

            // RICActionID
            pRICSubscriptionResponse->ricActionAdmittedList.ricActionID[index] =
              pRICaction_Admitted_ItemIEs->value.choice.RICaction_Admitted_Item.ricActionID;
            index++;
        }
        pRICSubscriptionResponse->ricActionAdmittedList.contentLength = index;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionResponseRICaction_Admitted_ListMissing;
    }

    // RICaction-NotAdmitted-List, OPTIONAL
    if (pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list.count > 3) {
        pRICsubscriptionResponse_IEs = pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse.protocolIEs.list.array[3];
        if (pRICsubscriptionResponse_IEs->value.present == RICsubscriptionResponse_IEs__value_PR_RICaction_NotAdmitted_List) {
            pRICSubscriptionResponse->ricActionNotAdmittedListPresent = true;
            pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength = 0;
            uint64_t index = 0;
            while ((index < maxofRICactionID) && (index < pRICsubscriptionResponse_IEs->value.choice.RICaction_NotAdmitted_List.list.count)) {
                RICaction_NotAdmitted_ItemIEs_t* pRICaction_NotAdmitted_ItemIEs =
                  (RICaction_NotAdmitted_ItemIEs_t*)pRICsubscriptionResponse_IEs->value.choice.RICaction_NotAdmitted_List.list.array[index];

                // RICActionID
                pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricActionID;

                //  RICcause
                if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present == RICcause_PR_radioNetwork) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = RICcause_PR_radioNetwork;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.radioNetwork;
                }
                else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present == RICcause_PR_transport) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = RICcause_PR_transport;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.transport;
                }
                else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present == RICcause_PR_protocol) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = RICcause_PR_protocol;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.protocol;
                }
                else if(pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present == RICcause_PR_misc) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = RICcause_PR_misc;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.misc;
                }
                else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present == RICcause_PR_ric) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = RICcause_PR_ric;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.ric;
                }
               index++;
            }
            pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength = index;
        }
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionResponseRICaction_NotAdmitted_ListMissing;
    }

    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICSubscriptionFailureData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionFailure_t* pRICSubscriptionFailure) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;

    // RICrequestID
    RICsubscriptionFailure_IEs_t* pRICsubscriptionFailure_IEs;
    if (pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure.protocolIEs.list.count > 0) {
        pRICsubscriptionFailure_IEs = pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure.protocolIEs.list.array[0];
        pRICSubscriptionFailure->ricRequestID.ricRequestorID = pRICsubscriptionFailure_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionFailure->ricRequestID.ricRequestSequenceNumber = pRICsubscriptionFailure_IEs->value.choice.RICrequestID.ricRequestSequenceNumber;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionFailureRICrequestIDMissing;
    }

    // RANfunctionID
    if (pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure.protocolIEs.list.count > 1) {
        pRICsubscriptionFailure_IEs = pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure.protocolIEs.list.array[1];
        pRICSubscriptionFailure->ranFunctionID = pRICsubscriptionFailure_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionFailureRANfunctionIDMissing;
    }

    // RICaction-NotAdmitted-List
    if (pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure.protocolIEs.list.count > 2) {
        pRICsubscriptionFailure_IEs = pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure.protocolIEs.list.array[2];
        uint64_t index = 0;
        while ((index < maxofRICactionID) && (index < pRICsubscriptionFailure_IEs->value.choice.RICaction_NotAdmitted_List.list.count)) {
            RICaction_NotAdmitted_ItemIEs_t* pRICaction_NotAdmitted_ItemIEs =
              (RICaction_NotAdmitted_ItemIEs_t*)pRICsubscriptionFailure_IEs->value.choice.RICaction_NotAdmitted_List.list.array[index];

            // RICActionID
            pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID =
              pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricActionID;

            //  RICcause
            if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present == RICcause_PR_radioNetwork) {
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = RICcause_PR_radioNetwork;
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.radioNetwork;
            }
            else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present == RICcause_PR_transport) {
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = RICcause_PR_transport;
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.transport;
            }
            else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present == RICcause_PR_protocol) {
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = RICcause_PR_protocol;
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.protocol;
            }
            else if(pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present == RICcause_PR_misc) {
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = RICcause_PR_misc;
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.misc;
            }
            else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.present == RICcause_PR_ric) {
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = RICcause_PR_ric;
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricCause.choice.ric;
            }
            index++;
        }
        pRICSubscriptionFailure->ricActionNotAdmittedList.contentLength = index;

        // CriticalityDiagnostics. OPTIONAL

    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionFailureRICaction_NotAdmitted_ListMissing;
    }

    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICIndicationData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICIndication_t* pRICIndication) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;

    // RICrequestID
    RICindication_IEs_t* pRICindication_IEs;
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.count > 0) {
        pRICindication_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.array[0];
        pRICIndication->ricRequestID.ricRequestorID = pRICindication_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICIndication->ricRequestID.ricRequestSequenceNumber = pRICindication_IEs->value.choice.RICrequestID.ricRequestSequenceNumber;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICindicationRICrequestIDMissing;
    }

    // RANfunctionID
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.count > 1) {
        pRICindication_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.array[1];
        pRICIndication->ranFunctionID = pRICindication_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICindicationRANfunctionIDMissing;
    }

    // RICactionID
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.count > 2) {
        pRICindication_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.array[2];
        pRICIndication->ricActionID = pRICindication_IEs->value.choice.RICactionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICindicationRICactionIDMissing;
    }

    // RICindicationSN
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.count > 3) {
        pRICindication_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.array[3];
        pRICIndication->ricIndicationSN = pRICindication_IEs->value.choice.RICindicationSN;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICindicationRICindicationSNMissing;
    }

    // RICindicationType
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.count > 4) {
        pRICindication_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.array[4];
        pRICIndication->ricIndicationType = pRICindication_IEs->value.choice.RICindicationType;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICindicationRICindicationTypeMissing;
    }

    // RICindicationHeader
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.count > 5) {
        pRICindication_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.array[5];

        pRICIndication->ricIndicationHeader.octetString.contentLength = pRICindication_IEs->value.choice.RICindicationHeader.size;
        if (pRICIndication->ricIndicationHeader.octetString.contentLength < cMaxSizeOfOctetString) {
            memcpy(pRICIndication->ricIndicationHeader.octetString.data, pRICindication_IEs->value.choice.RICindicationHeader.buf,
              pRICIndication->ricIndicationHeader.octetString.contentLength);

              uint64_t returnCode;
              if ((returnCode = getRICIndicationHeaderData(&pRICIndication->ricIndicationHeader) != e2err_OK)) {
                ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
                return returnCode;
              }
        }
        else {
            ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
            return e2err_RICIndicationHeaderContentLengthFail;
        }
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICindicationRICindicationHeaderMissing;
    }

    // RICindicationMessage
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.count > 6) {
        pRICindication_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICindication.protocolIEs.list.array[6];

        pRICIndication->ricIndicationMessage.octetString.contentLength = pRICindication_IEs->value.choice.RICindicationMessage.size;
        if (pRICIndication->ricIndicationMessage.octetString.contentLength < cMaxSizeOfOctetString) {
            memcpy(pRICIndication->ricIndicationMessage.octetString.data, pRICindication_IEs->value.choice.RICindicationMessage.buf,
              pRICIndication->ricIndicationMessage.octetString.contentLength);

              uint64_t returnCode;
              if ((returnCode = getRICIndicationMessageData(&pRICIndication->ricIndicationMessage) != e2err_OK)) {
                ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
                return returnCode;
              }
        }
        else {
            ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
            return e2err_RICIndicationMessageContentLengthFail;
        }
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICindicationRICindicationMessageMissing;
    }

    // RICcallProcessID, OPTIONAL. Not used in RIC.

    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICIndicationHeaderData(RICIndicationHeader_t* pRICIndicationHeader) {

    E2SM_gNB_X2_indicationHeader_t* pE2SM_gNB_X2_indicationHeader = 0;
    asn_dec_rval_t rval;
    rval = asn_decode(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2SM_gNB_X2_indicationHeader, (void **)&pE2SM_gNB_X2_indicationHeader,
                      pRICIndicationHeader->octetString.data, pRICIndicationHeader->octetString.contentLength);
    switch(rval.code) {
    case RC_OK:
        // Debug print
        if (debug) {
            printf("\nSuccessfully decoded E2SM_gNB_X2_indicationHeader\n\n");
            asn_fprint(stdout, &asn_DEF_E2SM_gNB_X2_indicationHeader, pE2SM_gNB_X2_indicationHeader);
        }

        // InterfaceID, GlobalENB-ID or GlobalGNB-ID
        if (pE2SM_gNB_X2_indicationHeader->interface_ID.present == Interface_ID_PR_global_eNB_ID) {

            // GlobalENB-ID
            pRICIndicationHeader->interfaceID.globalENBIDPresent = true;

            // PLMN-Identity
            pRICIndicationHeader->interfaceID.globalENBID.pLMNIdentity.contentLength =
              pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.pLMN_Identity.size;
            memcpy(pRICIndicationHeader->interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal,
              pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf,
              pRICIndicationHeader->interfaceID.globalENBID.pLMNIdentity.contentLength);

            //  ENB-ID
            IdOctects_t eNBOctects;
            memset(eNBOctects.octets, 0, sizeof(eNBOctects));
            if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.present == ENB_ID_PR_macro_eNB_ID) {
                // BIT STRING (SIZE (20)
                pRICIndicationHeader->interfaceID.globalENBID.nodeID.bits = cMacroENBIDP_20Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf,
                  pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.size);
                pRICIndicationHeader->interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.present == ENB_ID_PR_home_eNB_ID) {
                // BIT STRING (SIZE (28)
                pRICIndicationHeader->interfaceID.globalENBID.nodeID.bits = cHomeENBID_28Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf,
                  pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.size);
                pRICIndicationHeader->interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.present == ENB_ID_PR_short_Macro_eNB_ID) {
                // BIT STRING (SIZE(18)
                pRICIndicationHeader->interfaceID.globalENBID.nodeID.bits = cShortMacroENBID_18Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf,
                  pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.size);
                pRICIndicationHeader->interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.present == ENB_ID_PR_long_Macro_eNB_ID) {
                // BIT STRING (SIZE(21)
                pRICIndicationHeader->interfaceID.globalENBID.nodeID.bits =  clongMacroENBIDP_21Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf,
                  pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.size);
                pRICIndicationHeader->interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else {
                pRICIndicationHeader->interfaceID.globalENBIDPresent = false;
                pRICIndicationHeader->interfaceID.globalGNBIDPresent = false;
                ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_indicationHeader, pE2SM_gNB_X2_indicationHeader);
                return e2err_RICEventTriggerDefinitionIEValueFail_9;
            }
        }
        else if (pE2SM_gNB_X2_indicationHeader->interface_ID.present == Interface_ID_PR_global_gNB_ID) {
            // GlobalGNB-ID
            pRICIndicationHeader->interfaceID.globalGNBIDPresent = true;

            // PLMN-Identity
            pRICIndicationHeader->interfaceID.globalGNBID.pLMNIdentity.contentLength =
              pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.pLMN_Identity.size;
            memcpy(pRICIndicationHeader->interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal,
              pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf,
              pRICIndicationHeader->interfaceID.globalGNBID.pLMNIdentity.contentLength);

            // GNB-ID
            IdOctects_t gNBOctects;
            memset(gNBOctects.octets, 0, sizeof(gNBOctects));
            if (pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.gNB_ID.present == GNB_ID_PR_gNB_ID) {
                pRICIndicationHeader->interfaceID.globalGNBID.nodeID.bits = pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.size;
                memcpy(gNBOctects.octets, pE2SM_gNB_X2_indicationHeader->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf,
                   pRICIndicationHeader->interfaceID.globalGNBID.nodeID.bits);
                pRICIndicationHeader->interfaceID.globalGNBID.nodeID.nodeID = gNBOctects.nodeID;
            }
            else {
                pRICIndicationHeader->interfaceID.globalENBIDPresent = false;
                pRICIndicationHeader->interfaceID.globalGNBIDPresent = false;
                ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_indicationHeader, pE2SM_gNB_X2_indicationHeader);
                return e2err_RICEventTriggerDefinitionIEValueFail_10;
            }
        }
        else {
            pRICIndicationHeader->interfaceID.globalENBIDPresent = false;
            pRICIndicationHeader->interfaceID.globalGNBIDPresent = false;
            ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_indicationHeader, pE2SM_gNB_X2_indicationHeader);
            return e2err_RICEventTriggerDefinitionIEValueFail_11;
        }

        // InterfaceDirection
        pRICIndicationHeader->interfaceDirection = pE2SM_gNB_X2_indicationHeader->interfaceDirection;

        // TimeStamp OPTIONAL. Not used in RIC.

        ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_indicationHeader, pE2SM_gNB_X2_indicationHeader);
        return e2err_OK;
    case RC_WMORE:
        if (debug)
            printf("\nDecode failed. More data needed. Buffer size %zu, %s, consumed %zu\n",pRICIndicationHeader->octetString.contentLength,
                   asn_DEF_E2SM_gNB_X2_indicationHeader.name, rval.consumed);
        return e2err_RICIndicationHeaderDecodeWMOREFail;
    case RC_FAIL:
        if (debug)
            printf("\nDecode failed. Buffer size %zu, %s, consumed %zu\n",pRICIndicationHeader->octetString.contentLength,
                   asn_DEF_E2SM_gNB_X2_indicationHeader.name, rval.consumed);

        return e2err_RICIndicationHeaderDecodeFAIL;
    default:
        return e2err_RICIndicationHeaderDecodeDefaultFail;
    }
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICIndicationMessageData(RICIndicationMessage_t* pRICIndicationMessage) {

    E2SM_gNB_X2_indicationMessage_t* pE2SM_gNB_X2_indicationMessage = 0;
    asn_dec_rval_t rval;
    rval = asn_decode(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2SM_gNB_X2_indicationMessage, (void **)&pE2SM_gNB_X2_indicationMessage,
                      pRICIndicationMessage->octetString.data, pRICIndicationMessage->octetString.contentLength);
    switch(rval.code) {
    case RC_OK:
        // Debug print
        if (debug) {
            printf("\nSuccessfully decoded E2SM_gNB_X2_indicationMessage\n\n");
            asn_fprint(stdout, &asn_DEF_E2SM_gNB_X2_indicationMessage, pE2SM_gNB_X2_indicationMessage);
        }

        // InterfaceMessage
        pRICIndicationMessage->interfaceMessage.contentLength = pE2SM_gNB_X2_indicationMessage->interfaceMessage.size;
        if(pRICIndicationMessage->octetString.contentLength < cMaxSizeOfOctetString) {
            memcpy(pRICIndicationMessage->interfaceMessage.data,pE2SM_gNB_X2_indicationMessage->interfaceMessage.buf,
              pRICIndicationMessage->interfaceMessage.contentLength);
            ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_indicationMessage, pE2SM_gNB_X2_indicationMessage);
            return e2err_OK;
        }
        else {
            ASN_STRUCT_FREE(asn_DEF_E2SM_gNB_X2_indicationMessage, pE2SM_gNB_X2_indicationMessage);
            return e2err_RICIndicationMessageIEContentLengthFail;
        }
    case RC_WMORE:
        if (debug)
            printf("\nDecode failed. More data needed. Buffer size %zu, %s, consumed %zu\n",pRICIndicationMessage->octetString.contentLength,
                   asn_DEF_E2SM_gNB_X2_indicationMessage.name, rval.consumed);

        return e2err_RICIndicationMessageDecodeWMOREFail;
    case RC_FAIL:
        if (debug)
            printf("\nDecode failed. Buffer size %zu, %s, consumed %zu\n",pRICIndicationMessage->octetString.contentLength,
                   asn_DEF_E2SM_gNB_X2_indicationMessage.name, rval.consumed);

        return e2err_RICIndicationMessageDecodeFAIL;
    default:
        return e2err_RICIndicationMessageDecodeDefaultFail;
    }
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICSubscriptionDeleteRequestData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionDeleteRequest_t* pRICSubscriptionDeleteRequest) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;

    // RICrequestID
    RICsubscriptionDeleteRequest_IEs_t* pRICsubscriptionDeleteRequest_IEs;
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionDeleteRequest.protocolIEs.list.count > 0) {
        pRICsubscriptionDeleteRequest_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionDeleteRequest.protocolIEs.list.array[0];
        pRICSubscriptionDeleteRequest->ricRequestID.ricRequestorID = pRICsubscriptionDeleteRequest_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionDeleteRequest->ricRequestID.ricRequestSequenceNumber = pRICsubscriptionDeleteRequest_IEs->value.choice.RICrequestID.ricRequestSequenceNumber;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteRequestRICrequestIDMissing;
    }

    // RANfunctionID
    if (pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionDeleteRequest.protocolIEs.list.count > 1) {
        pRICsubscriptionDeleteRequest_IEs = pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionDeleteRequest.protocolIEs.list.array[1];
        pRICSubscriptionDeleteRequest->ranFunctionID = pRICsubscriptionDeleteRequest_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteRequestRANfunctionIDMissing;
    }

    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICSubscriptionDeleteResponseData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionDeleteResponse_t* pRICSubscriptionDeleteResponse) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;

    // RICrequestID
    RICsubscriptionDeleteResponse_IEs_t* pRICsubscriptionDeleteResponse_IEs;
    if (pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionDeleteResponse.protocolIEs.list.count > 0) {
        pRICsubscriptionDeleteResponse_IEs = pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionDeleteResponse.protocolIEs.list.array[0];
        pRICSubscriptionDeleteResponse->ricRequestID.ricRequestorID = pRICsubscriptionDeleteResponse_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionDeleteResponse->ricRequestID.ricRequestSequenceNumber = pRICsubscriptionDeleteResponse_IEs->value.choice.RICrequestID.ricRequestSequenceNumber;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteResponseRICrequestIDMissing;
    }

    // RANfunctionID
    if (pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionDeleteResponse.protocolIEs.list.count > 1) {
        pRICsubscriptionDeleteResponse_IEs = pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionDeleteResponse.protocolIEs.list.array[1];
        pRICSubscriptionDeleteResponse->ranFunctionID = pRICsubscriptionDeleteResponse_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteResponseRANfunctionIDMissing;
    }

    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICSubscriptionDeleteFailureData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionDeleteFailure_t* pRICSubscriptionDeleteFailure) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;

    // RICrequestID
    RICsubscriptionDeleteFailure_IEs_t* pRICsubscriptionDeleteFailure_IEs;
    if (pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionDeleteFailure.protocolIEs.list.count > 0) {
        pRICsubscriptionDeleteFailure_IEs = pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionDeleteFailure.protocolIEs.list.array[0];
        pRICSubscriptionDeleteFailure->ricRequestID.ricRequestorID = pRICsubscriptionDeleteFailure_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionDeleteFailure->ricRequestID.ricRequestSequenceNumber = pRICsubscriptionDeleteFailure_IEs->value.choice.RICrequestID.ricRequestSequenceNumber;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteFailureRICrequestIDMissing;
    }

    // RANfunctionID
    if (pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionDeleteFailure.protocolIEs.list.count > 1) {
        pRICsubscriptionDeleteFailure_IEs = pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionDeleteFailure.protocolIEs.list.array[1];
        pRICSubscriptionDeleteFailure->ranFunctionID = pRICsubscriptionDeleteFailure_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteFailureRANfunctionIDMissing;
    }

    // RICcause
    if (pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionDeleteFailure.protocolIEs.list.count > 2) {
        pRICsubscriptionDeleteFailure_IEs = pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionDeleteFailure.protocolIEs.list.array[2];
        if (pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.present == RICcause_PR_radioNetwork) {
            pRICSubscriptionDeleteFailure->ricCause.content = RICcause_PR_radioNetwork;
            pRICSubscriptionDeleteFailure->ricCause.cause =
              pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.choice.radioNetwork;
        }
        else if (pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.present == RICcause_PR_transport) {
            pRICSubscriptionDeleteFailure->ricCause.content = RICcause_PR_transport;
            pRICSubscriptionDeleteFailure->ricCause.cause =
              pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.choice.transport;
        }
        else if (pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.present == RICcause_PR_protocol) {
            pRICSubscriptionDeleteFailure->ricCause.content = RICcause_PR_protocol;
            pRICSubscriptionDeleteFailure->ricCause.cause =
              pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.choice.protocol;
        }
        else if(pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.present == RICcause_PR_misc) {
            pRICSubscriptionDeleteFailure->ricCause.content = RICcause_PR_misc;
            pRICSubscriptionDeleteFailure->ricCause.cause =
              pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.choice.misc;
        }
        else if (pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.present == RICcause_PR_ric) {
            pRICSubscriptionDeleteFailure->ricCause.content = RICcause_PR_ric;
            pRICSubscriptionDeleteFailure->ricCause.cause =
              pRICsubscriptionDeleteFailure_IEs->value.choice.RICcause.choice.ric;
        }
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteFailureRICcauseMissing;
    }
    // CriticalityDiagnostics, OPTIONAL

    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
    return e2err_OK;
}
