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
#cgo LDFLAGS: -le2ap_wrapper -le2ap
*/
import "C"

import (
	"encoding/hex"
	"fmt"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap_wrapper"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	"strconv"
)

var packerif e2ap.E2APPackerIf = e2ap_wrapper.NewAsn1E2Packer()

type E2ap struct {
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) FillReportSubReqMsgs(stype models.SubscriptionType, params interface{}, subreqList *e2ap.SubscriptionRequestList, restSubscription *RESTSubscription) error {
	xapp.Logger.Info("FillReportSubReqMsgs")

	p := params.(*models.ReportParams)

	lengthEventTriggers := len(p.EventTriggers)
	if lengthEventTriggers == 0 {
		err := fmt.Errorf("Error in content element count. Count of EventTriggers=%v", lengthEventTriggers)
		return err
	}

	var lengthActionParameters int = 0
	if p.ReportActionDefinitions != nil && p.ReportActionDefinitions.ActionDefinitionFormat1 != nil && p.ReportActionDefinitions.ActionDefinitionFormat1.ActionParameters != nil {
		lengthActionParameters = len(p.ReportActionDefinitions.ActionDefinitionFormat1.ActionParameters)
	}

	xapp.Logger.Info("EventTrigger count=%v, ActionParameter count=%v", lengthEventTriggers, lengthActionParameters)

	// 1..
	for index, restEventTriggerItem := range p.EventTriggers {
		subReqMsg := e2ap.E2APSubscriptionRequest{}
		if p.RANFunctionID != nil {
			subReqMsg.FunctionId = (e2ap.FunctionId)(*p.RANFunctionID)
		}

		if restEventTriggerItem.TriggerNature == "" {
			// E2SM-gNB-X2
			subReqMsg.EventTriggerDefinition.NBX2EventTriggerDefinitionPresent = true
			// Interface-ID is either ENBID or GNBID. GNBID is not present in REST definition currently
			if len(restEventTriggerItem.ENBID) > 0 {
				subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
				plmId64, _ := strconv.Atoi(restEventTriggerItem.PlmnID)
				subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDHomeBits28 // Bit length should be set based calculation. Only 28bit length works in ASN1C currently
				subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = (uint32)(plmId64)
				subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Set(restEventTriggerItem.ENBID)
			} else {
				xapp.Logger.Info("Missing manadatory element. ReportParams.EventTriggers[%v].ENBID=%v", index, restEventTriggerItem.ENBID)
				// Using some value as not mandatory value for xapp now
				subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
				subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDHomeBits28 // Bit length should be set based calculation. Only 28bit length works in ASN1C currently
				subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = 123456
				subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Set("12345")
			}

			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceDirection = (uint32)(restEventTriggerItem.InterfaceDirection)
			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.ProcedureCode = (uint32)(restEventTriggerItem.ProcedureCode)
			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.TypeOfMessage = (uint64)(restEventTriggerItem.TypeOfMessage)

			actionToBeSetupItem := e2ap.ActionToBeSetupItem{}
			actionToBeSetupItem.ActionId = 0 // REST definition does not yet have this
			actionToBeSetupItem.ActionType = e2ap.E2AP_ActionTypeReport
			actionToBeSetupItem.RicActionDefinitionPresent = true
			actionToBeSetupItem.ActionDefinitionChoice.ActionDefinitionX2Format1Present = true // This is mandator, but can be empty
			if p.ReportActionDefinitions != nil && p.ReportActionDefinitions.ActionDefinitionFormat1 != nil && p.ReportActionDefinitions.ActionDefinitionFormat1.StyleID != nil {
				actionToBeSetupItem.ActionDefinitionChoice.ActionDefinitionX2Format1.StyleID = (uint64)(*p.ReportActionDefinitions.ActionDefinitionFormat1.StyleID)
			}

			// 0.. OPTIONAL
			if lengthActionParameters > 0 {
				for _, restActionParameterItem := range p.ReportActionDefinitions.ActionDefinitionFormat1.ActionParameters {
					actionParameterItem := e2ap.ActionParameterItem{}
					if restActionParameterItem.ActionParameterID != nil {
						actionParameterItem.ParameterID = (uint32)(*restActionParameterItem.ActionParameterID)
					}
					actionParameterItem.ActionParameterValue.ValueBoolPresent = true
					if restActionParameterItem.ActionParameterValue != nil {
						actionParameterItem.ActionParameterValue.ValueBool = (bool)(*restActionParameterItem.ActionParameterValue)
					}
					actionToBeSetupItem.ActionDefinitionChoice.ActionDefinitionX2Format1.ActionParameterItems =
						append(actionToBeSetupItem.ActionDefinitionChoice.ActionDefinitionX2Format1.ActionParameterItems, actionParameterItem)

					// OPTIONAL
					actionToBeSetupItem.SubsequentActionPresent = false // SubseguentAction not pressent in REST definition
				}
			}
			subReqMsg.ActionSetups = append(subReqMsg.ActionSetups, actionToBeSetupItem)
			subreqList.E2APSubscriptionRequests = append(subreqList.E2APSubscriptionRequests, subReqMsg)
		} else {
			// E2SM-gNB-NRT
			subReqMsg.EventTriggerDefinition.NBNRTEventTriggerDefinitionPresent = true
			if restEventTriggerItem.TriggerNature == "now" {
				subReqMsg.EventTriggerDefinition.NBNRTEventTriggerDefinition.TriggerNature = e2ap.NRTTriggerNature_now
			} else if restEventTriggerItem.TriggerNature == "on change" {
				subReqMsg.EventTriggerDefinition.NBNRTEventTriggerDefinition.TriggerNature = e2ap.NRTTriggerNature_onchange
			}

			actionToBeSetupItem := e2ap.ActionToBeSetupItem{}
			actionToBeSetupItem.ActionId = 0 // REST definition does not yet have this
			actionToBeSetupItem.ActionType = e2ap.E2AP_ActionTypeReport
			actionToBeSetupItem.RicActionDefinitionPresent = true
			actionToBeSetupItem.ActionDefinitionChoice.ActionDefinitionNRTFormat1Present = true // This is mandator, but can be empty
			subReqMsg.ActionSetups = append(subReqMsg.ActionSetups, actionToBeSetupItem)
			subreqList.E2APSubscriptionRequests = append(subreqList.E2APSubscriptionRequests, subReqMsg)
		}
	}
	return nil
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *E2ap) FillPolicySubReqMsgs(stype models.SubscriptionType, params interface{}, subreqList *e2ap.SubscriptionRequestList, restSubscription *RESTSubscription) error {
	xapp.Logger.Info("FillPolicySubReqMsgs")

	p := params.(*models.PolicyParams)
	lengthEventTriggers := len(p.EventTriggers)
	var lengthRANUeGroupParameters int = 0
	if p.PolicyActionDefinitions != nil && p.PolicyActionDefinitions.ActionDefinitionFormat2 != nil && p.PolicyActionDefinitions.ActionDefinitionFormat2.RANUeGroupParameters != nil {
		lengthRANUeGroupParameters = len(p.PolicyActionDefinitions.ActionDefinitionFormat2.RANUeGroupParameters)
	}
	if lengthEventTriggers == 0 {
		err := fmt.Errorf("Error in content element count. Count of EventTriggers=%v", lengthEventTriggers)
		return err
	}

	xapp.Logger.Info("EventTrigger count=%v, RANUeGroupParameter count=%v", lengthEventTriggers, lengthRANUeGroupParameters)

	// 1..
	for index, restEventTriggerItem := range p.EventTriggers {
		subReqMsg := e2ap.E2APSubscriptionRequest{}
		if p.RANFunctionID != nil {
			subReqMsg.FunctionId = (e2ap.FunctionId)(*p.RANFunctionID)
		}
		subReqMsg.EventTriggerDefinition.NBX2EventTriggerDefinitionPresent = true

		// Interface-ID is either ENBID or GNBID. GNBID is not present in REST definition currently
		if len(restEventTriggerItem.ENBID) > 0 {
			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
			plmId64, _ := strconv.Atoi(restEventTriggerItem.PlmnID)
			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDHomeBits28 // Bit length should be set based calculation. Only 28bit length works in ASN1C currently
			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = (uint32)(plmId64)
			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Set(restEventTriggerItem.ENBID)
		} else {
			xapp.Logger.Info("Missing manadatory element. PolicyParams.EventTriggers[%v].ENBID=%v", index, restEventTriggerItem.ENBID)
			// Using some value as not mandatory value for xapp now
			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.Present = true
			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Bits = e2ap.E2AP_ENBIDHomeBits28 // Bit length should be set based calculation. Only 28bit length works in ASN1C currently
			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.NodeId.Id = 123456
			subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceId.GlobalEnbId.PlmnIdentity.Set("12345")
		}
		subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.InterfaceDirection = (uint32)(restEventTriggerItem.InterfaceDirection)
		subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.ProcedureCode = (uint32)(restEventTriggerItem.ProcedureCode)
		subReqMsg.EventTriggerDefinition.X2EventTriggerDefinition.TypeOfMessage = (uint64)(restEventTriggerItem.TypeOfMessage)

		actionToBeSetupItem := e2ap.ActionToBeSetupItem{}
		actionToBeSetupItem.ActionId = 0 // REST definition does not yet have this
		actionToBeSetupItem.ActionType = e2ap.E2AP_ActionTypePolicy
		actionToBeSetupItem.RicActionDefinitionPresent = true

		actionToBeSetupItem.ActionDefinitionChoice.ActionDefinitionX2Format2Present = true

		// 0.. OPTIONAL
		if lengthRANUeGroupParameters > 0 {
			for _, restRANUeGroupParametersItem := range p.PolicyActionDefinitions.ActionDefinitionFormat2.RANUeGroupParameters {
				ranUEgroupItem := e2ap.RANueGroupItem{}
				ranUEgroupItem.RanUEgroupID = *restRANUeGroupParametersItem.RANUeGroupID

				// This is 1..255 in asn.1 spec but only one element in REST definition
				ranUEGroupDefItem := e2ap.RANueGroupDefItem{}
				if restRANUeGroupParametersItem.RANUeGroupDefinition.RANParameterID != nil {
					ranUEGroupDefItem.RanParameterID = (uint32)(*restRANUeGroupParametersItem.RANUeGroupDefinition.RANParameterID)
				}
				if restRANUeGroupParametersItem.RANUeGroupDefinition.RANParameterTestCondition == "equal" {
					ranUEGroupDefItem.RanParameterTest = e2ap.RANParameterTest_equal
				} else if restRANUeGroupParametersItem.RANUeGroupDefinition.RANParameterTestCondition == "greaterthan" {
					ranUEGroupDefItem.RanParameterTest = e2ap.RANParameterTest_greaterthan
				} else if restRANUeGroupParametersItem.RANUeGroupDefinition.RANParameterTestCondition == "lessthan" {
					ranUEGroupDefItem.RanParameterTest = e2ap.RANParameterTest_lessthan
				} else {
					return fmt.Errorf("Incorrect RANParameterTestCondition %s", restRANUeGroupParametersItem.RANUeGroupDefinition.RANParameterTestCondition)
				}
				ranUEGroupDefItem.RanParameterValue.ValueIntPresent = true
				if restRANUeGroupParametersItem.RANUeGroupDefinition.RANParameterValue != nil {
					ranUEGroupDefItem.RanParameterValue.ValueInt = *restRANUeGroupParametersItem.RANUeGroupDefinition.RANParameterValue
				}
				ranUEgroupItem.RanUEgroupDefinition.RanUEGroupDefItems = append(ranUEgroupItem.RanUEgroupDefinition.RanUEGroupDefItems, ranUEGroupDefItem)

				// This is 1..255 in asn.1 spec but only one element in REST definition
				ranParameterItem := e2ap.RANParameterItem{}
				if restRANUeGroupParametersItem.RANImperativePolicy.PolicyParameterID != nil {
					ranParameterItem.RanParameterID = (uint8)(*restRANUeGroupParametersItem.RANImperativePolicy.PolicyParameterID)
				}
				ranParameterItem.RanParameterValue.ValueIntPresent = true
				if restRANUeGroupParametersItem.RANImperativePolicy.PolicyParameterValue != nil {
					ranParameterItem.RanParameterValue.ValueInt = *restRANUeGroupParametersItem.RANImperativePolicy.PolicyParameterValue
				}
				ranUEgroupItem.RanPolicy.RanParameterItems = append(ranUEgroupItem.RanPolicy.RanParameterItems, ranParameterItem)
				actionToBeSetupItem.ActionDefinitionChoice.ActionDefinitionX2Format2.RanUEgroupItems =
					append(actionToBeSetupItem.ActionDefinitionChoice.ActionDefinitionX2Format2.RanUEgroupItems, ranUEgroupItem)

					// OPTIONAL
				actionToBeSetupItem.SubsequentActionPresent = false // SubseguentAction not pressent in REST definition
			}
		}
		subReqMsg.ActionSetups = append(subReqMsg.ActionSetups, actionToBeSetupItem)
		subreqList.E2APSubscriptionRequests = append(subreqList.E2APSubscriptionRequests, subReqMsg)
	}
	return nil
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------
func (c *E2ap) UnpackSubscriptionRequest(payload []byte, debugPrint bool) (*e2ap.E2APSubscriptionRequest, error) {
	e2SubReq := packerif.NewPackerSubscriptionRequest()
	err, subReq, msgString := e2SubReq.UnPack(&e2ap.PackedData{payload}, debugPrint)
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}

	if msgString != "" {
		xapp.Logger.Debug("%s", msgString)
	}
	return subReq, nil
}

func (c *E2ap) PackSubscriptionRequest(req *e2ap.E2APSubscriptionRequest, debugPrint bool) (int, *e2ap.PackedData, error) {
	e2SubReq := packerif.NewPackerSubscriptionRequest()
	err, packedData, msgString := e2SubReq.Pack(req, debugPrint)

	if msgString != "" {
		xapp.Logger.Debug("%s", msgString)
	}

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
	err, subResp := e2SubResp.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subResp, nil
}

func (c *E2ap) PackSubscriptionResponse(req *e2ap.E2APSubscriptionResponse) (int, *e2ap.PackedData, error) {
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
	err, subFail := e2SubFail.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subFail, nil
}

func (c *E2ap) PackSubscriptionFailure(req *e2ap.E2APSubscriptionFailure) (int, *e2ap.PackedData, error) {
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
	err, subDelReq := e2SubDelReq.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subDelReq, nil
}

func (c *E2ap) PackSubscriptionDeleteRequest(req *e2ap.E2APSubscriptionDeleteRequest) (int, *e2ap.PackedData, error) {
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
	err, subDelResp := e2SubDelResp.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subDelResp, nil
}

func (c *E2ap) PackSubscriptionDeleteResponse(req *e2ap.E2APSubscriptionDeleteResponse) (int, *e2ap.PackedData, error) {
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
	err, subDelFail := e2SubDelFail.UnPack(&e2ap.PackedData{payload})
	if err != nil {
		return nil, fmt.Errorf("%s buf[%s]", err.Error(), hex.EncodeToString(payload))
	}
	return subDelFail, nil
}

func (c *E2ap) PackSubscriptionDeleteFailure(req *e2ap.E2APSubscriptionDeleteFailure) (int, *e2ap.PackedData, error) {
	e2SubDelFail := packerif.NewPackerSubscriptionDeleteFailure()
	err, packedData := e2SubDelFail.Pack(req)
	if err != nil {
		return 0, nil, err
	}
	return xapp.RIC_SUB_DEL_FAILURE, packedData, nil
}
