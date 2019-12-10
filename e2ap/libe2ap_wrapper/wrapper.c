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

#include <errno.h>
#include "wrapper.h"

size_t encode_E2AP_PDU(E2AP_PDU_t* pdu, void* buffer, size_t buf_size)
{
    asn_enc_rval_t encode_result;
    encode_result = aper_encode_to_buffer(&asn_DEF_E2AP_PDU, NULL, pdu, buffer, buf_size);
    if (encode_result.encoded == -1) {
        return -1;
    }
    return encode_result.encoded;
}

E2AP_PDU_t* decode_E2AP_PDU(const void* buffer, size_t buf_size)
{
    asn_dec_rval_t decode_result;
    E2AP_PDU_t *pdu = 0;
    decode_result = aper_decode_complete(NULL, &asn_DEF_E2AP_PDU, (void **)&pdu, buffer, buf_size);
    if (decode_result.code == RC_OK) {
        return pdu;
    } else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
        return 0;
    }
}

long e2ap_get_ric_subscription_request_sequence_number(void *buffer, size_t buf_size)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if  (pdu != NULL && pdu->present == E2AP_PDU_PR_initiatingMessage)
    {
        RICInitiatingMessage_t* initiatingMessage = &pdu->choice.initiatingMessage;
        if ( initiatingMessage->procedureCode == ProcedureCode_id_ricSubscription
            && initiatingMessage->value.present == RICInitiatingMessage__value_PR_RICsubscriptionRequest)
        {
            RICsubscriptionRequest_t *ric_subscription_request = &(initiatingMessage->value.choice.RICsubscriptionRequest);
            for (int i = 0; i < ric_subscription_request->protocolIEs.list.count; ++i)
            {
                if (ric_subscription_request->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    long sequence_number = ric_subscription_request->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber;
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return sequence_number;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

ssize_t  e2ap_set_ric_subscription_request_sequence_number(void *buffer, size_t buf_size, long sequence_number)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_initiatingMessage)
    {
        RICInitiatingMessage_t* initiatingMessage = &pdu->choice.initiatingMessage;
        if ( initiatingMessage->procedureCode == ProcedureCode_id_ricSubscription
            && initiatingMessage->value.present == RICInitiatingMessage__value_PR_RICsubscriptionRequest)
        {
            RICsubscriptionRequest_t *ricSubscriptionRequest = &initiatingMessage->value.choice.RICsubscriptionRequest;
            for (int i = 0; i < ricSubscriptionRequest->protocolIEs.list.count; ++i)
            {
                if (ricSubscriptionRequest->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    ricSubscriptionRequest->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber = sequence_number;
                    size_t encode_size = encode_E2AP_PDU(pdu, buffer, buf_size);
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return encode_size;
                }
                else
                    errorCode = -3;
            }
        }
        else
            return -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

/* RICsubscriptionResponse */
long e2ap_get_ric_subscription_response_sequence_number(void *buffer, size_t buf_size)
{
    int errorCode = -1;
     E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_successfulOutcome)
    {
        RICSuccessfulOutcome_t* successfulOutcome = &pdu->choice.successfulOutcome;
        if ( successfulOutcome->procedureCode == ProcedureCode_id_ricSubscription
            && successfulOutcome->value.present == RICSuccessfulOutcome__value_PR_RICsubscriptionResponse)
        {
            RICsubscriptionResponse_t *ricSubscriptionResponse = &successfulOutcome->value.choice.RICsubscriptionResponse;
            for (int i = 0; i < ricSubscriptionResponse->protocolIEs.list.count; ++i)
            {
                if (ricSubscriptionResponse->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    long sequence_number = ricSubscriptionResponse->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber;
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return sequence_number;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

ssize_t  e2ap_set_ric_subscription_response_sequence_number(void *buffer, size_t buf_size, long sequence_number)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_successfulOutcome)
    {
        RICSuccessfulOutcome_t* successfulOutcome = &pdu->choice.successfulOutcome;
        if ( successfulOutcome->procedureCode == ProcedureCode_id_ricSubscription
            && successfulOutcome->value.present == RICSuccessfulOutcome__value_PR_RICsubscriptionResponse)
        {
            RICsubscriptionResponse_t *ricSubscriptionResponse = &successfulOutcome->value.choice.RICsubscriptionResponse;
            for (int i = 0; i < ricSubscriptionResponse->protocolIEs.list.count; ++i)
            {
                if (ricSubscriptionResponse->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    ricSubscriptionResponse->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber = sequence_number;
                    size_t encode_size = encode_E2AP_PDU(pdu, buffer, buf_size);
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return encode_size;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

/* RICsubscriptionDeleteRequest */
long e2ap_get_ric_subscription_delete_request_sequence_number(void *buffer, size_t buf_size)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_initiatingMessage)
    {
        RICInitiatingMessage_t* initiatingMessage = &pdu->choice.initiatingMessage;
        if ( initiatingMessage->procedureCode == ProcedureCode_id_ricSubscriptionDelete
            && initiatingMessage->value.present == RICInitiatingMessage__value_PR_RICsubscriptionDeleteRequest )
        {
            RICsubscriptionDeleteRequest_t *subscriptionDeleteRequest = &initiatingMessage->value.choice.RICsubscriptionDeleteRequest;
            for (int i = 0; i < subscriptionDeleteRequest->protocolIEs.list.count; ++i)
            {
                if (subscriptionDeleteRequest->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    long sequence_number = subscriptionDeleteRequest->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber;
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return sequence_number;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

ssize_t  e2ap_set_ric_subscription_delete_request_sequence_number(void *buffer, size_t buf_size, long sequence_number)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_initiatingMessage)
    {
        RICInitiatingMessage_t* initiatingMessage = &pdu->choice.initiatingMessage;
        if ( initiatingMessage->procedureCode == ProcedureCode_id_ricSubscriptionDelete
            && initiatingMessage->value.present == RICInitiatingMessage__value_PR_RICsubscriptionDeleteRequest )
        {
            RICsubscriptionDeleteRequest_t* subscriptionDeleteRequest = &initiatingMessage->value.choice.RICsubscriptionDeleteRequest;
            for (int i = 0; i < subscriptionDeleteRequest->protocolIEs.list.count; ++i)
            {
                if (subscriptionDeleteRequest->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    subscriptionDeleteRequest->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber = sequence_number;
                    size_t encode_size = encode_E2AP_PDU(pdu, buffer, buf_size);
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return encode_size;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

/* RICsubscriptionDeleteResponse */
long e2ap_get_ric_subscription_delete_response_sequence_number(void *buffer, size_t buf_size)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_successfulOutcome)
    {
        RICSuccessfulOutcome_t* successfulOutcome = &pdu->choice.successfulOutcome;
        if ( successfulOutcome->procedureCode == ProcedureCode_id_ricSubscriptionDelete
            && successfulOutcome->value.present == RICSuccessfulOutcome__value_PR_RICsubscriptionDeleteResponse )
        {
            RICsubscriptionDeleteResponse_t* subscriptionDeleteResponse = &successfulOutcome->value.choice.RICsubscriptionDeleteResponse;
            for (int i = 0; i < subscriptionDeleteResponse->protocolIEs.list.count; ++i)
            {
                if (subscriptionDeleteResponse->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    long sequence_number = subscriptionDeleteResponse->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber;
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return sequence_number;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

ssize_t  e2ap_set_ric_subscription_delete_response_sequence_number(void *buffer, size_t buf_size, long sequence_number)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_successfulOutcome)
    {
        RICSuccessfulOutcome_t* successfulOutcome = &pdu->choice.successfulOutcome;
        if ( successfulOutcome->procedureCode == ProcedureCode_id_ricSubscriptionDelete
            && successfulOutcome->value.present == RICSuccessfulOutcome__value_PR_RICsubscriptionDeleteResponse )
        {
            RICsubscriptionDeleteResponse_t* subscriptionDeleteResponse = &successfulOutcome->value.choice.RICsubscriptionDeleteResponse;
            for (int i = 0; i < subscriptionDeleteResponse->protocolIEs.list.count; ++i)
            {
                if (subscriptionDeleteResponse->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    subscriptionDeleteResponse->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber = sequence_number;
                    size_t encode_size = encode_E2AP_PDU(pdu, buffer, buf_size);
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return encode_size;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

// This function is not used currently. Can be deleted if not needed
ssize_t  e2ap_set_ric_subscription_failure_sequence_number(void *buffer, size_t buf_size, long sequence_number)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_unsuccessfulOutcome)
    {
        RICUnsuccessfulOutcome_t* unsuccessfulOutcome = &pdu->choice.unsuccessfulOutcome;
        if (unsuccessfulOutcome->procedureCode == ProcedureCode_id_ricSubscription
            && unsuccessfulOutcome->value.present == RICUnsuccessfulOutcome__value_PR_RICsubscriptionFailure)
        {
            RICsubscriptionFailure_t* subscriptionFailure = &unsuccessfulOutcome->value.choice.RICsubscriptionFailure;
            for (int i = 0; i < subscriptionFailure->protocolIEs.list.count; ++i)
            {
                if (subscriptionFailure->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    subscriptionFailure->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber = sequence_number;
                    size_t encode_size = encode_E2AP_PDU(pdu, buffer, buf_size);
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return encode_size;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

long e2ap_get_ric_subscription_failure_sequence_number(void *buffer, size_t buf_size)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_unsuccessfulOutcome)
    {
        RICUnsuccessfulOutcome_t* unsuccessfulOutcome = &pdu->choice.unsuccessfulOutcome;
        if (unsuccessfulOutcome->procedureCode == ProcedureCode_id_ricSubscription
            && unsuccessfulOutcome->value.present == RICUnsuccessfulOutcome__value_PR_RICsubscriptionFailure)
        {
            RICsubscriptionFailure_t* subscriptionFailure = &unsuccessfulOutcome->value.choice.RICsubscriptionFailure;;
            for (int i = 0; i < subscriptionFailure->protocolIEs.list.count; ++i)
            {
                if (subscriptionFailure->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    long sequence_number = subscriptionFailure->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber;
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return sequence_number;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

// This function is not used currently. Can be deleted if not needed
ssize_t  e2ap_set_ric_subscription_delete_failure_sequence_number(void *buffer, size_t buf_size, long sequence_number)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_unsuccessfulOutcome)
    {
        RICUnsuccessfulOutcome_t* unsuccessfulOutcome = &pdu->choice.unsuccessfulOutcome;
        if (unsuccessfulOutcome->procedureCode == ProcedureCode_id_ricSubscriptionDelete
            && unsuccessfulOutcome->value.present == RICUnsuccessfulOutcome__value_PR_RICsubscriptionDeleteFailure)
        {
            RICsubscriptionDeleteFailure_t* subscriptionDeleteFailure = &unsuccessfulOutcome->value.choice.RICsubscriptionDeleteFailure;
            for (int i = 0; i < subscriptionDeleteFailure->protocolIEs.list.count; ++i)
            {
                if (subscriptionDeleteFailure->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    subscriptionDeleteFailure->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber = sequence_number;
                    size_t encode_size = encode_E2AP_PDU(pdu, buffer, buf_size);
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return encode_size;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}

long e2ap_get_ric_subscription_delete_failure_sequence_number(void *buffer, size_t buf_size)
{
    int errorCode = -1;
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if (pdu != NULL && pdu->present == E2AP_PDU_PR_unsuccessfulOutcome)
    {
        RICUnsuccessfulOutcome_t* unsuccessfulOutcome = &pdu->choice.unsuccessfulOutcome;
        if (unsuccessfulOutcome->procedureCode == ProcedureCode_id_ricSubscriptionDelete
            && unsuccessfulOutcome->value.present == RICUnsuccessfulOutcome__value_PR_RICsubscriptionDeleteFailure)
        {
            RICsubscriptionDeleteFailure_t* subscriptionDeleteFailure = &unsuccessfulOutcome->value.choice.RICsubscriptionDeleteFailure;;
            for (int i = 0; i < subscriptionDeleteFailure->protocolIEs.list.count; ++i)
            {
                if (subscriptionDeleteFailure->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID)
                {
                    long sequence_number = subscriptionDeleteFailure->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber;
                    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
                    return sequence_number;
                }
                else
                    errorCode = -3;
            }
        }
        else
            errorCode = -2;
    }
    ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
    return errorCode;
}
