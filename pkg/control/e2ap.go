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
	"encoding/hex"
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/packer"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
)

var packerif e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

type E2ap struct {
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionRequest(payload []byte) (*e2ap.E2APSubscriptionRequest, error) {
	e2SubReq := packerif.NewPackerSubscriptionRequest()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err, subReq := e2SubReq.UnPack(packedData)
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subReq, nil
}

func (c *E2ap) PackSubscriptionRequest(req *e2ap.E2APSubscriptionRequest) (int, *packer.PackedData, error) {
	e2SubReq := packerif.NewPackerSubscriptionRequest()
	err, packedData := e2SubReq.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_REQ, packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionResponse(payload []byte) (*e2ap.E2APSubscriptionResponse, error) {
	e2SubResp := packerif.NewPackerSubscriptionResponse()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err, subResp := e2SubResp.UnPack(packedData)
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subResp, nil
}

func (c *E2ap) PackSubscriptionResponse(req *e2ap.E2APSubscriptionResponse) (int, *packer.PackedData, error) {
	e2SubResp := packerif.NewPackerSubscriptionResponse()
	err, packedData := e2SubResp.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_RESP, packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionFailure(payload []byte) (*e2ap.E2APSubscriptionFailure, error) {
	e2SubFail := packerif.NewPackerSubscriptionFailure()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err, subFail := e2SubFail.UnPack(packedData)
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subFail, nil
}

func (c *E2ap) PackSubscriptionFailure(req *e2ap.E2APSubscriptionFailure) (int, *packer.PackedData, error) {
	e2SubFail := packerif.NewPackerSubscriptionFailure()
	err, packedData := e2SubFail.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_FAILURE, packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionDeleteRequest(payload []byte) (*e2ap.E2APSubscriptionDeleteRequest, error) {
	e2SubDelReq := packerif.NewPackerSubscriptionDeleteRequest()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err, subDelReq := e2SubDelReq.UnPack(packedData)
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subDelReq, nil
}

func (c *E2ap) PackSubscriptionDeleteRequest(req *e2ap.E2APSubscriptionDeleteRequest) (int, *packer.PackedData, error) {
	e2SubDelReq := packerif.NewPackerSubscriptionDeleteRequest()
	err, packedData := e2SubDelReq.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_DEL_REQ, packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionDeleteResponse(payload []byte) (*e2ap.E2APSubscriptionDeleteResponse, error) {
	e2SubDelResp := packerif.NewPackerSubscriptionDeleteResponse()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err, subDelResp := e2SubDelResp.UnPack(packedData)
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subDelResp, nil
}

func (c *E2ap) PackSubscriptionDeleteResponse(req *e2ap.E2APSubscriptionDeleteResponse) (int, *packer.PackedData, error) {
	e2SubDelResp := packerif.NewPackerSubscriptionDeleteResponse()
	err, packedData := e2SubDelResp.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_DEL_RESP, packedData, nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionDeleteFailure(payload []byte) (*e2ap.E2APSubscriptionDeleteFailure, error) {
	e2SubDelFail := packerif.NewPackerSubscriptionDeleteFailure()
	packedData := &packer.PackedData{}
	packedData.Buf = payload
	err, subDelFail := e2SubDelFail.UnPack(packedData)
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subDelFail, nil
}

func (c *E2ap) PackSubscriptionDeleteFailure(req *e2ap.E2APSubscriptionDeleteFailure) (int, *packer.PackedData, error) {
	e2SubDelFail := packerif.NewPackerSubscriptionDeleteFailure()
	err, packedData := e2SubDelFail.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_DEL_FAILURE, packedData, nil
}
