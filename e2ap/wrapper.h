#ifndef	_WRAPPER_H_
#define	_WRAPPER_H_

#include "RICsubscriptionRequest.h"
#include "RICsubscriptionResponse.h"
#include "RICsubscriptionDeleteRequest.h"
#include "RICsubscriptionDeleteResponse.h"
#include "E2AP-PDU.h"
#include "InitiatingMessageE2.h"
#include "ProtocolIE-Container.h"
#include "ProtocolIE-Field.h"

size_t encode_E2AP_PDU(E2AP_PDU_t* pdu, void* buffer, size_t buf_size);
E2AP_PDU_t* decode_E2AP_PDU(const void* buffer, size_t buf_size);

/* RICsubscriptionRequest */
ssize_t encode_RIC_subscription_request(RICsubscriptionRequest_t* pdu, void* buffer, size_t buf_size);
RICsubscriptionRequest_t* decode_RIC_subscription_request(const void *buffer, size_t buf_size);

long e2ap_get_ric_subscription_request_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_request_sequence_number(void *buffer, size_t buf_size, long sequence_number);
RICsubscription_t* e2ap_get_ric_subscription_request_ric_subscription(void *buffer, size_t buffer_size);

/* RICsubscriptionResponse */
ssize_t encode_RIC_subscription_response(RICsubscriptionResponse_t* pdu, void* buffer, size_t buf_size);
RICsubscriptionResponse_t* decode_RIC_subscription_response(const void *buffer, size_t buf_size);

long e2ap_get_ric_subscription_response_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_response_sequence_number(void *buffer, size_t buf_size, long sequence_number);

/* RICsubscriptionDeleteRequest */
ssize_t encode_RIC_subscription_delete_request(RICsubscriptionDeleteRequest_t* pdu, void* buffer, size_t buf_size);
RICsubscriptionDeleteRequest_t* decode_RIC_subscription_delete_request(const void *buffer, size_t buf_size);

long e2ap_get_ric_subscription_delete_request_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_delete_request_sequence_number(void *buffer, size_t buf_size, long sequence_number);

/* RICsubscriptionDeleteResponse */
ssize_t encode_RIC_subscription_delete_response(RICsubscriptionDeleteResponse_t* pdu, void* buffer, size_t buf_size);
RICsubscriptionDeleteResponse_t* decode_RIC_subscription_delete_response(const void *buffer, size_t buf_size);

long e2ap_get_ric_subscription_delete_response_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_delete_response_sequence_number(void *buffer, size_t buf_size, long sequence_number);



#endif /* _WRAPPER_H_ */
