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

func (c *E2ap) GetSubscriptionRequestSequenceNumber(payload []byte) (sub_id uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_request_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, errors.New("e2ap wrapper is unable to get Subscirption Request Sequence Number due to wrong or invalid payload")
	}
	sub_id = uint16(cret)
	return
}

func (c *E2ap) SetSubscriptionRequestSequenceNumber(payload []byte, newSubscriptionid uint16) (new_payload []byte, err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_request_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return make([]byte, 0), errors.New("e2ap wrapper is unable to set Subscirption Request Sequence Number due to wrong or invalid payload")
	}
	new_payload = C.GoBytes(cptr, C.int(size))
	return
}

func (c *E2ap) GetSubscriptionResponseSequenceNumber(payload []byte) (sub_id uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_response_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, errors.New("e2ap wrapper is unable to get Subscirption Response Sequence Number due to wrong or invalid payload")
	}
	sub_id = uint16(cret)
	return
}

func (c *E2ap) SetSubscriptionResponseSequenceNumber(payload []byte, newSubscriptionid uint16) (new_payload []byte, err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_response_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return make([]byte, 0), errors.New("e2ap wrapper is unable to set Subscirption Reponse Sequence Number due to wrong or invalid payload")
	}
	new_payload = C.GoBytes(cptr, C.int(size))
	return
}
