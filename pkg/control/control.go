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
	"gerrit.o-ran-sc.org/r/ric-plt/e2ap/pkg/e2ap"
	rtmgrclient "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client"
	rtmgrhandle "gerrit.o-ran-sc.org/r/ric-plt/submgr/pkg/rtmgr_client/handle"
	"gerrit.o-ran-sc.org/r/ric-plt/xapp-frame/pkg/xapp"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/spf13/viper"
	"sync"
	"time"
)

//-----------------------------------------------------------------------------
//
//-----------------------------------------------------------------------------

var e2tSubReqTimeout time.Duration = 5 * time.Second
var e2tSubDelReqTime time.Duration = 5 * time.Second
var e2tMaxSubReqTryCount uint64 = 2    // Initial try + retry
var e2tMaxSubDelReqTryCount uint64 = 2 // Initial try + retry

type Control struct {
	e2ap         *E2ap
	registry     *Registry
	tracker      *Tracker
	timerMap     *TimerMap
	rmrSendMutex sync.Mutex
	msgCounter   uint64
}

type RMRMeid struct {
	PlmnID  string
	EnbID   string
	RanName string
}

const (
	CREATE Action = 0
	MERGE  Action = 1
	NONE   Action = 2
	DELETE Action = 3
)

func init() {
	xapp.Logger.Info("SUBMGR")
	viper.AutomaticEnv()
	viper.SetEnvPrefix("submgr")
	viper.AllowEmptyEnv(true)
}

func NewControl() *Control {

	transport := httptransport.New(viper.GetString("rtmgr.HostAddr")+":"+viper.GetString("rtmgr.port"), viper.GetString("rtmgr.baseUrl"), []string{"http"})
	client := rtmgrclient.New(transport, strfmt.Default)
	handle := rtmgrhandle.NewProvideXappSubscriptionHandleParamsWithTimeout(10 * time.Second)
	deleteHandle := rtmgrhandle.NewDeleteXappSubscriptionHandleParamsWithTimeout(10 * time.Second)
	rtmgrClient := RtmgrClient{client, handle, deleteHandle}

	registry := new(Registry)
	registry.Initialize()
	registry.rtmgrClient = &rtmgrClient

	tracker := new(Tracker)
	tracker.Init()

	timerMap := new(TimerMap)
	timerMap.Init()

	return &Control{e2ap: new(E2ap),
		registry:   registry,
		tracker:    tracker,
		timerMap:   timerMap,
		msgCounter: 0,
	}
}

func (c *Control) Run() {
	xapp.Run(c)
}

func (c *Control) rmrSendRaw(desc string, params *RMRParams) (err error) {

	xapp.Logger.Info("%s: %s", desc, params.String())
	status := false
	i := 1
	for ; i <= 10 && status == false; i++ {
		c.rmrSendMutex.Lock()
		status = xapp.Rmr.Send(params.RMRParams, false)
		c.rmrSendMutex.Unlock()
		if status == false {
			xapp.Logger.Info("rmr.Send() failed. Retry count %d, %s", i, params.String())
			time.Sleep(500 * time.Millisecond)
		}
	}
	if status == false {
		err = fmt.Errorf("rmr.Send() failed. Retry count %d, %s", i, params.String())
		xapp.Logger.Error("%s: %s", desc, err.Error())
		xapp.Rmr.Free(params.Mbuf)
	}
	return
}

func (c *Control) rmrSend(desc string, subs *Subscription, trans *Transaction) (err error) {
	params := &RMRParams{&xapp.RMRParams{}}
	params.Mtype = trans.GetMtype()
	params.SubId = int(subs.GetSubId())
	params.Xid = ""
	params.Meid = subs.GetMeid()
	params.Src = ""
	params.PayloadLen = len(trans.Payload.Buf)
	params.Payload = trans.Payload.Buf
	params.Mbuf = nil

	return c.rmrSendRaw(desc, params)
}

func (c *Control) rmrReplyToSender(desc string, subs *Subscription, trans *Transaction) (err error) {
	params := &RMRParams{&xapp.RMRParams{}}
	params.Mtype = trans.GetMtype()
	params.SubId = int(subs.GetSubId())
	params.Xid = trans.GetXid()
	params.Meid = trans.GetMeid()
	params.Src = ""
	params.PayloadLen = len(trans.Payload.Buf)
	params.Payload = trans.Payload.Buf
	params.Mbuf = nil

	return c.rmrSendRaw(desc, params)
}

func (c *Control) Consume(params *xapp.RMRParams) (err error) {
	xapp.Rmr.Free(params.Mbuf)
	params.Mbuf = nil
	msg := &RMRParams{params}
	c.msgCounter++
	switch msg.Mtype {
	case xapp.RICMessageTypes["RIC_SUB_REQ"]:
		go c.handleXAPPSubscriptionRequest(msg)
	case xapp.RICMessageTypes["RIC_SUB_RESP"]:
		go c.handleE2TSubscriptionResponse(msg)
	case xapp.RICMessageTypes["RIC_SUB_FAILURE"]:
		go c.handleE2TSubscriptionFailure(msg)
	case xapp.RICMessageTypes["RIC_SUB_DEL_REQ"]:
		go c.handleXAPPSubscriptionDeleteRequest(msg)
	case xapp.RICMessageTypes["RIC_SUB_DEL_RESP"]:
		go c.handleE2TSubscriptionDeleteResponse(msg)
	case xapp.RICMessageTypes["RIC_SUB_DEL_FAILURE"]:
		go c.handleE2TSubscriptionDeleteFailure(msg)
	default:
		xapp.Logger.Info("Unknown Message Type '%d', discarding", msg.Mtype)
	}

	return nil
}
func idstring(trans fmt.Stringer, subs fmt.Stringer, err error) string {
	var retval string = ""
	var filler string = ""
	if trans != nil {
		retval += filler + trans.String()
		filler = " "
	}
	if subs != nil {
		retval += filler + subs.String()
		filler = " "
	}
	if err != nil {
		retval += filler + "err(" + err.Error() + ")"
		filler = " "

	}
	return retval
}

//-------------------------------------------------------------------
//
// XAPP->SUBS REQ
//
//-------------------------------------------------------------------
func (c *Control) handleXAPPSubscriptionRequest(params *RMRParams) {
	xapp.Logger.Info("XAPP-SubReq from xapp: %s", params.String())

	subReqMsg, err := c.e2ap.UnpackSubscriptionRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(params, nil, err))
		return
	}

	trans, err := c.tracker.TrackTransaction(NewRmrEndpoint(params.Src), params.Xid, params.Meid)
	if err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(params, nil, err))
		return
	}
	defer trans.Release()

	subs, err := c.registry.AssignToSubscription(trans, subReqMsg)
	if err != nil {
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(trans, nil, err))
		return
	}

	if subs.IsTransactionReserved() {
		err := fmt.Errorf("Currently parallel or queued transactions are not allowed")
		xapp.Logger.Error("XAPP-SubReq: %s", idstring(trans, subs, err))
		return
	}

	//
	// Wake subs request
	//
	go c.handleSubscriptionCreate(subs, trans)
	event, timedOut := trans.WaitEvent(0) //blocked wait as timeout is handled in subs side

	err = nil
	if timedOut == false {
		if event != nil {
			switch themsg := event.(type) {
			case *e2ap.E2APSubscriptionResponse:
				trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionResponse(themsg)
				if err == nil {
					c.rmrReplyToSender("XAPP-SubReq: SubResp to xapp", subs, trans)
					return
				}
			case *e2ap.E2APSubscriptionFailure:
				trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionFailure(themsg)
				if err == nil {
					c.rmrReplyToSender("XAPP-SubReq: SubFail to xapp", subs, trans)
				}
				go func() {
					time.Sleep(5 * time.Second)
					xapp.Logger.Info("XAPP-SubReq: SubFail cleaning: %s", idstring(trans, subs, err))
					subs.DelEndpoint(trans.GetEndpoint())
				}()
				return
			default:
				break
			}
		}
	}

	//
	//Generate internal delete and clean stuff
	//
	xapp.Logger.Info("XAPP-SubReq: internal delete due event(%T) timedOut(%t) %s", event, timedOut, idstring(trans, subs, err))
	go c.handleSubscriptionDelete(subs, trans)
	trans.WaitEvent(0) //blocked wait as timeout is handled in subs side

	//go func() {
	//	time.Sleep(5 * time.Second)
	xapp.Logger.Info("XAPP-SubReq: internal delete cleaning %s", idstring(trans, subs, err))
	subs.DelEndpoint(trans.GetEndpoint())
	//}()
}

//-------------------------------------------------------------------
//
// XAPP->SUBS DEL REQ
//
//-------------------------------------------------------------------

func (c *Control) handleXAPPSubscriptionDeleteRequest(params *RMRParams) {
	xapp.Logger.Info("XAPP-SubDelReq from xapp: %s", params.String())

	subDelReqMsg, err := c.e2ap.UnpackSubscriptionDeleteRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("XAPP-SubDelReq %s", idstring(params, nil, err))
		return
	}

	trans, err := c.tracker.TrackTransaction(NewRmrEndpoint(params.Src), params.Xid, params.Meid)
	if err != nil {
		xapp.Logger.Error("XAPP-SubDelReq %s", idstring(params, nil, err))
		return
	}
	defer trans.Release()

	subs, err := c.registry.GetSubscriptionFirstMatch([]uint16{uint16(subDelReqMsg.RequestId.Seq), uint16(params.SubId)})
	if err != nil {
		xapp.Logger.Error("XAPP-SubDelReq: %s", idstring(trans, nil, err))
		return
	}

	if subs.IsTransactionReserved() {
		err := fmt.Errorf("Currently parallel or queued transactions are not allowed")
		xapp.Logger.Error("XAPP-SubDelReq: %s", idstring(trans, subs, err))
		return
	}

	//
	// Wake subs side transaction and wait its response
	//
	go c.handleSubscriptionDelete(subs, trans)
	trans.WaitEvent(0) //blocked wait as timeout is handled in subs side

	// Whatever is received send ok delete response
	subDelRespMsg := &e2ap.E2APSubscriptionDeleteResponse{}
	subDelRespMsg.RequestId.Id = subs.SubReqMsg.RequestId.Id
	subDelRespMsg.RequestId.Seq = uint32(subs.GetSubId())
	subDelRespMsg.FunctionId = subs.SubReqMsg.FunctionId
	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionDeleteResponse(subDelRespMsg)
	if err == nil {
		c.rmrReplyToSender("XAPP-SubDelReq: SubDelResp to xapp", subs, trans)
	}
	go func() {
		time.Sleep(5 * time.Second)
		xapp.Logger.Info("XAPP-SubDelReq: SubDelResp cleaning: %s", idstring(trans, subs, err))
		subs.DelEndpoint(trans.GetEndpoint())
	}()
}

//-------------------------------------------------------------------
//
// SUBS->RAN
//
//-------------------------------------------------------------------

func (c *Control) handleSubscriptionCreate(subs *Subscription, parentTrans TransactionIf) {
	var err error
	var event interface{}
	var timedOut bool

	trans := c.tracker.NewTransaction(subs.GetMeid())
	subs.WaitTransactionTurn(trans)
	defer subs.ReleaseTransactionTurn(trans)
	defer trans.Release()

	if subs.SubRespMsg != nil {
		xapp.Logger.Debug("SUBS-SubReq: Handling (immediate response) %s parent %s", idstring(nil, subs, nil), parentTrans.String())
		parentTrans.SendEvent(subs.SubRespMsg, 0)
		return
	}

	xapp.Logger.Debug("SUBS-SubReq: Handling %s parent %s", idstring(trans, subs, nil), parentTrans.String())

	subReqMsg := subs.SubReqMsg
	subReqMsg.RequestId.Id = 123
	subReqMsg.RequestId.Seq = uint32(subs.GetSubId())
	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionRequest(subReqMsg)
	if err != nil {
		xapp.Logger.Error("SUBS-SubReq: %s parent %s", idstring(trans, subs, err), parentTrans.String())
		parentTrans.SendEvent(nil, 0)
		return
	}

	for retries := uint64(0); retries < e2tMaxSubReqTryCount; retries++ {
		c.rmrSend("SUBS-SubReq: SubReq to E2T", subs, trans)
		event, timedOut = trans.WaitEvent(e2tSubReqTimeout)
		if timedOut {
			xapp.Logger.Info("SUBS-SubReq: Timeout: Handling (retries=%d) %s parent %s", retries, idstring(trans, subs, nil), parentTrans.String())
			continue
		}
		break
	}

	switch themsg := event.(type) {
	case *e2ap.E2APSubscriptionResponse:
		xapp.Logger.Debug("SUBS-SubReq: SubResp: Handling %s parent %s", idstring(trans, subs, nil), parentTrans.String())
		subs.SubRespMsg = themsg
	case *e2ap.E2APSubscriptionFailure:
		xapp.Logger.Debug("SUBS-SubReq: SubFail: Handling %s parent %s", idstring(trans, subs, nil), parentTrans.String())
	default:
		xapp.Logger.Error("SUBS-SubReq: WaitEvent failure event(%T) timedOut(%t) %s parent %s", themsg, timedOut, idstring(trans, subs, nil), parentTrans.String())
	}
	parentTrans.SendEvent(event, 0)
	return
}

//-------------------------------------------------------------------
//
// SUBS->RAN DEL REQ
//
//-------------------------------------------------------------------

func (c *Control) handleSubscriptionDelete(subs *Subscription, parentTrans TransactionIf) {
	var err error
	var event interface{}
	var timedOut bool

	trans := c.tracker.NewTransaction(subs.GetMeid())
	subs.WaitTransactionTurn(trans)
	defer subs.ReleaseTransactionTurn(trans)
	defer trans.Release()

	xapp.Logger.Debug("SUBS-SubDelReq: Handling %s parent %s", idstring(trans, subs, nil), parentTrans.String())

	subDelReqMsg := &e2ap.E2APSubscriptionDeleteRequest{}
	subDelReqMsg.RequestId.Id = 123
	subDelReqMsg.RequestId.Seq = uint32(subs.GetSubId())
	subDelReqMsg.FunctionId = 0
	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionDeleteRequest(subDelReqMsg)
	if err != nil {
		xapp.Logger.Error("SUBS-SubDelReq: %s parent %s", idstring(trans, subs, err), parentTrans.String())
		parentTrans.SendEvent(nil, 0)
		return
	}

	for retries := uint64(0); retries < e2tMaxSubDelReqTryCount; retries++ {
		c.rmrSend("SUBS-SubDelReq: SubDelReq to E2T", subs, trans)
		event, timedOut = trans.WaitEvent(e2tSubDelReqTime)
		if timedOut {
			xapp.Logger.Info("SUBS-SubDelReq: Timeout: Handling (retries=%d) %s parent %s", retries, idstring(trans, subs, nil), parentTrans.String())
			continue
		}
		break
	}

	switch themsg := event.(type) {
	case *e2ap.E2APSubscriptionDeleteResponse:
		xapp.Logger.Debug("SUBS-SubDelReq: SubDelResp: Handling %s parent %s", idstring(trans, subs, nil), parentTrans.String())
	case *e2ap.E2APSubscriptionDeleteFailure:
		xapp.Logger.Debug("SUBS-SubDelReq: SubDelFail: Handling %s parent %s", idstring(trans, subs, nil), parentTrans.String())
	default:
		xapp.Logger.Error("SUBS-SubDelReq: WaitEvent failure event(%T) timedOut(%t) %s parent %s", themsg, timedOut, idstring(trans, subs, nil), parentTrans.String())
	}

	parentTrans.SendEvent(event, 0)
}

//-------------------------------------------------------------------
// E2T -> SUBS
//-------------------------------------------------------------------
func (c *Control) handleE2TSubscriptionResponse(params *RMRParams) {
	xapp.Logger.Info("MSG-SubResp from E2T: %s", params.String())
	subRespMsg, err := c.e2ap.UnpackSubscriptionResponse(params.Payload)
	if err != nil {
		xapp.Logger.Error("MSG-SubResp %s", idstring(params, nil, err))
		return
	}
	subs, err := c.registry.GetSubscriptionFirstMatch([]uint16{uint16(subRespMsg.RequestId.Seq), uint16(params.SubId)})
	if err != nil {
		xapp.Logger.Error("MSG-SubResp: %s", idstring(params, nil, err))
		return
	}
	trans := subs.GetTransaction()
	if trans == nil {
		err = fmt.Errorf("Ongoing transaction not found")
		xapp.Logger.Error("MSG-SubResp: %s", idstring(params, subs, err))
		return
	}
	sendOk, timedOut := trans.SendEvent(subRespMsg, 5)
	if sendOk == false {
		err = fmt.Errorf("Passing event to transaction failed: sendOk(%t) timedOut(%t)", sendOk, timedOut)
		xapp.Logger.Error("MSG-SubResp: %s", idstring(trans, subs, err))
	}
	return
}

func (c *Control) handleE2TSubscriptionFailure(params *RMRParams) {
	xapp.Logger.Info("MSG-SubFail from E2T: %s", params.String())
	subFailMsg, err := c.e2ap.UnpackSubscriptionFailure(params.Payload)
	if err != nil {
		xapp.Logger.Error("MSG-SubFail %s", idstring(params, nil, err))
		return
	}
	subs, err := c.registry.GetSubscriptionFirstMatch([]uint16{uint16(subFailMsg.RequestId.Seq), uint16(params.SubId)})
	if err != nil {
		xapp.Logger.Error("MSG-SubFail: %s", idstring(params, nil, err))
		return
	}
	trans := subs.GetTransaction()
	if trans == nil {
		err = fmt.Errorf("Ongoing transaction not found")
		xapp.Logger.Error("MSG-SubFail: %s", idstring(params, subs, err))
		return
	}
	sendOk, timedOut := trans.SendEvent(subFailMsg, 5)
	if sendOk == false {
		err = fmt.Errorf("Passing event to transaction failed: sendOk(%t) timedOut(%t)", sendOk, timedOut)
		xapp.Logger.Error("MSG-SubFail: %s", idstring(trans, subs, err))
	}
	return
}

func (c *Control) handleE2TSubscriptionDeleteResponse(params *RMRParams) (err error) {
	xapp.Logger.Info("SUBS-SubDelResp from E2T:%s", params.String())
	subDelRespMsg, err := c.e2ap.UnpackSubscriptionDeleteResponse(params.Payload)
	if err != nil {
		xapp.Logger.Error("SUBS-SubDelResp: %s", idstring(params, nil, err))
		return
	}
	subs, err := c.registry.GetSubscriptionFirstMatch([]uint16{uint16(subDelRespMsg.RequestId.Seq), uint16(params.SubId)})
	if err != nil {
		xapp.Logger.Error("SUBS-SubDelResp: %s", idstring(params, nil, err))
		return
	}
	trans := subs.GetTransaction()
	if trans == nil {
		err = fmt.Errorf("Ongoing transaction not found")
		xapp.Logger.Error("SUBS-SubDelResp: %s", idstring(params, subs, err))
		return
	}
	sendOk, timedOut := trans.SendEvent(subDelRespMsg, 5)
	if sendOk == false {
		err = fmt.Errorf("Passing event to transaction failed: sendOk(%t) timedOut(%t)", sendOk, timedOut)
		xapp.Logger.Error("MSG-SubDelResp: %s", idstring(trans, subs, err))
	}
	return
}

func (c *Control) handleE2TSubscriptionDeleteFailure(params *RMRParams) {
	xapp.Logger.Info("MSG-SubDelFail from E2T:%s", params.String())
	subDelFailMsg, err := c.e2ap.UnpackSubscriptionDeleteFailure(params.Payload)
	if err != nil {
		xapp.Logger.Error("MSG-SubDelFail: %s", idstring(params, nil, err))
		return
	}
	subs, err := c.registry.GetSubscriptionFirstMatch([]uint16{uint16(subDelFailMsg.RequestId.Seq), uint16(params.SubId)})
	if err != nil {
		xapp.Logger.Error("MSG-SubDelFail: %s", idstring(params, nil, err))
		return
	}
	trans := subs.GetTransaction()
	if trans == nil {
		err = fmt.Errorf("Ongoing transaction not found")
		xapp.Logger.Error("MSG-SubDelFail: %s", idstring(params, subs, err))
		return
	}
	sendOk, timedOut := trans.SendEvent(subDelFailMsg, 5)
	if sendOk == false {
		err = fmt.Errorf("Passing event to transaction failed: sendOk(%t) timedOut(%t)", sendOk, timedOut)
		xapp.Logger.Error("MSG-SubDelFail: %s", idstring(trans, subs, err))
	}
	return
}
