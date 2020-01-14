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
func (c *E2ap) PackSubscriptionDeleteResponseFromSubDelReq(payload []byte, newSubscriptionid uint16) (newPayload []byte, err error) {

	subDelReq, err := c.UnpackSubscriptionDeleteRequest(payload)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("PackSubDelRespFromSubDelReq: SubDelReq unpack failed: %s", err.Error())
	}

	subDelResp := &e2ap.E2APSubscriptionDeleteResponse{}
	subDelResp.RequestId.Id = subDelReq.RequestId.Id
	subDelResp.RequestId.Seq = uint32(newSubscriptionid)
	subDelResp.FunctionId = subDelReq.FunctionId

	packedData, err := c.PackSubscriptionDeleteResponse(subDelResp)
	if err != nil {
		return make([]byte, 0), fmt.Errorf("PackSubDelRespFromSubDelReq: SubDelResp pack failed: %s", err.Error())
	}
	return packedData.Buf, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionRequest(payload []byte) (*e2ap.E2APSubscriptionRequest, error) {
	e2SubReq := packerif.NewPackerSubscriptionRequest()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err := e2SubReq.UnPack(packedData)
	if err != nil {
		return nil, err
	}
	err, subReq := e2SubReq.Get()
	if err != nil {
		return nil, err
	}
	return subReq, nil
}

func (c *E2ap) PackSubscriptionRequest(req *e2ap.E2APSubscriptionRequest) (*packer.PackedData, error) {
	e2SubReq := packerif.NewPackerSubscriptionRequest()
	err := e2SubReq.Set(req)
	if err != nil {
		return nil, err
	}
	err, packedData := e2SubReq.Pack(nil)
	if err != nil {
		return nil, err
	}
	return packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionResponse(payload []byte) (*e2ap.E2APSubscriptionResponse, error) {
	e2SubResp := packerif.NewPackerSubscriptionResponse()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err := e2SubResp.UnPack(packedData)
	if err != nil {
		return nil, err
	}
	err, subResp := e2SubResp.Get()
	if err != nil {
		return nil, err
	}
	return subResp, nil
}

func (c *E2ap) PackSubscriptionResponse(req *e2ap.E2APSubscriptionResponse) (*packer.PackedData, error) {
	e2SubResp := packerif.NewPackerSubscriptionResponse()
	err := e2SubResp.Set(req)
	if err != nil {
		return nil, err
	}
	err, packedData := e2SubResp.Pack(nil)
	if err != nil {
		return nil, err
	}
	return packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionDeleteRequest(payload []byte) (*e2ap.E2APSubscriptionDeleteRequest, error) {
	e2SubDelReq := packerif.NewPackerSubscriptionDeleteRequest()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err := e2SubDelReq.UnPack(packedData)
	if err != nil {
		return nil, err
	}
	err, subDelReq := e2SubDelReq.Get()
	if err != nil {
		return nil, err
	}
	return subDelReq, nil
}

func (c *E2ap) PackSubscriptionDeleteRequest(req *e2ap.E2APSubscriptionDeleteRequest) (*packer.PackedData, error) {
	e2SubDelReq := packerif.NewPackerSubscriptionDeleteRequest()
	err := e2SubDelReq.Set(req)
	if err != nil {
		return nil, err
	}
	err, packedData := e2SubDelReq.Pack(nil)
	if err != nil {
		return nil, err
	}
	return packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionDeleteResponse(payload []byte) (*e2ap.E2APSubscriptionDeleteResponse, error) {
	e2SubDelResp := packerif.NewPackerSubscriptionDeleteResponse()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err := e2SubDelResp.UnPack(packedData)
	if err != nil {
		return nil, err
	}
	err, subDelResp := e2SubDelResp.Get()
	if err != nil {
		return nil, err
	}
	return subDelResp, nil
}

func (c *E2ap) PackSubscriptionDeleteResponse(req *e2ap.E2APSubscriptionDeleteResponse) (*packer.PackedData, error) {
	e2SubDelResp := packerif.NewPackerSubscriptionDeleteResponse()
	err := e2SubDelResp.Set(req)
	if err != nil {
		return nil, err
	}
	err, packedData := e2SubDelResp.Pack(nil)
	if err != nil {
		return nil, err
	}
	return packedData, nil
}
