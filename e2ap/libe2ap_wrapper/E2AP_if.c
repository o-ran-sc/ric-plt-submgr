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

#include "E2_asn_constant.h"
#include "E2AP-PDU.h"
#include "ProtocolIE-Field.h"
#include "RICsubsequentAction.h"
#include "E2_E2SM-gNB-X2-eventTriggerDefinition.h"
#include "E2_E2SM-gNB-NRT-EventTriggerDefinition.h"
#include "E2_E2SM-gNB-X2-ActionDefinitionChoice.h"
#include "E2_E2SM-gNB-X2-actionDefinition.h"
#include "E2_ActionParameter-Item.h"
#include "E2_E2SM-gNB-X2-ActionDefinition-Format2.h"
#include "E2_RANueGroup-Item.h"
#include "E2_RANueGroupDef-Item.h"
#include "E2_RANimperativePolicy.h"
#include "E2_RANParameter-Item.h"

// E2SM-gNB-NRT
#include "E2_E2SM-gNB-NRT-ActionDefinition.h"
#include "E2_E2SM-gNB-NRT-ActionDefinition-Format1.h"
#include "E2_RANparameter-Item.h"

#include "asn_constant.h"
#include "E2_asn_constant.h"
#include "E2AP_if.h"


#ifdef DEBUG
    static const bool debug = true;
#else
    static const bool debug = true; //false;
#endif

const int64_t cMaxNrOfErrors = 256;
const uint64_t cMaxSizeOfOctetString = 1024;

const size_t cMacroENBIDP_20Bits = 20;
const size_t cHomeENBID_28Bits = 28;
const size_t cShortMacroENBID_18Bits = 18;
const size_t clongMacroENBIDP_21Bits = 21;

const int cCauseRICRequest = 1;
const int cCauseRICService = 2;
const int cCauseTransport = 3;
const int cCauseProtocol = 4;
const int cCauseMisc = 5;

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
        sprintf(pLogBuffer,"Serialization of %s failed", asn_DEF_E2AP_PDU.name);
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return false;
    }
    else if (rval.encoded > *dataBufferSize) {
        sprintf(pLogBuffer,"Buffer of size %zu is too small for %s, need %zu",*dataBufferSize, asn_DEF_E2AP_PDU.name, rval.encoded);
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return false;
    }
    else {
        if (debug)
            sprintf(pLogBuffer,"Successfully encoded %s. Buffer size %zu, encoded size %zu",asn_DEF_E2AP_PDU.name, *dataBufferSize, rval.encoded);

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
        pE2AP_PDU->choice.initiatingMessage.procedureCode = ProcedureCode_id_RICsubscription;
        pE2AP_PDU->choice.initiatingMessage.criticality = Criticality_ignore;
        pE2AP_PDU->choice.initiatingMessage.value.present = InitiatingMessage__value_PR_RICsubscriptionRequest;

        // RICrequestID
        RICsubscriptionRequest_IEs_t* pRICsubscriptionRequest_IEs = calloc(1, sizeof(RICsubscriptionRequest_IEs_t));
        if (pRICsubscriptionRequest_IEs) {
            pRICsubscriptionRequest_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionRequest_IEs->criticality = Criticality_reject;
            pRICsubscriptionRequest_IEs->value.present = RICsubscriptionRequest_IEs__value_PR_RICrequestID;
            pRICsubscriptionRequest_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionRequest->ricRequestID.ricRequestorID;
            pRICsubscriptionRequest_IEs->value.choice.RICrequestID.ricInstanceID = pRICSubscriptionRequest->ricRequestID.ricInstanceID;
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

        // RICsubscriptionDetails
        pRICsubscriptionRequest_IEs = calloc(1, sizeof(RICsubscriptionRequest_IEs_t));
        if (pRICsubscriptionRequest_IEs) {
            pRICsubscriptionRequest_IEs->id = ProtocolIE_ID_id_RICsubscriptionDetails;
            pRICsubscriptionRequest_IEs->criticality = Criticality_reject;
            pRICsubscriptionRequest_IEs->value.present = RICsubscriptionRequest_IEs__value_PR_RICsubscriptionDetails;

            // RICeventTriggerDefinition
            uint64_t returnCode;
            if ((returnCode = packRICEventTriggerDefinition(pLogBuffer, &pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition)) != e2err_OK) {
                return returnCode;
            }

            pRICsubscriptionRequest_IEs->value.choice.RICsubscriptionDetails.ricEventTriggerDefinition.buf =
              calloc(1, pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.contentLength);
            if (pRICsubscriptionRequest_IEs->value.choice.RICsubscriptionDetails.ricEventTriggerDefinition.buf) {
                pRICsubscriptionRequest_IEs->value.choice.RICsubscriptionDetails.ricEventTriggerDefinition.size =
                  pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.contentLength;
                memcpy(pRICsubscriptionRequest_IEs->value.choice.RICsubscriptionDetails.ricEventTriggerDefinition.buf,
                       pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.data,
                       pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.contentLength);
            }
            else
                return e2err_RICSubscriptionRequestAllocRICeventTriggerDefinitionBufFail;

            // RICactions-ToBeSetup-List
            uint64_t index = 0;
            while (index < pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength && index < maxofRICactionID) {
                RICaction_ToBeSetup_ItemIEs_t* pRICaction_ToBeSetup_ItemIEs = calloc(1, sizeof(RICaction_ToBeSetup_ItemIEs_t));
                if (pRICaction_ToBeSetup_ItemIEs) {
                    pRICaction_ToBeSetup_ItemIEs->id = ProtocolIE_ID_id_RICaction_ToBeSetup_Item;
                    pRICaction_ToBeSetup_ItemIEs->criticality = Criticality_reject;
                    pRICaction_ToBeSetup_ItemIEs->value.present = RICaction_ToBeSetup_ItemIEs__value_PR_RICaction_ToBeSetup_Item;

                    // RICActionID
                    pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionID =
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID;

                    // RICActionType
                    pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionType =
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType;

                    // RICactionDefinition, OPTIONAL
                    if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent) {
                        uint64_t returnCode;
                        if ((returnCode = packRICActionDefinition(pLogBuffer, &pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice)) != e2err_OK) {
                            return returnCode;
                        }

                        pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition = calloc(1, sizeof (RICactionDefinition_t));
                        if (pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition) {
                            pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition->buf =
                              calloc(1, pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.contentLength);
                            if (pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition->buf) {
                                pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition->size =
                                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.contentLength;
                                memcpy(pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition->buf,
                                       pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.data,
                                       pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.contentLength);
                            }
                            else
                                return e2err_RICSubscriptionRequestAllocRICactionDefinitionBufFail;
                        }
                        else
                            return e2err_RICSubscriptionRequestAllocRICactionDefinitionFail;
                    }

                    // RICsubsequentAction, OPTIONAL
                    if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent) {
                        RICsubsequentAction_t* pRICsubsequentAction = calloc(1, sizeof(RICsubsequentAction_t));
                        if (pRICsubsequentAction) {
                            pRICsubsequentAction->ricSubsequentActionType =
                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType;
                            pRICsubsequentAction->ricTimeToWait =
                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait;
                            pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricSubsequentAction = pRICsubsequentAction;
                        }
                        else
                            return e2err_RICSubscriptionRequestAllocRICsubsequentActionFail;
                    }
                }
                else
                    return e2err_RICSubscriptionRequestAllocRICaction_ToBeSetup_ItemIEsFail;
                ASN_SEQUENCE_ADD(&pRICsubscriptionRequest_IEs->value.choice.RICsubscriptionDetails.ricAction_ToBeSetup_List.list, pRICaction_ToBeSetup_ItemIEs);
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

    if (pRICEventTriggerDefinition->E2SMgNBX2EventTriggerDefinitionPresent)
        return packRICEventTriggerDefinitionX2Format(pLogBuffer, pRICEventTriggerDefinition);
    else if(pRICEventTriggerDefinition->E2SMgNBNRTEventTriggerDefinitionPresent)
        return packRICEventTriggerDefinitionNRTFormat(pLogBuffer, pRICEventTriggerDefinition);
    else
        return e2err_RICEventTriggerDefinitionAllocEventTriggerDefinitionEmptyFail;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICEventTriggerDefinitionX2Format(char* pLogBuffer, RICEventTriggerDefinition_t* pRICEventTriggerDefinition) {

    E2_E2SM_gNB_X2_eventTriggerDefinition_t* pE2SM_gNB_X2_eventTriggerDefinition = calloc(1, sizeof(E2_E2SM_gNB_X2_eventTriggerDefinition_t));
    if(pE2SM_gNB_X2_eventTriggerDefinition == NULL)
        return e2err_RICEventTriggerDefinitionAllocE2SM_gNB_X2_eventTriggerDefinitionFail;

    // RICeventTriggerDefinition
    // InterfaceID
    if ((pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBIDPresent == true &&
         pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBIDPresent == true) ||
        (pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBIDPresent == false &&
         pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBIDPresent == false))
        return e2err_RICEventTriggerDefinitionIEValueFail_1;

    // GlobalENB-ID or GlobalGNB-ID
    if (pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBIDPresent)
    {
        pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.present = E2_Interface_ID_PR_global_eNB_ID;

        // GlobalENB-ID
        // PLMN-Identity
        pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.size =
        pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.contentLength;
        pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf = calloc(1,3);
        if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf) {
            memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf,
                   pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal,
                   pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.contentLength);
        }
        else
            return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDpLMN_IdentityBufFail;

        // Add ENB-ID
        if (pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits == cMacroENBIDP_20Bits){
            // BIT STRING, SIZE 20
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_macro_eNB_ID;
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf = calloc(1,3);
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf) {
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.size = 3; // bytes
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.bits_unused = 4; // trailing unused bits
                memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf,
                       (void*)&pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID,3);
            }
            else
                return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDmacro_eNB_IDBufFail;
        }
        else if (pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits == cHomeENBID_28Bits) {
            // BIT STRING, SIZE 28
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_home_eNB_ID;
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf = calloc(1,4);
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf) {
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.size = 4; // bytes
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.bits_unused = 4; // trailing unused bits
                memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf,
                       (void*)&pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID,4);
            }
            else
                return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDhome_eNB_IDBufFail;
        }
        else if (pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits == cShortMacroENBID_18Bits) {
            // BIT STRING, SIZE 18
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_short_Macro_eNB_ID;
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf = calloc(1,3);
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf) {
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.size = 3;
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.bits_unused = 6; // trailing unused bits
                memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf,
                       (void*)&pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID,3);
            }
            else
                return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDshort_Macro_eNB_IDBufFail;
        }
        else if (pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits == clongMacroENBIDP_21Bits) {
            // BIT STRING, SIZE 21
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present = ENB_ID_PR_long_Macro_eNB_ID;
            pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf = calloc(1,3);
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf) {
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.size = 3; // bytes
                pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.bits_unused = 3; // trailing unused bits
                memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf,
                       (void*)&pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID,3);
            }
            else
                return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_eNB_IDeNB_IDlong_Macro_eNB_IDBufFail;
        }
        else
            return e2err_RICEventTriggerDefinitionIEValueFail_2;
    }
    else if (pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBIDPresent) {
        // GlobalGNB-ID
        pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.present = E2_Interface_ID_PR_global_gNB_ID;

        // PLMN-Identity
        pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.pLMN_Identity.size =
          pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.contentLength;
        pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.pLMN_Identity.buf =
          calloc(1,pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.contentLength);
        if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.pLMN_Identity.buf) {
            memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.pLMN_Identity.buf,
                   (void*)&pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal,
                    pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.contentLength);
        }
        else
            return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_gNB_IDpLMN_IdentityBufFail;
        // GNB-ID, BIT STRING, SIZE 22..32
        pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.size = 4;  //32bits
        pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf = calloc(1, 4);
        if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf) {
            memcpy(pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf,
                   (void*)&pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID,4); //32bits
        }
        else
            return e2err_RICIndicationAllocRICEventTriggerDefinitionglobal_gNB_IDgNB_IDBufFail;
    }
    else
        return e2err_RICEventTriggerDefinitionIEValueFail_3;

    // InterfaceDirection
    pE2SM_gNB_X2_eventTriggerDefinition->interfaceDirection = pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceDirection;

    // InterfaceMessageType
    // ProcedureCode
    pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.procedureCode = pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceMessageType.procedureCode;

    // TypeOfMessage
    if(pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceMessageType.typeOfMessage == cE2InitiatingMessage)
        pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage = E2_TypeOfMessage_initiating_message;
    else if(pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceMessageType.typeOfMessage == cE2SuccessfulOutcome)
        pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage = E2_TypeOfMessage_successful_outcome;
    else if(pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceMessageType.typeOfMessage == cE2UnsuccessfulOutcome)
        pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage = E2_TypeOfMessage_unsuccessful_outcome;
    else
        return e2err_RICEventTriggerDefinitionIEValueFail_4;

    // InterfaceProtocolIE-List, OPTIONAL. Not used in RIC currently
    if (pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceProtocolIEListPresent == true) {}

    // Debug print
    if (debug)
        asn_fprint(stdout, &asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);

    // Encode
    size_t bufferSize = sizeof(pRICEventTriggerDefinition->octetString.data);
    asn_enc_rval_t rval;
    rval = asn_encode_to_buffer(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition,
                                pRICEventTriggerDefinition->octetString.data, bufferSize);

    if(rval.encoded == -1) {
        sprintf(pLogBuffer,"Serialization of %s failed", asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition.name);
        return e2err_RICEventTriggerDefinitionPackFail_1;
    }
    else if(rval.encoded > bufferSize) {
       sprintf(pLogBuffer,"Buffer of size %zu is too small for %s, need %zu",bufferSize, asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition.name, rval.encoded);
        return e2err_RICEventTriggerDefinitionPackFail_2;
    }
    else
    if (debug)
           sprintf(pLogBuffer,"Successfully encoded %s. Buffer size %zu, encoded size %zu",asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition.name, bufferSize, rval.encoded);

    ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);

    pRICEventTriggerDefinition->octetString.contentLength = rval.encoded;
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICEventTriggerDefinitionNRTFormat(char* pLogBuffer, RICEventTriggerDefinition_t* pRICEventTriggerDefinition) {

    E2_E2SM_gNB_NRT_EventTriggerDefinition_t* pE2_E2SM_gNB_NRT_EventTriggerDefinition = calloc(1, sizeof(E2_E2SM_gNB_NRT_EventTriggerDefinition_t));
    if(pE2_E2SM_gNB_NRT_EventTriggerDefinition == NULL)
        return e2err_RICEventTriggerDefinitionAllocE2SM_gNB_NRT_eventTriggerDefinitionFail;

    pE2_E2SM_gNB_NRT_EventTriggerDefinition->present = E2_E2SM_gNB_NRT_EventTriggerDefinition_PR_eventDefinition_Format1;
    pE2_E2SM_gNB_NRT_EventTriggerDefinition->choice.eventDefinition_Format1.triggerNature =
      pRICEventTriggerDefinition->e2SMgNBNRTEventTriggerDefinition.eventDefinitionFormat1.triggerNature;

    // Debug print
    if (debug)
        asn_fprint(stdout, &asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition, pE2_E2SM_gNB_NRT_EventTriggerDefinition);

    // Encode
    size_t bufferSize = sizeof(pRICEventTriggerDefinition->octetString.data);
    asn_enc_rval_t rval;
    rval = asn_encode_to_buffer(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition, pE2_E2SM_gNB_NRT_EventTriggerDefinition,
                                pRICEventTriggerDefinition->octetString.data, bufferSize);

    if(rval.encoded == -1) {
        sprintf(pLogBuffer,"Serialization of %s failed", asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition.name);
        return e2err_RICENRTventTriggerDefinitionPackFail_1;
    }
    else if(rval.encoded > bufferSize) {
       sprintf(pLogBuffer,"Buffer of size %zu is too small for %s, need %zu",bufferSize, asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition.name, rval.encoded);
        return e2err_RICNRTEventTriggerDefinitionPackFail_2;
    }
    else
    if (debug)
           sprintf(pLogBuffer,"Successfully encoded %s. Buffer size %zu, encoded size %zu",asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition.name, bufferSize, rval.encoded);

    ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition, pE2_E2SM_gNB_NRT_EventTriggerDefinition);

    pRICEventTriggerDefinition->octetString.contentLength = rval.encoded;
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICActionDefinition(char* pLogBuffer, RICActionDefinitionChoice_t* pRICActionDefinitionChoice) {

    if (pRICActionDefinitionChoice->actionDefinitionX2Format1Present ||
        pRICActionDefinitionChoice->actionDefinitionX2Format2Present) {
        // E2SM-gNB-X2-actionDefinition
        return packActionDefinitionX2Format(pLogBuffer,pRICActionDefinitionChoice);
    }
    else if (pRICActionDefinitionChoice->actionDefinitionNRTFormat1Present) {
        // E2SM-gNB-NRT-actionDefinition
        return packActionDefinitionNRTFormat(pLogBuffer,pRICActionDefinitionChoice);
    }
    else
        return e2err_RICSubscriptionRequestRICActionDefinitionEmpty;
}

//////////////////////////////////////////////////////////////////////
uint64_t packActionDefinitionX2Format(char* pLogBuffer, RICActionDefinitionChoice_t* pRICActionDefinitionChoice) {

    int result;

    // E2SM-gNB-X2-actionDefinition
    E2_E2SM_gNB_X2_ActionDefinitionChoice_t* pE2_E2SM_gNB_X2_ActionDefinitionChoice = calloc(1, sizeof(E2_E2SM_gNB_X2_ActionDefinitionChoice_t));
    if (pE2_E2SM_gNB_X2_ActionDefinitionChoice == NULL) {
        return e2err_RICSubscriptionRequestAllocE2_E2SM_gNB_X2_ActionDefinitionChoiceFail;
    }

    if (pRICActionDefinitionChoice->actionDefinitionX2Format1Present) {

        // E2SM-gNB-X2-actionDefinition
        pE2_E2SM_gNB_X2_ActionDefinitionChoice->present = E2_E2SM_gNB_X2_ActionDefinitionChoice_PR_actionDefinition_Format1;
        pE2_E2SM_gNB_X2_ActionDefinitionChoice->choice.actionDefinition_Format1.style_ID = pRICActionDefinitionChoice->actionDefinitionX2Format1->styleID;

        if (pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterCount > 0) {
            struct E2_E2SM_gNB_X2_actionDefinition__actionParameter_List* pE2_E2SM_gNB_X2_actionDefinition__actionParameter_List =
            calloc(1, sizeof (struct E2_E2SM_gNB_X2_actionDefinition__actionParameter_List));
            uint64_t index = 0;
            while (index < pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterCount && index < E2_maxofRANParameters) {
                E2_ActionParameter_Item_t* pE2_ActionParameter_Item = calloc(1, sizeof(E2_ActionParameter_Item_t));
                if (pE2_ActionParameter_Item) {
                    pE2_ActionParameter_Item->actionParameter_ID = pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].parameterID;
                    if (pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueIntPresent) {
                        pE2_ActionParameter_Item->actionParameter_Value.present = E2_ActionParameter_Value_PR_valueInt;
                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueInt =
                        pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueInt;
                    }
                    else if (pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueEnumPresent) {
                        pE2_ActionParameter_Item->actionParameter_Value.present = E2_ActionParameter_Value_PR_valueEnum;
                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueEnum =
                        pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueIntPresent;
                    }
                    else if (pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBoolPresent) {
                        pE2_ActionParameter_Item->actionParameter_Value.present = E2_ActionParameter_Value_PR_valueBool;
                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueBool =
                        pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBool;
                    }
                    else if (pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBitSPresent) {
                        pE2_ActionParameter_Item->actionParameter_Value.present = E2_ActionParameter_Value_PR_valueBitS;
                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueBitS.size =
                        pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBitS.byteLength;
                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueBitS.bits_unused =
                        pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBitS.unusedBits;
                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueBitS.buf =
                        calloc(pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBitS.byteLength, 1);
                        if (pE2_ActionParameter_Item->actionParameter_Value.choice.valueBitS.buf) {
                            memcpy(pE2_ActionParameter_Item->actionParameter_Value.choice.valueBitS.buf,
                                pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBitS.data,
                                pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBitS.byteLength);
                        }
                        else
                            return e2err_RICSubscriptionRequestAllocactionParameterValueValueBitSFail;
                    }
                    else if (pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueOctSPresent) {
                        pE2_ActionParameter_Item->actionParameter_Value.present = E2_ActionParameter_Value_PR_valueOctS;
                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueOctS.size =
                        pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueOctS.length;
                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueOctS.buf =
                        calloc(pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueOctS.length, 1);
                        if (pE2_ActionParameter_Item->actionParameter_Value.choice.valueOctS.buf) {
                            memcpy(pE2_ActionParameter_Item->actionParameter_Value.choice.valueOctS.buf,
                                pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueOctS.data,
                                pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueOctS.length);
                        }
                        else
                            return e2err_RICSubscriptionRequestAllocactionParameterValueValueOctSFail;
                    }
                    else if (pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valuePrtSPresent) {
                        pE2_ActionParameter_Item->actionParameter_Value.present = E2_ActionParameter_Value_PR_valuePrtS;
                        pE2_ActionParameter_Item->actionParameter_Value.choice.valuePrtS.size =
                        pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valuePrtS.length;
                        pE2_ActionParameter_Item->actionParameter_Value.choice.valuePrtS.buf =
                        calloc(pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valuePrtS.length ,1);
                        if (pE2_ActionParameter_Item->actionParameter_Value.choice.valuePrtS.buf) {
                            memcpy(pE2_ActionParameter_Item->actionParameter_Value.choice.valuePrtS.buf,
                                pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valuePrtS.data,
                                pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valuePrtS.length);
                        }
                        else
                            return e2err_RICSubscriptionRequestAllocactionParameterValueValuePrtsSFail;
                    }
                    else
                        return e2err_RICSubscriptionRequestActionParameterItemFail;

                    if ((result = asn_set_add(pE2_E2SM_gNB_X2_actionDefinition__actionParameter_List, pE2_ActionParameter_Item)) != 0)
                        return e2err_RICSubscriptionRequestAsn_set_addE2_ActionParameter_ItemFail;
                }
                else
                    return e2err_RICSubscriptionRequestAllocActionDefinitionFail;
                index++;
            }
            pE2_E2SM_gNB_X2_ActionDefinitionChoice->choice.actionDefinition_Format1.actionParameter_List = pE2_E2SM_gNB_X2_actionDefinition__actionParameter_List;
        }
    }
    else if (pRICActionDefinitionChoice->actionDefinitionX2Format2Present) {

        // E2SM-gNB-X2-ActionDefinition-Format2
        pE2_E2SM_gNB_X2_ActionDefinitionChoice->present = E2_E2SM_gNB_X2_ActionDefinitionChoice_PR_actionDefinition_Format2;

        if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupCount > 0) {
            struct E2_E2SM_gNB_X2_ActionDefinition_Format2__ranUEgroup_List* pE2_E2SM_gNB_X2_ActionDefinition_Format2__ranUEgroup_List =
            calloc(1, sizeof(struct E2_E2SM_gNB_X2_ActionDefinition_Format2__ranUEgroup_List));

            uint64_t index = 0;
            while (index < pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupCount && index < E2_maxofUEgroup) {

                E2_RANueGroup_Item_t* pE2_RANueGroup_Item = calloc(1, sizeof(E2_RANueGroup_Item_t));
                if (pE2_RANueGroup_Item) {

                    struct E2_RANueGroupDefinition__ranUEgroupDef_List* pE2_RANueGroupDefinition__ranUEgroupDef_List =
                    calloc(1, sizeof (struct E2_RANueGroupDefinition__ranUEgroupDef_List));

                    pE2_RANueGroup_Item->ranUEgroupID = pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupID;

                    uint64_t index2 = 0;
                    while (index2 < pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefCount && index2 < E2_maxofRANParameters) {
                        E2_RANueGroupDef_Item_t* pE2_RANueGroupDef_Item = calloc(1, sizeof(E2_RANueGroupDef_Item_t));
                        if(pE2_RANueGroupDef_Item) {
                            pE2_RANueGroupDef_Item->ranParameter_ID = pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterID;
                            pE2_RANueGroupDef_Item->ranParameter_Test = pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterTest;
                            if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueIntPresent) {
                                pE2_RANueGroupDef_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueInt;
                                pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueInt =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueInt;
                            }
                            else if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueEnum) {
                                pE2_RANueGroupDef_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueEnum;
                                pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueEnum =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueEnum;
                            }
                            else if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBoolPresent) {
                                pE2_RANueGroupDef_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueBool;
                                pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueBool =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBool;
                            }
                            else if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBitSPresent) {
                                pE2_RANueGroupDef_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueBitS;
                                pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueBitS.size =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBitS.byteLength;
                                pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueBitS.bits_unused =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBitS.unusedBits;
                                pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueBitS.buf =
                                calloc(pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBitS.byteLength, 1);
                                if (pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueBitS.buf) {
                                    memcpy(pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueBitS.buf,
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBitS.data,
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBitS.byteLength);
                                }
                                else
                                    return e2err_RICSubscriptionRequestAllocactionRanParameterValueValueBitSFail;
                            }
                            else if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueOctSPresent) {
                                pE2_RANueGroupDef_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueOctS;
                                pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueOctS.size =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueOctS.length;
                                pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueOctS.buf =
                                calloc(pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueOctS.length, 1);
                                if (pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueOctS.buf) {
                                    memcpy(pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueOctS.buf,
                                            pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueOctS.data,
                                            pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueOctS.length);
                                }
                                else
                                    return e2err_RICSubscriptionRequestAllocactionRanParameterValueValueOctSFail;
                            }
                            else if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valuePrtSPresent) {
                                pE2_RANueGroupDef_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valuePrtS;
                                pE2_RANueGroupDef_Item->ranParameter_Value.choice.valuePrtS.size =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valuePrtS.length;
                                pE2_RANueGroupDef_Item->ranParameter_Value.choice.valuePrtS.buf =
                                calloc(pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valuePrtS.length, 1);
                                if (pE2_RANueGroupDef_Item->ranParameter_Value.choice.valuePrtS.buf) {
                                    memcpy(pE2_RANueGroupDef_Item->ranParameter_Value.choice.valuePrtS.buf,
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valuePrtS.data,
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valuePrtS.length);
                                }
                                else
                                    return e2err_RICSubscriptionRequestAllocactionRanParameterValueValuePrtsSFail;
                            }
                            else
                                return e2err_RICSubscriptionRequestRanranUeGroupDefItemParameterValueEmptyFail;

                            if ((result = asn_set_add(pE2_RANueGroupDefinition__ranUEgroupDef_List, pE2_RANueGroupDef_Item)) != 0)
                                return e2err_RICSubscriptionRequestAsn_set_addRANueGroupDef_ItemFail;
                            pE2_RANueGroup_Item->ranUEgroupDefinition.ranUEgroupDef_List = pE2_RANueGroupDefinition__ranUEgroupDef_List;
                        }
                        else
                            return e2err_RICSubscriptionRequestAllocE2_RANueGroupDef_ItemFail;
                        index2++;
                    }

                    struct E2_RANimperativePolicy__ranImperativePolicy_List* pE2_RANimperativePolicy__ranImperativePolicy_List =
                    calloc(1, sizeof (struct E2_RANimperativePolicy__ranImperativePolicy_List));

                    uint64_t index3 = 0;
                    while (index3 < pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterCount && index3 < E2_maxofRANParameters) {
                        E2_RANParameter_Item_t* pE2_RANParameter_Item = calloc(1, sizeof(E2_RANParameter_Item_t));
                        if (pE2_RANParameter_Item) {
                            pE2_RANParameter_Item->ranParameter_ID = pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterID;
                            if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueIntPresent) {
                                pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueInt;
                                pE2_RANParameter_Item->ranParameter_Value.choice.valueInt =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueInt;
                            }
                            else if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueEnum) {
                                pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueEnum;
                                pE2_RANParameter_Item->ranParameter_Value.choice.valueEnum =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueEnum;
                            }
                            else if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBoolPresent) {
                                pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueBool;
                                pE2_RANParameter_Item->ranParameter_Value.choice.valueBool =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBool;
                            }
                            else if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitSPresent) {
                                pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueBitS;
                                pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.size =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS.byteLength;
                                pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.bits_unused =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS.unusedBits;
                                pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.buf =
                                calloc(pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS.byteLength, 1);
                                if (pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.buf) {
                                    memcpy(pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.buf,
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS.data,
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS.byteLength);
                                }
                                else
                                    return e2err_RICSubscriptionRequestAllocactionRanParameterValue2ValueBitSFail;
                            }
                            else if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctSPresent) {
                                pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueOctS;
                                pE2_RANParameter_Item->ranParameter_Value.choice.valueOctS.size =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctS.length;
                                pE2_RANParameter_Item->ranParameter_Value.choice.valueOctS.buf =
                                calloc(pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctS.length, 1);
                                if (pE2_RANParameter_Item->ranParameter_Value.choice.valueOctS.buf) {
                                    memcpy(pE2_RANParameter_Item->ranParameter_Value.choice.valueOctS.buf,
                                            pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctS.data,
                                            pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctS.length);
                                }
                                else
                                    return e2err_RICSubscriptionRequestAllocactionRanParameterValue2ValueOctSFail;
                            }
                            else if (pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtSPresent) {
                                pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valuePrtS;
                                pE2_RANParameter_Item->ranParameter_Value.choice.valuePrtS.size =
                                pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtS.length;
                                pE2_RANParameter_Item->ranParameter_Value.choice.valuePrtS.buf =
                                calloc(pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtS.length, 1);
                                if (pE2_RANParameter_Item->ranParameter_Value.choice.valuePrtS.buf) {
                                    memcpy(pE2_RANParameter_Item->ranParameter_Value.choice.valuePrtS.buf,
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtS.data,
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtS.length);
                                }
                                else
                                    return e2err_RICSubscriptionRequestAllocactionRanParameterValue2ValuePrtsSFail;
                            }
                            else
                                return e2err_RICSubscriptionRequestRanParameterItemRanParameterValueEmptyFail;

                            if ((result = asn_set_add(pE2_RANimperativePolicy__ranImperativePolicy_List, pE2_RANParameter_Item)) != 0)
                                return e2err_RICSubscriptionRequestAsn_set_addE2_RANParameter_ItemFail;
                            pE2_RANueGroup_Item->ranPolicy.ranImperativePolicy_List = pE2_RANimperativePolicy__ranImperativePolicy_List;
                        }
                        else
                            return e2err_RICSubscriptionRequestAllocActionDefinitionFail;
                        index3++;
                    }

                    const int result = asn_set_add(pE2_E2SM_gNB_X2_ActionDefinition_Format2__ranUEgroup_List, pE2_RANueGroup_Item);
                    if (result != 0)
                        sprintf(pLogBuffer,"asn_set_add() failed");
                }
                else
                    return e2err_RICSubscriptionRequestAllocRANParameter_ItemFail;
                index++;
            }
            pE2_E2SM_gNB_X2_ActionDefinitionChoice->choice.actionDefinition_Format2.ranUEgroup_List = pE2_E2SM_gNB_X2_ActionDefinition_Format2__ranUEgroup_List;
        }
    }
    else {
        return e2err_RICSubscriptionRequestRICActionDefinitionEmptyE2_E2SM_gNB_X2_actionDefinition;
    }

    // Debug print
    if (debug)
        asn_fprint(stdout, &asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice);

    // Encode
    size_t bufferSize = sizeof(pRICActionDefinitionChoice->octetString.data);
    asn_enc_rval_t rval;
    rval = asn_encode_to_buffer(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice,
                                pRICActionDefinitionChoice->octetString.data, bufferSize);
    if(rval.encoded == -1) {
        sprintf(pLogBuffer,"Serialization of %s failed", asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice.name);
        return e2err_RICActionDefinitionChoicePackFail_1;
    }
    else if(rval.encoded > bufferSize) {
       sprintf(pLogBuffer,"Buffer of size %zu is too small for %s, need %zu",bufferSize, asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice.name, rval.encoded);
        return e2err_RICActionDefinitionChoicePackFail_2;
    }
    else
    if (debug)
           sprintf(pLogBuffer,"Successfully encoded %s. Buffer size %zu, encoded size %zu",asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice.name, bufferSize, rval.encoded);

    ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice);

    pRICActionDefinitionChoice->octetString.contentLength = rval.encoded;
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t packActionDefinitionNRTFormat(char* pLogBuffer, RICActionDefinitionChoice_t* pRICActionDefinitionChoice) {

    // E2SM-gNB-NRT-actionDefinition
    E2_E2SM_gNB_NRT_ActionDefinition_t* pE2_E2SM_gNB_NRT_ActionDefinition = calloc(1, sizeof(E2_E2SM_gNB_NRT_ActionDefinition_t));
    if (pE2_E2SM_gNB_NRT_ActionDefinition == NULL)
        return e2err_RICSubscriptionRequestAllocE2_E2SM_gNB_NRT_ActionDefinitionFail;

    if (pRICActionDefinitionChoice->actionDefinitionNRTFormat1Present) {

        // E2SM-gNB-NRT-ActionDefinition-Format1
        pE2_E2SM_gNB_NRT_ActionDefinition->present = E2_E2SM_gNB_NRT_ActionDefinition_PR_actionDefinition_Format1;

        struct E2_E2SM_gNB_NRT_ActionDefinition_Format1__ranParameter_List* pE2_E2SM_gNB_NRT_ActionDefinition_Format1__ranParameter_List =
          calloc(1, sizeof(struct E2_E2SM_gNB_NRT_ActionDefinition_Format1__ranParameter_List));

        uint64_t index = 0;
        while (index <  pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterCount && index < E2_maxofRANParameters) {
            E2_RANParameter_Item_t* pE2_RANParameter_Item = calloc(1, sizeof(E2_RANParameter_Item_t));
            if (pE2_RANParameter_Item) {
                pE2_RANParameter_Item->ranParameter_ID = pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterID;
                if (pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueIntPresent) {
                    pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueInt;
                    pE2_RANParameter_Item->ranParameter_Value.choice.valueInt =
                      pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueInt;
                }
                else if (pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueEnum) {
                    pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueEnum;
                    pE2_RANParameter_Item->ranParameter_Value.choice.valueEnum =
                      pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueEnum;
                }
                else if (pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBoolPresent) {
                    pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueBool;
                    pE2_RANParameter_Item->ranParameter_Value.choice.valueBool =
                      pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBool;
                }
                else if (pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBitSPresent) {
                    pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueBitS;
                    pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.size =
                      pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBitS.byteLength;
                    pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.bits_unused =
                      pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBitS.unusedBits;
                    pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.buf =
                      calloc(pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBitS.byteLength, 1);
                    if (pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.buf) {
                        memcpy(pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.buf,
                               pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBitS.data,
                               pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBitS.byteLength);
                      }
                      else
                        return e2err_RICSubscriptionRequestAllocactionNRTRanParameterValue2ValueBitSFail;
                }
                else if (pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueOctSPresent) {
                    pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valueOctS;
                    pE2_RANParameter_Item->ranParameter_Value.choice.valueOctS.size =
                      pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueOctS.length;
                      pE2_RANParameter_Item->ranParameter_Value.choice.valueOctS.buf =
                      calloc(pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueOctS.length, 1);
                      if (pE2_RANParameter_Item->ranParameter_Value.choice.valueOctS.buf) {
                          memcpy(pE2_RANParameter_Item->ranParameter_Value.choice.valueOctS.buf,
                                 pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueOctS.data,
                                 pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueOctS.length);
                      }
                      else
                        return e2err_RICSubscriptionRequestAllocactionNRTRanParameterValue2ValueOctSFail;
                }
                else if (pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valuePrtSPresent) {
                    pE2_RANParameter_Item->ranParameter_Value.present = E2_RANParameter_Value_PR_valuePrtS;
                    pE2_RANParameter_Item->ranParameter_Value.choice.valuePrtS.size =
                      pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valuePrtS.length;
                    pE2_RANParameter_Item->ranParameter_Value.choice.valuePrtS.buf =
                      calloc(pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valuePrtS.length, 1);
                    if (pE2_RANParameter_Item->ranParameter_Value.choice.valuePrtS.buf) {
                        memcpy(pE2_RANParameter_Item->ranParameter_Value.choice.valuePrtS.buf,
                               pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valuePrtS.data,
                               pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valuePrtS.length);
                      }
                      else
                        return e2err_RICSubscriptionRequestAllocactionNRTRanParameterValue2ValuePrtsSFail;
                }
                else
                    return e2err_RICSubscriptionRequestRanParameterItemNRTRanParameterValueEmptyFail;

                int result;
                if ((result = asn_set_add(pE2_E2SM_gNB_NRT_ActionDefinition_Format1__ranParameter_List, pE2_RANParameter_Item)) != 0)
                    return e2err_RICSubscriptionRequestAsn_set_addE2_NRTRANParameter_ItemFail;
                pE2_E2SM_gNB_NRT_ActionDefinition->choice.actionDefinition_Format1.ranParameter_List = pE2_E2SM_gNB_NRT_ActionDefinition_Format1__ranParameter_List;
            }
            else
                return e2err_RICSubscriptionRequestAllocNRTRANParameter_ItemFail;
            index++;
        }
    }
    else
        return e2err_RICSubscriptionRequestRICActionDefinitionEmptyE2_E2SM_gNB_NRT_actionDefinition;

    // Debug print
    if (debug)
        asn_fprint(stdout, &asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition, pE2_E2SM_gNB_NRT_ActionDefinition);

    // Encode
    size_t bufferSize = sizeof(pRICActionDefinitionChoice->octetString.data);
    asn_enc_rval_t rval;
    rval = asn_encode_to_buffer(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition, pE2_E2SM_gNB_NRT_ActionDefinition,
                                pRICActionDefinitionChoice->octetString.data, bufferSize);
    if(rval.encoded == -1) {
        sprintf(pLogBuffer,"Serialization of %s failed", asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition.name);
        return e2err_RICActionDefinitionChoicePackFail_1;
    }
    else if(rval.encoded > bufferSize) {
       sprintf(pLogBuffer,"Buffer of size %zu is too small for %s, need %zu",bufferSize, asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition.name, rval.encoded);
        return e2err_RICActionDefinitionChoicePackFail_2;
    }
    else
    if (debug)
           sprintf(pLogBuffer,"Successfully encoded %s. Buffer size %zu, encoded size %zu",asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition.name, bufferSize, rval.encoded);

    ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition, pE2_E2SM_gNB_NRT_ActionDefinition);

    pRICActionDefinitionChoice->octetString.contentLength = rval.encoded;
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t packRICSubscriptionResponse(size_t* pDataBufferSize, byte* pDataBuffer, char* pLogBuffer, RICSubscriptionResponse_t* pRICSubscriptionResponse) {

    E2AP_PDU_t* pE2AP_PDU = calloc(1, sizeof(E2AP_PDU_t));
    if(pE2AP_PDU)
	{
        pE2AP_PDU->present = E2AP_PDU_PR_successfulOutcome;
        pE2AP_PDU->choice.initiatingMessage.procedureCode = ProcedureCode_id_RICsubscription;
        pE2AP_PDU->choice.initiatingMessage.criticality = Criticality_ignore;
        pE2AP_PDU->choice.initiatingMessage.value.present = SuccessfulOutcome__value_PR_RICsubscriptionResponse;

        // RICrequestID
        RICsubscriptionResponse_IEs_t* pRICsubscriptionResponse_IEs = calloc(1, sizeof(RICsubscriptionResponse_IEs_t));
        if (pRICsubscriptionResponse_IEs) {
            pRICsubscriptionResponse_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionResponse_IEs->criticality = Criticality_reject;
            pRICsubscriptionResponse_IEs->value.present = RICsubscriptionResponse_IEs__value_PR_RICrequestID;
            pRICsubscriptionResponse_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionResponse->ricRequestID.ricRequestorID;
            pRICsubscriptionResponse_IEs->value.choice.RICrequestID.ricInstanceID = pRICSubscriptionResponse->ricRequestID.ricInstanceID;
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

        // RICaction-NotAdmitted list, OPTIONAL
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
                        if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_ricRequest) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_ricRequest;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.ricRequest =
                              pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal;
                        }
                        else if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_ricService) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_ricService;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.ricService =
                              pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal;
                        }
                        else if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_transport) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_transport;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.transport =
                              pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal;
                        }
                        else if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_protocol) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_protocol;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.protocol =
                              pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal;
                        }
                        else if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_misc) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_misc;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.misc =
                              pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal;
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
        pE2AP_PDU->choice.unsuccessfulOutcome.procedureCode = ProcedureCode_id_RICsubscription;
        pE2AP_PDU->choice.unsuccessfulOutcome.criticality = Criticality_ignore;
        pE2AP_PDU->choice.unsuccessfulOutcome.value.present = UnsuccessfulOutcome__value_PR_RICsubscriptionFailure;

        // RICrequestID
        RICsubscriptionFailure_IEs_t* pRICsubscriptionFailure_IEs = calloc(1, sizeof(RICsubscriptionFailure_IEs_t));
        if (pRICsubscriptionFailure_IEs) {
            pRICsubscriptionFailure_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionFailure_IEs->criticality = Criticality_reject;
            pRICsubscriptionFailure_IEs->value.present = RICsubscriptionFailure_IEs__value_PR_RICrequestID;
            pRICsubscriptionFailure_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionFailure->ricRequestID.ricRequestorID;
            pRICsubscriptionFailure_IEs->value.choice.RICrequestID.ricInstanceID = pRICSubscriptionFailure->ricRequestID.ricInstanceID;
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
                    if (pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_ricRequest) {
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_ricRequest;
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.ricRequest =
                          pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal;
                    }
                    else if (pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_ricService) {
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_ricService;
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.ricService =
                          pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal;
                    }
                    else if (pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_transport) {
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_transport;
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.transport =
                          pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal;
                    }
                    else if (pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_protocol) {
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_protocol;
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.protocol =
                          pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal;
                    }
                    else if (pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_misc) {
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_misc;
                        pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.misc =
                          pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal;
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
uint64_t packRICSubscriptionDeleteRequest(size_t* pDataBufferSize, byte* pDataBuffer, char* pLogBuffer, RICSubscriptionDeleteRequest_t* pRICSubscriptionDeleteRequest) {

    E2AP_PDU_t* pE2AP_PDU = calloc(1, sizeof(E2AP_PDU_t));
    if(pE2AP_PDU)
	{
        pE2AP_PDU->present = E2AP_PDU_PR_initiatingMessage;
        pE2AP_PDU->choice.initiatingMessage.procedureCode = ProcedureCode_id_RICsubscriptionDelete;
        pE2AP_PDU->choice.initiatingMessage.criticality = Criticality_ignore;
        pE2AP_PDU->choice.initiatingMessage.value.present = InitiatingMessage__value_PR_RICsubscriptionDeleteRequest;

        // RICrequestID
        RICsubscriptionDeleteRequest_IEs_t* pRICsubscriptionDeleteRequest_IEs = calloc(1, sizeof(RICsubscriptionDeleteRequest_IEs_t));
        if (pRICsubscriptionDeleteRequest_IEs) {
            pRICsubscriptionDeleteRequest_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionDeleteRequest_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteRequest_IEs->value.present = RICsubscriptionDeleteRequest_IEs__value_PR_RICrequestID;
            pRICsubscriptionDeleteRequest_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionDeleteRequest->ricRequestID.ricRequestorID;
            pRICsubscriptionDeleteRequest_IEs->value.choice.RICrequestID.ricInstanceID = pRICSubscriptionDeleteRequest->ricRequestID.ricInstanceID;
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
        pE2AP_PDU->choice.successfulOutcome.procedureCode = ProcedureCode_id_RICsubscriptionDelete;
        pE2AP_PDU->choice.successfulOutcome.criticality = Criticality_ignore;
        pE2AP_PDU->choice.successfulOutcome.value.present = SuccessfulOutcome__value_PR_RICsubscriptionDeleteResponse;

        // RICrequestID
        RICsubscriptionDeleteResponse_IEs_t* pRICsubscriptionDeleteResponse_IEs = calloc(1, sizeof(RICsubscriptionDeleteResponse_IEs_t));
        if (pRICsubscriptionDeleteResponse_IEs) {
            pRICsubscriptionDeleteResponse_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionDeleteResponse_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteResponse_IEs->value.present = RICsubscriptionDeleteResponse_IEs__value_PR_RICrequestID;
            pRICsubscriptionDeleteResponse_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionDeleteResponse->ricRequestID.ricRequestorID;
            pRICsubscriptionDeleteResponse_IEs->value.choice.RICrequestID.ricInstanceID = pRICSubscriptionDeleteResponse->ricRequestID.ricInstanceID;
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
        pE2AP_PDU->choice.unsuccessfulOutcome.procedureCode = ProcedureCode_id_RICsubscriptionDelete;
        pE2AP_PDU->choice.unsuccessfulOutcome.criticality = Criticality_ignore;
        pE2AP_PDU->choice.unsuccessfulOutcome.value.present = UnsuccessfulOutcome__value_PR_RICsubscriptionDeleteFailure;

        // RICrequestID
        RICsubscriptionDeleteFailure_IEs_t* pRICsubscriptionDeleteFailure_IEs = calloc(1, sizeof(RICsubscriptionDeleteFailure_IEs_t));
        if (pRICsubscriptionDeleteFailure_IEs) {
            pRICsubscriptionDeleteFailure_IEs->id = ProtocolIE_ID_id_RICrequestID;
            pRICsubscriptionDeleteFailure_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteFailure_IEs->value.present = RICsubscriptionDeleteFailure_IEs__value_PR_RICrequestID;
            pRICsubscriptionDeleteFailure_IEs->value.choice.RICrequestID.ricRequestorID = pRICSubscriptionDeleteFailure->ricRequestID.ricRequestorID;
            pRICsubscriptionDeleteFailure_IEs->value.choice.RICrequestID.ricInstanceID = pRICSubscriptionDeleteFailure->ricRequestID.ricInstanceID;
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
            pRICsubscriptionDeleteFailure_IEs->id = ProtocolIE_ID_id_Cause;
            pRICsubscriptionDeleteFailure_IEs->criticality = Criticality_reject;
            pRICsubscriptionDeleteFailure_IEs->value.present = RICsubscriptionDeleteFailure_IEs__value_PR_Cause;
            if (pRICSubscriptionDeleteFailure->cause.content == Cause_PR_ricRequest) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present = Cause_PR_ricRequest;
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.ricRequest =
                  pRICSubscriptionDeleteFailure->cause.causeVal;
            }
            else if (pRICSubscriptionDeleteFailure->cause.content == Cause_PR_ricService) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present = Cause_PR_ricService;
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.ricService =
                  pRICSubscriptionDeleteFailure->cause.causeVal;
            }
            else if (pRICSubscriptionDeleteFailure->cause.content == Cause_PR_transport) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present = Cause_PR_transport;
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.transport =
                  pRICSubscriptionDeleteFailure->cause.causeVal;
            }
            else if (pRICSubscriptionDeleteFailure->cause.content == Cause_PR_protocol) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present = Cause_PR_protocol;
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.protocol =
                  pRICSubscriptionDeleteFailure->cause.causeVal;
            }
            else if (pRICSubscriptionDeleteFailure->cause.content == Cause_PR_misc) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present = Cause_PR_misc;
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.misc =
                  pRICSubscriptionDeleteFailure->cause.causeVal;
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
            sprintf(pLogBuffer,"Successfully decoded E2AP-PDU");
            asn_fprint(stdout, &asn_DEF_E2AP_PDU, pE2AP_PDU);
        }

        if (pE2AP_PDU->present == E2AP_PDU_PR_initiatingMessage) {
            if (pE2AP_PDU->choice.initiatingMessage.procedureCode == ProcedureCode_id_RICsubscription) {
                if (pE2AP_PDU->choice.initiatingMessage.value.present == InitiatingMessage__value_PR_RICsubscriptionRequest) {
                    pMessageInfo->messageType = cE2InitiatingMessage;
                    pMessageInfo->messageId = cRICSubscriptionRequest;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"Error. Not supported initiatingMessage MessageId = %u",pE2AP_PDU->choice.initiatingMessage.value.present);
                    return 0;
                }
            }
            else if (pE2AP_PDU->choice.initiatingMessage.procedureCode == ProcedureCode_id_RICsubscriptionDelete) {
                if (pE2AP_PDU->choice.initiatingMessage.value.present == InitiatingMessage__value_PR_RICsubscriptionDeleteRequest) {
                    pMessageInfo->messageType = cE2InitiatingMessage;
                    pMessageInfo->messageId = cRICSubscriptionDeleteRequest;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"Error. Not supported initiatingMessage MessageId = %u",pE2AP_PDU->choice.initiatingMessage.value.present);
                    return 0;
                }
            }
            else {
                sprintf(pLogBuffer,"Error. Procedure not supported. ProcedureCode = %li",pE2AP_PDU->choice.initiatingMessage.procedureCode);
                return 0;
            }
        }
        else if (pE2AP_PDU->present == E2AP_PDU_PR_successfulOutcome) {
            if (pE2AP_PDU->choice.successfulOutcome.procedureCode == ProcedureCode_id_RICsubscription) {
                if (pE2AP_PDU->choice.successfulOutcome.value.present == SuccessfulOutcome__value_PR_RICsubscriptionResponse) {
                    pMessageInfo->messageType = cE2SuccessfulOutcome;
                    pMessageInfo->messageId = cRICSubscriptionResponse;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"Error. Not supported successfulOutcome MessageId = %u",pE2AP_PDU->choice.successfulOutcome.value.present);
                    return 0;
                }
            }
            else if (pE2AP_PDU->choice.successfulOutcome.procedureCode == ProcedureCode_id_RICsubscriptionDelete) {
                if (pE2AP_PDU->choice.successfulOutcome.value.present == SuccessfulOutcome__value_PR_RICsubscriptionDeleteResponse) {
                    pMessageInfo->messageType = cE2SuccessfulOutcome;
                    pMessageInfo->messageId = cRICsubscriptionDeleteResponse;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"Error. Not supported successfulOutcome MessageId = %u",pE2AP_PDU->choice.successfulOutcome.value.present);
                    return 0;
                }
            }
            else {
                sprintf(pLogBuffer,"Error. Procedure not supported. ProcedureCode = %li",pE2AP_PDU->choice.successfulOutcome.procedureCode);
                return 0;
            }
        }
        else if (pE2AP_PDU->present == E2AP_PDU_PR_unsuccessfulOutcome) {
            if (pE2AP_PDU->choice.unsuccessfulOutcome.procedureCode == ProcedureCode_id_RICsubscription) {
                if (pE2AP_PDU->choice.unsuccessfulOutcome.value.present == UnsuccessfulOutcome__value_PR_RICsubscriptionFailure) {
                    pMessageInfo->messageType = cE2UnsuccessfulOutcome;
                    pMessageInfo->messageId = cRICSubscriptionFailure;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"Error. Not supported unsuccessfulOutcome MessageId = %u",pE2AP_PDU->choice.unsuccessfulOutcome.value.present);
                    return 0;
                }
            }
            else if (pE2AP_PDU->choice.unsuccessfulOutcome.procedureCode == ProcedureCode_id_RICsubscriptionDelete) {
                if (pE2AP_PDU->choice.unsuccessfulOutcome.value.present == UnsuccessfulOutcome__value_PR_RICsubscriptionDeleteFailure) {
                    pMessageInfo->messageType = cE2UnsuccessfulOutcome;
                    pMessageInfo->messageId = cRICsubscriptionDeleteFailure;
                    return (e2ap_pdu_ptr_t*)pE2AP_PDU;
                }
                else {
                    sprintf(pLogBuffer,"Error. Not supported unsuccessfulOutcome MessageId = %u",pE2AP_PDU->choice.unsuccessfulOutcome.value.present);
                    return 0;
                }
            }
        }
        else
            sprintf(pLogBuffer,"Decode failed. Invalid message type %u",pE2AP_PDU->present);
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return 0;
    case RC_WMORE:
        sprintf(pLogBuffer,"Decode failed. More data needed. Buffer size %zu, %s, consumed %zu",dataBufferSize, asn_DEF_E2AP_PDU.name, rval.consumed);
        return 0;
    case RC_FAIL:
        sprintf(pLogBuffer,"Decode failed. Buffer size %zu, %s, consumed %zu",dataBufferSize, asn_DEF_E2AP_PDU.name, rval.consumed);
        return 0;
    default:
        return 0;
    }
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICSubscriptionRequestData(mem_track_hdr_t * pDynMemHead, e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionRequest_t* pRICSubscriptionRequest) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;

    RICsubscriptionRequest_t *asnRicSubscriptionRequest = &pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionRequest;
    RICsubscriptionRequest_IEs_t* pRICsubscriptionRequest_IEs;

    // RICrequestID
    if (asnRicSubscriptionRequest->protocolIEs.list.count > 0 &&
        asnRicSubscriptionRequest->protocolIEs.list.array[0]->id == ProtocolIE_ID_id_RICrequestID) {
        pRICsubscriptionRequest_IEs = asnRicSubscriptionRequest->protocolIEs.list.array[0];
        pRICSubscriptionRequest->ricRequestID.ricRequestorID = pRICsubscriptionRequest_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionRequest->ricRequestID.ricInstanceID = pRICsubscriptionRequest_IEs->value.choice.RICrequestID.ricInstanceID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionRequestRICrequestIDMissing;
    }

    // RANfunctionID
    if (asnRicSubscriptionRequest->protocolIEs.list.count > 1 &&
        asnRicSubscriptionRequest->protocolIEs.list.array[1]->id == ProtocolIE_ID_id_RANfunctionID) {
        pRICsubscriptionRequest_IEs = asnRicSubscriptionRequest->protocolIEs.list.array[1];
        pRICSubscriptionRequest->ranFunctionID = pRICsubscriptionRequest_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionRequestRANfunctionIDMissing;
    }

    // RICsubscriptionDetails
    if (asnRicSubscriptionRequest->protocolIEs.list.count > 2 &&
        asnRicSubscriptionRequest->protocolIEs.list.array[2]->id == ProtocolIE_ID_id_RICsubscriptionDetails) {
        pRICsubscriptionRequest_IEs = asnRicSubscriptionRequest->protocolIEs.list.array[2];

        // Unpack EventTriggerDefinition
        RICeventTriggerDefinition_t* pRICeventTriggerDefinition =
          (RICeventTriggerDefinition_t*)&pRICsubscriptionRequest_IEs->value.choice.RICsubscriptionDetails.ricEventTriggerDefinition;
        pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.contentLength = pRICeventTriggerDefinition->size;
        memcpy(pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.octetString.data, pRICeventTriggerDefinition->buf, pRICeventTriggerDefinition->size);

        // Workaround to spec problem. E2AP spec does not specify what speck (gNB-X2 or gNB-NRT) should be used when decoded EventTriggerDefinition and ActionDefinition
        // received received. Here we know that length of gNB-NRT EventTriggerDefinition octet string is always 1 at the moment.
        if (pRICeventTriggerDefinition->size == 1) {
            pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.E2SMgNBNRTEventTriggerDefinitionPresent = true;
            pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.E2SMgNBX2EventTriggerDefinitionPresent = false;
        }
        else {
            pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.E2SMgNBNRTEventTriggerDefinitionPresent = false;
            pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.E2SMgNBX2EventTriggerDefinitionPresent = true;
        }

        uint64_t returnCode;
        if ((returnCode = getRICEventTriggerDefinitionData(&pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition)) != e2err_OK) {
            ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
            return returnCode;
        }

        // RICactions-ToBeSetup-List
        RICaction_ToBeSetup_ItemIEs_t* pRICaction_ToBeSetup_ItemIEs;
        uint64_t index = 0;
        while (index < pRICsubscriptionRequest_IEs->value.choice.RICsubscriptionDetails.ricAction_ToBeSetup_List.list.count)
        {
            pRICaction_ToBeSetup_ItemIEs = (RICaction_ToBeSetup_ItemIEs_t*)pRICsubscriptionRequest_IEs->value.choice.RICsubscriptionDetails.ricAction_ToBeSetup_List.list.array[index];

            // RICActionID
            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID =
              pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionID;

            // RICActionType
            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType =
              pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionType;

            // RICactionDefinition, OPTIONAL
            if (pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition)
            {
                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.contentLength =
                pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition->size;
                memcpy(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.octetString.data,
                       pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition->buf,
                       pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionDefinition->size);

                // Workaround to spec problem. E2AP spec does not specify what speck (gNB-X2 or gNB-NRT) should be used when decoded EventTriggerDefinition and ActionDefinition
                // received received. Here we know that length of gNB-NRT EventTriggerDefinition octet string is always 1 at the moment.
                if (pRICeventTriggerDefinition->size == 1)
                    pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionNRTFormat1Present = true;
                else
                    pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionNRTFormat1Present = false;

                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = true;
                if ((returnCode = getRICActionDefinitionData(pDynMemHead, &pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice)) != e2err_OK) {
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
                    return returnCode;
                }
            }
            else
                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = false;

            // RICsubsequentAction, OPTIONAL
            RICsubsequentAction_t* pRICsubsequentAction;
            if (pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricSubsequentAction)
            {
                pRICsubsequentAction = pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricSubsequentAction;
                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent = true;
                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType =
                  pRICsubsequentAction->ricSubsequentActionType;
                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait =
                  pRICsubsequentAction->ricTimeToWait;
            }
            else
                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent = false;
            index++;
        }
        pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength = index;
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

    if (pRICEventTriggerDefinition->E2SMgNBX2EventTriggerDefinitionPresent)
        return getRICEventTriggerDefinitionDataX2Format(pRICEventTriggerDefinition);
    else if (pRICEventTriggerDefinition->E2SMgNBNRTEventTriggerDefinitionPresent)
        return getRICEventTriggerDefinitionDataNRTFormat(pRICEventTriggerDefinition);
    else
        return e2err_RICEventTriggerDefinitionEmptyDecodeDefaultFail;
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICEventTriggerDefinitionDataX2Format(RICEventTriggerDefinition_t* pRICEventTriggerDefinition) {

    E2_E2SM_gNB_X2_eventTriggerDefinition_t* pE2SM_gNB_X2_eventTriggerDefinition = 0;
    asn_dec_rval_t rval;
    rval = asn_decode(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition, (void **)&pE2SM_gNB_X2_eventTriggerDefinition,
                      pRICEventTriggerDefinition->octetString.data, pRICEventTriggerDefinition->octetString.contentLength);
    switch(rval.code) {
    case RC_OK:
        // Debug print
        if (debug) {
            printf("Successfully decoded E2SM_gNB_X2_eventTriggerDefinition\n");
            asn_fprint(stdout, &asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
        }

        // InterfaceID, GlobalENB-ID or GlobalGNB-ID
        if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.present == E2_Interface_ID_PR_global_eNB_ID) {

            // GlobalENB-ID
            pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBIDPresent = true;

            // PLMN-Identity
            pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.contentLength =
              pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.size;
            memcpy(pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal,
              pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf,
              pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.contentLength);

            //  ENB-ID
            IdOctects_t eNBOctects;
            memset(eNBOctects.octets, 0, sizeof(eNBOctects));
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present == E2_ENB_ID_PR_macro_eNB_ID) {
                // BIT STRING, SIZE 20
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits = cMacroENBIDP_20Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.buf,
                  pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.macro_eNB_ID.size);
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present == E2_ENB_ID_PR_home_eNB_ID) {
                // BIT STRING, SIZE 28
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits = cHomeENBID_28Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.buf,
                  pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.home_eNB_ID.size);
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present == E2_ENB_ID_PR_short_Macro_eNB_ID) {
                // BIT STRING, SIZE 18
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits = cShortMacroENBID_18Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.buf,
                  pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.short_Macro_eNB_ID.size);
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.present == E2_ENB_ID_PR_long_Macro_eNB_ID) {
                // BIT STRING, SIZE 21
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits =  clongMacroENBIDP_21Bits;
                memcpy(eNBOctects.octets,pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.buf,
                  pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.eNB_ID.choice.long_Macro_eNB_ID.size);
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
            }
            else {
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBIDPresent = false;
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBIDPresent = false;
                ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
                return e2err_RICEventTriggerDefinitionIEValueFail_5;
            }
        }
        else if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.present == E2_Interface_ID_PR_global_gNB_ID) {
            // GlobalGNB-ID
            pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBIDPresent = true;

            // PLMN-Identity
            pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.contentLength =
              pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.size;
            memcpy(pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.pLMNIdentityVal,
              pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_eNB_ID.pLMN_Identity.buf,
              pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.pLMNIdentity.contentLength);

            // GNB-ID
            IdOctects_t gNBOctects;
            memset(gNBOctects.octets, 0, sizeof(gNBOctects));
            if (pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.present == E2_GNB_ID_PR_gNB_ID) {
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.nodeID.bits = pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.size;
                memcpy(gNBOctects.octets, pE2SM_gNB_X2_eventTriggerDefinition->interface_ID.choice.global_gNB_ID.gNB_ID.choice.gNB_ID.buf,
                   pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.nodeID.bits);
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBID.nodeID.nodeID = gNBOctects.nodeID;
            }
            else {
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBIDPresent = false;
                pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBIDPresent = false;
                ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
                return e2err_RICEventTriggerDefinitionIEValueFail_6;
            }
        }
        else {
            pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBIDPresent = false;
            pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBIDPresent = false;
            ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
            return e2err_RICEventTriggerDefinitionIEValueFail_7;
        }

        // InterfaceDirection
        pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceDirection = pE2SM_gNB_X2_eventTriggerDefinition->interfaceDirection;

        // InterfaceMessageType
        pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceMessageType.procedureCode = pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.procedureCode;

        if (pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage == E2_TypeOfMessage_initiating_message)
            pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceMessageType.typeOfMessage = cE2InitiatingMessage;
        else if (pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage == E2_TypeOfMessage_successful_outcome)
            pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceMessageType.typeOfMessage = cE2SuccessfulOutcome;
        else if (pE2SM_gNB_X2_eventTriggerDefinition->interfaceMessageType.typeOfMessage == E2_TypeOfMessage_unsuccessful_outcome)
            pRICEventTriggerDefinition->e2SMgNBX2eventTriggerDefinition.interfaceMessageType.typeOfMessage = cE2UnsuccessfulOutcome;
        else {
            ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
            return e2err_RICEventTriggerDefinitionIEValueFail_8;
        }
        ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition, pE2SM_gNB_X2_eventTriggerDefinition);
        return e2err_OK;
    case RC_WMORE:
        if (debug)
            printf("Decode failed. More data needed. Buffer size %zu, %s, consumed %zu\n",pRICEventTriggerDefinition->octetString.contentLength,
                   asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition.name, rval.consumed);

        return e2err_RICEventTriggerDefinitionDecodeWMOREFail;
    case RC_FAIL:
        if (debug)
            printf("Decode failed. Buffer size %zu, %s, consumed %zu\n",pRICEventTriggerDefinition->octetString.contentLength,
                   asn_DEF_E2_E2SM_gNB_X2_eventTriggerDefinition.name, rval.consumed);
        return e2err_RICEventTriggerDefinitionDecodeFAIL;
    default:
        return e2err_RICEventTriggerDefinitionDecodeDefaultFail;
    }
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICEventTriggerDefinitionDataNRTFormat(RICEventTriggerDefinition_t* pRICEventTriggerDefinition) {

    E2_E2SM_gNB_NRT_EventTriggerDefinition_t* pE2_E2SM_gNB_NRT_EventTriggerDefinition = 0;
    asn_dec_rval_t rval;
    rval = asn_decode(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition, (void **)&pE2_E2SM_gNB_NRT_EventTriggerDefinition,
                      pRICEventTriggerDefinition->octetString.data, pRICEventTriggerDefinition->octetString.contentLength);
    switch(rval.code) {
    case RC_OK:
        // Debug print
        if (debug) {
            printf("Successfully decoded E2SM_gNB_X2_eventTriggerDefinition\n");
            asn_fprint(stdout, &asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition, pE2_E2SM_gNB_NRT_EventTriggerDefinition);
        }

        // NRT-TriggerNature
        if (pE2_E2SM_gNB_NRT_EventTriggerDefinition->present == E2_E2SM_gNB_NRT_EventTriggerDefinition_PR_eventDefinition_Format1) {
            pRICEventTriggerDefinition->E2SMgNBNRTEventTriggerDefinitionPresent = true;
            pRICEventTriggerDefinition->e2SMgNBNRTEventTriggerDefinition.eventDefinitionFormat1.triggerNature =
              pE2_E2SM_gNB_NRT_EventTriggerDefinition->choice.eventDefinition_Format1.triggerNature;
        }
        ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition, pE2_E2SM_gNB_NRT_EventTriggerDefinition);
        return e2err_OK;
    case RC_WMORE:
        if (debug)
            printf("Decode failed. More data needed. Buffer size %zu, %s, consumed %zu\n",pRICEventTriggerDefinition->octetString.contentLength,
                   asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition.name, rval.consumed);

        return e2err_RICNRTEventTriggerDefinitionDecodeWMOREFail;
    case RC_FAIL:
        if (debug)
            printf("Decode failed. Buffer size %zu, %s, consumed %zu\n",pRICEventTriggerDefinition->octetString.contentLength,
                   asn_DEF_E2_E2SM_gNB_NRT_EventTriggerDefinition.name, rval.consumed);
        return e2err_RICNRTEventTriggerDefinitionDecodeFAIL;
    default:
        return e2err_RICNRTEventTriggerDefinitionDecodeDefaultFail;
    }
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICActionDefinitionData(mem_track_hdr_t *pDynMemHead, RICActionDefinitionChoice_t* pRICActionDefinitionChoice) {

    if (pRICActionDefinitionChoice->actionDefinitionNRTFormat1Present)
        return getRICActionDefinitionDataNRTFormat(pDynMemHead, pRICActionDefinitionChoice);
//    if (pRICActionDefinitionChoice->actionDefinitionX2Format1Present || pRICActionDefinitionChoice->actionDefinitionX2Format2Present)
    else
        return getRICActionDefinitionDataX2Format(pDynMemHead, pRICActionDefinitionChoice);
//    else
//        return e2err_RICActionDefinitionChoiceEmptyFAIL;
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICActionDefinitionDataX2Format(mem_track_hdr_t* pDynMemHead, RICActionDefinitionChoice_t* pRICActionDefinitionChoice) {

    E2_E2SM_gNB_X2_ActionDefinitionChoice_t* pE2_E2SM_gNB_X2_ActionDefinitionChoice = 0;
    asn_dec_rval_t rval;
    rval = asn_decode(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, (void **)&pE2_E2SM_gNB_X2_ActionDefinitionChoice,
                      pRICActionDefinitionChoice->octetString.data, pRICActionDefinitionChoice->octetString.contentLength);
    switch(rval.code) {
    case RC_OK:
        // Debug print
        if (debug) {
            printf("Successfully decoded E2SM_gNB_X2_ActionDefinitionChoice\n");
            asn_fprint(stdout, &asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice);
        }
        // ActionDefinitionChoice
        if (pE2_E2SM_gNB_X2_ActionDefinitionChoice->present == E2_E2SM_gNB_X2_ActionDefinitionChoice_PR_actionDefinition_Format1) {

            // E2SM-gNB-X2-actionDefinition
            uint64_t status;
            if ((status = allocActionDefinitionX2Format1(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format1)) != e2err_OK)
                return status;
            pRICActionDefinitionChoice->actionDefinitionX2Format1Present = true;
            pRICActionDefinitionChoice->actionDefinitionX2Format2Present = false;
            pRICActionDefinitionChoice->actionDefinitionNRTFormat1Present = false;
            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterCount = 0;
            pRICActionDefinitionChoice->actionDefinitionX2Format1->styleID = pE2_E2SM_gNB_X2_ActionDefinitionChoice->choice.actionDefinition_Format1.style_ID;

            uint64_t index = 0;
            if (pE2_E2SM_gNB_X2_ActionDefinitionChoice->choice.actionDefinition_Format1.actionParameter_List) {
                while (index < pE2_E2SM_gNB_X2_ActionDefinitionChoice->choice.actionDefinition_Format1.actionParameter_List->list.count) {
                    E2_ActionParameter_Item_t* pE2_ActionParameter_Item = pE2_E2SM_gNB_X2_ActionDefinitionChoice->choice.actionDefinition_Format1.actionParameter_List->list.array[index];
                    if (pE2_ActionParameter_Item) {
                        pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].parameterID = pE2_ActionParameter_Item->actionParameter_ID;
                        if (pE2_ActionParameter_Item->actionParameter_Value.present == E2_ActionParameter_Value_PR_valueInt) {
                            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueIntPresent = true;
                            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueInt =
                            pE2_ActionParameter_Item->actionParameter_Value.choice.valueInt;
                        }
                        else if (pE2_ActionParameter_Item->actionParameter_Value.present == E2_ActionParameter_Value_PR_valueEnum) {
                            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueEnumPresent = true;
                            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueIntPresent =
                            pE2_ActionParameter_Item->actionParameter_Value.choice.valueEnum;
                        }
                        else if (pE2_ActionParameter_Item->actionParameter_Value.present == E2_ActionParameter_Value_PR_valueBool) {
                            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBoolPresent = true;
                            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBool =
                            pE2_ActionParameter_Item->actionParameter_Value.choice.valueBool;
                        }
                        else if (pE2_ActionParameter_Item->actionParameter_Value.present == E2_ActionParameter_Value_PR_valueBitS) {
                            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBitSPresent = true;
                            addBitString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueBitS,
                                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueBitS.size,
                                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueBitS.buf,
                                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueBitS.bits_unused);
                        }
                        else if (pE2_ActionParameter_Item->actionParameter_Value.present == E2_ActionParameter_Value_PR_valueOctS) {
                            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueOctSPresent = true;
                            addOctetString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valueOctS,
                                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueOctS.size,
                                        pE2_ActionParameter_Item->actionParameter_Value.choice.valueOctS.buf);
                        }
                        else if (pE2_ActionParameter_Item->actionParameter_Value.present == E2_ActionParameter_Value_PR_valuePrtS) {
                            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valuePrtSPresent = true;
                            addOctetString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterItem[index].actionParameterValue.valuePrtS,
                                        pE2_ActionParameter_Item->actionParameter_Value.choice.valuePrtS.size,
                                        pE2_ActionParameter_Item->actionParameter_Value.choice.valuePrtS.buf);
                        }
                        else {
                            ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice);
                            return e2err_RICSubscriptionRequestActionParameterItemFail;
                        }
                    }
                    index++;
                }
            }
            pRICActionDefinitionChoice->actionDefinitionX2Format1->actionParameterCount = index;
        }
        else if (pE2_E2SM_gNB_X2_ActionDefinitionChoice->present == E2_E2SM_gNB_X2_ActionDefinitionChoice_PR_actionDefinition_Format2) {

            // E2SM-gNB-X2-ActionDefinition-Format2
            uint64_t status;
            if ((status = allocActionDefinitionX2Format2(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format2)) != e2err_OK)
                return status;

            pRICActionDefinitionChoice->actionDefinitionX2Format2Present = true;
            pRICActionDefinitionChoice->actionDefinitionX2Format1Present = false;
            pRICActionDefinitionChoice->actionDefinitionNRTFormat1Present = false;
            pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupCount = 0;
            pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem->ranUEgroupDefinition.ranUeGroupDefCount = 0;
            pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem->ranPolicy.ranParameterCount = 0;

            E2_E2SM_gNB_X2_ActionDefinition_Format2_t* pE2SM_gNB_X2_actionDefinition = &pE2_E2SM_gNB_X2_ActionDefinitionChoice->choice.actionDefinition_Format2;
            if(pE2SM_gNB_X2_actionDefinition) {
                uint64_t index = 0;
                if (pE2SM_gNB_X2_actionDefinition->ranUEgroup_List) {
                    while (index < pE2SM_gNB_X2_actionDefinition->ranUEgroup_List->list.count) {
                        E2_RANueGroup_Item_t* pE2_RANueGroup_Item = pE2SM_gNB_X2_actionDefinition->ranUEgroup_List->list.array[index];
                        if (pE2_RANueGroup_Item) {
                            pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupID = pE2_RANueGroup_Item->ranUEgroupID;
                            uint64_t index2 = 0;
                            while (index2 < pE2_RANueGroup_Item->ranUEgroupDefinition.ranUEgroupDef_List->list.count) {
                                E2_RANueGroupDef_Item_t* pE2_RANueGroupDef_Item = pE2_RANueGroup_Item->ranUEgroupDefinition.ranUEgroupDef_List->list.array[index2];
                                if(pE2_RANueGroupDef_Item) {
                                    pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterID = pE2_RANueGroupDef_Item->ranParameter_ID;
                                    pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterTest = pE2_RANueGroupDef_Item->ranParameter_Test;
                                    if (pE2_RANueGroupDef_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valueInt) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueIntPresent = true;
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueInt =
                                          pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueInt;
                                    }
                                    else if (pE2_RANueGroupDef_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valueEnum) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueEnum = true;
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueEnum =
                                          pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueEnum;
                                    }
                                    else if (pE2_RANueGroupDef_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valueBool) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBoolPresent = true;
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBool =
                                          pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueBool;
                                    }
                                    else if (pE2_RANueGroupDef_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valueBitS) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBitSPresent = true;
                                        addBitString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueBitS,
                                                     pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueBitS.size,
                                                     pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueBitS.buf,
                                                     pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueBitS.bits_unused);
                                    }
                                    else if (pE2_RANueGroupDef_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valueOctS) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueOctSPresent = true;
                                        addOctetString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valueOctS,
                                                     pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueOctS.size,
                                                     pE2_RANueGroupDef_Item->ranParameter_Value.choice.valueOctS.buf);
                                    }
                                    else if (pE2_RANueGroupDef_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valuePrtS) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valuePrtSPresent = true;
                                        addOctetString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefItem[index2].ranParameterValue.valuePrtS,
                                                     pE2_RANueGroupDef_Item->ranParameter_Value.choice.valuePrtS.size,
                                                     pE2_RANueGroupDef_Item->ranParameter_Value.choice.valuePrtS.buf);
                                    }
                                    else {
                                        ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice);
                                        return e2err_RICSubscriptionRequestRanranUeGroupDefItemParameterValueEmptyFail;
                                    }
                                }
                                else {
                                    ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice);
                                    return e2err_RICSubscriptionRequestAllocE2_RANueGroupDef_ItemFail;
                                }
                                index2++;
                            }
                            pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranUEgroupDefinition.ranUeGroupDefCount = index2;

                            uint64_t index3 = 0;
                            while (index3 < pE2_RANueGroup_Item->ranPolicy.ranImperativePolicy_List->list.count) {
                                E2_RANParameter_Item_t* pE2_RANParameter_Item = pE2_RANueGroup_Item->ranPolicy.ranImperativePolicy_List->list.array[index3];
                                if (pE2_RANParameter_Item) {
                                    pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterID = pE2_RANParameter_Item->ranParameter_ID;
                                    if (pE2_RANParameter_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valueInt) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueIntPresent = true;
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueInt =
                                          pE2_RANParameter_Item->ranParameter_Value.choice.valueInt;
                                    }
                                    else if (pE2_RANParameter_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valueEnum) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueEnum = true;
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueEnum =
                                          pE2_RANParameter_Item->ranParameter_Value.choice.valueEnum;
                                    }
                                    else if (pE2_RANParameter_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valueBool) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBoolPresent = true;
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBool =
                                          pE2_RANParameter_Item->ranParameter_Value.choice.valueBool;
                                    }
                                    else if (pE2_RANParameter_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valueBitS) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitSPresent = true;
                                        addBitString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS,
                                                     pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.size,
                                                     pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.buf,
                                                     pE2_RANParameter_Item->ranParameter_Value.choice.valueBitS.bits_unused);
                                    }
                                    else if (pE2_RANParameter_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valueOctS) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctSPresent = true;
                                        addOctetString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctS,
                                                     pE2_RANParameter_Item->ranParameter_Value.choice.valueOctS.size,
                                                     pE2_RANParameter_Item->ranParameter_Value.choice.valueOctS.buf);
                                    }
                                    else if (pE2_RANParameter_Item->ranParameter_Value.present == E2_RANParameter_Value_PR_valuePrtS) {
                                        pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtSPresent = true;
                                        addOctetString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtS,
                                                     pE2_RANParameter_Item->ranParameter_Value.choice.valuePrtS.size,
                                                     pE2_RANParameter_Item->ranParameter_Value.choice.valuePrtS.buf);
                                    }
                                    else {
                                        ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice);
                                        return e2err_RICSubscriptionRequestRanParameterItemRanParameterValueEmptyFail;
                                    }
                                }
                                else {
                                    ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice);
                                    return e2err_RICSubscriptionRequestAllocActionDefinitionFail;
                                }
                                index3++;
                            }
                            pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupItem[index].ranPolicy.ranParameterCount = index3;
                        }
                        else {
                            ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice);
                            return e2err_RICSubscriptionRequestAllocRANParameter_ItemFail;
                        }
                        index++;
                    }
               }
               pRICActionDefinitionChoice->actionDefinitionX2Format2->ranUeGroupCount = index;
            }
        }
        ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice, pE2_E2SM_gNB_X2_ActionDefinitionChoice);
        return e2err_OK;
    case RC_WMORE:
        if (debug)
            printf("Decode failed. More data needed. Buffer size %zu, %s, consumed %zu\n",pRICActionDefinitionChoice->octetString.contentLength,
                   asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice.name, rval.consumed);

        return e2err_RICActionDefinitionChoiceWMOREFail;
    case RC_FAIL:
        if (debug)
            printf("Decode failed. Buffer size %zu, %s, consumed %zu\n",pRICActionDefinitionChoice->octetString.contentLength,
                   asn_DEF_E2_E2SM_gNB_X2_ActionDefinitionChoice.name, rval.consumed);

        return e2err_RICActionDefinitionChoiceDecodeFAIL;
    default:
        return e2err_RICActionDefinitionChoiceDecodeDefaultFail;
    }
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICActionDefinitionDataNRTFormat(mem_track_hdr_t* pDynMemHead, RICActionDefinitionChoice_t* pRICActionDefinitionChoice) {

    E2_E2SM_gNB_NRT_ActionDefinition_t* pE2_E2SM_gNB_NRT_ActionDefinition = 0;
    asn_dec_rval_t rval;
    rval = asn_decode(0, ATS_ALIGNED_BASIC_PER, &asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition, (void **)&pE2_E2SM_gNB_NRT_ActionDefinition,
                      pRICActionDefinitionChoice->octetString.data, pRICActionDefinitionChoice->octetString.contentLength);
    switch(rval.code) {
    case RC_OK:
        // Debug print
        if (debug) {
            printf("Successfully decoded E2SM_gNB_NRT_ActionDefinition\n");
            asn_fprint(stdout, &asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition, pE2_E2SM_gNB_NRT_ActionDefinition);
        }

        // ActionDefinitionChoice
        if (pE2_E2SM_gNB_NRT_ActionDefinition->present == E2_E2SM_gNB_NRT_ActionDefinition_PR_actionDefinition_Format1) {

            // E2SM-gNB-NRT-actionDefinition
            uint64_t status;
            if ((status = allocActionDefinitionNRTFormat1(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionNRTFormat1)) != e2err_OK)
                return status;

            pRICActionDefinitionChoice->actionDefinitionNRTFormat1Present = true;
            pRICActionDefinitionChoice->actionDefinitionX2Format1Present = false;
            pRICActionDefinitionChoice->actionDefinitionX2Format2Present = false;
            pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterCount = 0;
            uint64_t index = 0;
            while (index < pE2_E2SM_gNB_NRT_ActionDefinition->choice.actionDefinition_Format1.ranParameter_List->list.count) {
                E2_RANparameter_Item_t* pE2_RANparameter_Item = pE2_E2SM_gNB_NRT_ActionDefinition->choice.actionDefinition_Format1.ranParameter_List->list.array[index];
                if (pE2_RANparameter_Item) {
                    pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterID = pE2_RANparameter_Item->ranParameter_ID;

                    if (pE2_RANparameter_Item->ranParameter_Value.present == E2_RANparameter_Value_PR_valueInt) {
                        pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueIntPresent = true;
                        pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueInt =
                          pE2_RANparameter_Item->ranParameter_Value.choice.valueInt;
                    }
                    else if (pE2_RANparameter_Item->ranParameter_Value.present == E2_RANparameter_Value_PR_valueEnum) {
                        pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueEnumPresent = true;
                        pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueEnum =
                          pE2_RANparameter_Item->ranParameter_Value.choice.valueEnum;
                    }
                    else if (pE2_RANparameter_Item->ranParameter_Value.present == E2_RANparameter_Value_PR_valueBool) {
                        pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBoolPresent = true;
                        pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBool =
                          pE2_RANparameter_Item->ranParameter_Value.choice.valueBool;
                    }
                    else if (pE2_RANparameter_Item->ranParameter_Value.present == E2_RANparameter_Value_PR_valueBitS) {
                        pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBitSPresent = true;
                        addBitString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueBitS,
                                     pE2_RANparameter_Item->ranParameter_Value.choice.valueBitS.size,
                                     pE2_RANparameter_Item->ranParameter_Value.choice.valueBitS.buf,
                                     pE2_RANparameter_Item->ranParameter_Value.choice.valueBitS.bits_unused);
                    }
                    else if (pE2_RANparameter_Item->ranParameter_Value.present == E2_RANparameter_Value_PR_valueOctS) {
                        pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueOctSPresent = true;
                        addOctetString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valueOctS,
                                     pE2_RANparameter_Item->ranParameter_Value.choice.valueOctS.size,
                                     pE2_RANparameter_Item->ranParameter_Value.choice.valueOctS.buf);
                    }
                    else if (pE2_RANparameter_Item->ranParameter_Value.present == E2_RANparameter_Value_PR_valuePrtS) {
                        pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valuePrtSPresent = true;
                        addOctetString(pDynMemHead, &pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterList[index].ranParameterValue.valuePrtS,
                                     pE2_RANparameter_Item->ranParameter_Value.choice.valuePrtS.size,
                                     pE2_RANparameter_Item->ranParameter_Value.choice.valuePrtS.buf);
                    }
                    else {
                        ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition, pE2_E2SM_gNB_NRT_ActionDefinition);
                        return e2err_RICSubscriptionRequestNRTRanParameterItemRanParameterValueEmptyFail;
                    }
                }
                else {
                    ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition, pE2_E2SM_gNB_NRT_ActionDefinition);
                    return e2err_RICSubscriptionRequestNRTAllocActionDefinitionFail;
                }
                index++;
            }
            pRICActionDefinitionChoice->actionDefinitionNRTFormat1->ranParameterCount = index;
        }
        ASN_STRUCT_FREE(asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition, pE2_E2SM_gNB_NRT_ActionDefinition);
        return e2err_OK;
    case RC_WMORE:
        if (debug)
            printf("Decode failed. More data needed. Buffer size %zu, %s, consumed %zu\n",pRICActionDefinitionChoice->octetString.contentLength,
                   asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition.name, rval.consumed);

        return e2err_RICNRTActionDefinitionChoiceWMOREFail;
    case RC_FAIL:
        if (debug)
            printf("Decode failed. Buffer size %zu, %s, consumed %zu\n",pRICActionDefinitionChoice->octetString.contentLength,
                   asn_DEF_E2_E2SM_gNB_NRT_ActionDefinition.name, rval.consumed);

        return e2err_RICNRTActionDefinitionChoiceDecodeFAIL;
    default:
        return e2err_RICNRTActionDefinitionChoiceDecodeDefaultFail;
    }
}
//////////////////////////////////////////////////////////////////////
uint64_t getRICSubscriptionResponseData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionResponse_t* pRICSubscriptionResponse) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;

    RICsubscriptionResponse_t *asnRicSubscriptionResponse = &pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionResponse;
    RICsubscriptionResponse_IEs_t* pRICsubscriptionResponse_IEs;

    // RICrequestID
    if (asnRicSubscriptionResponse->protocolIEs.list.count > 0 &&
        asnRicSubscriptionResponse->protocolIEs.list.array[0]->id == ProtocolIE_ID_id_RICrequestID) {
        pRICsubscriptionResponse_IEs = asnRicSubscriptionResponse->protocolIEs.list.array[0];
        pRICSubscriptionResponse->ricRequestID.ricRequestorID = pRICsubscriptionResponse_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionResponse->ricRequestID.ricInstanceID = pRICsubscriptionResponse_IEs->value.choice.RICrequestID.ricInstanceID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionResponseRICrequestIDMissing;
    }

    // RANfunctionID
    if (asnRicSubscriptionResponse->protocolIEs.list.count > 1 &&
        asnRicSubscriptionResponse->protocolIEs.list.array[1]->id == ProtocolIE_ID_id_RANfunctionID) {
        pRICsubscriptionResponse_IEs = asnRicSubscriptionResponse->protocolIEs.list.array[1];
        pRICSubscriptionResponse->ranFunctionID = pRICsubscriptionResponse_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionResponseRANfunctionIDMissing;
    }

    // RICaction-Admitted-List
    if (asnRicSubscriptionResponse->protocolIEs.list.count > 2  &&
        asnRicSubscriptionResponse->protocolIEs.list.array[2]->id == ProtocolIE_ID_id_RICactions_Admitted) {
        pRICsubscriptionResponse_IEs = asnRicSubscriptionResponse->protocolIEs.list.array[2];
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
    if (asnRicSubscriptionResponse->protocolIEs.list.count > 3 &&
        asnRicSubscriptionResponse->protocolIEs.list.array[3]->id == ProtocolIE_ID_id_RICactions_NotAdmitted) {
        pRICsubscriptionResponse_IEs = asnRicSubscriptionResponse->protocolIEs.list.array[3];
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
                if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_ricRequest) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_ricRequest;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.ricRequest;
                }
                else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_ricService) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_ricService;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.ricService;
                }
                else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_transport) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_transport;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.transport;
                }
                else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_protocol) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_protocol;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.protocol;
                }
                else if(pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_misc) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_misc;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.misc;
                }
               index++;
            }
            pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength = index;
        }
    }
    else {
        pRICSubscriptionResponse->ricActionNotAdmittedListPresent = false;
        pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength = 0;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
    return e2err_OK;
}

//////////////////////////////////////////////////////////////////////
uint64_t getRICSubscriptionFailureData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionFailure_t* pRICSubscriptionFailure) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;

    RICsubscriptionFailure_t *asnRicSubscriptionFailure = &pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure;
    RICsubscriptionFailure_IEs_t* pRICsubscriptionFailure_IEs;

    // RICrequestID
    if (asnRicSubscriptionFailure->protocolIEs.list.count > 0 &&
        asnRicSubscriptionFailure->protocolIEs.list.array[0]->id == ProtocolIE_ID_id_RICrequestID) {
        pRICsubscriptionFailure_IEs = asnRicSubscriptionFailure->protocolIEs.list.array[0];
        pRICSubscriptionFailure->ricRequestID.ricRequestorID = pRICsubscriptionFailure_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionFailure->ricRequestID.ricInstanceID = pRICsubscriptionFailure_IEs->value.choice.RICrequestID.ricInstanceID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionFailureRICrequestIDMissing;
    }

    // RANfunctionID
    if (asnRicSubscriptionFailure->protocolIEs.list.count > 1 &&
        asnRicSubscriptionFailure->protocolIEs.list.array[1]->id == ProtocolIE_ID_id_RANfunctionID) {
        pRICsubscriptionFailure_IEs = asnRicSubscriptionFailure->protocolIEs.list.array[1];
        pRICSubscriptionFailure->ranFunctionID = pRICsubscriptionFailure_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionFailureRANfunctionIDMissing;
    }

    // RICaction-NotAdmitted-List
    if (asnRicSubscriptionFailure->protocolIEs.list.count > 2 &&
        asnRicSubscriptionFailure->protocolIEs.list.array[2]->id == ProtocolIE_ID_id_RICactions_NotAdmitted) {
        pRICsubscriptionFailure_IEs = asnRicSubscriptionFailure->protocolIEs.list.array[2];
        uint64_t index = 0;
        while ((index < maxofRICactionID) && (index < pRICsubscriptionFailure_IEs->value.choice.RICaction_NotAdmitted_List.list.count)) {
            RICaction_NotAdmitted_ItemIEs_t* pRICaction_NotAdmitted_ItemIEs =
              (RICaction_NotAdmitted_ItemIEs_t*)pRICsubscriptionFailure_IEs->value.choice.RICaction_NotAdmitted_List.list.array[index];

            // RICActionID
            pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID =
              pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.ricActionID;

            //  RICcause
            if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_ricRequest) {
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_ricRequest;
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.ricRequest;
            }
            else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_ricService) {
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_ricService;
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.ricService;
            }
            else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_transport) {
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_transport;
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.transport;
            }
            else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_protocol) {
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_protocol;
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.protocol;
            }
            else if(pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_misc) {
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_misc;
                pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                  pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.misc;
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
uint64_t getRICSubscriptionDeleteRequestData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionDeleteRequest_t* pRICSubscriptionDeleteRequest) {

    E2AP_PDU_t* pE2AP_PDU = (E2AP_PDU_t*)pE2AP_PDU_pointer;

    RICsubscriptionDeleteRequest_t *asnRicSubscriptionDeleteRequest = &pE2AP_PDU->choice.initiatingMessage.value.choice.RICsubscriptionDeleteRequest;
    RICsubscriptionDeleteRequest_IEs_t* pRICsubscriptionDeleteRequest_IEs;

    // RICrequestID
    if (asnRicSubscriptionDeleteRequest->protocolIEs.list.count > 0 &&
        asnRicSubscriptionDeleteRequest->protocolIEs.list.array[0]->id == ProtocolIE_ID_id_RICrequestID) {
        pRICsubscriptionDeleteRequest_IEs = asnRicSubscriptionDeleteRequest->protocolIEs.list.array[0];
        pRICSubscriptionDeleteRequest->ricRequestID.ricRequestorID = pRICsubscriptionDeleteRequest_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionDeleteRequest->ricRequestID.ricInstanceID = pRICsubscriptionDeleteRequest_IEs->value.choice.RICrequestID.ricInstanceID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteRequestRICrequestIDMissing;
    }

    // RANfunctionID
    if (asnRicSubscriptionDeleteRequest->protocolIEs.list.count > 1 &&
        asnRicSubscriptionDeleteRequest->protocolIEs.list.array[1]->id == ProtocolIE_ID_id_RANfunctionID) {
        pRICsubscriptionDeleteRequest_IEs = asnRicSubscriptionDeleteRequest->protocolIEs.list.array[1];
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

    RICsubscriptionDeleteResponse_t *asnRicSubscriptionDeleteResponse = &pE2AP_PDU->choice.successfulOutcome.value.choice.RICsubscriptionDeleteResponse;
    RICsubscriptionDeleteResponse_IEs_t* pRICsubscriptionDeleteResponse_IEs;

    // RICrequestID
    if (asnRicSubscriptionDeleteResponse->protocolIEs.list.count > 0 &&
        asnRicSubscriptionDeleteResponse->protocolIEs.list.array[0]->id == ProtocolIE_ID_id_RICrequestID) {
        pRICsubscriptionDeleteResponse_IEs = asnRicSubscriptionDeleteResponse->protocolIEs.list.array[0];
        pRICSubscriptionDeleteResponse->ricRequestID.ricRequestorID = pRICsubscriptionDeleteResponse_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionDeleteResponse->ricRequestID.ricInstanceID = pRICsubscriptionDeleteResponse_IEs->value.choice.RICrequestID.ricInstanceID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteResponseRICrequestIDMissing;
    }

    // RANfunctionID
    if (asnRicSubscriptionDeleteResponse->protocolIEs.list.count > 1 &&
        asnRicSubscriptionDeleteResponse->protocolIEs.list.array[1]->id == ProtocolIE_ID_id_RANfunctionID) {
        pRICsubscriptionDeleteResponse_IEs = asnRicSubscriptionDeleteResponse->protocolIEs.list.array[1];
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

    RICsubscriptionDeleteFailure_t *asnRicSubscriptionDeleteFailure = &pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionDeleteFailure;
    RICsubscriptionDeleteFailure_IEs_t* pRICsubscriptionDeleteFailure_IEs;

    // RICrequestID
    if (asnRicSubscriptionDeleteFailure->protocolIEs.list.count > 0 &&
        asnRicSubscriptionDeleteFailure->protocolIEs.list.array[0]->id == ProtocolIE_ID_id_RICrequestID) {
        pRICsubscriptionDeleteFailure_IEs = asnRicSubscriptionDeleteFailure->protocolIEs.list.array[0];
        pRICSubscriptionDeleteFailure->ricRequestID.ricRequestorID = pRICsubscriptionDeleteFailure_IEs->value.choice.RICrequestID.ricRequestorID;
        pRICSubscriptionDeleteFailure->ricRequestID.ricInstanceID = pRICsubscriptionDeleteFailure_IEs->value.choice.RICrequestID.ricInstanceID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteFailureRICrequestIDMissing;
    }

    // RANfunctionID
    if (asnRicSubscriptionDeleteFailure->protocolIEs.list.count > 1 &&
        asnRicSubscriptionDeleteFailure->protocolIEs.list.array[1]->id == ProtocolIE_ID_id_RANfunctionID) {
        pRICsubscriptionDeleteFailure_IEs = asnRicSubscriptionDeleteFailure->protocolIEs.list.array[1];
        pRICSubscriptionDeleteFailure->ranFunctionID = pRICsubscriptionDeleteFailure_IEs->value.choice.RANfunctionID;
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionDeleteFailureRANfunctionIDMissing;
    }

    // RICcause
    if (asnRicSubscriptionDeleteFailure->protocolIEs.list.count > 2 &&
        asnRicSubscriptionDeleteFailure->protocolIEs.list.array[2]->id == ProtocolIE_ID_id_Cause) {
        pRICsubscriptionDeleteFailure_IEs = asnRicSubscriptionDeleteFailure->protocolIEs.list.array[2];
        if (pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present == Cause_PR_ricRequest) {
            pRICSubscriptionDeleteFailure->cause.content = Cause_PR_ricRequest;
            pRICSubscriptionDeleteFailure->cause.causeVal =
              pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.ricRequest;
        }
        else if (pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present == Cause_PR_ricService) {
            pRICSubscriptionDeleteFailure->cause.content = Cause_PR_ricService;
            pRICSubscriptionDeleteFailure->cause.causeVal =
              pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.ricService;
        }
        else if (pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present == Cause_PR_transport) {
            pRICSubscriptionDeleteFailure->cause.content = Cause_PR_transport;
            pRICSubscriptionDeleteFailure->cause.causeVal =
              pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.transport;
        }
        else if (pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present == Cause_PR_protocol) {
            pRICSubscriptionDeleteFailure->cause.content = Cause_PR_protocol;
            pRICSubscriptionDeleteFailure->cause.causeVal =
              pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.protocol;
        }
        else if(pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present == Cause_PR_misc) {
            pRICSubscriptionDeleteFailure->cause.content = Cause_PR_misc;
            pRICSubscriptionDeleteFailure->cause.causeVal =
              pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.misc;
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

//////////////////////////////////////////////////////////////////////
uint64_t allocActionDefinitionX2Format1(mem_track_hdr_t* pDynMemHead, E2SMgNBX2actionDefinition_t** pActionDefinition) {
    *pActionDefinition = mem_track_alloc(pDynMemHead, sizeof(E2SMgNBX2actionDefinition_t));
    if(*pActionDefinition)
        return e2err_OK;
    else
        return e2err_RICSubscriptionRequestAllocactionDefinitionX2Format1Fail;
}

//////////////////////////////////////////////////////////////////////
uint64_t allocActionDefinitionX2Format2(mem_track_hdr_t* pDynMemHead, E2SMgNBX2ActionDefinitionFormat2_t** pActionDefinition) {
    *pActionDefinition = mem_track_alloc(pDynMemHead, sizeof(E2SMgNBX2ActionDefinitionFormat2_t));
    if(*pActionDefinition)
        return e2err_OK;
    else
        return e2err_RICSubscriptionRequestAllocactionDefinitionX2Format2Fail;
}

//////////////////////////////////////////////////////////////////////
uint64_t allocActionDefinitionNRTFormat1(mem_track_hdr_t* pDynMemHead, E2SMgNBNRTActionDefinitionFormat1_t** pActionDefinition) {
    *pActionDefinition = mem_track_alloc(pDynMemHead, sizeof(E2SMgNBNRTActionDefinitionFormat1_t));
    if(*pActionDefinition)
        return e2err_OK;
    else
        return e2err_RICSubscriptionRequestAllocactionDefinitionNRTFormat1Fail;
}

//////////////////////////////////////////////////////////////////////
bool addOctetString(mem_track_hdr_t* pDynMemHead, DynOctetString_t* pOctetString, uint64_t bufferSize, void* pData)
{
    pOctetString->data = mem_track_alloc(pDynMemHead, bufferSize);
    if (pOctetString->data) {
        pOctetString->length = bufferSize;
        memcpy(pOctetString->data,pData,bufferSize);
        return true;
    }
    else
        return false;
}

//////////////////////////////////////////////////////////////////////
bool addBitString(mem_track_hdr_t* pDynMemHead, DynBitString_t* pBitString, uint64_t bufferSize, void* pData, uint8_t unusedBits)
{
    pBitString->data = mem_track_alloc(pDynMemHead, bufferSize);
    if (pBitString->data) {
        pBitString->byteLength = bufferSize;
        pBitString->unusedBits = unusedBits; // Unused trailing bits in the last octet (0..7)
        memcpy(pBitString->data,pData,bufferSize);
        return true;
    }
    else
        return false;
}

//////////////////////////////////////////////////////////////////////
void mem_track_init(mem_track_hdr_t *curr)
{
    *curr=(mem_track_hdr_t)MEM_TRACK_HDR_INIT;
}

//////////////////////////////////////////////////////////////////////
void* mem_track_alloc(mem_track_hdr_t *curr, size_t sz)
{
    mem_track_t *newentry = (mem_track_t *)malloc(sizeof(mem_track_t)+sz);
    newentry->next=0;
    newentry->sz=sz;
    memset(newentry->ptr,0,newentry->sz);

    if (!curr->next) {
        curr->next = newentry;
    } else {
        mem_track_t *iter=curr->next;
        for(;iter->next;iter=iter->next);
        iter->next = newentry;
    }
    return newentry->ptr;
}

//////////////////////////////////////////////////////////////////////
void mem_track_free(mem_track_hdr_t *curr)
{
    mem_track_t *itecurr,*itenext;
    for(itecurr=curr->next; itecurr; itecurr=itenext){
        itenext = itecurr->next;
        free(itecurr);
    }
    mem_track_init(curr);
}
