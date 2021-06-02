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
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	rtmgrclient "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
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

var e2tSubReqTimeout time.Duration
var e2tSubDelReqTime time.Duration
var e2tRecvMsgTimeout time.Duration
var waitRouteCleanup_ms time.Duration
var e2tMaxSubReqTryCount uint64    // Initial try + retry
var e2tMaxSubDelReqTryCount uint64 // Initial try + retry
var readSubsFromDb string

type Control struct {
	*xapp.RMRClient
	e2ap     *E2ap
	registry *Registry
	tracker  *Tracker
	db       Sdlnterface
	//subscriber *xapp.Subscriber
	CntRecvMsg    uint64
	ResetTestFlag bool
	Counters      map[string]xapp.Counter
}

type RMRMeid struct {
	PlmnID  string
	EnbID   string
	RanName string
}

type SubmgrRestartTestEvent struct{}
type SubmgrRestartUpEvent struct{}

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
		db:       CreateSdl(),
		//subscriber: subscriber,
		Counters: xapp.Metric.RegisterCounterGroup(GetMetricsOpts(), "SUBMGR"),
	}
	c.ReadConfigParameters("")

	// Register REST handler for testing support
	xapp.Resource.InjectRoute("/ric/v1/test/{testId}", c.TestRestHandler, "POST")
	xapp.Resource.InjectRoute("/ric/v1/symptomdata", c.SymptomDataHandler, "GET")

	go xapp.Subscription.Listen(c.SubscriptionHandler, c.QueryHandler, c.SubscriptionDeleteHandlerCB)

	if readSubsFromDb == "false" {
		return c
	}

	// Read subscriptions from db
	xapp.Logger.Info("Reading subscriptions from db")
	subIds, register, err := c.ReadAllSubscriptionsFromSdl()
	if err != nil {
		xapp.Logger.Error("%v", err)
	} else {
		c.registry.subIds = subIds
		c.registry.register = register
		c.HandleUncompletedSubscriptions(register)
	}
	return c
}

func (c *Control) SymptomDataHandler(w http.ResponseWriter, r *http.Request) {
	subscriptions, _ := c.registry.QueryHandler()
	xapp.Resource.SendSymptomDataJson(w, r, subscriptions, "platform/subscriptions.json")
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) ReadConfigParameters(f string) {

	// viper.GetDuration returns nanoseconds
	e2tSubReqTimeout = viper.GetDuration("controls.e2tSubReqTimeout_ms") * 1000000
	if e2tSubReqTimeout == 0 {
		e2tSubReqTimeout = 2000 * 1000000
	}
	xapp.Logger.Info("e2tSubReqTimeout %v", e2tSubReqTimeout)
	e2tSubDelReqTime = viper.GetDuration("controls.e2tSubDelReqTime_ms") * 1000000
	if e2tSubDelReqTime == 0 {
		e2tSubDelReqTime = 2000 * 1000000
	}
	xapp.Logger.Info("e2tSubDelReqTime %v", e2tSubDelReqTime)
	e2tRecvMsgTimeout = viper.GetDuration("controls.e2tRecvMsgTimeout_ms") * 1000000
	if e2tRecvMsgTimeout == 0 {
		e2tRecvMsgTimeout = 2000 * 1000000
	}
	xapp.Logger.Info("e2tRecvMsgTimeout %v", e2tRecvMsgTimeout)

	// Internal cfg parameter, used to define a wait time for RMR route clean-up. None default
	// value 100ms used currently only in unittests.
	waitRouteCleanup_ms = viper.GetDuration("controls.waitRouteCleanup_ms") * 1000000
	if waitRouteCleanup_ms == 0 {
		waitRouteCleanup_ms = 5000 * 1000000
	}
	xapp.Logger.Info("waitRouteCleanup %v", waitRouteCleanup_ms)

	e2tMaxSubReqTryCount = viper.GetUint64("controls.e2tMaxSubReqTryCount")
	if e2tMaxSubReqTryCount == 0 {
		e2tMaxSubReqTryCount = 1
	}
	xapp.Logger.Info("e2tMaxSubReqTryCount %v", e2tMaxSubReqTryCount)

	e2tMaxSubDelReqTryCount = viper.GetUint64("controls.e2tMaxSubDelReqTryCount")
	if e2tMaxSubDelReqTryCount == 0 {
		e2tMaxSubDelReqTryCount = 1
	}
	xapp.Logger.Info("e2tMaxSubDelReqTryCount %v", e2tMaxSubDelReqTryCount)

	readSubsFromDb = viper.GetString("controls.readSubsFromDb")
	if readSubsFromDb == "" {
		readSubsFromDb = "true"
	}
	xapp.Logger.Info("readSubsFromDb %v", readSubsFromDb)
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) HandleUncompletedSubscriptions(register map[uint32]*Subscription) {

	xapp.Logger.Debug("HandleUncompletedSubscriptions. len(register) = %v", len(register))
	for subId, subs := range register {
		if subs.SubRespRcvd == false {
			subs.NoRespToXapp = true
			xapp.Logger.Debug("SendSubscriptionDeleteReq. subId = %v", subId)
			c.SendSubscriptionDeleteReq(subs)
		}
	}
}

func (c *Control) ReadyCB(data interface{}) {
	if c.RMRClient == nil {
		c.RMRClient = xapp.Rmr
	}
}

func (c *Control) Run() {
	xapp.SetReadyCB(c.ReadyCB, nil)
	xapp.AddConfigChangeListener(c.ReadConfigParameters)
	xapp.Run(c)
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) SubscriptionHandler(params interface{}) (*models.SubscriptionResponse, error) {

	c.CntRecvMsg++
	c.UpdateCounter(cRestSubReqFromXapp)

	restSubId := ksuid.New().String()
	subResp := models.SubscriptionResponse{}
	subResp.SubscriptionID = &restSubId
	p := params.(*models.SubscriptionParams)

	if p.ClientEndpoint == nil {
		xapp.Logger.Error("ClientEndpoint == nil")
		c.UpdateCounter(cRestSubFailToXapp)
		return nil, fmt.Errorf("")
	}

	_, xAppRmrEndpoint, err := ConstructEndpointAddresses(*p.ClientEndpoint)
	if err != nil {
		xapp.Logger.Error("%s", err.Error())
		c.UpdateCounter(cRestSubFailToXapp)
		return nil, err
	}

	restSubscription, err := c.registry.CreateRESTSubscription(&restSubId, &xAppRmrEndpoint, p.Meid)
	if err != nil {
		xapp.Logger.Error("%s", err.Error())
		c.UpdateCounter(cRestSubFailToXapp)
		return nil, err
	}

	subReqList := e2ap.SubscriptionRequestList{}
	err = c.e2ap.FillSubscriptionReqMsgs(params, &subReqList, restSubscription)
	if err != nil {
		xapp.Logger.Error("%s", err.Error())
		c.registry.DeleteRESTSubscription(&restSubId)
		c.UpdateCounter(cRestSubFailToXapp)
		return nil, err
	}

	go c.processSubscriptionRequests(restSubscription, &subReqList, p.ClientEndpoint, p.Meid, &restSubId)

	c.UpdateCounter(cRestSubRespToXapp)
	return &subResp, nil
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------

func (c *Control) processSubscriptionRequests(restSubscription *RESTSubscription, subReqList *e2ap.SubscriptionRequestList,
	clientEndpoint *models.SubscriptionParamsClientEndpoint, meid *string, restSubId *string) {

	xapp.Logger.Info("Subscription Request count=%v ", len(subReqList.E2APSubscriptionRequests))

	_, xAppRmrEndpoint, err := ConstructEndpointAddresses(*clientEndpoint)
	if err != nil {
		xapp.Logger.Error("%s", err.Error())
		return
	}

	var requestorID int64
	var instanceId int64
	for index := 0; index < len(subReqList.E2APSubscriptionRequests); index++ {
		subReqMsg := subReqList.E2APSubscriptionRequests[index]

		xid := *restSubId + "_" + strconv.FormatUint(uint64(subReqMsg.RequestId.InstanceId), 10)
		trans := c.tracker.NewXappTransaction(xapp.NewRmrEndpoint(xAppRmrEndpoint), xid, subReqMsg.RequestId, &xapp.RMRMeid{RanName: *meid})
		//trans := c.tracker.NewXappTransaction(xapp.NewRmrEndpoint(xAppRmrEndpoint), *restSubId, subReqMsg.RequestId, &xapp.RMRMeid{RanName: *meid})
		if trans == nil {
			c.registry.DeleteRESTSubscription(restSubId)
			xapp.Logger.Error("XAPP-SubReq transaction not created. RESTSubId=%s, EndPoint=%s, Meid=%s", *restSubId, xAppRmrEndpoint, *meid)
			return
		}

		defer trans.Release()
		xapp.Logger.Info("Handle SubscriptionRequest index=%v, %s", index, idstring(nil, trans))
		subRespMsg, err := c.handleSubscriptionRequest(trans, &subReqMsg, meid, restSubId)
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
			// Mark REST subscription request processed.
			restSubscription.SetProcessed()
			xapp.Logger.Info("Sending unsuccessful REST notification to endpoint=%v:%v, InstanceId=%v, %s", clientEndpoint.Host, clientEndpoint.HTTPPort, instanceId, idstring(nil, trans))
			xapp.Subscription.Notify(resp, *clientEndpoint)
			c.UpdateCounter(cRestSubFailNotifToXapp)
		} else {
			xapp.Logger.Info("SubscriptionRequest index=%v processed successfully. endpoint=%v, InstanceId=%v, %s", index, *clientEndpoint, instanceId, idstring(nil, trans))

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
			xapp.Logger.Info("Sending successful REST notification to endpoint=%v, InstanceId=%v, %s", *clientEndpoint, instanceId, idstring(nil, trans))
			xapp.Subscription.Notify(resp, *clientEndpoint)
			c.UpdateCounter(cRestSubNotifToXapp)

		}
	}
}

//-------------------------------------------------------------------
//
//------------------------------------------------------------------
func (c *Control) handleSubscriptionRequest(trans *TransactionXapp, subReqMsg *e2ap.E2APSubscriptionRequest, meid *string,
	restSubId *string) (*e2ap.E2APSubscriptionResponse, error) {

	err := c.tracker.Track(trans)
	if err != nil {
		err = fmt.Errorf("XAPP-SubReq: %s", idstring(err, trans))
		xapp.Logger.Error("%s", err.Error())
		return nil, err
	}

	subs, err := c.registry.AssignToSubscription(trans, subReqMsg, c.ResetTestFlag, c)
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
	c.registry.RemoveFromSubscription(subs, trans, waitRouteCleanup_ms, c)
	return nil, err
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) SubscriptionDeleteHandlerCB(restSubId string) error {

	c.CntRecvMsg++
	c.UpdateCounter(cRestSubDelReqFromXapp)

	xapp.Logger.Info("SubscriptionDeleteRequest from XAPP")

	restSubscription, err := c.registry.GetRESTSubscription(restSubId)
	if err != nil {
		xapp.Logger.Error("%s", err.Error())
		if restSubscription == nil {
			// Subscription was not found
			return nil
		} else {
			if restSubscription.SubReqOngoing == true {
				err := fmt.Errorf("Handling of the REST Subscription Request still ongoing %s", restSubId)
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

	c.UpdateCounter(cRestSubDelRespToXapp)

	return nil
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) SubscriptionDeleteHandler(restSubId *string, endPoint *string, meid *string, instanceId uint32) error {

	xid := *restSubId + "_" + strconv.FormatUint(uint64(instanceId), 10)
	trans := c.tracker.NewXappTransaction(xapp.NewRmrEndpoint(*endPoint), xid, e2ap.RequestId{0, 0}, &xapp.RMRMeid{RanName: *meid})
	//trans := c.tracker.NewXappTransaction(xapp.NewRmrEndpoint(*endPoint), *restSubId, e2ap.RequestId{0, 0}, &xapp.RMRMeid{RanName: *meid})
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

	c.registry.RemoveFromSubscription(subs, trans, waitRouteCleanup_ms, c)

	return nil
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) QueryHandler() (models.SubscriptionList, error) {
	xapp.Logger.Info("QueryHandler() called")

	c.CntRecvMsg++

	return c.registry.QueryHandler()
}

func (c *Control) TestRestHandler(w http.ResponseWriter, r *http.Request) {
	xapp.Logger.Info("TestRestHandler() called")

	pathParams := mux.Vars(r)
	s := pathParams["testId"]

	// This can be used to delete single subscription from db
	if contains := strings.Contains(s, "deletesubid="); contains == true {
		var splits = strings.Split(s, "=")
		if subId, err := strconv.ParseInt(splits[1], 10, 64); err == nil {
			xapp.Logger.Info("RemoveSubscriptionFromSdl() called. subId = %v", subId)
			c.RemoveSubscriptionFromSdl(uint32(subId))
			return
		}
	}

	// This can be used to remove all subscriptions db from
	if s == "emptydb" {
		xapp.Logger.Info("RemoveAllSubscriptionsFromSdl() called")
		c.RemoveAllSubscriptionsFromSdl()
		return
	}

	// This is meant to cause submgr's restart in testing
	if s == "restart" {
		xapp.Logger.Info("os.Exit(1) called")
		os.Exit(1)
	}

	xapp.Logger.Info("Unsupported rest command received %s", s)
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------

func (c *Control) rmrSendToE2T(desc string, subs *Subscription, trans *TransactionSubs) (err error) {
	params := &xapp.RMRParams{}
	params.Mtype = trans.GetMtype()
	params.SubId = int(subs.GetReqId().InstanceId)
	params.Xid = ""
	params.Meid = subs.GetMeid()
	params.Src = ""
	params.PayloadLen = len(trans.Payload.Buf)
	params.Payload = trans.Payload.Buf
	params.Mbuf = nil
	xapp.Logger.Info("MSG to E2T: %s %s %s", desc, trans.String(), params.String())
	err = c.SendWithRetry(params, false, 5)
	if err != nil {
		xapp.Logger.Error("rmrSendToE2T: Send failed: %+v", err)
	}
	return err
}

func (c *Control) rmrSendToXapp(desc string, subs *Subscription, trans *TransactionXapp) (err error) {

	params := &xapp.RMRParams{}
	params.Mtype = trans.GetMtype()
	params.SubId = int(subs.GetReqId().InstanceId)
	params.Xid = trans.GetXid()
	params.Meid = trans.GetMeid()
	params.Src = ""
	params.PayloadLen = len(trans.Payload.Buf)
	params.Payload = trans.Payload.Buf
	params.Mbuf = nil
	xapp.Logger.Info("MSG to XAPP: %s %s %s", desc, trans.String(), params.String())
	err = c.SendWithRetry(params, false, 5)
	if err != nil {
		xapp.Logger.Error("rmrSendToXapp: Send failed: %+v", err)
	}
	return err
}

func (c *Control) Consume(msg *xapp.RMRParams) (err error) {
	if c.RMRClient == nil {
		err = fmt.Errorf("Rmr object nil can handle %s", msg.String())
		xapp.Logger.Error("%s", err.Error())
		return
	}
	c.CntRecvMsg++

	defer c.RMRClient.Free(msg.Mbuf)

	// xapp-frame might use direct access to c buffer and
	// when msg.Mbuf is freed, someone might take it into use
	// and payload data might be invalid inside message handle function
	//
	// subscriptions won't load system a lot so there is no
	// real performance hit by cloning buffer into new go byte slice
	cPay := append(msg.Payload[:0:0], msg.Payload...)
	msg.Payload = cPay
	msg.PayloadLen = len(cPay)

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
func (c *Control) handleXAPPSubscriptionRequest(params *xapp.RMRParams) {
	xapp.Logger.Info("MSG from XAPP: %s", params.String())
	c.UpdateCounter(cSubReqFromXapp)

	subReqMsg, err := c.e2ap.UnpackSubscriptionRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(err, params))
		return
	}

	trans := c.tracker.NewXappTransaction(xapp.NewRmrEndpoint(params.Src), params.Xid, subReqMsg.RequestId, params.Meid)
	if trans == nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(fmt.Errorf("transaction not created"), params))
		return
	}
	defer trans.Release()

	if err = c.tracker.Track(trans); err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(err, trans))
		return
	}

	//TODO handle subscription toward e2term inside AssignToSubscription / hide handleSubscriptionCreate in it?
	subs, err := c.registry.AssignToSubscription(trans, subReqMsg, c.ResetTestFlag, c)
	if err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(err, trans))
		return
	}

	c.wakeSubscriptionRequest(subs, trans)
}

//-------------------------------------------------------------------
// Wake Subscription Request to E2node
//------------------------------------------------------------------
func (c *Control) wakeSubscriptionRequest(subs *Subscription, trans *TransactionXapp) {

	go c.handleSubscriptionCreate(subs, trans)
	event, _ := trans.WaitEvent(0) //blocked wait as timeout is handled in subs side
	var err error
	if event != nil {
		switch themsg := event.(type) {
		case *e2ap.E2APSubscriptionResponse:
			themsg.RequestId.Id = trans.RequestId.Id
			trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionResponse(themsg)
			if err == nil {
				trans.Release()
				c.UpdateCounter(cSubRespToXapp)
				c.rmrSendToXapp("", subs, trans)
				return
			}
		case *e2ap.E2APSubscriptionFailure:
			themsg.RequestId.Id = trans.RequestId.Id
			trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionFailure(themsg)
			if err == nil {
				c.UpdateCounter(cSubFailToXapp)
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
func (c *Control) handleXAPPSubscriptionDeleteRequest(params *xapp.RMRParams) {
	xapp.Logger.Info("MSG from XAPP: %s", params.String())
	c.UpdateCounter(cSubDelReqFromXapp)

	subDelReqMsg, err := c.e2ap.UnpackSubscriptionDeleteRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("XAPP-SubDelReq %s", idstring(err, params))
		return
	}

	trans := c.tracker.NewXappTransaction(xapp.NewRmrEndpoint(params.Src), params.Xid, subDelReqMsg.RequestId, params.Meid)
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

	if subs.NoRespToXapp == true {
		// Do no send delete responses to xapps due to submgr restart is deleting uncompleted subscriptions
		return
	}

	// Whatever is received success, fail or timeout, send successful delete response
	subDelRespMsg := &e2ap.E2APSubscriptionDeleteResponse{}
	subDelRespMsg.RequestId.Id = trans.RequestId.Id
	subDelRespMsg.RequestId.InstanceId = subs.GetReqId().RequestId.InstanceId
	subDelRespMsg.FunctionId = subs.SubReqMsg.FunctionId
	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionDeleteResponse(subDelRespMsg)
	if err == nil {
		c.UpdateCounter(cSubDelRespToXapp)
		c.rmrSendToXapp("", subs, trans)
	}

	//TODO handle subscription toward e2term insiged RemoveFromSubscription / hide handleSubscriptionDelete in it?
	//c.registry.RemoveFromSubscription(subs, trans, 5*time.Second)
}

//-------------------------------------------------------------------
// SUBS CREATE Handling
//-------------------------------------------------------------------
func (c *Control) handleSubscriptionCreate(subs *Subscription, parentTrans *TransactionXapp) {

	var removeSubscriptionFromDb bool = false
	trans := c.tracker.NewSubsTransaction(subs)
	subs.WaitTransactionTurn(trans)
	defer subs.ReleaseTransactionTurn(trans)
	defer trans.Release()

	xapp.Logger.Debug("SUBS-SubReq: Handling %s ", idstring(nil, trans, subs, parentTrans))

	subRfMsg, valid := subs.GetCachedResponse()
	if subRfMsg == nil && valid == true {
		event := c.sendE2TSubscriptionRequest(subs, trans, parentTrans)
		switch event.(type) {
		case *e2ap.E2APSubscriptionResponse:
			subRfMsg, valid = subs.SetCachedResponse(event, true)
			subs.SubRespRcvd = true
		case *e2ap.E2APSubscriptionFailure:
			removeSubscriptionFromDb = true
			subRfMsg, valid = subs.SetCachedResponse(event, false)
			xapp.Logger.Info("SUBS-SubReq: internal delete due event(%s) %s", typeofSubsMessage(event), idstring(nil, trans, subs, parentTrans))
			c.sendE2TSubscriptionDeleteRequest(subs, trans, parentTrans)
		case *SubmgrRestartTestEvent:
			// This simulates that no response has been received and after restart subscriptions are restored from db
			xapp.Logger.Debug("Test restart flag is active. Dropping this transaction to test restart case")
			return
		default:
			xapp.Logger.Info("SUBS-SubReq: internal delete due event(%s) %s", typeofSubsMessage(event), idstring(nil, trans, subs, parentTrans))
			removeSubscriptionFromDb = true
			subRfMsg, valid = subs.SetCachedResponse(nil, false)
			c.sendE2TSubscriptionDeleteRequest(subs, trans, parentTrans)
		}
		xapp.Logger.Debug("SUBS-SubReq: Handling (e2t response %s) %s", typeofSubsMessage(subRfMsg), idstring(nil, trans, subs, parentTrans))
	} else {
		xapp.Logger.Debug("SUBS-SubReq: Handling (cached response %s) %s", typeofSubsMessage(subRfMsg), idstring(nil, trans, subs, parentTrans))
	}

	//Now RemoveFromSubscription in here to avoid race conditions (mostly concerns delete)
	if valid == false {
		c.registry.RemoveFromSubscription(subs, parentTrans, waitRouteCleanup_ms, c)
	}

	c.UpdateSubscriptionInDB(subs, removeSubscriptionFromDb)
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
	c.registry.RemoveFromSubscription(subs, parentTrans, waitRouteCleanup_ms, c)
	c.registry.UpdateSubscriptionToDb(subs, c)
	parentTrans.SendEvent(nil, 0)
}

//-------------------------------------------------------------------
// send to E2T Subscription Request
//-------------------------------------------------------------------
func (c *Control) sendE2TSubscriptionRequest(subs *Subscription, trans *TransactionSubs, parentTrans *TransactionXapp) interface{} {
	var err error
	var event interface{} = nil
	var timedOut bool = false

	subReqMsg := subs.SubReqMsg
	subReqMsg.RequestId = subs.GetReqId().RequestId
	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionRequest(subReqMsg)
	if err != nil {
		xapp.Logger.Error("SUBS-SubReq: %s", idstring(err, trans, subs, parentTrans))
		return event
	}

	// Write uncompleted subscrition in db. If no response for subscrition it need to be re-processed (deleted) after restart
	c.WriteSubscriptionToDb(subs)

	for retries := uint64(0); retries < e2tMaxSubReqTryCount; retries++ {
		desc := fmt.Sprintf("(retry %d)", retries)
		if retries == 0 {
			c.UpdateCounter(cSubReqToE2)
		} else {
			c.UpdateCounter(cSubReReqToE2)
		}
		c.rmrSendToE2T(desc, subs, trans)
		if subs.DoNotWaitSubResp == false {
			event, timedOut = trans.WaitEvent(e2tSubReqTimeout)
			if timedOut {
				c.UpdateCounter(cSubReqTimerExpiry)
				continue
			}
		} else {
			// Simulating case where subscrition request has been sent but response has not been received before restart
			event = &SubmgrRestartTestEvent{}
		}
		break
	}
	xapp.Logger.Debug("SUBS-SubReq: Response handling event(%s) %s", typeofSubsMessage(event), idstring(nil, trans, subs, parentTrans))
	return event
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
		if retries == 0 {
			c.UpdateCounter(cSubDelReqToE2)
		} else {
			c.UpdateCounter(cSubDelReReqToE2)
		}
		c.rmrSendToE2T(desc, subs, trans)
		event, timedOut = trans.WaitEvent(e2tSubDelReqTime)
		if timedOut {
			c.UpdateCounter(cSubDelReqTimerExpiry)
			continue
		}
		break
	}
	xapp.Logger.Debug("SUBS-SubDelReq: Response handling event(%s) %s", typeofSubsMessage(event), idstring(nil, trans, subs, parentTrans))
	return event
}

//-------------------------------------------------------------------
// handle from E2T Subscription Response
//-------------------------------------------------------------------
func (c *Control) handleE2TSubscriptionResponse(params *xapp.RMRParams) {
	xapp.Logger.Info("MSG from E2T: %s", params.String())
	c.UpdateCounter(cSubRespFromE2)

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
func (c *Control) handleE2TSubscriptionFailure(params *xapp.RMRParams) {
	xapp.Logger.Info("MSG from E2T: %s", params.String())
	c.UpdateCounter(cSubFailFromE2)
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
func (c *Control) handleE2TSubscriptionDeleteResponse(params *xapp.RMRParams) (err error) {
	xapp.Logger.Info("MSG from E2T: %s", params.String())
	c.UpdateCounter(cSubDelRespFromE2)
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
func (c *Control) handleE2TSubscriptionDeleteFailure(params *xapp.RMRParams) {
	xapp.Logger.Info("MSG from E2T: %s", params.String())
	c.UpdateCounter(cSubDelFailFromE2)
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
	//case *e2ap.E2APSubscriptionRequest:
	//	return "SubReq"
	case *e2ap.E2APSubscriptionResponse:
		return "SubResp"
	case *e2ap.E2APSubscriptionFailure:
		return "SubFail"
	//case *e2ap.E2APSubscriptionDeleteRequest:
	//	return "SubDelReq"
	case *e2ap.E2APSubscriptionDeleteResponse:
		return "SubDelResp"
	case *e2ap.E2APSubscriptionDeleteFailure:
		return "SubDelFail"
	default:
		return "Unknown"
	}
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) WriteSubscriptionToDb(subs *Subscription) {
	xapp.Logger.Debug("WriteSubscriptionToDb() subId = %v", subs.ReqId.InstanceId)
	err := c.WriteSubscriptionToSdl(subs.ReqId.InstanceId, subs)
	if err != nil {
		xapp.Logger.Error("%v", err)
	}
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) UpdateSubscriptionInDB(subs *Subscription, removeSubscriptionFromDb bool) {

	if removeSubscriptionFromDb == true {
		// Subscription was written in db already when subscription request was sent to BTS, except for merged request
		c.RemoveSubscriptionFromDb(subs)
	} else {
		// Update is needed for successful response and merge case here
		if subs.RetryFromXapp == false {
			c.WriteSubscriptionToDb(subs)
		}
	}
	subs.RetryFromXapp = false
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) RemoveSubscriptionFromDb(subs *Subscription) {
	xapp.Logger.Debug("RemoveSubscriptionFromDb() subId = %v", subs.ReqId.InstanceId)
	err := c.RemoveSubscriptionFromSdl(subs.ReqId.InstanceId)
	if err != nil {
		xapp.Logger.Error("%v", err)
	}
}

func (c *Control) SendSubscriptionDeleteReq(subs *Subscription) {

	xapp.Logger.Debug("Sending subscription delete due to restart. subId = %v", subs.ReqId.InstanceId)

	// Send delete for every endpoint in the subscription
	subDelReqMsg := &e2ap.E2APSubscriptionDeleteRequest{}
	subDelReqMsg.RequestId = subs.GetReqId().RequestId
	subDelReqMsg.FunctionId = subs.SubReqMsg.FunctionId
	mType, payload, err := c.e2ap.PackSubscriptionDeleteRequest(subDelReqMsg)
	if err != nil {
		xapp.Logger.Error("SendSubscriptionDeleteReq() %s", idstring(err))
		return
	}
	for _, endPoint := range subs.EpList.Endpoints {
		params := &xapp.RMRParams{}
		params.Mtype = mType
		params.SubId = int(subs.GetReqId().InstanceId)
		params.Xid = ""
		params.Meid = subs.Meid
		params.Src = endPoint.String()
		params.PayloadLen = len(payload.Buf)
		params.Payload = payload.Buf
		params.Mbuf = nil
		subs.DeleteFromDb = true
		c.handleXAPPSubscriptionDeleteRequest(params)
	}
}