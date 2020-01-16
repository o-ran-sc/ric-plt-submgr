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

#if DEBUG

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "E2AP_if.h"

const size_t cDataBufferSize = 2048;

typedef union {
    uint32_t  nodeID;
    uint8_t   octets[4];
} IdOctects_t;

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionRequest() {
    RICSubscriptionRequest_t ricSubscriptionRequest;
    ricSubscriptionRequest.ricRequestID.ricRequestorID = 1;
    ricSubscriptionRequest.ricRequestID.ricRequestSequenceNumber = 22;
    ricSubscriptionRequest.ranFunctionID = 33;

    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.octetString.contentLength = 0;

    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBIDPresent = true;
    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalGNBIDPresent = false;
    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.contentLength = 3;
    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[0] = 1;
    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[1] = 2;
    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[2] = 3;

//    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.eNBID.bits = cMacroENBIDP_20Bits;
//    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.eNBID.bits = cHomeENBID_28Bits;
//    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.eNBID.bits = cShortMacroENBID_18Bits;
    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.nodeID.bits = clongMacroENBIDP_21Bits;

    IdOctects_t eNBOctects;
    memset(eNBOctects.octets, 0, sizeof(eNBOctects));
    eNBOctects.octets[0] = 11;
    eNBOctects.octets[1] = 22;
    eNBOctects.octets[2] = 31;
    eNBOctects.octets[3] = 1;
    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
    printf("eNBOctects.nodeID = %u\n\n",eNBOctects.nodeID);

    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceDirection = InterfaceDirection__incoming;
    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceMessageType.procedureCode = 35;  // id-rRCTransfer
    ricSubscriptionRequest.ricSubscription.ricEventTriggerDefinition.interfaceMessageType.typeOfMessage = cE2InitiatingMessage;

    ricSubscriptionRequest.ricSubscription.ricActionToBeSetupItemIEs.contentLength = 1;
    uint64_t index = 0;
    while (index < ricSubscriptionRequest.ricSubscription.ricActionToBeSetupItemIEs.contentLength) {
        ricSubscriptionRequest.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID = 255; //index;
        ricSubscriptionRequest.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType = RICActionType_insert;

        // ricActionDefinition, OPTIONAL. Not used in RIC
        ricSubscriptionRequest.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = false; //true;
        ricSubscriptionRequest.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinition.styleID = 255;
        ricSubscriptionRequest.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinition.sequenceOfActionParameters.parameterID = 222;

        // ricSubsequentActionPresent, OPTIONAL
        ricSubscriptionRequest.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent = true;
        ricSubscriptionRequest.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType = RICSubsequentActionType_Continue;
        ricSubscriptionRequest.ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait = RICTimeToWait_w100ms;
        index++;
    }

    printRICSubscriptionRequest(&ricSubscriptionRequest);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if (packRICSubscriptionRequest(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionRequest) == e2err_OK)
    {
        memset(&ricSubscriptionRequest,0, sizeof ricSubscriptionRequest);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2InitiatingMessage) {
                if (messageInfo.messageId == cRICSubscriptionRequest) {
                    if ((returnCode = getRICSubscriptionRequestData(pE2AP_PDU, &ricSubscriptionRequest)) == e2err_OK) {
                        printRICSubscriptionRequest(&ricSubscriptionRequest);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionRequestData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionRequest\n");
            }
            else
                printf("Not InitiatingMessage\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionResponse() {
    // Test RICSubscribeResponse
    RICSubscriptionResponse_t ricSubscriptionResponse;
    ricSubscriptionResponse.ricRequestID.ricRequestorID = 1;
    ricSubscriptionResponse.ricRequestID.ricRequestSequenceNumber = 22;
    ricSubscriptionResponse.ranFunctionID = 33;
    ricSubscriptionResponse.ricActionAdmittedList.contentLength = 16;
    uint64_t index = 0;
    while (index < ricSubscriptionResponse.ricActionAdmittedList.contentLength) {
        ricSubscriptionResponse.ricActionAdmittedList.ricActionID[index] = index;
        index++;
    }
    ricSubscriptionResponse.ricActionNotAdmittedListPresent = true;
    ricSubscriptionResponse.ricActionNotAdmittedList.contentLength = 16;
    index = 0;
    while (index < ricSubscriptionResponse.ricActionNotAdmittedList.contentLength) {
        ricSubscriptionResponse.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID = index;
        ricSubscriptionResponse.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = cRICCauseRadioNetwork;
        ricSubscriptionResponse.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause = index;
        index++;
    }

    printRICSubscriptionResponse(&ricSubscriptionResponse);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if (packRICSubscriptionResponse(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionResponse) == e2err_OK)
    {
        memset(&ricSubscriptionResponse,0, sizeof ricSubscriptionResponse);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2SuccessfulOutcome) {
                if (messageInfo.messageId == cRICSubscriptionResponse) {
                    if ((returnCode = getRICSubscriptionResponseData(pE2AP_PDU, &ricSubscriptionResponse)) == e2err_OK) {
                        printRICSubscriptionResponse(&ricSubscriptionResponse);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionResponseData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionResponse\n");
            }
            else
                printf("Not SuccessfulOutcome\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionFailure() {
    // Test RICSubscribeFailure
    RICSubscriptionFailure_t ricSubscriptionFailure;
    ricSubscriptionFailure.ricRequestID.ricRequestorID = 1;
    ricSubscriptionFailure.ricRequestID.ricRequestSequenceNumber = 22;
    ricSubscriptionFailure.ranFunctionID = 33;
    ricSubscriptionFailure.ricActionNotAdmittedList.contentLength = 16;
    uint64_t index = 0;
    while (index < ricSubscriptionFailure.ricActionNotAdmittedList.contentLength) {
        ricSubscriptionFailure.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID = index;
        ricSubscriptionFailure.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = cRICCauseRadioNetwork;
        ricSubscriptionFailure.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause = index;
        index++;
    }
    // CriticalityDiagnostics, OPTIONAL. Not used in RIC
    ricSubscriptionFailure.criticalityDiagnosticsPresent = false;
    ricSubscriptionFailure.criticalityDiagnostics.procedureCodePresent = true;
    ricSubscriptionFailure.criticalityDiagnostics.procedureCode = 1;
    ricSubscriptionFailure.criticalityDiagnostics.triggeringMessagePresent = true;
    ricSubscriptionFailure.criticalityDiagnostics.triggeringMessage = TriggeringMessage__initiating_message;
    ricSubscriptionFailure.criticalityDiagnostics.procedureCriticalityPresent = true;
    ricSubscriptionFailure.criticalityDiagnostics.procedureCriticality = Criticality__reject;

    ricSubscriptionFailure.criticalityDiagnostics.iEsCriticalityDiagnosticsPresent = false;
    ricSubscriptionFailure.criticalityDiagnostics.criticalityDiagnosticsIELength = 256;
    uint16_t index2 = 0;
    while (index2 < ricSubscriptionFailure.criticalityDiagnostics.criticalityDiagnosticsIELength) {
        ricSubscriptionFailure.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iECriticality = Criticality__reject;
        ricSubscriptionFailure.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].iE_ID = index2;
        ricSubscriptionFailure.criticalityDiagnostics.criticalityDiagnosticsIEListItem[index2].typeOfError = TypeOfError_missing;
        index2++;
    }

    printRICSubscriptionFailure(&ricSubscriptionFailure);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if (packRICSubscriptionFailure(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionFailure) == e2err_OK)
    {
        memset(&ricSubscriptionFailure,0, sizeof ricSubscriptionFailure);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2UnsuccessfulOutcome) {
                if (messageInfo.messageId == cRICSubscriptionFailure) {
                    if ((returnCode = getRICSubscriptionFailureData(pE2AP_PDU, &ricSubscriptionFailure)) == e2err_OK) {
                        printRICSubscriptionFailure(&ricSubscriptionFailure);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionFailureData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionFailure\n");
            }
            else
                printf("Not UnuccessfulOutcome\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionDeleteRequest() {

    RICSubscriptionDeleteRequest_t ricSubscriptionDeleteRequest;
    ricSubscriptionDeleteRequest.ricRequestID.ricRequestorID = 1;
    ricSubscriptionDeleteRequest.ricRequestID.ricRequestSequenceNumber = 22;
    ricSubscriptionDeleteRequest.ranFunctionID = 33;

    printRICSubscriptionDeleteRequest(&ricSubscriptionDeleteRequest);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[cDataBufferSize];
    if ((packRICSubscriptionDeleteRequest(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionDeleteRequest)) == e2err_OK)
    {
        memset(&ricSubscriptionDeleteRequest,0, sizeof ricSubscriptionDeleteRequest);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2InitiatingMessage) {
                if (messageInfo.messageId == cRICSubscriptionDeleteRequest) {
                    if ((returnCode = getRICSubscriptionDeleteRequestData(pE2AP_PDU, &ricSubscriptionDeleteRequest)) == e2err_OK) {
                        printRICSubscriptionDeleteRequest(&ricSubscriptionDeleteRequest);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionDeleteRequestData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionDeleteRequest\n");
            }
            else
                printf("Not InitiatingMessage\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionDeleteResponse() {

    RICSubscriptionDeleteResponse_t ricSubscriptionDeleteResponse;
    ricSubscriptionDeleteResponse.ricRequestID.ricRequestorID = 1;
    ricSubscriptionDeleteResponse.ricRequestID.ricRequestSequenceNumber = 22;
    ricSubscriptionDeleteResponse.ranFunctionID = 33;

    printRICSubscriptionDeleteResponse(&ricSubscriptionDeleteResponse);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if ((packRICSubscriptionDeleteResponse(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionDeleteResponse)) == e2err_OK)
    {
        memset(&ricSubscriptionDeleteResponse,0, sizeof ricSubscriptionDeleteResponse);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2SuccessfulOutcome) {
                if (messageInfo.messageId == cRICsubscriptionDeleteResponse) {
                    if ((returnCode = getRICSubscriptionDeleteResponseData(pE2AP_PDU, &ricSubscriptionDeleteResponse)) == e2err_OK) {
                        printRICSubscriptionDeleteResponse(&ricSubscriptionDeleteResponse);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionDeleteResponseData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionDeleteResponse\n");
            }
            else
                printf("Not SuccessfulOutcome\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionDeleteFailure() {

    RICSubscriptionDeleteFailure_t ricSubscriptionDeleteFailure;
    ricSubscriptionDeleteFailure.ricRequestID.ricRequestorID = 1;
    ricSubscriptionDeleteFailure.ricRequestID.ricRequestSequenceNumber = 22;
    ricSubscriptionDeleteFailure.ranFunctionID = 33;
    ricSubscriptionDeleteFailure.ricCause.content = cRICCauseRadioNetwork;
    ricSubscriptionDeleteFailure.ricCause.cause = 3;

    printRICSubscriptionDeleteFailure(&ricSubscriptionDeleteFailure);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if ((packRICSubscriptionDeleteFailure(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionDeleteFailure)) == e2err_OK)
    {
        memset(&ricSubscriptionDeleteFailure,0, sizeof ricSubscriptionDeleteFailure);
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2UnsuccessfulOutcome) {
                if (messageInfo.messageId == cRICsubscriptionDeleteFailure) {
                    if ((returnCode = getRICSubscriptionDeleteFailureData(pE2AP_PDU, &ricSubscriptionDeleteFailure)) == e2err_OK) {
                        printRICSubscriptionDeleteFailure(&ricSubscriptionDeleteFailure);
                        return true;
                    }
                    else
                        printf("Error in getRICSubscriptionDeleteFailureData. ReturnCode = %s",getE2ErrorString(returnCode));
                }
                else
                    printf("Not RICSubscriptionDeleteFailure\n");
            }
            else
                printf("Not UnuccessfulOutcome\n");
        }
        else
            printf("%s",logBuffer);
    }
    else
        printf("%s",logBuffer);
    return false;
}

//////////////////////////////////////////////////////////////////////
void printDataBuffer(const size_t byteCount, const uint8_t* pData) {

    uint64_t index = 0;
    while (index < byteCount) {
        if (index % 50 == 0) {
            printf("\n");
        }
        printf("%u ",pData[index]);
        index++;
    }
}

//////////////////////////////////////////////////////////////////////
void printRICSubscriptionRequest(const RICSubscriptionRequest_t* pRICSubscriptionRequest) {
    printf("pRICSubscriptionRequest->ricRequestID.ricRequestorID = %u\n", pRICSubscriptionRequest->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionRequest->ricRequestID.ricRequestSequenceNumber = %u\n", pRICSubscriptionRequest->ricRequestID.ricRequestSequenceNumber);
    printf("pRICSubscriptionRequest->ranFunctionID = %u\n",pRICSubscriptionRequest->ranFunctionID);

    printf("pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.nodeIDbits = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.nodeID.bits);
    printf("pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID = %u\n",
        (unsigned)pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID);
    printf("pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.interfaceDirection = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.interfaceDirection);
    printf("pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.interfaceMessageType.procedureCode = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.interfaceMessageType.procedureCode);
    printf("pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.interfaceMessageType.typeOfMessage = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscription.ricEventTriggerDefinition.interfaceMessageType.typeOfMessage);
    printf("pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.contentLength = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.contentLength);

    uint64_t index = 0;
    while (index < pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.contentLength) {
        printf("pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID = %li\n",
             pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID);
        printf("pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType = %li\n",
             pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType);
        printf("pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = %i\n",
             pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent);
        if(pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent)
        {
            printf("pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinition.styleID = %li\n",
                 pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinition.styleID);
            printf("pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinition.sequenceOfActionParameters.parameterID = %i\n",
                 pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinition.sequenceOfActionParameters.parameterID);
        }
        printf("pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent = %i\n",
          pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent);
        if(pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent)
        {
            printf("pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType = %li\n",
                 pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType);
            printf("pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait = %li\n",
                 pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait);
        }
        printf("\n\n");
        index++;
    }
    printf("\n\n");
}

//////////////////////////////////////////////////////////////////////
void printRICSubscriptionResponse(const RICSubscriptionResponse_t* pRICSubscriptionResponse) {

    printf("pRICSubscriptionResponse->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionResponse->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionResponse->ricRequestID.ricRequestSequenceNumber = %u\n", pRICSubscriptionResponse->ricRequestID.ricRequestSequenceNumber);
    printf("pRICSubscriptionResponse->ranFunctionID = %u\n",pRICSubscriptionResponse->ranFunctionID);
    printf("pRICSubscriptionResponse->ricActionAdmittedList.contentLength = %u\n",(unsigned)pRICSubscriptionResponse->ricActionAdmittedList.contentLength);
    uint64_t index = 0;
    while (index < pRICSubscriptionResponse->ricActionAdmittedList.contentLength) {
        printf("pRICSubscriptionResponse->ricActionAdmittedList.ricActionID[index] = %lu\n",pRICSubscriptionResponse->ricActionAdmittedList.ricActionID[index]);
        index++;
    }
    printf("pRICSubscriptionResponse->ricActionNotAdmittedListPresent = %u\n",pRICSubscriptionResponse->ricActionNotAdmittedListPresent);
    printf("pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength = %u\n",(unsigned)pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength);
    index = 0;
    while (index < pRICSubscriptionResponse->ricActionNotAdmittedList.contentLength) {
        printf("pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID = %lu\n",
             pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID);
        printf("pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = %u\n",
             (unsigned)pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content);
        printf("pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause = %u\n",
             (unsigned)pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause);
        index++;
    }
    printf("\n");
}

//////////////////////////////////////////////////////////////////////
void printRICSubscriptionFailure(const RICSubscriptionFailure_t* pRICSubscriptionFailure) {

    printf("pRICSubscriptionFailure->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionFailure->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionFailure->ricRequestID.ricRequestSequenceNumber = %u\n",pRICSubscriptionFailure->ricRequestID.ricRequestSequenceNumber);
    printf("pRICSubscriptionFailure->ranFunctionID = %i\n",pRICSubscriptionFailure->ranFunctionID);
    printf("pRICSubscriptionFailure->ricActionNotAdmittedList.contentLength = %u\n",(unsigned)pRICSubscriptionFailure->ricActionNotAdmittedList.contentLength);
    uint64_t index = 0;
    while (index < pRICSubscriptionFailure->ricActionNotAdmittedList.contentLength) {
        printf("pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID = %lu\n",
             pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID);
        printf("pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content = %u\n",
            (unsigned)pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.content);
        printf("pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause = %u\n",
             (unsigned)pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricCause.cause);
        index++;
    }
    if (pRICSubscriptionFailure->criticalityDiagnosticsPresent) {
        printf("pRICSubscriptionFailure->criticalityDiagnosticsPresent = %u\n",pRICSubscriptionFailure->criticalityDiagnosticsPresent);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.procedureCodePresent = %u\n",pRICSubscriptionFailure->criticalityDiagnostics.procedureCodePresent);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.procedureCode = %u\n",(unsigned)pRICSubscriptionFailure->criticalityDiagnostics.procedureCode);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.triggeringMessagePresent = %u\n",pRICSubscriptionFailure->criticalityDiagnostics.triggeringMessagePresent);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.triggeringMessage = %u\n",(unsigned)pRICSubscriptionFailure->criticalityDiagnostics.triggeringMessage);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.procedureCriticalityPresent = %u\n",pRICSubscriptionFailure->criticalityDiagnostics.procedureCriticalityPresent);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.procedureCriticality = %u\n",(unsigned)pRICSubscriptionFailure->criticalityDiagnostics.procedureCriticality);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.iEsCriticalityDiagnosticsPresent = %u\n",pRICSubscriptionFailure->criticalityDiagnostics.iEsCriticalityDiagnosticsPresent);
        printf("pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIELength = %u\n",pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIELength);
        index = 0;
        while (index < pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIELength) {
            printf("pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].iECriticality = %u\n",
                 (unsigned)pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].iECriticality);
            printf("pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].iE_ID = %u\n",
                 pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].iE_ID);
            printf("pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].typeOfError = %u\n",
                 (unsigned)pRICSubscriptionFailure->criticalityDiagnostics.criticalityDiagnosticsIEListItem[index].typeOfError);
            index++;
        }
    }
    printf("\n");
}

void printRICSubscriptionDeleteRequest(const RICSubscriptionDeleteRequest_t* pRICSubscriptionDeleteRequest) {

    printf("\npRICSubscriptionDeleteRequest->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionDeleteRequest->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionDeleteRequest->ricRequestID.ricRequestSequenceNumber = %u\n",pRICSubscriptionDeleteRequest->ricRequestID.ricRequestSequenceNumber);
    printf("pRICSubscriptionDeleteRequest->ranFunctionID = %i\n",pRICSubscriptionDeleteRequest->ranFunctionID);
    printf("\n");
}

void printRICSubscriptionDeleteResponse(const RICSubscriptionDeleteResponse_t* pRICSubscriptionDeleteResponse) {

    printf("\npRICSubscriptionDeleteResponse->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionDeleteResponse->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionDeleteResponse->ricRequestID.ricRequestSequenceNumber = %u\n",pRICSubscriptionDeleteResponse->ricRequestID.ricRequestSequenceNumber);
    printf("pRICSubscriptionDeleteResponse->ranFunctionID = %i\n",pRICSubscriptionDeleteResponse->ranFunctionID);
    printf("\n");
}

void printRICSubscriptionDeleteFailure(const RICSubscriptionDeleteFailure_t* pRICSubscriptionDeleteFailure) {

    printf("\npRICSubscriptionDeleteFailure->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionDeleteFailure->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionDeleteFailure->ricRequestID.ricRequestSequenceNumber = %u\n",pRICSubscriptionDeleteFailure->ricRequestID.ricRequestSequenceNumber);
    printf("pRICSubscriptionDeleteFailure->ranFunctionID = %i\n",pRICSubscriptionDeleteFailure->ranFunctionID);
    printf("pRICSubscriptionDeleteFailure->ricCause.content = %i\n",pRICSubscriptionDeleteFailure->ricCause.content);
    printf("pRICSubscriptionDeleteFailure->ricCause.cause = %i\n",pRICSubscriptionDeleteFailure->ricCause.cause);
    printf("\n");
}

#endif
