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

#ifndef	_WRAPPER_H_
#define	_WRAPPER_H_

#ifndef ASN_DISABLE_OER_SUPPORT
#define ASN_DISABLE_OER_SUPPORT
#endif

#include "RICsubscriptionRequest.h"
#include "RICsubscriptionResponse.h"
#include "RICsubscriptionDeleteRequest.h"
#include "RICsubscriptionDeleteResponse.h"
#include "RICsubscriptionFailure.h"
#include "RICsubscriptionDeleteFailure.h"
#include "E2AP-PDU.h"
#include "RICInitiatingMessage.h"
#include "RICSuccessfulOutcome.h"
#include "RICUnsuccessfulOutcome.h"
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
ssize_t  e2ap_set_ric_subscription_failure_sequence_number(void *buffer, size_t buf_size, long sequence_number);

/* RICsubscriptionFailure */
long e2ap_get_ric_subscription_delete_failure_sequence_number(void *buffer, size_t buf_size);
ssize_t  e2ap_set_ric_subscription_delete_failure_sequence_number(void *buffer, size_t buf_size, long sequence_number);


#endif /* _WRAPPER_H_ */
