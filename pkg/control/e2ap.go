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

#cgo LDFLAGS: -le2ap_wrapper -le2ap
*/
import "C"

import (
	"fmt"
	"unsafe"
)

type E2ap struct {
}

/* RICsubscriptionRequest */

// Used by e2t test stub
func (c *E2ap) GetSubscriptionRequestSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_request_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Request Sequence Number due to wrong or invalid payload. ErrorCode: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by submgr, rco test stub
func (c *E2ap) SetSubscriptionRequestSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_request_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Request Sequence Number due to wrong or invalid payload. ErrorCode: %v", size)
	}
	return
}

// Used by submgr, rco test stub
func (c *E2ap) GetSubscriptionResponseSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_response_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Response Sequence Number due to wrong or invalid payload. ErrorCode: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by e2t test stub
func (c *E2ap) SetSubscriptionResponseSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_response_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Response Sequence Number due to wrong or invalid payload. ErrorCode: %v", size)
	}
	return
}

/* RICsubscriptionDeleteRequest */

// Used by submgr, e2t test stub
func (c *E2ap) GetSubscriptionDeleteRequestSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_delete_request_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Delete Request Sequence Number due to wrong or invalid payload. ErrorCode: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by rco test stub
func (c *E2ap) SetSubscriptionDeleteRequestSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_delete_request_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Delete Request Sequence Number due to wrong or invalid payload. ErrorCode: %v", size)
	}
	return
}

/* RICsubscriptionDeleteResponse */

// Used by submgr, rco test stub
func (c *E2ap) GetSubscriptionDeleteResponseSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_delete_response_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Delete Response Sequence Number due to wrong or invalid payload. ErrorCode: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by e2t test stub
func (c *E2ap) SetSubscriptionDeleteResponseSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_delete_response_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Delete Response Sequence Number due to wrong or invalid payload. ErrorCode: %v", size)
	}
	return
}

/* RICsubscriptionRequestFailure */

// Used by submgr
func (c *E2ap) GetSubscriptionFailureSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_failure_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Failure Sequence Number due to wrong or invalid payload. ErrorCode: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by submgr
func (c *E2ap) SetSubscriptionFailureSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_failure_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Failure Sequence Number due to wrong or invalid payload. ErrorCode: %v", size)
	}
	return
}

/* RICsubscriptionDeleteFailure */

// Used by submgr
func (c *E2ap) GetSubscriptionDeleteFailureSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_delete_failure_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Delete Failure Sequence Number due to wrong or invalid payload. ErrorCode: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by submgr
func (c *E2ap) SetSubscriptionDeleteFailureSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_delete_failure_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Delete Failure Sequence Number due to wrong or invalid payload. ErrorCode: %v", size)
	}
	return
}
