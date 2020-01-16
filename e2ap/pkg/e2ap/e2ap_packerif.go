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

package e2ap

import (
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/packer"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerIf interface {
	Pack(*packer.PackedData) (error, *packer.PackedData)
	UnPack(msg *packer.PackedData) error
	String() string
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionRequestIf interface {
	E2APMsgPackerIf
	Set(*E2APSubscriptionRequest) error
	Get() (error, *E2APSubscriptionRequest)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionResponseIf interface {
	E2APMsgPackerIf
	Set(*E2APSubscriptionResponse) error
	Get() (error, *E2APSubscriptionResponse)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionFailureIf interface {
	E2APMsgPackerIf
	Set(*E2APSubscriptionFailure) error
	Get() (error, *E2APSubscriptionFailure)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionDeleteRequestIf interface {
	E2APMsgPackerIf
	Set(*E2APSubscriptionDeleteRequest) error
	Get() (error, *E2APSubscriptionDeleteRequest)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionDeleteResponseIf interface {
	E2APMsgPackerIf
	Set(*E2APSubscriptionDeleteResponse) error
	Get() (error, *E2APSubscriptionDeleteResponse)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APMsgPackerSubscriptionDeleteFailureIf interface {
	E2APMsgPackerIf
	Set(*E2APSubscriptionDeleteFailure) error
	Get() (error, *E2APSubscriptionDeleteFailure)
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APPackerIf interface {
	NewPackerSubscriptionRequest() E2APMsgPackerSubscriptionRequestIf
	NewPackerSubscriptionResponse() E2APMsgPackerSubscriptionResponseIf
	NewPackerSubscriptionFailure() E2APMsgPackerSubscriptionFailureIf
	NewPackerSubscriptionDeleteRequest() E2APMsgPackerSubscriptionDeleteRequestIf
	NewPackerSubscriptionDeleteResponse() E2APMsgPackerSubscriptionDeleteResponseIf
	NewPackerSubscriptionDeleteFailure() E2APMsgPackerSubscriptionDeleteFailureIf
	MessageInfo(msg *packer.PackedData) *packer.MessageInfo
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
type E2APAutoPacker struct {
	packer E2APPackerIf
}

func NewE2APAutoPacker(packer E2APPackerIf) *E2APAutoPacker {
	return &E2APAutoPacker{packer: packer}
}

// TODO improve openasn handling to reuse PDU etc...
// Now practically decodes two times each E2/X2 message, as first round solves message type
func (autopacker *E2APAutoPacker) UnPack(msg *packer.PackedData) (error, interface{}) {
	var err error = nil
	msgInfo := autopacker.packer.MessageInfo(msg)
	if msgInfo != nil {
		switch msgInfo.MsgType {
		case E2AP_InitiatingMessage:
			switch msgInfo.MsgId {
			case E2AP_RICSubscriptionRequest:
				unpa := autopacker.packer.NewPackerSubscriptionRequest()
				err = unpa.UnPack(msg)
				if err == nil {
					return unpa.Get()
				}
			case E2AP_RICSubscriptionDeleteRequest:
				unpa := autopacker.packer.NewPackerSubscriptionDeleteRequest()
				err = unpa.UnPack(msg)
				if err == nil {
					return unpa.Get()
				}
			default:
				err = fmt.Errorf("MsgType: E2AP_InitiatingMessage => MsgId:%d unknown", msgInfo.MsgId)
			}
		case E2AP_SuccessfulOutcome:
			switch msgInfo.MsgId {
			case E2AP_RICSubscriptionResponse:
				unpa := autopacker.packer.NewPackerSubscriptionResponse()
				err = unpa.UnPack(msg)
				if err == nil {
					return unpa.Get()
				}
			case E2AP_RICSubscriptionDeleteResponse:
				unpa := autopacker.packer.NewPackerSubscriptionDeleteResponse()
				err = unpa.UnPack(msg)
				if err == nil {
					return unpa.Get()
				}
			default:
				err = fmt.Errorf("MsgType: E2AP_SuccessfulOutcome => MsgId:%d unknown", msgInfo.MsgId)
			}
		case E2AP_UnsuccessfulOutcome:
			switch msgInfo.MsgId {
			case E2AP_RICSubscriptionFailure:
				unpa := autopacker.packer.NewPackerSubscriptionFailure()
				err = unpa.UnPack(msg)
				if err == nil {
					return unpa.Get()
				}
			case E2AP_RICSubscriptionDeleteFailure:
				unpa := autopacker.packer.NewPackerSubscriptionDeleteFailure()
				err = unpa.UnPack(msg)
				if err == nil {
					return unpa.Get()
				}
			default:
				err = fmt.Errorf("MsgType: E2AP_UnsuccessfulOutcome => MsgId:%d unknown", msgInfo.MsgId)
			}
		default:
			err = fmt.Errorf("MsgType: %d and MsgId:%d unknown", msgInfo.MsgType, msgInfo.MsgId)
		}
	} else {
		err = fmt.Errorf("MsgInfo not received")
	}
	return err, nil
}

func (autopacker *E2APAutoPacker) Pack(data interface{}, trg *packer.PackedData) (error, *packer.PackedData) {
	var err error = nil
	switch themsg := data.(type) {
	case *E2APSubscriptionRequest:
		pa := autopacker.packer.NewPackerSubscriptionRequest()
		err = pa.Set(themsg)
		if err == nil {
			return pa.Pack(trg)
		}
	case *E2APSubscriptionResponse:
		pa := autopacker.packer.NewPackerSubscriptionResponse()
		err = pa.Set(themsg)
		if err == nil {
			return pa.Pack(trg)
		}
	case *E2APSubscriptionFailure:
		pa := autopacker.packer.NewPackerSubscriptionFailure()
		err = pa.Set(themsg)
		if err == nil {
			return pa.Pack(trg)
		}
	case *E2APSubscriptionDeleteRequest:
		pa := autopacker.packer.NewPackerSubscriptionDeleteRequest()
		err = pa.Set(themsg)
		if err == nil {
			return pa.Pack(trg)
		}
	case *E2APSubscriptionDeleteResponse:
		pa := autopacker.packer.NewPackerSubscriptionDeleteResponse()
		err = pa.Set(themsg)
		if err == nil {
			return pa.Pack(trg)
		}
	case *E2APSubscriptionDeleteFailure:
		pa := autopacker.packer.NewPackerSubscriptionDeleteFailure()
		err = pa.Set(themsg)
		if err == nil {
			return pa.Pack(trg)
		}
	default:
		err = fmt.Errorf("unknown message")
	}
	return err, nil
}
