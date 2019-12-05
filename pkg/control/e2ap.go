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

package control

/*
#include <wrapper.h>

#cgo LDFLAGS: -lwrapper
*/
import "C"

import (
	"errors"
	"unsafe"
)

type E2ap struct {
}

func (c *E2ap) GetSubscriptionRequestSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_request_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, errors.New("e2ap wrapper is unable to get Subscirption Request Sequence Number due to wrong or invalid payload")
	}
	subId = uint16(cret)
	return
}

func (c *E2ap) SetSubscriptionRequestSequenceNumber(payload []byte, newSubscriptionid uint16) (newPayload []byte, err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_request_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return make([]byte, 0), errors.New("e2ap wrapper is unable to set Subscription Request Sequence Number due to wrong or invalid payload")
	}
	newPayload = C.GoBytes(cptr, C.int(size))
	return
}

func (c *E2ap) GetSubscriptionResponseSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_response_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, errors.New("e2ap wrapper is unable to get Subscirption Response Sequence Number due to wrong or invalid payload")
	}
	subId = uint16(cret)
	return
}

func (c *E2ap) SetSubscriptionResponseSequenceNumber(payload []byte, newSubscriptionid uint16) (newPayload []byte, err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_response_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return make([]byte, 0), errors.New("e2ap wrapper is unable to set Subscription Response Sequence Number due to wrong or invalid payload")
	}
	newPayload = C.GoBytes(cptr, C.int(size))
	return
}

/* RICsubscriptionDeleteRequest */

func (c *E2ap) GetSubscriptionDeleteRequestSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_delete_request_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, errors.New("e2ap wrapper is unable to get Subscirption Delete Request Sequence Number due to wrong or invalid payload")
	}
	subId = uint16(cret)
	return
}

func (c *E2ap) SetSubscriptionDeleteRequestSequenceNumber(payload []byte, newSubscriptionid uint16) (newPayload []byte, err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_delete_request_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return make([]byte, 0), errors.New("e2ap wrapper is unable to set Subscription Delete Request Sequence Number due to wrong or invalid payload")
	}
	newPayload = C.GoBytes(cptr, C.int(size))
	return
}

/* RICsubscriptionDeleteResponse */

func (c *E2ap) GetSubscriptionDeleteResponseSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_delete_response_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, errors.New("e2ap wrapper is unable to get Subscirption Delete Response Sequence Number due to wrong or invalid payload")
	}
	subId = uint16(cret)
	return
}

func (c *E2ap) SetSubscriptionDeleteResponseSequenceNumber(payload []byte, newSubscriptionid uint16) (newPayload []byte, err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_delete_response_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return make([]byte, 0), errors.New("e2ap wrapper is unable to set Subscription Delete Response Sequence Number due to wrong or invalid payload")
	}
	newPayload = C.GoBytes(cptr, C.int(size))
	return
}

/* RICsubscriptionRequestFailure */

func (c *E2ap) GetSubscriptionFailureSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_failure_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, errors.New("e2ap wrapper is unable to get Subscirption Failure Sequence Number due to wrong or invalid payload")
	}
	subId = uint16(cret)
	return
}

// This function is not used currently. Can be deleted if not needed
func (c *E2ap) SetSubscriptionFailureSequenceNumber(payload []byte, newSubscriptionid uint16) (newPayload []byte, err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_failure_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return make([]byte, 0), errors.New("e2ap wrapper is unable to set Subscription Failure Sequence Number due to wrong or invalid payload")
	}
	newPayload = C.GoBytes(cptr, C.int(size))
	return
}

/* RICsubscriptionDeleteFailure */

func (c *E2ap) GetSubscriptionDeleteFailureSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_delete_failure_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, errors.New("e2ap wrapper is unable to get Subscirption Delete Failure Sequence Number due to wrong or invalid payload")
	}
	subId = uint16(cret)
	return
}

// This function is not used currently. Can be deleted if not needed
func (c *E2ap) SetSubscriptionDeleteFailureSequenceNumber(payload []byte, newSubscriptionid uint16) (newPayload []byte, err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_delete_failure_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return make([]byte, 0), errors.New("e2ap wrapper is unable to set Subscription Delete Failure Sequence Number due to wrong or invalid payload")
	}
	newPayload = C.GoBytes(cptr, C.int(size))
	return
}
