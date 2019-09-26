#include <errno.h>
#include "wrapper.h"

size_t encode_E2AP_PDU(E2AP_PDU_t* pdu, void* buffer, size_t buf_size)
{
    asn_enc_rval_t encode_result;
    encode_result = aper_encode_to_buffer(&asn_DEF_E2AP_PDU, NULL, pdu, buffer, buf_size);
    if(encode_result.encoded == -1) {
        return -1;
    }
    return encode_result.encoded;
}

E2AP_PDU_t* decode_E2AP_PDU(const void* buffer, size_t buf_size)
{
    asn_dec_rval_t decode_result;
    E2AP_PDU_t *pdu = 0;
    decode_result = aper_decode_complete(NULL, &asn_DEF_E2AP_PDU, (void **)&pdu, buffer, buf_size);
    if(decode_result.code == RC_OK) {
        return pdu;
    } else {
        ASN_STRUCT_FREE(asn_DEF_E2AP_PDU, pdu);
        return 0;
    }
}

long e2ap_get_ric_subscription_request_sequence_number(void *buffer, size_t buf_size)
{
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if ( pdu != NULL && pdu->present == E2AP_PDU_PR_initiatingMessage)
    {
        InitiatingMessageE2_t* initiatingMessage = pdu->choice.initiatingMessage;
        if ( initiatingMessage->procedureCode == ProcedureCode_id_ricSubscription
            && initiatingMessage->value.present == InitiatingMessageE2__value_PR_RICsubscriptionRequest)
        {
            RICsubscriptionRequest_t *ric_subscription_request = &(initiatingMessage->value.choice.RICsubscriptionRequest);
            for (int i = 0; i < ric_subscription_request->protocolIEs.list.count; ++i )
            {
                if ( ric_subscription_request->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID )
                {
                    return ric_subscription_request->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber;
                }
            }
        }
    }
    return -1;
}

ssize_t  e2ap_set_ric_subscription_request_sequence_number(void *buffer, size_t buf_size, long sequence_number)
{
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if ( pdu != NULL && pdu->present == E2AP_PDU_PR_initiatingMessage)
    {
        InitiatingMessageE2_t* initiatingMessage = pdu->choice.initiatingMessage;
        if ( initiatingMessage->procedureCode == ProcedureCode_id_ricSubscription
            && initiatingMessage->value.present == InitiatingMessageE2__value_PR_RICsubscriptionRequest)
        {
            RICsubscriptionRequest_t *ricSubscriptionRequest = &initiatingMessage->value.choice.RICsubscriptionRequest;
            for (int i = 0; i < ricSubscriptionRequest->protocolIEs.list.count; ++i )
            {
                if ( ricSubscriptionRequest->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID )
                {
                    ricSubscriptionRequest->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber = sequence_number;
                    return encode_E2AP_PDU(pdu, buffer, buf_size);
                }
            }
        }
    }
    return -1;
}

/* RICsubscriptionResponse */
long e2ap_get_ric_subscription_response_sequence_number(void *buffer, size_t buf_size)
{
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if ( pdu != NULL && pdu->present == E2AP_PDU_PR_successfulOutcome )
    {
        SuccessfulOutcomeE2_t* successfulOutcome = pdu->choice.successfulOutcome;
        if ( successfulOutcome->procedureCode == ProcedureCode_id_ricSubscription
            && successfulOutcome->value.present == SuccessfulOutcomeE2__value_PR_RICsubscriptionResponse)
        {
            RICsubscriptionResponse_t *ricSubscriptionResponse = &successfulOutcome->value.choice.RICsubscriptionResponse;
            for (int i = 0; i < ricSubscriptionResponse->protocolIEs.list.count; ++i )
            {
                if ( ricSubscriptionResponse->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID )
                {
                    return ricSubscriptionResponse->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber;
                }
            }
        }
    }
    return -1;
}

ssize_t  e2ap_set_ric_subscription_response_sequence_number(void *buffer, size_t buf_size, long sequence_number)
{
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if ( pdu != NULL && pdu->present == E2AP_PDU_PR_successfulOutcome )
    {
        SuccessfulOutcomeE2_t* successfulOutcome = pdu->choice.successfulOutcome;
        if ( successfulOutcome->procedureCode == ProcedureCode_id_ricSubscription
            && successfulOutcome->value.present == SuccessfulOutcomeE2__value_PR_RICsubscriptionResponse)
        {
            RICsubscriptionResponse_t *ricSubscriptionResponse = &successfulOutcome->value.choice.RICsubscriptionResponse;
            for (int i = 0; i < ricSubscriptionResponse->protocolIEs.list.count; ++i )
            {
                if ( ricSubscriptionResponse->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID )
                {
                    ricSubscriptionResponse->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber = sequence_number;
                    return encode_E2AP_PDU(pdu, buffer, buf_size);
                }
            }
        }
    }
    return -1;
}

/* RICsubscriptionDeleteRequest */
long e2ap_get_ric_subscription_delete_request_sequence_number(void *buffer, size_t buf_size)
{
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if ( pdu != NULL && pdu->present == E2AP_PDU_PR_initiatingMessage )
    {
        InitiatingMessageE2_t* initiatingMessage = pdu->choice.initiatingMessage;
        if ( initiatingMessage->procedureCode == ProcedureCode_id_ricSubscriptionDelete
            && initiatingMessage->value.present == InitiatingMessageE2__value_PR_RICsubscriptionDeleteRequest )
        {
            RICsubscriptionDeleteRequest_t *subscriptionDeleteRequest = &initiatingMessage->value.choice.RICsubscriptionDeleteRequest;
            for (int i = 0; i < subscriptionDeleteRequest->protocolIEs.list.count; ++i )
            {
                if ( subscriptionDeleteRequest->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID )
                {
                    return subscriptionDeleteRequest->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber;
                }
            }
        }
    }
    return -1;
}

ssize_t  e2ap_set_ric_subscription_delete_request_sequence_number(void *buffer, size_t buf_size, long sequence_number)
{
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if ( pdu != NULL && pdu->present == E2AP_PDU_PR_initiatingMessage )
    {
        InitiatingMessageE2_t* initiatingMessage = pdu->choice.initiatingMessage;
        if ( initiatingMessage->procedureCode == ProcedureCode_id_ricSubscriptionDelete
            && initiatingMessage->value.present == InitiatingMessageE2__value_PR_RICsubscriptionDeleteRequest )
        {
            RICsubscriptionDeleteRequest_t* subscriptionDeleteRequest = &initiatingMessage->value.choice.RICsubscriptionDeleteRequest;
            for (int i = 0; i < subscriptionDeleteRequest->protocolIEs.list.count; ++i )
            {
                if ( subscriptionDeleteRequest->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID )
                {
                    subscriptionDeleteRequest->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber = sequence_number;
                    return encode_E2AP_PDU(pdu, buffer, buf_size);
                }
            }
        }
    }
    return -1;
}

/* RICsubscriptionDeleteResponse */
long e2ap_get_ric_subscription_delete_response_sequence_number(void *buffer, size_t buf_size)
{
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if ( pdu != NULL && pdu->present == E2AP_PDU_PR_successfulOutcome )
    {
        SuccessfulOutcomeE2_t* successfulOutcome = pdu->choice.successfulOutcome;
        if ( successfulOutcome->procedureCode == ProcedureCode_id_ricSubscriptionDelete
            && successfulOutcome->value.present == SuccessfulOutcomeE2__value_PR_RICsubscriptionDeleteResponse )
        {
            RICsubscriptionDeleteResponse_t* subscriptionDeleteResponse = &successfulOutcome->value.choice.RICsubscriptionDeleteResponse;
            for (int i = 0; i < subscriptionDeleteResponse->protocolIEs.list.count; ++i )
            {
                if ( subscriptionDeleteResponse->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID )
                {
                    return subscriptionDeleteResponse->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber;
                }
            }
        }
    }
    return -1;
}

ssize_t  e2ap_set_ric_subscription_delete_response_sequence_number(void *buffer, size_t buf_size, long sequence_number)
{
    E2AP_PDU_t *pdu = decode_E2AP_PDU(buffer, buf_size);
    if ( pdu != NULL && pdu->present == E2AP_PDU_PR_successfulOutcome )
    {
        SuccessfulOutcomeE2_t* successfulOutcome = pdu->choice.successfulOutcome;
        if ( successfulOutcome->procedureCode == ProcedureCode_id_ricSubscriptionDelete
            && successfulOutcome->value.present == SuccessfulOutcomeE2__value_PR_RICsubscriptionDeleteResponse )
        {
            RICsubscriptionDeleteResponse_t* subscriptionDeleteResponse;
            for (int i = 0; i < subscriptionDeleteResponse->protocolIEs.list.count; ++i )
            {
                if ( subscriptionDeleteResponse->protocolIEs.list.array[i]->id == ProtocolIE_ID_id_RICrequestID )
                {
                    subscriptionDeleteResponse->protocolIEs.list.array[i]->value.choice.RICrequestID.ricRequestSequenceNumber = sequence_number;
                    return encode_E2AP_PDU(pdu, buffer, buf_size);
                }
            }
        }
    }
    return -1;
}
