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
    mem_track_hdr_t memHead;

    printf(" sizeof RICSubscriptionRequest_t = %li\n", sizeof (RICSubscriptionRequest_t));

    ricSubscriptionRequest.ricRequestID.ricRequestorID = 1;
    ricSubscriptionRequest.ricRequestID.ricInstanceID = 22;
    ricSubscriptionRequest.ranFunctionID = 33;

    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.octetString.contentLength = 0;

    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBIDPresent = true;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalGNBIDPresent = false;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.contentLength = 3;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[0] = 1;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[1] = 2;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.pLMNIdentity.pLMNIdentityVal[2] = 3;

//    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.eNBID.bits = cMacroENBIDP_20Bits;
//    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.eNBID.bits = cHomeENBID_28Bits;
//    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.eNBID.bits = cShortMacroENBID_18Bits;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits = clongMacroENBIDP_21Bits;

    IdOctects_t eNBOctects;
    memset(eNBOctects.octets, 0, sizeof(eNBOctects));
    eNBOctects.octets[0] = 11;
    eNBOctects.octets[1] = 22;
    eNBOctects.octets[2] = 31;
    eNBOctects.octets[3] = 1;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID = eNBOctects.nodeID;
    printf("eNBOctects.nodeID = %u\n\n",eNBOctects.nodeID);

    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceDirection = InterfaceDirection__incoming;
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceMessageType.procedureCode = 35;  // id-rRCTransfer
    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceMessageType.typeOfMessage = cE2InitiatingMessage;

    ricSubscriptionRequest.ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceProtocolIEListPresent = false;

    ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength = 1;
    uint64_t index = 0;
    while (index < ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength) {
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID = 255; //index;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType = RICActionType_insert;

        // ricActionDefinition, OPTIONAL.
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = true;

        char data[3] = {'A','B','C'};

        // Format 1
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1Present = false;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.styleID = 255;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterCount = 1;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem->parameterID = 1;  // index
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[0].actionParameterValue.valueIntPresent = false;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[0].actionParameterValue.valueInt = 111;

        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[0].actionParameterValue.valueOctSPresent = true;
        if (addOctetString(&memHead,&ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[0].actionParameterValue.valueOctS,sizeof(data),data) == NULL) {
            printf("addOctetString failure\n");
            return false;
        }

        // Format 2
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2Present = true;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupCount = 1;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranUEgroupID = 0;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranUEgroupDefinitionCount = 1;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranUEgroupDefinition.ranUeGroupDefItem[0].ranParameterID = 0;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranUEgroupDefinition.ranUeGroupDefItem[0].ranParameterTest = RANParameterTest_equal;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranUEgroupDefinition.ranUeGroupDefItem[0].ranParameterValue.valueIntPresent = false;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranUEgroupDefinition.ranUeGroupDefItem[0].ranParameterValue.valueInt = 1;

        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranUEgroupDefinition.ranUeGroupDefItem[0].ranParameterValue.valueOctSPresent = true;
        addOctetString(&memHead,&ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranUEgroupDefinition.ranUeGroupDefItem[0].ranParameterValue.valueOctS,sizeof(data),data);

        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranPolicyCount = 1;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranPolicy.ranParameterItem[0].ranParameterID = 0;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranPolicy.ranParameterItem[0].ranParameterValue.valueIntPresent = true;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[0].ranPolicy.ranParameterItem[0].ranParameterValue.valueInt = 2;

        // ricSubsequentActionPresent, OPTIONAL
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent = true;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType = RICSubsequentActionType_Continue;
        ricSubscriptionRequest.ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait = RICTimeToWait_w100ms;
        index++;
    }

    printf(" sizeof RICSubscriptionRequest_t = %li\n", sizeof (RICSubscriptionRequest_t));

    printRICSubscriptionRequest(&ricSubscriptionRequest);

    uint64_t logBufferSize = 1024;
    char logBuffer[logBufferSize];
    uint64_t dataBufferSize = cDataBufferSize;
    byte dataBuffer[dataBufferSize];
    if (packRICSubscriptionRequest(&dataBufferSize, dataBuffer, logBuffer, &ricSubscriptionRequest) == e2err_OK)
    {
        memset(&ricSubscriptionRequest,0, sizeof (RICSubscriptionRequest_t));
        uint64_t returnCode;
        E2MessageInfo_t messageInfo;
        e2ap_pdu_ptr_t* pE2AP_PDU = unpackE2AP_pdu(dataBufferSize, dataBuffer, logBuffer, &messageInfo);
        if (pE2AP_PDU != 0) {
            if (messageInfo.messageType == cE2InitiatingMessage) {
                if (messageInfo.messageId == cRICSubscriptionRequest) {
                    if ((returnCode = getRICSubscriptionRequestData(&memHead, pE2AP_PDU, &ricSubscriptionRequest)) == e2err_OK) {
                        printRICSubscriptionRequest(&ricSubscriptionRequest);
                        mem_track_free(&memHead);
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

    mem_track_free(&memHead);
    return false;
}

//////////////////////////////////////////////////////////////////////
bool TestRICSubscriptionResponse() {
    // Test RICSubscribeResponse
    RICSubscriptionResponse_t ricSubscriptionResponse;
    ricSubscriptionResponse.ricRequestID.ricRequestorID = 1;
    ricSubscriptionResponse.ricRequestID.ricInstanceID = 22;
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
        ricSubscriptionResponse.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID = index;  //0..255
        ricSubscriptionResponse.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = cCauseRICService;
        ricSubscriptionResponse.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal = 2;
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
    ricSubscriptionFailure.ricRequestID.ricInstanceID = 22;
    ricSubscriptionFailure.ranFunctionID = 33;
    ricSubscriptionFailure.ricActionNotAdmittedList.contentLength = 16;
    uint64_t index = 0;
    while (index < ricSubscriptionFailure.ricActionNotAdmittedList.contentLength) {
        ricSubscriptionFailure.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID = index;
        ricSubscriptionFailure.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = cCauseRICService;
        ricSubscriptionFailure.ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal = 2;
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
    ricSubscriptionDeleteRequest.ricRequestID.ricInstanceID = 22;
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
    ricSubscriptionDeleteResponse.ricRequestID.ricInstanceID = 22;
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
    ricSubscriptionDeleteFailure.ricRequestID.ricInstanceID = 22;
    ricSubscriptionDeleteFailure.ranFunctionID = 33;
    ricSubscriptionDeleteFailure.cause.content = cCauseRICService;
    ricSubscriptionDeleteFailure.cause.causeVal = 2;

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
        if (index > 0 && index % 50 == 0) {
            printf("\n");
        }
        printf("%u ",pData[index]);
        index++;
    }
}

//////////////////////////////////////////////////////////////////////
void printRICSubscriptionRequest(const RICSubscriptionRequest_t* pRICSubscriptionRequest) {
    printf("pRICSubscriptionRequest->ricRequestID.ricRequestorID = %u\n", pRICSubscriptionRequest->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionRequest->ricRequestID.ricInstanceID = %u\n", pRICSubscriptionRequest->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionRequest->ranFunctionID = %u\n",pRICSubscriptionRequest->ranFunctionID);

    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeIDbits = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.bits);
    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID = %u\n",
        (unsigned)pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceID.globalENBID.nodeID.nodeID);
    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceDirection = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceDirection);
    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceMessageType.procedureCode = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceMessageType.procedureCode);
    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceMessageType.typeOfMessage = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscriptionDetails.ricEventTriggerDefinition.e2SMgNBX2eventTriggerDefinition.interfaceMessageType.typeOfMessage);
    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength = %u\n",
         (unsigned)pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength);

    uint64_t index = 0;
    while (index < pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.contentLength) {
        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID = %li\n",
             pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionID);
        printf("pRICSubscriptionRequest->ricSubscription.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType = %li\n",
             pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionType);
        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent = %i\n",
             pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent);
        if(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionPresent)
        {

            if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1Present) {
                // Format 1
                printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1Present = %i\n",
                    pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1Present);
                printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.styleID = %li\n",
                    pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.styleID);
                printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterCount = %i\n",
                    pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterCount);
                printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index]ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem->parameterID = %i\n",
                     pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem->parameterID);

                uint64_t index2 = 0;
                while (index2 < pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterCount) {
                    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueIntPresent = %i\n",
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueIntPresent);
                    if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueIntPresent) {
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueInt = %li\n",
                          pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueInt);
                    }
                    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueEnumPresent = %i\n",
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueEnumPresent);
                    if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueEnumPresent) {
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueEnum = %li\n",
                          pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueEnum);
                    }
                    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBoolPresent = %i\n",
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBoolPresent);
                    if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBoolPresent) {
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBool = %i\n",
                          pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBool);
                    }
                    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBitSPresent = %i\n",
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBitSPresent);
                    if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBitSPresent) {
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBitS.byteLength = %li\n",
                          pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBitS.byteLength);
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBitS.data = ");
                        printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBitS.byteLength,
                                        pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueBitS.data);
                        printf("\n");
                    }
                    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueOctSPresent = %i\n",
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueOctSPresent);
                    if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueOctSPresent) {
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueOctS.contentLength = %li\n",
                          pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueOctS.contentLength);
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueOctS.data = ");
                          printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueOctS.contentLength,
                                          pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueOctS.data);
                        printf("\n");
                    }
                    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valuePrtSPresent = %i\n",
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valuePrtSPresent);
                    if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valuePrtSPresent) {
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueOctS.contentLength = %li\n",
                          pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valueOctS.contentLength);
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valuePrtS.data = ");
                        printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valuePrtS.contentLength,
                                        pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat1.actionParameterItem[index2].actionParameterValue.valuePrtS.data);
                        printf("\n");
                    }
                    index2++;
                }
            }
            else if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2Present) {
                printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2Present = %i\n",
                  pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2Present);

                // Format 2
                printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupCount = %i\n",
                    pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupCount);

                uint64_t index2 = 0;
                while (index2 < pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupCount) {
                    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupID = %li\n",
                        pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupID);
                    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinitionCount = %i\n",
                        pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinitionCount);

                    uint64_t index3 = 0;
                    while (index3 < pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinitionCount) {

                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterID = %i\n",
                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterID);
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterTest = %i\n",
                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterTest);
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueIntPresent = %i\n",
                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueIntPresent);
                        if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueIntPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueInt = %li\n",
                                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueInt);
                        }
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueEnumPresent = %i\n",
                               pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueEnumPresent);
                        if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueEnumPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueEnum = %li\n",
                                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueEnum);
                        }
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBoolPresent = %i\n",
                               pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBoolPresent);
                        if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBoolPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBool = %i\n",
                                pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBool);
                        }
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBitSPresent = %i\n",
                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBitSPresent);
                        if ( pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBitSPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue,valueBitS.byteLength = %li\n",
                                   pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBitS.byteLength);
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBitS.data = ");
                            printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBitS.byteLength,
                                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueBitS.data);
                            printf("\n");
                        }
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueOctSPresent = %i\n",
                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueOctSPresent);
                        if(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueOctSPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue,valueOctS.contentLength = %li\n",
                                   pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueOctS.contentLength);
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueOctS.data = ");
                            printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueOctS.contentLength,
                                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valueOctS.data);
                            printf("\n");
                        }
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valuePrtSPresent = %i\n",
                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valuePrtSPresent);
                        if(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valuePrtSPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue,valuePrtS.contentLength = %li\n",
                                   pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valuePrtS.contentLength);
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valuePrtS.data = ");
                            printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valuePrtS.contentLength,
                                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranUEgroupDefinition.ranUeGroupDefItem[index3].ranParameterValue.valuePrtS.data);
                            printf("\n");
                        }
                        index3++;
                    }

                    printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicyCount %i\n",
                      pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicyCount);

                    index3 = 0;
                    while (index3 < pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicyCount) {
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterID = %i\n",
                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterID);
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueIntPresent = %i\n",
                               pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueIntPresent);
                        if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueIntPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueInt = %li\n",
                                   pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueInt);
                        }
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueEnumPresent = %i\n",
                               pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueEnumPresent);
                        if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueEnumPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueEnum = %li\n",
                                   pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueEnum);
                        }
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBoolPresent = %i\n",
                               pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBoolPresent);
                        if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBoolPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBool = %i\n",
                                   pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBool);
                        }
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitSPresent = %i\n",
                               pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitSPresent);
                        if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitSPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS.byteLength, = %li\n",
                                   pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS.byteLength);
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS.data = ");
                            printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS.byteLength,
                                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueBitS.data);
                            printf("\n");
                        }
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctSPresent = %i\n",
                               pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctSPresent);
                        if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctSPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctS.contentLength = %li\n",
                                   pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctS.contentLength);
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctS.data = ");
                            printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctS.contentLength,
                                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valueOctS.data);
                            printf("\n");
                        }
                        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtSPresent = %i\n",
                               pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtSPresent);
                        if (pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtSPresent) {
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtS.contentLength = %li\n",
                                   pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtS.contentLength);
                            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtS.data = ");
                            printDataBuffer(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtS.contentLength,
                                            pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricActionDefinitionChoice.actionDefinitionFormat2.ranUeGroupItem[index2].ranPolicy.ranParameterItem[index3].ranParameterValue.valuePrtS.data);
                            printf("\n");
                        }
                        index3++;
                    }
                    index2++;
                }
            }
        }
        printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent = %i\n",
          pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent);
        if(pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentActionPresent)
        {
            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType = %li\n",
                 pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricSubsequentActionType);
            printf("pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait = %li\n",
                 pRICSubscriptionRequest->ricSubscriptionDetails.ricActionToBeSetupItemIEs.ricActionToBeSetupItem[index].ricSubsequentAction.ricTimeToWait);
        }
        printf("\n\n");
        index++;
    }
    printf("\n\n");
}

//////////////////////////////////////////////////////////////////////
void printRICSubscriptionResponse(const RICSubscriptionResponse_t* pRICSubscriptionResponse) {

    printf("pRICSubscriptionResponse->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionResponse->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionResponse->ricRequestID.ricInstanceID = %u\n", pRICSubscriptionResponse->ricRequestID.ricInstanceID);
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
        printf("pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = %u\n",
             (unsigned)pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content);
        printf("pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.cause = %u\n",
             (unsigned)pRICSubscriptionResponse->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal);
        index++;
    }
    printf("\n");
}

//////////////////////////////////////////////////////////////////////
void printRICSubscriptionFailure(const RICSubscriptionFailure_t* pRICSubscriptionFailure) {

    printf("pRICSubscriptionFailure->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionFailure->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionFailure->ricRequestID.ricInstanceID = %u\n",pRICSubscriptionFailure->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionFailure->ranFunctionID = %i\n",pRICSubscriptionFailure->ranFunctionID);
    printf("pRICSubscriptionFailure->ricActionNotAdmittedList.contentLength = %u\n",(unsigned)pRICSubscriptionFailure->ricActionNotAdmittedList.contentLength);
    uint64_t index = 0;
    while (index < pRICSubscriptionFailure->ricActionNotAdmittedList.contentLength) {
        printf("pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID = %lu\n",
             pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].ricActionID);
        printf("pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content = %u\n",
            (unsigned)pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.content);
        printf("pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.cause = %u\n",
             (unsigned)pRICSubscriptionFailure->ricActionNotAdmittedList.RICActionNotAdmittedItem[index].cause.causeVal);
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
    printf("pRICSubscriptionDeleteRequest->ricRequestID.ricInstanceID = %u\n",pRICSubscriptionDeleteRequest->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionDeleteRequest->ranFunctionID = %i\n",pRICSubscriptionDeleteRequest->ranFunctionID);
    printf("\n");
}

void printRICSubscriptionDeleteResponse(const RICSubscriptionDeleteResponse_t* pRICSubscriptionDeleteResponse) {

    printf("\npRICSubscriptionDeleteResponse->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionDeleteResponse->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionDeleteResponse->ricRequestID.ricInstanceID = %u\n",pRICSubscriptionDeleteResponse->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionDeleteResponse->ranFunctionID = %i\n",pRICSubscriptionDeleteResponse->ranFunctionID);
    printf("\n");
}

void printRICSubscriptionDeleteFailure(const RICSubscriptionDeleteFailure_t* pRICSubscriptionDeleteFailure) {

    printf("\npRICSubscriptionDeleteFailure->ricRequestID.ricRequestorID = %u\n",pRICSubscriptionDeleteFailure->ricRequestID.ricRequestorID);
    printf("pRICSubscriptionDeleteFailure->ricRequestID.ricInstanceID = %u\n",pRICSubscriptionDeleteFailure->ricRequestID.ricInstanceID);
    printf("pRICSubscriptionDeleteFailure->ranFunctionID = %i\n",pRICSubscriptionDeleteFailure->ranFunctionID);
    printf("pRICSubscriptionDeleteFailure->ricCause.content = %i\n",pRICSubscriptionDeleteFailure->cause.content);
    printf("pRICSubscriptionDeleteFailure->ricCause.cause = %i\n",pRICSubscriptionDeleteFailure->cause.causeVal);
    printf("\n");
}

#endif
