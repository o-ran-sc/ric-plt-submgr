#ifndef	_WRAPPER_H_
#define	_WRAPPER_H_

#include "RICsubscriptionRequest.h"
#include "RICsubscriptionResponse.h"
#include "RICsubscriptionDeleteRequest.h"
#include "RICsubscriptionDeleteResponse.h"
#include "RICsubscriptionFailure.h"
#include "RICsubscriptionDeleteFailure.h"
#include "E2AP-PDU.h"
#include "InitiatingMessage.h"
#include "SuccessfulOutcome.h"
#include "UnsuccessfulOutcome.h"
#include "ProtocolIE-Container.h"
#include "ProtocolIE-Field.h"

size_t encode_E2AP_PDU(E2AP_PDU_t* pdu, void* buffer, size_t buf_size);
E2AP_PDU_t* decode_E2AP_PDU(const void* buffer, size_t buf_size);

long e2ap_get_ric_subscription_request_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_request_sequence_number(void *buffer, size_t buf_size, long sequence_number);
RICsubscription_t* e2ap_get_ric_subscription_request_ric_subscription(void *buffer, size_t buffer_size);

/* RICsubscriptionResponse */
long e2ap_get_ric_subscription_response_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_response_sequence_number(void *buffer, size_t buf_size, long sequence_number);

/* RICsubscriptionDeleteRequest */
long e2ap_get_ric_subscription_delete_request_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_delete_request_sequence_number(void *buffer, size_t buf_size, long sequence_number);

/* RICsubscriptionDeleteResponse */
long e2ap_get_ric_subscription_delete_response_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_delete_response_sequence_number(void *buffer, size_t buf_size, long sequence_number);

/* RICsubscriptionFailure */
long e2ap_get_ric_subscription_failure_sequence_number(void *buffer, size_t buf_size);
// This function is not used currently. Can be deleted if not needed
ssize_t  e2ap_set_ric_subscription_failure_sequence_number(void *buffer, size_t buf_size, long sequence_number);

/* RICsubscriptionFailure */
long e2ap_get_ric_subscription_delete_failure_sequence_number(void *buffer, size_t buf_size);
// This function is not used currently. Can be deleted if not needed
ssize_t  e2ap_set_ric_subscription_delete_failure_sequence_number(void *buffer, size_t buf_size, long sequence_number);

#endif /* _WRAPPER_H_ */
