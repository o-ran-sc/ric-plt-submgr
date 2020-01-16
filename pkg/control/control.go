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

var subReqTime time.Duration = 5 * time.Second
var subDelReqTime time.Duration = 5 * time.Second
var maxSubReqTryCount uint64 = 2    // Initial try + retry
var maxSubDelReqTryCount uint64 = 2 // Initial try + retry

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
		go c.handleSubscriptionRequest(msg)
	case xapp.RICMessageTypes["RIC_SUB_RESP"]:
		go c.handleSubscriptionResponse(msg)
	case xapp.RICMessageTypes["RIC_SUB_FAILURE"]:
		go c.handleSubscriptionFailure(msg)
	case xapp.RICMessageTypes["RIC_SUB_DEL_REQ"]:
		go c.handleSubscriptionDeleteRequest(msg)
	case xapp.RICMessageTypes["RIC_SUB_DEL_RESP"]:
		go c.handleSubscriptionDeleteResponse(msg)
	case xapp.RICMessageTypes["RIC_SUB_DEL_FAILURE"]:
		go c.handleSubscriptionDeleteFailure(msg)
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

func (c *Control) findSubs(ids []int) (*Subscription, error) {
	var subs *Subscription = nil
	for _, id := range ids {
		if id >= 0 {
			subs = c.registry.GetSubscription(uint16(id))
		}
		if subs != nil {
			break
		}
	}
	if subs == nil {
		return nil, fmt.Errorf("No valid subscription found with ids %v", ids)
	}
	return subs, nil
}

func (c *Control) findSubsAndTrans(ids []int) (*Subscription, *Transaction, error) {
	subs, err := c.findSubs(ids)
	if err != nil {
		return nil, nil, err
	}
	trans := subs.GetTransaction()
	if trans == nil {
		return subs, nil, fmt.Errorf("No ongoing transaction found from %s", idstring(nil, subs, nil))
	}
	return subs, trans, nil
}

func (c *Control) handleSubscriptionRequest(params *RMRParams) {
	xapp.Logger.Info("SubReq from xapp: %s", params.String())

	SubReqMsg, err := c.e2ap.UnpackSubscriptionRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubReq Drop: %s", idstring(params, nil, err))
		return
	}

	trans, err := c.tracker.TrackTransaction(NewRmrEndpoint(params.Src), params.Xid, params.Meid, false, true)
	if err != nil {
		xapp.Logger.Error("SubReq Drop: %s", idstring(params, nil, err))
		return
	}
	trans.SubReqMsg = SubReqMsg

	subs, err := c.registry.ReserveSubscription(trans.Meid)
	if err != nil {
		xapp.Logger.Error("SubReq Drop: %s", idstring(trans, nil, err))
		trans.Release()
		return
	}

	err = subs.SetTransaction(trans)
	if err != nil {
		xapp.Logger.Error("SubReq Drop: %s", idstring(trans, subs, err))
		subs.Release()
		trans.Release()
		return
	}
	trans.SubReqMsg.RequestId.Seq = uint32(subs.GetSubId())

	xapp.Logger.Debug("SubReq: Handling %s", idstring(trans, subs, nil))

	//
	// TODO: subscription create is in fact owned by subscription and not transaction.
	//       Transaction is toward xapp while Subscription is toward ran.
	//       In merge several xapps may wake transactions, while only one subscription
	//       toward ran occurs -> subscription owns subscription creation toward ran
	//
	//       This is intermediate solution while improving message handling
	//
	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionRequest(trans.SubReqMsg)
	if err != nil {
		xapp.Logger.Error("SubResp Drop: %s", idstring(trans, subs, err))
		subs.Release()
		trans.Release()
		return
	}

	c.rmrSend("SubReq: SubReq to E2T", subs, trans)
	c.timerMap.StartTimer("RIC_SUB_REQ", int(subs.GetSubId()), subReqTime, FirstTry, c.handleSubscriptionRequestTimer)
	return
}

func (c *Control) handleSubscriptionResponse(params *RMRParams) {
	xapp.Logger.Info("SubResp from E2T: %s", params.String())

	SubRespMsg, err := c.e2ap.UnpackSubscriptionResponse(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubResp Drop %s", idstring(params, nil, err))
		return
	}

	subs, trans, err := c.findSubsAndTrans([]int{int(SubRespMsg.RequestId.Seq), params.SubId})
	if err != nil {
		xapp.Logger.Error("SubResp: %s", idstring(params, nil, err))
		return
	}
	trans.SubRespMsg = SubRespMsg
	xapp.Logger.Debug("SubResp: Handling %s", idstring(trans, subs, nil))

	c.timerMap.StopTimer("RIC_SUB_REQ", int(subs.GetSubId()))

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription timer already received
		return
	}

	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionResponse(trans.SubRespMsg)
	if err != nil {
		xapp.Logger.Error("SubResp: %s", idstring(trans, subs, err))
		trans.Release()
		return
	}

	subs.Confirmed()
	trans.Release()
	c.rmrReplyToSender("SubResp: SubResp to xapp", subs, trans)
	return
}

func (c *Control) handleSubscriptionFailure(params *RMRParams) {
	xapp.Logger.Info("SubFail from E2T: %s", params.String())

	SubFailMsg, err := c.e2ap.UnpackSubscriptionFailure(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubFail Drop %s", idstring(params, nil, err))
		return
	}

	subs, trans, err := c.findSubsAndTrans([]int{int(SubFailMsg.RequestId.Seq), params.SubId})
	if err != nil {
		xapp.Logger.Error("SubFail: %s", idstring(params, nil, err))
		return
	}
	trans.SubFailMsg = SubFailMsg
	xapp.Logger.Debug("SubFail: Handling %s", idstring(trans, subs, nil))

	c.timerMap.StopTimer("RIC_SUB_REQ", int(subs.GetSubId()))
	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription timer already received
		return
	}

	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionFailure(trans.SubFailMsg)
	if err == nil {
		c.rmrReplyToSender("SubFail: SubFail to xapp", subs, trans)
		time.Sleep(3 * time.Second)
	} else {
		//TODO error handling improvement
		xapp.Logger.Error("SubFail: (continue cleaning) %s", idstring(trans, subs, err))
	}

	trans.Release()
	subs.Release()
	return
}

func (c *Control) handleSubscriptionRequestTimer(strId string, nbrId int, tryCount uint64) {
	xapp.Logger.Info("SubReq timeout: subId: %v,  tryCount: %v", nbrId, tryCount)

	subs, trans, err := c.findSubsAndTrans(([]int{nbrId}))
	if err != nil {
		xapp.Logger.Error("SubReq timeout: %s", idstring(nil, nil, err))
		return
	}
	xapp.Logger.Debug("SubReq timeout: Handling %s", idstring(trans, subs, nil))

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription Response or Failure already received
		return
	}

	if tryCount < maxSubReqTryCount {
		xapp.Logger.Info("SubReq timeout: %s", idstring(trans, subs, nil))

		trans.RetryTransaction()

		c.rmrSend("SubReq timeout: SubReq to E2T", subs, trans)

		tryCount++
		c.timerMap.StartTimer("RIC_SUB_REQ", int(subs.GetSubId()), subReqTime, tryCount, c.handleSubscriptionRequestTimer)
		return
	}

	// Release CREATE transaction
	trans.Release()

	// Create DELETE transaction (internal and no messages toward xapp)
	deltrans, err := c.tracker.TrackTransaction(&trans.RmrEndpoint, trans.GetXid(), trans.GetMeid(), false, false)
	if err != nil {
		xapp.Logger.Error("SubReq timeout: %s", idstring(trans, subs, err))
		//TODO improve error handling. Important at least in merge
		subs.Release()
		return
	}

	deltrans.SubDelReqMsg = &e2ap.E2APSubscriptionDeleteRequest{}
	deltrans.SubDelReqMsg.RequestId.Id = trans.SubReqMsg.RequestId.Id
	deltrans.SubDelReqMsg.RequestId.Seq = uint32(subs.GetSubId())
	deltrans.SubDelReqMsg.FunctionId = trans.SubReqMsg.FunctionId
	deltrans.Mtype, deltrans.Payload, err = c.e2ap.PackSubscriptionDeleteRequest(deltrans.SubDelReqMsg)
	if err != nil {
		xapp.Logger.Error("SubReq timeout: %s", idstring(trans, subs, err))
		//TODO improve error handling. Important at least in merge
		deltrans.Release()
		subs.Release()
		return
	}

	err = subs.SetTransaction(deltrans)
	if err != nil {
		xapp.Logger.Error("SubReq timeout: %s", idstring(trans, subs, err))
		//TODO improve error handling. Important at least in merge
		deltrans.Release()
		return
	}

	c.rmrSend("SubReq timer: SubDelReq to E2T", subs, deltrans)
	c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()), subDelReqTime, FirstTry, c.handleSubscriptionDeleteRequestTimer)
	return
}

func (c *Control) handleSubscriptionDeleteRequest(params *RMRParams) {
	xapp.Logger.Info("SubDelReq from xapp: %s", params.String())

	SubDelReqMsg, err := c.e2ap.UnpackSubscriptionDeleteRequest(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelReq Drop %s", idstring(params, nil, err))
		return
	}

	trans, err := c.tracker.TrackTransaction(NewRmrEndpoint(params.Src), params.Xid, params.Meid, false, true)
	if err != nil {
		xapp.Logger.Error("SubDelReq Drop %s", idstring(params, nil, err))
		return
	}
	trans.SubDelReqMsg = SubDelReqMsg

	subs, err := c.findSubs([]int{int(trans.SubDelReqMsg.RequestId.Seq), params.SubId})
	if err != nil {
		xapp.Logger.Error("SubDelReq: %s", idstring(params, nil, err))
		trans.Release()
		return
	}

	err = subs.SetTransaction(trans)
	if err != nil {
		xapp.Logger.Error("SubDelReq: %s", idstring(trans, subs, err))
		trans.Release()
		return
	}

	xapp.Logger.Debug("SubDelReq: Handling %s", idstring(trans, subs, nil))

	//
	// TODO: subscription delete is in fact owned by subscription and not transaction.
	//       Transaction is toward xapp while Subscription is toward ran.
	//       In merge several xapps may wake transactions, while only one subscription
	//       toward ran occurs -> subscription owns subscription creation toward ran
	//
	//       This is intermediate solution while improving message handling
	//
	trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionDeleteRequest(trans.SubDelReqMsg)
	if err != nil {
		xapp.Logger.Error("SubDelReq: %s", idstring(trans, subs, err))
		trans.Release()
		return
	}

	subs.UnConfirmed()

	c.rmrSend("SubDelReq: SubDelReq to E2T", subs, trans)

	c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()), subDelReqTime, FirstTry, c.handleSubscriptionDeleteRequestTimer)
	return
}

func (c *Control) handleSubscriptionDeleteResponse(params *RMRParams) (err error) {
	xapp.Logger.Info("SubDelResp from E2T:%s", params.String())

	SubDelRespMsg, err := c.e2ap.UnpackSubscriptionDeleteResponse(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelResp: Dropping this msg. %s", idstring(params, nil, err))
		return
	}

	subs, trans, err := c.findSubsAndTrans([]int{int(SubDelRespMsg.RequestId.Seq), params.SubId})
	if err != nil {
		xapp.Logger.Error("SubDelResp: %s", idstring(params, nil, err))
		return
	}
	trans.SubDelRespMsg = SubDelRespMsg
	xapp.Logger.Debug("SubDelResp: Handling %s", idstring(trans, subs, nil))

	c.timerMap.StopTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()))

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription Delete timer already received
		return
	}

	c.sendSubscriptionDeleteResponse("SubDelResp", trans, subs)
	return
}

func (c *Control) handleSubscriptionDeleteFailure(params *RMRParams) {
	xapp.Logger.Info("SubDelFail from E2T:%s", params.String())

	SubDelFailMsg, err := c.e2ap.UnpackSubscriptionDeleteFailure(params.Payload)
	if err != nil {
		xapp.Logger.Error("SubDelFail: Dropping this msg. %s", idstring(params, nil, err))
		return
	}

	subs, trans, err := c.findSubsAndTrans([]int{int(SubDelFailMsg.RequestId.Seq), params.SubId})
	if err != nil {
		xapp.Logger.Error("SubDelFail: %s", idstring(params, nil, err))
		return
	}
	trans.SubDelFailMsg = SubDelFailMsg
	xapp.Logger.Debug("SubDelFail: Handling %s", idstring(trans, subs, nil))

	c.timerMap.StopTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()))

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription Delete timer already received
		return
	}

	c.sendSubscriptionDeleteResponse("SubDelFail", trans, subs)
	return
}

func (c *Control) handleSubscriptionDeleteRequestTimer(strId string, nbrId int, tryCount uint64) {
	xapp.Logger.Info("SubDelReq timeout: subId: %v, tryCount: %v", nbrId, tryCount)

	subs, trans, err := c.findSubsAndTrans([]int{nbrId})
	if err != nil {
		xapp.Logger.Error("SubDelReq timeout: %s", idstring(nil, nil, err))
		return
	}
	xapp.Logger.Debug("SubDelReq timeout: Handling %s", idstring(trans, subs, nil))

	responseReceived := trans.CheckResponseReceived()
	if responseReceived == true {
		// Subscription Delete Response or Failure already received
		return
	}

	if tryCount < maxSubDelReqTryCount {
		// Set possible to handle new response for the subId
		trans.RetryTransaction()
		c.rmrSend("SubDelReq timeout: SubDelReq to E2T", subs, trans)
		tryCount++
		c.timerMap.StartTimer("RIC_SUB_DEL_REQ", int(subs.GetSubId()), subReqTime, tryCount, c.handleSubscriptionDeleteRequestTimer)
		return
	}

	c.sendSubscriptionDeleteResponse("SubDelReq(timer)", trans, subs)
	return
}

func (c *Control) sendSubscriptionDeleteResponse(desc string, trans *Transaction, subs *Subscription) {

	if trans.ForwardRespToXapp == true {
		//Always generate SubDelResp
		trans.SubDelRespMsg = &e2ap.E2APSubscriptionDeleteResponse{}
		trans.SubDelRespMsg.RequestId.Id = trans.SubDelReqMsg.RequestId.Id
		trans.SubDelRespMsg.RequestId.Seq = uint32(subs.GetSubId())
		trans.SubDelRespMsg.FunctionId = trans.SubDelReqMsg.FunctionId

		var err error
		trans.Mtype, trans.Payload, err = c.e2ap.PackSubscriptionDeleteResponse(trans.SubDelRespMsg)
		if err == nil {
			c.rmrReplyToSender(desc+": SubDelResp to xapp", subs, trans)
			time.Sleep(3 * time.Second)
		} else {
			//TODO error handling improvement
			xapp.Logger.Error("%s: (continue cleaning) %s", desc, idstring(trans, subs, err))
		}
	}

	trans.Release()
	subs.Release()
}
