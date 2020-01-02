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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/packer"
	"unsafe"
)

var packerif e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

type E2ap struct {
}

/* RICsubscriptionRequest */

// Used by e2t test stub
func (c *E2ap) GetSubscriptionRequestSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_request_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Request Sequence Number due to wrong or invalid payload. Erroxappde: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by submgr, xapp test stub
func (c *E2ap) SetSubscriptionRequestSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_request_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Request Sequence Number due to wrong or invalid payload. Erroxappde: %v", size)
	}
	return
}

// Used by submgr, xapp test stub
func (c *E2ap) GetSubscriptionResponseSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_response_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Response Sequence Number due to wrong or invalid payload. Erroxappde: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by e2t test stub
func (c *E2ap) SetSubscriptionResponseSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_response_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Response Sequence Number due to wrong or invalid payload. Erroxappde: %v", size)
	}
	return
}

/* RICsubscriptionDeleteRequest */

// Used by submgr, e2t test stub
func (c *E2ap) GetSubscriptionDeleteRequestSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_delete_request_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Delete Request Sequence Number due to wrong or invalid payload. Erroxappde: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by xapp test stub
func (c *E2ap) SetSubscriptionDeleteRequestSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_delete_request_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Delete Request Sequence Number due to wrong or invalid payload. Erroxappde: %v", size)
	}
	return
}

/* RICsubscriptionDeleteResponse */

// Used by submgr, e2t test stub
func (c *E2ap) GetSubscriptionDeleteResponseSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_delete_response_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Delete Response Sequence Number due to wrong or invalid payload. Erroxappde: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by e2t test stub
func (c *E2ap) SetSubscriptionDeleteResponseSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_delete_response_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Delete Response Sequence Number due to wrong or invalid payload. Erroxappde: %v", size)
	}
	return
}

/* RICsubscriptionRequestFailure */

// Used by submgr
func (c *E2ap) GetSubscriptionFailureSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_failure_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Failure Sequence Number due to wrong or invalid payload. Erroxappde: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by e2t test stub
func (c *E2ap) SetSubscriptionFailureSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_failure_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Failure Sequence Number due to wrong or invalid payload. Erroxappde: %v", size)
	}
	return
}

/* RICsubscriptionDeleteFailure */

// Used by submgr
func (c *E2ap) GetSubscriptionDeleteFailureSequenceNumber(payload []byte) (subId uint16, err error) {
	cptr := unsafe.Pointer(&payload[0])
	cret := C.e2ap_get_ric_subscription_delete_failure_sequence_number(cptr, C.size_t(len(payload)))
	if cret < 0 {
		return 0, fmt.Errorf("e2ap wrapper is unable to get Subscirption Delete Failure Sequence Number due to wrong or invalid payload. Erroxappde: %v", cret)
	}
	subId = uint16(cret)
	return
}

// Used by submgr
func (c *E2ap) SetSubscriptionDeleteFailureSequenceNumber(payload []byte, newSubscriptionid uint16) (err error) {
	cptr := unsafe.Pointer(&payload[0])
	size := C.e2ap_set_ric_subscription_delete_failure_sequence_number(cptr, C.size_t(len(payload)), C.long(newSubscriptionid))
	if size < 0 {
		return fmt.Errorf("e2ap wrapper is unable to set Subscription Delete Failure Sequence Number due to wrong or invalid payload. Erroxappde: %v", size)
	}
	return
}

// Used by submgr
func (c *E2ap) PackSubscriptionDeleteResponse(payload []byte, newSubscriptionid uint16) (newPayload []byte, err error) {
	e2SubDelReq := packerif.NewPackerSubscriptionDeleteRequest()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err = e2SubDelReq.UnPack(packedData)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("PackSubDelResp: UnPack() failed: %s", err.Error())
	}
	getErr, subDelReq := e2SubDelReq.Get()
	if getErr != nil {
		return make([]byte, 0), fmt.Errorf("PackSubDelResp: Get() failed: %s", getErr.Error())
	}

	e2SubDelResp := packerif.NewPackerSubscriptionDeleteResponse()
	subDelResp := e2ap.E2APSubscriptionDeleteResponse{}
	subDelResp.RequestId.Id = subDelReq.RequestId.Id
	subDelResp.RequestId.Seq = uint32(newSubscriptionid)
	subDelResp.FunctionId = subDelReq.FunctionId
	err = e2SubDelResp.Set(&subDelResp)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("PackSubDelResp: Set() failed: %s", err.Error())
	}
	err, packedData = e2SubDelResp.Pack(nil)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("PackSubDelResp: Pack() failed: %s", err.Error())
	}
	return packedData.Buf, nil
}

// Used by submgr
func (c *E2ap) PackSubscriptionDeleteRequest(payload []byte, newSubscriptionid uint16) (newPayload []byte, err error) {
	e2SubReq := packerif.NewPackerSubscriptionRequest()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err = e2SubReq.UnPack(packedData)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("PackSubDelReq: UnPack() failed: %s", err.Error())
	}
	getErr, subReq := e2SubReq.Get()
	if getErr != nil {
		return make([]byte, 0), fmt.Errorf("PackSubDelReq: Get() failed: %s", getErr.Error())
	}

	e2SubDel := packerif.NewPackerSubscriptionDeleteRequest()
	subDelReq := e2ap.E2APSubscriptionDeleteRequest{}
	subDelReq.RequestId.Id = subReq.RequestId.Id
	subDelReq.RequestId.Seq = uint32(newSubscriptionid)
	subDelReq.FunctionId = subReq.FunctionId
	err = e2SubDel.Set(&subDelReq)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("PackSubDelReq: Set() failed: %s", err.Error())
	}
	err, packedData = e2SubDel.Pack(nil)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("PackSubDelReq: Pack() failed: %s", err.Error())
	}
	return packedData.Buf, nil
}
