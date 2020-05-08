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

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	rtmgrclient "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client"
	"gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/xapptweaks"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/segmentio/ksuid"
	"github.com/spf13/viper"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

func idstring(err error, entries ...fmt.Stringer) string {
	var retval string = ""
	var filler string = ""
	for _, entry := range entries {
		retval += filler + entry.String()
		filler = " "
	}
	if err != nil {
		retval += filler + "err(" + err.Error() + ")"
		filler = " "

	}
	return retval
}

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

var e2tSubReqTimeout time.Duration = 5 * time.Second
var e2tSubDelReqTime time.Duration = 5 * time.Second
var e2tMaxSubReqTryCount uint64 = 2    // Initial try + retry
var e2tMaxSubDelReqTryCount uint64 = 2 // Initial try + retry

var e2tRecvMsgTimeout time.Duration = 5 * time.Second

type Control struct {
	xapptweaks.XappWrapper
	e2ap     *E2ap
	registry *Registry
	tracker  *Tracker
}

type RMRMeid struct {
	PlmnID  string
	EnbID   string
	RanName string
}

func init() {
	xapp.Logger.Info("SUBMGR")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("submgr")
	viper.AllowEmptyEnv(true)
}

func NewControl() *Control {

	transport := httptransport.New(viper.GetString("rtmgr.HostAddr")+":"+viper.GetString("rtmgr.port"), viper.GetString("rtmgr.baseUrl"), []string{"http"})
	rtmgrClient := RtmgrClient{rtClient: rtmgrclient.New(transport, strfmt.Default)}

	registry := new(Registry)
	registry.Initialize()
	registry.rtmgrClient = &rtmgrClient

	tracker := new(Tracker)
	tracker.Init()

	c := &Control{e2ap: new(E2ap),
		registry: registry,
		tracker:  tracker,
	}
	c.XappWrapper.Init("")
	go xapp.Subscription.Listen(c.RESTSubscriptionRequestHandler, c.QueryHandler, c.RESTSubscriptionDeleteHandler)
	return c
}

func (c *Control) ReadyCB(data interface{}) {
	if c.Rmr == nil {
		c.Rmr = xapp.Rmr
	}
}

func (c *Control) Run() {
	xapp.SetReadyCB(c.ReadyCB, nil)
	xapp.Run(c)
}

// This function should be moved to packer_e2ap.go file
//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) FillReportSubReqMsgs(stype models.SubscriptionType, params interface{}, subreqList *e2ap.SubscriptionRequestList, restSubscription *RESTSubscription) error {
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
		actionToBeSetupItem.ActionDefinitionChoice.ActionDefinitionX2Format1Present = true // Only this choice present in REST definition currently
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
	}
	return nil
}

// This function should be moved to packer_e2ap.go file
//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) FillPolicySubReqMsgs(stype models.SubscriptionType, params interface{}, subreqList *e2ap.SubscriptionRequestList, restSubscription *RESTSubscription) error {
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

// This functionality should be moved to somewhere else. It would be better if xapp would send address without port
// i.e., in this form: service-ricxapp-xappname-http.ricxapp
//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) ConstructEndpointAddresses(clientEndpoint string) (string, string, error) {

	// Received clientEndpoint addres could be either: service-ricxapp-xappname-http.ricxapp:8080 or
	// service-ricxapp-xappname-rmr.ricxapp:4560
	if i := strings.Index(clientEndpoint, ":"); i == -1 {
		err := fmt.Errorf("Incorrect ClientEndpoint address format=%s. It should be address:port", clientEndpoint)
		return "", "", err
	}

	// xApp's http address need to be in this form: service-ricxapp-xappname-http.ricxapp
	xAppHttpEndPoint := clientEndpoint
	if i := strings.Index(xAppHttpEndPoint, ":"); i != -1 {
		// Remove port form the address
		xAppHttpEndPoint = xAppHttpEndPoint[0:i]
	}

	// Submgr's test address need to be in this form: localhost:13560
	if i := strings.Index(clientEndpoint, "localhost"); i != -1 {
		// Test address is used. clientEndpoint contains already the RMR address we need
		return xAppHttpEndPoint, clientEndpoint, nil
	}

	// xApp's RMR address should be in this form: service-ricxapp-xappname-rmr.ricxapp:4560
	var xAppRrmEndPoint string
	if i := strings.Index(clientEndpoint, "http"); i != -1 {
		// Fix http -> rmr
		xAppRrmEndPoint = strings.Replace(clientEndpoint, "http", "rmr", -1)
	}

	if i := strings.Index(xAppRrmEndPoint, "8080"); i != -1 {
		// Fix RMR port 8080 -> 4560
		xAppRrmEndPoint = strings.Replace(xAppRrmEndPoint, "8080", "4560", -1)
	}

	xapp.Logger.Info("xAppHttpEndPoint=%v, xAppRrmEndPoint=%v", xAppHttpEndPoint, xAppRrmEndPoint)

	return xAppHttpEndPoint, xAppRrmEndPoint, nil
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) RESTSubscriptionRequestHandler(stype models.SubscriptionType, params interface{}) (*models.SubscriptionResponse, error) {

	xapp.Logger.Info("SubscriptionRequest from XAPP")

	restSubId := ksuid.New().String()
	subResp := models.SubscriptionResponse{}
	subResp.SubscriptionID = &restSubId
	switch stype {
	case models.SubscriptionTypeReport:
		p := params.(*models.ReportParams)

		if p.ClientEndpoint == nil {
			xapp.Logger.Error("ClientEndpoint == nil")
			return nil, fmt.Errorf("")
		}

		xAppHttpEndPoint, xAppRmrEndpoint, err := c.ConstructEndpointAddresses(*p.ClientEndpoint)
		if err != nil {
			xapp.Logger.Error("%s", err.Error())
			return nil, err
		}

		restSubscription, err := c.registry.CreateRESTSubscription(&restSubId, &xAppRmrEndpoint, &p.Meid)
		if err != nil {
			xapp.Logger.Error("%s", err.Error())
			return nil, err
		}

		subReqList := e2ap.SubscriptionRequestList{}
		err = c.FillReportSubReqMsgs(stype, params, &subReqList, restSubscription)
		if err != nil {
			xapp.Logger.Error("%s", err.Error())
			c.registry.DeleteRESTSubscription(&restSubId)
			return nil, err
		}

		trans := c.tracker.NewXappTransaction(xapptweaks.NewRmrEndpoint(xAppRmrEndpoint), restSubId, 0, &xapp.RMRMeid{RanName: p.Meid})
		if trans == nil {
			c.registry.DeleteRESTSubscription(&restSubId)
			xapp.Logger.Error("XAPP-SubReq transaction not created. RESTSubId=%s, EndPoint=%s, Meid=%s", restSubId, xAppRmrEndpoint, p.Meid)
			return nil, fmt.Errorf("")
		}

		go c.processSubscriptionRequests(trans, restSubscription, &subReqList, &xAppHttpEndPoint, &xAppRmrEndpoint, &p.Meid, &restSubId, e2ap.E2AP_ActionTypeReport)

		// Respond to xapp
		return &subResp, nil
	case models.SubscriptionTypePolicy:
		p := params.(*models.PolicyParams)

		if p.ClientEndpoint == nil {
			xapp.Logger.Error("ClientEndpoint == nil")
			return nil, fmt.Errorf("")
		}

		xAppHttpEndPoint, xAppRmrEndpoint, err := c.ConstructEndpointAddresses(*p.ClientEndpoint)
		if err != nil {
			xapp.Logger.Error("%s", err.Error())
			return nil, err
		}

		restSubscription, err := c.registry.CreateRESTSubscription(&restSubId, &xAppRmrEndpoint, p.Meid)
		if err != nil {
			return nil, err
		}
		subReqList := e2ap.SubscriptionRequestList{}
		err = c.FillPolicySubReqMsgs(stype, params, &subReqList, restSubscription)
		if err != nil {
			xapp.Logger.Error("%s", err.Error())
			c.registry.DeleteRESTSubscription(&restSubId)
			return nil, err
		}

		trans := c.tracker.NewXappTransaction(xapptweaks.NewRmrEndpoint(xAppRmrEndpoint), restSubId, 0, &xapp.RMRMeid{RanName: *p.Meid})
		if trans == nil {
			c.registry.DeleteRESTSubscription(&restSubId)
			xapp.Logger.Error("XAPP-SubReq transaction not created. RESTSubId=%s, EndPoint=%s, Meid=%s", restSubId, xAppRmrEndpoint, *p.Meid)
			return nil, fmt.Errorf("")
		}

		go c.processSubscriptionRequests(trans, restSubscription, &subReqList, &xAppHttpEndPoint, &xAppRmrEndpoint, p.Meid, &restSubId, e2ap.E2AP_ActionTypePolicy)

		// Respond to xapp
		return &subResp, nil
	}
	// Respond to xapp
	return &subResp, fmt.Errorf("Not supported subscription type=%v", stype)
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------

func (c *Control) processSubscriptionRequests(trans *TransactionXapp, restSubscription *RESTSubscription, subReqList *e2ap.SubscriptionRequestList,
	xAppHttpEndPoint *string, xAppRmrpEndPoint *string, meid *string, restSubId *string, actionType uint64) {

	xapp.Logger.Info("Subscription Request count=%v, %s", len(subReqList.E2APSubscriptionRequests), idstring(nil, trans))
	defer trans.Release()

	var requestorID int64
	var instanceId int64
	for index, subReqMsg := range subReqList.E2APSubscriptionRequests {
		xapp.Logger.Info("Handle SubscriptionRequest index=%v, %s", index, idstring(nil, trans))

		subRespMsg, err := c.handleSubscriptionRequest(trans, &subReqMsg, xAppRmrpEndPoint, meid, restSubId, actionType)
		if err != nil {
			// Send notification to xApp that prosessing of a Subscription Request has failed. Currently it is not possible
			// to indicate error. Such possibility should be added. As a workaround requestorID and instanceId are set to zero value
			requestorID = (int64)(0)
			instanceId = (int64)(0)
			resp := &models.SubscriptionResponse{
				SubscriptionID: restSubId,
				SubscriptionInstances: []*models.SubscriptionInstance{
					&models.SubscriptionInstance{RequestorID: &requestorID, InstanceID: &instanceId},
				},
			}
			// Mark REST subscription request processesd.
			restSubscription.SetProcessed()
			xapp.Logger.Info("Sending unsuccessful REST notification to endpoint=%v, InstanceId=%v, %s", *xAppHttpEndPoint, instanceId, idstring(nil, trans))
			xapp.Subscription.Notify(resp, *xAppHttpEndPoint)
		} else {
			xapp.Logger.Info("SubscriptionRequest index=%v processed successfully. endpoint=%v, InstanceId=%v, %s", index, *xAppHttpEndPoint, instanceId, idstring(nil, trans))

			// Store successfully processed InstanceId for deletion
			restSubscription.AddInstanceId(subRespMsg.RequestId.InstanceId)

			// Send notification to xApp that a Subscription Request has been processed.
			requestorID = (int64)(subRespMsg.RequestId.Id)
			instanceId = (int64)(subRespMsg.RequestId.InstanceId)
			resp := &models.SubscriptionResponse{
				SubscriptionID: restSubId,
				SubscriptionInstances: []*models.SubscriptionInstance{
					&models.SubscriptionInstance{RequestorID: &requestorID, InstanceID: &instanceId},
				},
			}
			// Mark REST subscription request processesd.
			restSubscription.SetProcessed()
			xapp.Logger.Info("Sending successful REST notification to endpoint=%v, InstanceId=%v, %s", *xAppHttpEndPoint, instanceId, idstring(nil, trans))
			xapp.Subscription.Notify(resp, *xAppHttpEndPoint)
		}
	}
}

//-------------------------------------------------------------------
//
//------------------------------------------------------------------
func (c *Control) handleSubscriptionRequest(trans *TransactionXapp, subReqMsg *e2ap.E2APSubscriptionRequest, clientEndpoint *string, meid *string,
	restSubId *string, actionType uint64) (*e2ap.E2APSubscriptionResponse, error) {

	err := c.tracker.Track(trans)
	if err != nil {
		err = fmt.Errorf("XAPP-SubReq: %s", idstring(err, trans))
		xapp.Logger.Error("%s", err.Error())
		return nil, err
	}

	subs, err := c.registry.AssignToSubscription(trans, subReqMsg, actionType)
	if err != nil {
		err = fmt.Errorf("XAPP-SubReq: %s", idstring(err, trans))
		xapp.Logger.Error("%s", err.Error())
		return nil, err
	}

	//
	// Wake subs request
	//
	go c.handleSubscriptionCreate(subs, trans)
	event, _ := trans.WaitEvent(0) //blocked wait as timeout is handled in subs side

	err = nil
	if event != nil {
		switch themsg := event.(type) {
		case *e2ap.E2APSubscriptionResponse:
			trans.Release()
			return themsg, nil
		case *e2ap.E2APSubscriptionFailure:
			err = fmt.Errorf("SubscriptionFailure received")
			return nil, err
		default:
			break
		}
	}
	err = fmt.Errorf("XAPP-SubReq: failed %s", idstring(err, trans, subs))
	xapp.Logger.Error("%s", err.Error())
	c.registry.RemoveFromSubscription(subs, trans, 5*time.Second)
	return nil, err
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) RESTSubscriptionDeleteHandler(restSubId string) error {
	xapp.Logger.Info("SubscriptionDeleteRequest from XAPP")

	restSubscription, err := c.registry.GetRESTSubscription(restSubId)
	if err != nil {
		xapp.Logger.Error("%s", err.Error())
		if restSubscription == nil {
			// Subscription was not found
			return nil
		} else {
			if restSubscription.SubReqOngoing == true {
				err := fmt.Errorf("Handling of the REST Subscription Request still ongoing %s:", restSubId)
				xapp.Logger.Error("%s", err.Error())
				return err
			} else if restSubscription.SubDelReqOngoing == true {
				// Previous request for same restSubId still ongoing
				return nil
			}
		}
	}

	xAppRmrEndPoint := restSubscription.xAppRmrEndPoint
	go func() {
		for _, instanceId := range restSubscription.InstanceIds {
			err := c.SubscriptionDeleteHandler(&restSubId, &xAppRmrEndPoint, &restSubscription.Meid, instanceId)
			if err != nil {
				xapp.Logger.Error("%s", err.Error())
				//return err
			}
			xapp.Logger.Info("Deleteting instanceId = %v", instanceId)
			restSubscription.DeleteInstanceId(instanceId)
		}
		c.registry.DeleteRESTSubscription(&restSubId)
	}()
	return nil
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) SubscriptionDeleteHandler(restSubId *string, endPoint *string, meid *string, instanceId uint32) error {

	trans := c.tracker.NewXappTransaction(xapptweaks.NewRmrEndpoint(*endPoint), *restSubId, 0, &xapp.RMRMeid{RanName: *meid})
	if trans == nil {
		err := fmt.Errorf("XAPP-SubDelReq transaction not created. restSubId %s, endPoint %s, meid %s, instanceId %v", *restSubId, *endPoint, *meid, instanceId)
		xapp.Logger.Error("%s", err.Error())
	}
	defer trans.Release()

	err := c.tracker.Track(trans)
	if err != nil {
		err := fmt.Errorf("XAPP-SubDelReq %s:", idstring(err, trans))
		xapp.Logger.Error("%s", err.Error())
		return &time.ParseError{}
	}

	subs, err := c.registry.GetSubscriptionFirstMatch([]uint32{instanceId})
	if err != nil {
		err := fmt.Errorf("XAPP-SubDelReq %s:", idstring(err, trans))
		xapp.Logger.Error("%s", err.Error())
		return err
	}
	//
	// Wake subs delete
	//
	go c.handleSubscriptionDelete(subs, trans)
	trans.WaitEvent(0) //blocked wait as timeout is handled in subs side

	xapp.Logger.Debug("XAPP-SubDelReq: Handling event %s ", idstring(nil, trans, subs))

	// Whatever is received send ok delete response
	if err == nil {
		return nil
	}

	c.registry.RemoveFromSubscription(subs, trans, 5*time.Second)

	return nil
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) QueryHandler() (models.SubscriptionList, error) {
	return c.registry.QueryHandler()
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) rmrSendToE2T(desc string, subs *Subscription, trans *TransactionSubs) (err error) {
	params := xapptweaks.NewParams(nil)
	params.Mtype = trans.GetMtype()
	params.SubId = int(subs.GetReqId().InstanceId)
	params.Xid = ""
	params.Meid = subs.GetMeid()
	params.Src = ""
	params.PayloadLen = len(trans.Payload.Buf)
	params.Payload = trans.Payload.Buf
	params.Mbuf = nil
	xapp.Logger.Info("MSG to E2T: %s %s %s", desc, trans.String(), params.String())
	return c.RmrSend(params, 5)
}

func (c *Control) rmrSendToXapp(desc string, subs *Subscription, trans *TransactionXapp) (err error) {
	params := xapptweaks.NewParams(nil)
	params.Mtype = trans.GetMtype()
	params.SubId = int(subs.GetReqId().InstanceId)
	params.Xid = trans.GetXid()
	params.Meid = trans.GetMeid()
	params.Src = ""
	params.PayloadLen = len(trans.Payload.Buf)
	params.Payload = trans.Payload.Buf
	params.Mbuf = nil
	xapp.Logger.Info("MSG to XAPP: %s %s %s", desc, trans.String(), params.String())
	return c.RmrSend(params, 5)
}

func (c *Control) Consume(params *xapp.RMRParams) (err error) {
	msg := xapptweaks.NewParams(params)
	if c.Rmr == nil {
		err = fmt.Errorf("Rmr object nil can handle %s", msg.String())
		xapp.Logger.Error("%s", err.Error())
		return
	}
	c.CntRecvMsg++

	defer c.Rmr.Free(msg.Mbuf)

	switch msg.Mtype {
	case xapp.RIC_SUB_REQ:
		go c.handleXAPPSubscriptionRequest(msg)
	case xapp.RIC_SUB_RESP:
		go c.handleE2TSubscriptionResponse(msg)
	case xapp.RIC_SUB_FAILURE:
		go c.handleE2TSubscriptionFailure(msg)
	case xapp.RIC_SUB_DEL_REQ:
		go c.handleXAPPSubscriptionDeleteRequest(msg)
	case xapp.RIC_SUB_DEL_RESP:
		go c.handleE2TSubscriptionDeleteResponse(msg)
	case xapp.RIC_SUB_DEL_FAILURE:
		go c.handleE2TSubscriptionDeleteFailure(msg)
	default:
		xapp.Logger.Info("Unknown Message Type '%d', discarding", msg.Mtype)
	}
	return
}

//-------------------------------------------------------------------
// handle from XAPP Subscription Request
//------------------------------------------------------------------
func (c *Control) handleXAPPSubscriptionRequest(params *xapptweaks.RMRParams) {
	xapp.Logger.Info("MSG from XAPP: %s", params.String())

	subReqMsg, err := c.e2ap.UnpackSubscriptionRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(err, params))
		return
	}

	trans := c.tracker.NewXappTransaction(xapptweaks.NewRmrEndpoint(params.Src), params.Xid, subReqMsg.RequestId.InstanceId, params.Meid)
	if trans == nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(fmt.Errorf("transaction not created"), params))
		return
	}
	defer trans.Release()

	err = c.tracker.Track(trans)
	if err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(err, trans))
		return
	}

	//TODO handle subscription toward e2term inside AssignToSubscription / hide handleSubscriptionCreate in it?
	subs, err := c.registry.AssignToSubscription(trans, subReqMsg, e2ap.E2AP_ActionTypeInvalid)
	if err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(err, trans))
		return
	}

	//
	// Wake subs request
	//
	go c.handleSubscriptionCreate(subs, trans)
	event, _ := trans.WaitEvent(0) //blocked wait as timeout is handled in subs side

	err = nil
	if event != nil {
		switch themsg := event.(type) {
		case *e2ap.E2APSubscriptionResponse:
			trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionResponse(themsg)
			if err == nil {
				trans.Release()
				c.rmrSendToXapp("", subs, trans)
				return
			}
		case *e2ap.E2APSubscriptionFailure:
			trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionFailure(themsg)
			if err == nil {
				c.rmrSendToXapp("", subs, trans)
			}
		default:
			break
		}
	}
	xapp.Logger.Info("XAPP-SubReq: failed %s", idstring(err, trans, subs))
	//c.registry.RemoveFromSubscription(subs, trans, 5*time.Second)
}

//-------------------------------------------------------------------
// handle from XAPP Subscription Delete Request
//------------------------------------------------------------------
func (c *Control) handleXAPPSubscriptionDeleteRequest(params *xapptweaks.RMRParams) {
	xapp.Logger.Info("MSG from XAPP: %s", params.String())

	subDelReqMsg, err := c.e2ap.UnpackSubscriptionDeleteRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("XAPP-SubDelReq %s", idstring(err, params))
		return
	}

	trans := c.tracker.NewXappTransaction(xapptweaks.NewRmrEndpoint(params.Src), params.Xid, subDelReqMsg.RequestId.InstanceId, params.Meid)
	if trans == nil {
		xapp.Logger.Error("XAPP-SubDelReq: %s", idstring(fmt.Errorf("transaction not created"), params))
		return
	}
	defer trans.Release()

	err = c.tracker.Track(trans)
	if err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(err, trans))
		return
	}

	subs, err := c.registry.GetSubscriptionFirstMatch([]uint32{trans.GetSubId()})
	if err != nil {
		xapp.Logger.Error("XAPP-SubDelReq: %s", idstring(err, trans))
		return
	}

	//
	// Wake subs delete
	//
	go c.handleSubscriptionDelete(subs, trans)
	trans.WaitEvent(0) //blocked wait as timeout is handled in subs side

	xapp.Logger.Debug("XAPP-SubDelReq: Handling event %s ", idstring(nil, trans, subs))

	// Whatever is received send ok delete response
	subDelRespMsg := &e2ap.E2APSubscriptionDeleteResponse{}
	subDelRespMsg.RequestId = subs.GetReqId().RequestId
	subDelRespMsg.FunctionId = subs.SubReqMsg.FunctionId
	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionDeleteResponse(subDelRespMsg)
	if err == nil {
		c.rmrSendToXapp("", subs, trans)
	}

	//TODO handle subscription toward e2term insiged RemoveFromSubscription / hide handleSubscriptionDelete in it?
	//c.registry.RemoveFromSubscription(subs, trans, 5*time.Second)
}

//-------------------------------------------------------------------
// SUBS CREATE Handling
//-------------------------------------------------------------------
func (c *Control) handleSubscriptionCreate(subs *Subscription, parentTrans *TransactionXapp) {

	trans := c.tracker.NewSubsTransaction(subs)
	subs.WaitTransactionTurn(trans)
	defer subs.ReleaseTransactionTurn(trans)
	defer trans.Release()

	xapp.Logger.Debug("SUBS-SubReq: Handling %s ", idstring(nil, trans, subs, parentTrans))

	subRfMsg, valid := subs.GetCachedResponse()
	if subRfMsg == nil && valid == true {

		//
		// In case of failure
		// - make internal delete
		// - in case duplicate cause, retry (currently max 1 retry)
		//
		maxRetries := uint64(1)
		doRetry := true
		for retries := uint64(0); retries <= maxRetries && doRetry; retries++ {
			doRetry = false

			event, err := c.sendE2TSubscriptionRequest(subs, trans, parentTrans)
			switch themsg := event.(type) {
			case *e2ap.E2APSubscriptionResponse:
				subRfMsg, valid = subs.SetCachedResponse(event, true)
			case *e2ap.E2APSubscriptionFailure:
				subRfMsg, valid = subs.SetCachedResponse(event, false)
				doRetry = true
				for _, item := range themsg.ActionNotAdmittedList.Items {
					if item.Cause.Content != e2ap.E2AP_CauseContent_Ric || (item.Cause.Value != e2ap.E2AP_CauseValue_Ric_duplicate_action && item.Cause.Value != e2ap.E2AP_CauseValue_Ric_duplicate_event) {
						doRetry = false
						break
					}
				}
				xapp.Logger.Info("SUBS-SubReq: internal delete and possible retry due event(%s) retry(%t,%d/%d) %s", typeofSubsMessage(event), doRetry, retries, maxRetries, idstring(nil, trans, subs, parentTrans))
				c.sendE2TSubscriptionDeleteRequest(subs, trans, parentTrans)
			default:
				xapp.Logger.Info("SUBS-SubReq: internal delete due event(%s) %s", typeofSubsMessage(event), idstring(nil, trans, subs, parentTrans))
				subRfMsg, valid = subs.SetCachedResponse(nil, false)
				if err == nil {
					c.sendE2TSubscriptionDeleteRequest(subs, trans, parentTrans)
				}
			}
		}

		xapp.Logger.Debug("SUBS-SubReq: Handling (e2t response %s) %s", typeofSubsMessage(subRfMsg), idstring(nil, trans, subs, parentTrans))
	} else {
		xapp.Logger.Debug("SUBS-SubReq: Handling (cached response %s) %s", typeofSubsMessage(subRfMsg), idstring(nil, trans, subs, parentTrans))
	}

	//Now RemoveFromSubscription in here to avoid race conditions (mostly concerns delete)
	if valid == false {
		c.registry.RemoveFromSubscription(subs, parentTrans, 5*time.Second)
	}
	parentTrans.SendEvent(subRfMsg, 0)
}

//-------------------------------------------------------------------
// SUBS DELETE Handling
//-------------------------------------------------------------------

func (c *Control) handleSubscriptionDelete(subs *Subscription, parentTrans *TransactionXapp) {

	trans := c.tracker.NewSubsTransaction(subs)
	subs.WaitTransactionTurn(trans)
	defer subs.ReleaseTransactionTurn(trans)
	defer trans.Release()

	xapp.Logger.Debug("SUBS-SubDelReq: Handling %s", idstring(nil, trans, subs, parentTrans))

	subs.mutex.Lock()
	if subs.valid && subs.EpList.HasEndpoint(parentTrans.GetEndpoint()) && subs.EpList.Size() == 1 {
		subs.valid = false
		subs.mutex.Unlock()
		c.sendE2TSubscriptionDeleteRequest(subs, trans, parentTrans)
	} else {
		subs.mutex.Unlock()
	}
	//Now RemoveFromSubscription in here to avoid race conditions (mostly concerns delete)
	//  If parallel deletes ongoing both might pass earlier sendE2TSubscriptionDeleteRequest(...) if
	//  RemoveFromSubscription locates in caller side (now in handleXAPPSubscriptionDeleteRequest(...))
	c.registry.RemoveFromSubscription(subs, parentTrans, 5*time.Second)
	parentTrans.SendEvent(nil, 0)
}

//-------------------------------------------------------------------
// send to E2T Subscription Request
//-------------------------------------------------------------------
func (c *Control) sendE2TSubscriptionRequest(subs *Subscription, trans *TransactionSubs, parentTrans *TransactionXapp) (interface{}, error) {
	var err error
	var event interface{} = nil
	var timedOut bool = false

	subReqMsg := subs.SubReqMsg
	subReqMsg.RequestId = subs.GetReqId().RequestId

	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionRequest(subReqMsg)
	if err != nil {
		xapp.Logger.Error("SUBS-SubReq: %s", idstring(err, trans, subs, parentTrans))
		return nil, err
	}

	for retries := uint64(0); retries < e2tMaxSubReqTryCount; retries++ {
		desc := fmt.Sprintf("(retry %d)", retries)
		c.rmrSendToE2T(desc, subs, trans)
		event, timedOut = trans.WaitEvent(e2tSubReqTimeout)
		if timedOut {
			continue
		}
		break
	}
	xapp.Logger.Debug("SUBS-SubReq: Response handling event(%s) %s", typeofSubsMessage(event), idstring(nil, trans, subs, parentTrans))
	return event, nil
}

//-------------------------------------------------------------------
// send to E2T Subscription Delete Request
//-------------------------------------------------------------------

func (c *Control) sendE2TSubscriptionDeleteRequest(subs *Subscription, trans *TransactionSubs, parentTrans *TransactionXapp) interface{} {
	var err error
	var event interface{}
	var timedOut bool

	subDelReqMsg := &e2ap.E2APSubscriptionDeleteRequest{}
	subDelReqMsg.RequestId = subs.GetReqId().RequestId
	subDelReqMsg.FunctionId = subs.SubReqMsg.FunctionId
	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionDeleteRequest(subDelReqMsg)
	if err != nil {
		xapp.Logger.Error("SUBS-SubDelReq: %s", idstring(err, trans, subs, parentTrans))
		return event
	}

	for retries := uint64(0); retries < e2tMaxSubDelReqTryCount; retries++ {
		desc := fmt.Sprintf("(retry %d)", retries)
		c.rmrSendToE2T(desc, subs, trans)
		event, timedOut = trans.WaitEvent(e2tSubDelReqTime)
		if timedOut {
			continue
		}
		break
	}
	xapp.Logger.Debug("SUBS-SubDelReq: Response handling event(%s) %s", typeofSubsMessage(event), idstring(nil, trans, subs, parentTrans))
	return event
}

//-------------------------------------------------------------------
// handle from E2T Subscription Reponse
//-------------------------------------------------------------------
func (c *Control) handleE2TSubscriptionResponse(params *xapptweaks.RMRParams) {
	xapp.Logger.Info("MSG from E2T: %s", params.String())
	subRespMsg, err := c.e2ap.UnpackSubscriptionResponse(params.Payload)
	if err != nil {
		xapp.Logger.Error("MSG-SubResp %s", idstring(err, params))
		return
	}
	subs, err := c.registry.GetSubscriptionFirstMatch([]uint32{subRespMsg.RequestId.InstanceId})
	if err != nil {
		xapp.Logger.Error("MSG-SubResp: %s", idstring(err, params))
		return
	}
	trans := subs.GetTransaction()
	if trans == nil {
		err = fmt.Errorf("Ongoing transaction not found")
		xapp.Logger.Error("MSG-SubResp: %s", idstring(err, params, subs))
		return
	}
	sendOk, timedOut := trans.SendEvent(subRespMsg, e2tRecvMsgTimeout)
	if sendOk == false {
		err = fmt.Errorf("Passing event to transaction failed: sendOk(%t) timedOut(%t)", sendOk, timedOut)
		xapp.Logger.Error("MSG-SubResp: %s", idstring(err, trans, subs))
	}
	return
}

//-------------------------------------------------------------------
// handle from E2T Subscription Failure
//-------------------------------------------------------------------
func (c *Control) handleE2TSubscriptionFailure(params *xapptweaks.RMRParams) {
	xapp.Logger.Info("MSG from E2T: %s", params.String())
	subFailMsg, err := c.e2ap.UnpackSubscriptionFailure(params.Payload)
	if err != nil {
		xapp.Logger.Error("MSG-SubFail %s", idstring(err, params))
		return
	}
	subs, err := c.registry.GetSubscriptionFirstMatch([]uint32{subFailMsg.RequestId.InstanceId})
	if err != nil {
		xapp.Logger.Error("MSG-SubFail: %s", idstring(err, params))
		return
	}
	trans := subs.GetTransaction()
	if trans == nil {
		err = fmt.Errorf("Ongoing transaction not found")
		xapp.Logger.Error("MSG-SubFail: %s", idstring(err, params, subs))
		return
	}
	sendOk, timedOut := trans.SendEvent(subFailMsg, e2tRecvMsgTimeout)
	if sendOk == false {
		err = fmt.Errorf("Passing event to transaction failed: sendOk(%t) timedOut(%t)", sendOk, timedOut)
		xapp.Logger.Error("MSG-SubFail: %s", idstring(err, trans, subs))
	}
	return
}

//-------------------------------------------------------------------
// handle from E2T Subscription Delete Response
//-------------------------------------------------------------------
func (c *Control) handleE2TSubscriptionDeleteResponse(params *xapptweaks.RMRParams) (err error) {
	xapp.Logger.Info("MSG from E2T: %s", params.String())
	subDelRespMsg, err := c.e2ap.UnpackSubscriptionDeleteResponse(params.Payload)
	if err != nil {
		xapp.Logger.Error("MSG-SubDelResp: %s", idstring(err, params))
		return
	}
	subs, err := c.registry.GetSubscriptionFirstMatch([]uint32{subDelRespMsg.RequestId.InstanceId})
	if err != nil {
		xapp.Logger.Error("MSG-SubDelResp: %s", idstring(err, params))
		return
	}
	trans := subs.GetTransaction()
	if trans == nil {
		err = fmt.Errorf("Ongoing transaction not found")
		xapp.Logger.Error("MSG-SubDelResp: %s", idstring(err, params, subs))
		return
	}
	sendOk, timedOut := trans.SendEvent(subDelRespMsg, e2tRecvMsgTimeout)
	if sendOk == false {
		err = fmt.Errorf("Passing event to transaction failed: sendOk(%t) timedOut(%t)", sendOk, timedOut)
		xapp.Logger.Error("MSG-SubDelResp: %s", idstring(err, trans, subs))
	}
	return
}

//-------------------------------------------------------------------
// handle from E2T Subscription Delete Failure
//-------------------------------------------------------------------
func (c *Control) handleE2TSubscriptionDeleteFailure(params *xapptweaks.RMRParams) {
	xapp.Logger.Info("MSG from E2T: %s", params.String())
	subDelFailMsg, err := c.e2ap.UnpackSubscriptionDeleteFailure(params.Payload)
	if err != nil {
		xapp.Logger.Error("MSG-SubDelFail: %s", idstring(err, params))
		return
	}
	subs, err := c.registry.GetSubscriptionFirstMatch([]uint32{subDelFailMsg.RequestId.InstanceId})
	if err != nil {
		xapp.Logger.Error("MSG-SubDelFail: %s", idstring(err, params))
		return
	}
	trans := subs.GetTransaction()
	if trans == nil {
		err = fmt.Errorf("Ongoing transaction not found")
		xapp.Logger.Error("MSG-SubDelFail: %s", idstring(err, params, subs))
		return
	}
	sendOk, timedOut := trans.SendEvent(subDelFailMsg, e2tRecvMsgTimeout)
	if sendOk == false {
		err = fmt.Errorf("Passing event to transaction failed: sendOk(%t) timedOut(%t)", sendOk, timedOut)
		xapp.Logger.Error("MSG-SubDelFail: %s", idstring(err, trans, subs))
	}
	return
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func typeofSubsMessage(v interface{}) string {
	if v == nil {
		return "NIL"
	}
	switch v.(type) {
	case *e2ap.E2APSubscriptionRequest:
		return "SubReq"
	case *e2ap.E2APSubscriptionResponse:
		return "SubResp"
	case *e2ap.E2APSubscriptionFailure:
		return "SubFail"
	case *e2ap.E2APSubscriptionDeleteRequest:
		return "SubDelReq"
	case *e2ap.E2APSubscriptionDeleteResponse:
		return "SubDelResp"
	case *e2ap.E2APSubscriptionDeleteFailure:
		return "SubDelFail"
	default:
		return "Unknown"
	}
}
