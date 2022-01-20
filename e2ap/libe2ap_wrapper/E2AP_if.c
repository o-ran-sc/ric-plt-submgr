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

#include "asn_constant.h"
#include "E2AP_if.h"

static bool debugPrints = false;

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
void allowASN1DebugPrints(bool allowASN1DebugPrints) {
    debugPrints = allowASN1DebugPrints;
}

//////////////////////////////////////////////////////////////////////
const char* getE2ErrorString(uint64_t errorCode) {

    return E2ErrorStrings[errorCode];
}

/////////////////////////////////////////////////////////////////////
bool E2encode(E2AP_PDU_t* pE2AP_PDU, size_t* dataBufferSize, byte* dataBuffer, char* pLogBuffer) {

    if (debugPrints)
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
        if (debugPrints)
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
                    pRICaction_ToBeSetup_ItemIEs->criticality = Criticality_ignore;
                    pRICaction_ToBeSetup_ItemIEs->value.present = RICaction_ToBeSetup_ItemIEs__value_PR_RICaction_ToBeSetup_Item;

                    // RICActionID
                    pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionID =
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID;

                    // RICActionType
                    pRICaction_ToBeSetup_ItemIEs->value.choice.RICaction_ToBeSetup_Item.ricActionType =
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType;

                    // RICactionDefinition, OPTIONAL
                    if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent) {
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

                        // Cause
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
                        else if (pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content == Cause_PR_e2Node) {
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present = Cause_PR_e2Node;
                            pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.e2Node =
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

        // Cause
        pRICsubscriptionFailure_IEs = calloc(1, sizeof(RICsubscriptionFailure_IEs_t));
        if (pRICsubscriptionFailure_IEs) {
            pRICsubscriptionFailure_IEs->id = ProtocolIE_ID_id_Cause;
            pRICsubscriptionFailure_IEs->criticality = Criticality_reject;
            pRICsubscriptionFailure_IEs->value.present = RICsubscriptionFailure_IEs__value_PR_Cause;
            if (pRICSubscriptionFailure->cause.content == Cause_PR_ricRequest) {
                pRICsubscriptionFailure_IEs->value.choice.Cause.present = Cause_PR_ricRequest;
                pRICsubscriptionFailure_IEs->value.choice.Cause.choice.ricRequest =
                  pRICSubscriptionFailure->cause.causeVal;
            }
            else if (pRICSubscriptionFailure->cause.content == Cause_PR_ricService) {
                pRICsubscriptionFailure_IEs->value.choice.Cause.present = Cause_PR_ricService;
                pRICsubscriptionFailure_IEs->value.choice.Cause.choice.ricService =
                  pRICSubscriptionFailure->cause.causeVal;
            }
            else if (pRICSubscriptionFailure->cause.content == Cause_PR_e2Node) {
                pRICsubscriptionFailure_IEs->value.choice.Cause.present = Cause_PR_e2Node;
                pRICsubscriptionFailure_IEs->value.choice.Cause.choice.e2Node =
                  pRICSubscriptionFailure->cause.causeVal;
            }
            else if (pRICSubscriptionFailure->cause.content == Cause_PR_transport) {
                pRICsubscriptionFailure_IEs->value.choice.Cause.present = Cause_PR_transport;
                pRICsubscriptionFailure_IEs->value.choice.Cause.choice.transport =
                  pRICSubscriptionFailure->cause.causeVal;
            }
            else if (pRICSubscriptionFailure->cause.content == Cause_PR_protocol) {
                pRICsubscriptionFailure_IEs->value.choice.Cause.present = Cause_PR_protocol;
                pRICsubscriptionFailure_IEs->value.choice.Cause.choice.protocol =
                  pRICSubscriptionFailure->cause.causeVal;
            }
            else if (pRICSubscriptionFailure->cause.content == Cause_PR_misc) {
                pRICsubscriptionFailure_IEs->value.choice.Cause.present = Cause_PR_misc;
                pRICsubscriptionFailure_IEs->value.choice.Cause.choice.misc =
                  pRICSubscriptionFailure->cause.causeVal;
            }
            ASN_SEQUENCE_ADD(&pE2AP_PDU->choice.unsuccessfulOutcome.value.choice.RICsubscriptionFailure.protocolIEs.list, pRICsubscriptionFailure_IEs);
        }
        else
            return e2err_RICSubscriptionFailureAllocCauseFail;

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

        // Cause
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
            else if (pRICSubscriptionDeleteFailure->cause.content == Cause_PR_e2Node) {
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present = Cause_PR_e2Node;
                pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.e2Node =
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
        if (debugPrints) {
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
uint64_t getRICSubscriptionRequestData(e2ap_pdu_ptr_t* pE2AP_PDU_pointer, RICSubscriptionRequest_t* pRICSubscriptionRequest) {

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

                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = true;
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

                //  Cause
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
                else if (pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.present == Cause_PR_e2Node) {
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = Cause_PR_e2Node;
                    pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal =
                      pRICaction_NotAdmitted_ItemIEs->value.choice.RICaction_NotAdmitted_Item.cause.choice.e2Node;
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

    // Cause
    if (asnRicSubscriptionFailure->protocolIEs.list.count > 2 &&
        asnRicSubscriptionFailure->protocolIEs.list.array[2]->id == ProtocolIE_ID_id_Cause) {
        pRICsubscriptionFailure_IEs = asnRicSubscriptionFailure->protocolIEs.list.array[2];
        if (pRICsubscriptionFailure_IEs->value.choice.Cause.present == Cause_PR_ricRequest) {
            pRICSubscriptionFailure->cause.content = Cause_PR_ricRequest;
            pRICSubscriptionFailure->cause.causeVal =
              pRICsubscriptionFailure_IEs->value.choice.Cause.choice.ricRequest;
        }
        else if (pRICsubscriptionFailure_IEs->value.choice.Cause.present == Cause_PR_ricService) {
            pRICSubscriptionFailure->cause.content = Cause_PR_ricService;
            pRICSubscriptionFailure->cause.causeVal =
              pRICsubscriptionFailure_IEs->value.choice.Cause.choice.ricService;
        }
        else if (pRICsubscriptionFailure_IEs->value.choice.Cause.present == Cause_PR_e2Node) {
            pRICSubscriptionFailure->cause.content = Cause_PR_e2Node;
            pRICSubscriptionFailure->cause.causeVal =
              pRICsubscriptionFailure_IEs->value.choice.Cause.choice.e2Node;
        }
        else if (pRICsubscriptionFailure_IEs->value.choice.Cause.present == Cause_PR_transport) {
            pRICSubscriptionFailure->cause.content = Cause_PR_transport;
            pRICSubscriptionFailure->cause.causeVal =
              pRICsubscriptionFailure_IEs->value.choice.Cause.choice.transport;
        }
        else if (pRICsubscriptionFailure_IEs->value.choice.Cause.present == Cause_PR_protocol) {
            pRICSubscriptionFailure->cause.content = Cause_PR_protocol;
            pRICSubscriptionFailure->cause.causeVal =
              pRICsubscriptionFailure_IEs->value.choice.Cause.choice.protocol;
        }
        else if(pRICsubscriptionFailure_IEs->value.choice.Cause.present == Cause_PR_misc) {
            pRICSubscriptionFailure->cause.content = Cause_PR_misc;
            pRICSubscriptionFailure->cause.causeVal =
              pRICsubscriptionFailure_IEs->value.choice.Cause.choice.misc;
        }
    }
    else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pE2AP_PDU);
        return e2err_RICsubscriptionFailureCauseMissing;
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

    // Cause
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
        else if (pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.present == Cause_PR_e2Node) {
            pRICSubscriptionDeleteFailure->cause.content = Cause_PR_e2Node;
            pRICSubscriptionDeleteFailure->cause.causeVal =
              pRICsubscriptionDeleteFailure_IEs->value.choice.Cause.choice.e2Node;
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
