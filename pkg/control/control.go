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
	"time"

	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	rtmgrclient "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/models"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/gorilla/mux"
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
var e2tMaxSubReqTryCount uint64    // Initial try + retry
var e2tMaxSubDelReqTryCount uint64 // Initial try + retry

type Control struct {
	*xapp.RMRClient
	e2ap     *E2ap
	registry *Registry
	tracker  *Tracker
	db       Sdlnterface
	//subscriber *xapp.Subscriber
	CntRecvMsg    uint64
	ResetTestFlag bool
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

	transport := httptransport.New(viper.GetString("rtmgr.HostAddr")+":"+viper.GetString("rtmgr.port"), viper.GetString("rtmgr.baseUrl"), []string{"http"})
	rtmgrClient := RtmgrClient{rtClient: rtmgrclient.New(transport, strfmt.Default)}

	registry := new(Registry)
	registry.Initialize()
	registry.rtmgrClient = &rtmgrClient

	tracker := new(Tracker)
	tracker.Init()

	//subscriber := xapp.NewSubscriber(viper.GetString("subscription.host"), viper.GetInt("subscription.timeout"))

	c := &Control{e2ap: new(E2ap),
		registry: registry,
		tracker:  tracker,
		db:       CreateSdl(),
		//subscriber: subscriber,
	}

	// Register REST handler for testing support
	xapp.Resource.InjectRoute("/ric/v1/test/{testId}", c.TestRestHandler, "POST")

	go xapp.Subscription.Listen(c.SubscriptionHandler, c.QueryHandler, c.SubscriptionDeleteHandler)
	//go c.subscriber.Listen(c.SubscriptionHandler, c.QueryHandler)

	readSubsFromDb := viper.GetString("controls.readSubsFromDb")
	xapp.Logger.Info("readSubsFromDb %v", readSubsFromDb)
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
	xapp.Run(c)
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) SubscriptionHandler(stype models.SubscriptionType, params interface{}) (*models.SubscriptionResponse, error) {
	/*
	   switch p := params.(type) {
	   case *models.ReportParams:
	       trans := c.tracker.NewXappTransaction(NewRmrEndpoint(p.ClientEndpoint),"" , 0, &xapp.RMRMeid{RanName: p.Meid})
	       if trans == nil {
	             xapp.Logger.Error("XAPP-SubReq: %s", idstring(fmt.Errorf("transaction not created"), params))
	             return
	       }
	       defer trans.Release()
	   case *models.ControlParams:
	   case *models.PolicyParams:
	   }
	*/
	return &models.SubscriptionResponse{}, fmt.Errorf("Subscription rest interface not implemented")
}

func (c *Control) SubscriptionDeleteHandler(s string) error {
	return nil
}

func (c *Control) QueryHandler() (models.SubscriptionList, error) {
	return c.registry.QueryHandler()
}

func (c *Control) TestRestHandler(w http.ResponseWriter, r *http.Request) {

	xapp.Logger.Info("TestRestHandler() called")

	pathParams := mux.Vars(r)
	s := pathParams["testId"]

	// This can be used to empty db from subscriptions
	if s == "emptydb" {
		xapp.Logger.Info("RemoveAllSubscriptionsFromSdl() called")
		c.RemoveAllSubscriptionsFromSdl()
		return
	}

	// This is meant for submgr restart testing
	if s == "exit" {
		xapp.Logger.Info("Exit(1) called")
		os.Exit(1)
	}
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
	return c.SendWithRetry(params, false, 5)
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
	return c.SendWithRetry(params, false, 5)
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

	subReqMsg, err := c.e2ap.UnpackSubscriptionRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(err, params))
		return
	}

	trans := c.tracker.NewXappTransaction(xapp.NewRmrEndpoint(params.Src), params.Xid, subReqMsg.RequestId.InstanceId, params.Meid)
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
	subs, err := c.registry.AssignToSubscription(trans, subReqMsg, c.ResetTestFlag)
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
func (c *Control) handleXAPPSubscriptionDeleteRequest(params *xapp.RMRParams) {
	xapp.Logger.Info("MSG from XAPP: %s", params.String())

	subDelReqMsg, err := c.e2ap.UnpackSubscriptionDeleteRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("XAPP-SubDelReq %s", idstring(err, params))
		return
	}

	trans := c.tracker.NewXappTransaction(xapp.NewRmrEndpoint(params.Src), params.Xid, subDelReqMsg.RequestId.InstanceId, params.Meid)
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

	var removeSubscriptionFromDb bool = false
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

			event := c.sendE2TSubscriptionRequest(subs, trans, parentTrans)
			switch themsg := event.(type) {
			case *e2ap.E2APSubscriptionResponse:
				subRfMsg, valid = subs.SetCachedResponse(event, true)
				subs.SubRespRcvd = true
			case *e2ap.E2APSubscriptionFailure:
				removeSubscriptionFromDb = true
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
			case *SubmgrRestartTestEvent:
				// This simulates that no response has been received and after restart subscriptions are restored from db
				xapp.Logger.Info("Test restart flag is active. Dropping this transaction to test restart case")
				return
			default:
				xapp.Logger.Info("SUBS-SubReq: internal delete due event(%s) %s", typeofSubsMessage(event), idstring(nil, trans, subs, parentTrans))
				removeSubscriptionFromDb = true
				subRfMsg, valid = subs.SetCachedResponse(nil, false)
				c.sendE2TSubscriptionDeleteRequest(subs, trans, parentTrans)
			}
		}
		xapp.Logger.Debug("SUBS-SubReq: Handling (e2t response %s) %s", typeofSubsMessage(subRfMsg), idstring(nil, trans, subs, parentTrans))
	} else {
		xapp.Logger.Debug("SUBS-SubReq: Handling (cached response %s) %s", typeofSubsMessage(subRfMsg), idstring(nil, trans, subs, parentTrans))
	}

	//Now RemoveFromSubscription in here to avoid race conditions (mostly concerns delete)
	if valid == false {
		c.registry.RemoveFromSubscription(subs, parentTrans, 5*time.Second, c)
	}

	if removeSubscriptionFromDb == true {
		// Subscription was written in db already when subscription request was sent to BTS, except for merged request
		c.RemoveSubscriptionFromDb(subs)
	} else {
		// Write is needed for merge case here. Also update is needed for successful response
		c.WriteSubscriptionToDb(subs)
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
	c.registry.RemoveFromSubscription(subs, parentTrans, 5*time.Second, c)
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
		c.rmrSendToE2T(desc, subs, trans)
		if subs.DoNotWaitSubResp == false {
			event, timedOut = trans.WaitEvent(e2tSubReqTimeout)
			if timedOut {
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
func (c *Control) handleE2TSubscriptionResponse(params *xapp.RMRParams) {
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
func (c *Control) handleE2TSubscriptionFailure(params *xapp.RMRParams) {
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
func (c *Control) handleE2TSubscriptionDeleteResponse(params *xapp.RMRParams) (err error) {
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
func (c *Control) handleE2TSubscriptionDeleteFailure(params *xapp.RMRParams) {
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

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) WriteSubscriptionToDb(subs *Subscription) {
	xapp.Logger.Debug("Write subscription to db. subId = %v", subs.ReqId.InstanceId)
	err := c.WriteSubscriptionToSdl(subs.ReqId.InstanceId, subs)
	if err != nil {
		xapp.Logger.Error("%v", err)
	}
}

//-------------------------------------------------------------------
//
//-------------------------------------------------------------------
func (c *Control) RemoveSubscriptionFromDb(subs *Subscription) {
	xapp.Logger.Debug("Remove subscription from db. subId = %v", subs.ReqId.InstanceId)
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
		xapp.Logger.Error("SubDelReq: %s", idstring(err))
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

		if params == nil {
			xapp.Logger.Error("params == nil")
			os.Exit(1)
		}

		subs.DeleteFromDb = true
		c.handleXAPPSubscriptionDeleteRequest(params)
	}
}
