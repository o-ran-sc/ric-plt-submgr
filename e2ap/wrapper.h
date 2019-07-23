#ifndef	_WRAPPER_H_
#define	_WRAPPER_H_

#include "RICsubscriptionRequest.h"
#include "RICsubscriptionResponse.h"
#include "ProtocolIE-Container.h"
#include "ProtocolIE-Field.h"


ssize_t encode_RIC_subscription_request(RICsubscriptionRequest_t* pdu, void* buffer, size_t buf_size);
RICsubscriptionRequest_t* decode_RIC_subscription_request(const void *buffer, size_t buf_size);

long e2ap_get_ric_subscription_request_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_request_sequence_number(void *buffer, size_t buf_size, long sequence_number);

ssize_t encode_RIC_subscription_response(RICsubscriptionResponse_t* pdu, void* buffer, size_t buf_size);
RICsubscriptionResponse_t* decode_RIC_subscription_response(const void *buffer, size_t buf_size);

long e2ap_get_ric_subscription_response_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_response_sequence_number(void *buffer, size_t buf_size, long sequence_number);


#endif /* _WRAPPER_H_ */